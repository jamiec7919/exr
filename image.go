package exr

import (
	"bufio"
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
	"io"
	"sort"
)

type Pixels struct {
	Kind                 int // One of PixelTypeUint...
	Data                 interface{}
	Base                 int32 // Pixel is found at Data[Base+x*XStride+y*YStride]
	XStride, YStride     int32
	XSampling, YSampling int // only for sub-sampled images
	FillValue            float64
}

type TileDescription struct {
	Width, Height int
	Kind          int // One of TileOneLevel...TileRipMapLevels
}

type Header struct {
	dataWindow      [4]int32
	displayWindow   [4]int32
	channels        []Channel
	tiled           bool
	tileDescription TileDescription
}

func NewHeader(width, height int) Header {
	return NewHeaderWindow(0, 0, int32(width-1), int32(height-1))
}

func NewHeaderWindow(xMin, yMin, xMax, yMax int32) Header {
	return Header{
		dataWindow:    [4]int32{xMin, yMin, xMax, yMax},
		displayWindow: [4]int32{xMin, yMin, xMax, yMax},
	}

}

// Deep etc.
func (h *Header) SetType() {}

func (h *Header) AddChannel(ch Channel) {
	h.channels = append(h.channels, ch)
}

func (h *Header) FindChannel(name string) *Channel {
	for i := range h.channels {
		if h.channels[i].Name == name {
			return &h.channels[i]
		}
	}

	return nil
}

func (h *Header) SetTileDescription(td TileDescription) {
	h.tileDescription = td
	h.tiled = true
}

type fbChannel struct {
	name   string
	pixels Pixels
}

type fbChannels []fbChannel

func (s fbChannels) Len() int {
	return len(s)
}

func (s fbChannels) Less(i, j int) bool {
	return s[i].name < s[j].name
}

func (s fbChannels) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type Framebuffer struct {
	channels fbChannels
}

// Insert a slice of pixel data for a given channe;
func (fb *Framebuffer) Insert(ch string, pixels Pixels) {
	fb.channels = append(fb.channels, struct {
		name   string
		pixels Pixels
	}{ch, pixels})
}

type OutputFile struct {
	header      Header
	framebuffer Framebuffer

	w io.WriteSeeker

	versionWritten bool
	headerOfs      int64
	headerWritten  bool
	offsetTableOfs int64
	offsetTable    []uint64
	chunkOfs       int64

	numChunks int

	currentScanline int
}

func NewOutputFile(w io.WriteSeeker, h Header) *OutputFile {
	return &OutputFile{
		header: h,
		w:      w,
	}
}

func (o *OutputFile) SetFramebuffer(fb Framebuffer) {
	o.framebuffer = fb
}

func (o *OutputFile) stdAttribs() []attrib {
	var attribs []attrib

	attribs = append(attribs, attrib{"dataWindow", Box2i{o.header.dataWindow[0],
		o.header.dataWindow[1],
		o.header.dataWindow[2],
		o.header.dataWindow[3]}})

	attribs = append(attribs, attrib{"displayWindow", Box2i{o.header.displayWindow[0],
		o.header.displayWindow[1],
		o.header.displayWindow[2],
		o.header.displayWindow[3]}})

	attribs = append(attribs, attrib{"pixelAspectRatio", float32(1)})

	attribs = append(attribs, attrib{"screenWindowWidth", float32(1)})

	attribs = append(attribs, attrib{"screenWindowCenter", V2f{}})

	attribs = append(attribs, attrib{"compression", Compression(CompressionTypeNone)})
	attribs = append(attribs, attrib{"lineOrder", LineOrder(LineOrderIncreasingY)})

	attribs = append(attribs, attrib{"chunkCount", int32(o.numChunks)})

	attribs = append(attribs, attrib{"channels", Chlist(o.header.channels)})
	fmt.Printf("chlist:%v\n", Chlist(o.header.channels))
	return attribs
}

func writeAttrib(w io.Writer, v interface{}) error {
	switch t := v.(type) {
	case encoding.BinaryMarshaler:
		b, err := t.MarshalBinary()

		if err != nil {
			return fmt.Errorf("error calling MarshalBinary: writeAttrib: %v", err)
		}

		w.Write(b)

	default:
		binary.Write(w, binary.LittleEndian, v)
	}
	return nil
}

func writeFloat32AsHalf(buf *bytes.Buffer, v float32) error {
	return binary.Write(buf, binary.LittleEndian, Half(Float32ToFloat16(v)))
}

// WritePixels writes scanlines between y=start and y=end
func (o *OutputFile) WritePixels(count int) error {

	if o.header.tiled {
		return fmt.Errorf("attempting to write scanlines to a tiled image")
	}

	bufW := bufio.NewWriter(o.w)

	if !o.versionWritten {
		version := EXRVersion{}
		// fill version details.
		WriteVersion(&version, bufW)
		bufW.Flush()
	}

	if !o.headerWritten {
		linesPerChunk := 1
		numChunks := int(((o.header.dataWindow[3] - o.header.dataWindow[1] + 1) / int32(linesPerChunk))) // yMax-yMin

		o.numChunks = numChunks

		// Write header
		ofs, err := o.w.Seek(0, io.SeekCurrent)

		if err != nil {
			return fmt.Errorf("finding current file position: %v", err)
		}

		o.headerOfs = ofs

		attribs := o.stdAttribs()

		for _, attrib := range attribs {
			buf := bytes.Buffer{}
			writeAttrib(&buf, attrib.val)
			//binary.Write(&buf, binary.LittleEndian, attrib.val)

			a := &EXRAttribute{name: attrib.name,
				attribType: attribType(attrib.val),
				value:      buf.Bytes(),
			}

			fmt.Printf("Write %v: %v: %v\n", a.name, attrib.val, a.value)
			WriteAttrib(a, bufW)
		}

		WriteAttrib(nil, bufW)
		bufW.Flush()

	}

	ofs, err := o.w.Seek(0, io.SeekCurrent)

	if err != nil {
		return fmt.Errorf("finding current file position: %v", err)
	}

	o.offsetTableOfs = ofs

	// The number of scan lines in a block depends on the compression.
	// NO_COMPRESSION  1
	// RLE_COMPRESSION 1
	// ZIPS_COMPRESSION 1
	// ZIP_COMPRESSION 16
	// PIZ_COMPRESSION 32
	// PXR24_COMPRESSION 16
	// B44_COMPRESSION 32
	// B44A_COMPRESSION 32

	// Then chunk layout is
	// [part number]  (if multipart file)
	// y coordinate
	// pixel data size  (int, in bytes)
	// pixel data
	linesPerChunk := 1                                                                          // Assume NO_COMPRESSION to start
	numChunks := ((o.header.dataWindow[3] - o.header.dataWindow[1] + 1) / int32(linesPerChunk)) // yMax-yMin

	// Write offset table

	// For scan line blocks the line offset table is a sequence of scan line offsets with
	// one offset per scan line block.

	o.offsetTable = make([]uint64, numChunks)

	fmt.Printf("%v chunks\n", numChunks)

	// Initially write the chunk slice to reserve space even though we don't know the offsets
	binary.Write(o.w, binary.LittleEndian, o.offsetTable)

	ofs, err = o.w.Seek(0, io.SeekCurrent)

	if err != nil {
		return fmt.Errorf("finding current file position: %v", err)
	}

	o.chunkOfs = ofs

	y := int32(o.currentScanline) // o.header.dataWindow[1]

	// Need to make sure the framebuffer channels are sorted.
	sort.Sort(o.framebuffer.channels)

	for chunk := int32(0); chunk < numChunks; chunk++ {

		buf := &bytes.Buffer{}

		ofs, err = o.w.Seek(0, io.SeekCurrent)

		if err != nil {
			return fmt.Errorf("finding current file position: %v", err)
		}

		o.offsetTable[chunk] = uint64(ofs)

		fmt.Printf("Writing chunk %v (ofs %v)\n", chunk, ofs)

		// If framebuffer doesn't contain Pixels for a given Channel then the channel is filled with default value in file.
		// If framebuffer has Pixels for non-existent Channel then the pixels are skipped.
		// Unlike Ilm library there will be a seperate base value from the slice.

		// Pixel data is channels in alphabetical order of either byte, half or float
		for _, ch := range o.framebuffer.channels {

			// 1) Check channel exists in file header.
			headerChan := o.header.FindChannel(ch.name)

			if headerChan != nil {

			}

			//pixelOfs := ch.pixels.Base + 0*ch.pixels.XStride + y*ch.pixels.YStride
			pixelBaseOfs := ch.pixels.Base + y*ch.pixels.YStride

			// This isn't right, strides need to be taken into account for each pixel (rather than just writing whole lot)
			// OK, done that but should be converting & writing the given data into the type specified in HEADER, not that
			// of the passed in buffer.
			switch t := ch.pixels.Data.(type) {
			case []float32:
				for x := o.header.dataWindow[0]; x <= o.header.dataWindow[2]; x++ {
					pixelOfs := pixelBaseOfs + x*ch.pixels.XStride

					switch headerChan.PixelType {
					case PixelTypeUInt:
						//			writeFloat32AsUInt(buf, t[pixelOfs])
					case PixelTypeHalf:
						writeFloat32AsHalf(buf, t[pixelOfs])
					default:
						binary.Write(buf, binary.LittleEndian, t[pixelOfs])
					}
					//fmt.Printf("Pixel %v %v %v\n", ch.name, x, t[pixelOfs])
				}
				//binary.Write(o.w, binary.LittleEndian, t[pixelOfs:pixelOfs+(o.header.dataWindow[2]-o.header.dataWindow[0])])
			case []uint8:
				for x := o.header.dataWindow[0]; x <= o.header.dataWindow[2]; x++ {
					pixelOfs := pixelBaseOfs + x*ch.pixels.XStride
					binary.Write(buf, binary.LittleEndian, t[pixelOfs])
				}
				///binary.Write(o.w, binary.LittleEndian, t[pixelOfs:pixelOfs+(o.header.dataWindow[2]-o.header.dataWindow[0])])
			case []Half:
				for x := o.header.dataWindow[0]; x <= o.header.dataWindow[2]; x++ {
					pixelOfs := pixelBaseOfs + x*ch.pixels.XStride
					binary.Write(buf, binary.LittleEndian, t[pixelOfs])
				}
				//binary.Write(o.w, binary.LittleEndian, t[pixelOfs:pixelOfs+(o.header.dataWindow[2]-o.header.dataWindow[0])])
			default:
				return fmt.Errorf("invalid pixel type (%T) for channel %v", t, ch.name)
			}
		}

		if false { // multipart
			binary.Write(o.w, binary.LittleEndian, int32(0))
		}

		binary.Write(o.w, binary.LittleEndian, int32(y))
		binary.Write(o.w, binary.LittleEndian, int32(buf.Len())) // data size
		o.w.Write(buf.Bytes())

		y += int32(linesPerChunk)
	}

	o.currentScanline = int(y)

	// Offset table has been updated, seek to the start and write it out.
	_, err = o.w.Seek(o.offsetTableOfs, io.SeekStart)

	if err != nil {
		return fmt.Errorf("finding current file position: %v", err)
	}

	binary.Write(o.w, binary.LittleEndian, o.offsetTable)

	// At this point we should have a complete file

	return nil

}

func (o *OutputFile) WriteTiles(start, end int) error {
	return nil
}

type DeepFramebuffer struct {
}
