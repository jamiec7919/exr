package exr

import (
	"fmt"
)

//NOTES:
// The EXR format is nothing like the structs here, you need to parse the attributes one by one, probably from memory.
// Headers are composed from attributes with variable sized null-terminated strings for name and type.

// OpenEXR file format: http://www.openexr.com/openexrfilelayout.pdf

// Note that according to the layout.pdf the string type
// should have an int for the length followed by the chars.
// I'm guessing that for attribs this is assumed to consist of
// the data length of the attribute, but for an array of strings
// it won't be.
func FindStringAttribute(a []*EXRAttribute, name string) string {
	for i := range a {
		if a[i].name == name {
			if a[i].attribType == "string" {
				return string(a[i].value)
			}
		}
	}

	return ""
}

// pixel type: possible values are: UINT = 0 HALF = 1 FLOAT = 2
const (
	PixelTypeUInt = iota
	PixelTypeHalf
	PixelTypeFloat
)

const MaxAttributes = 128

const (
	CompressionTypeNone  = iota // 0
	CompressionTypeRLE          // 1
	CompressionTypeZipS         // 2
	CompressionTypeZip          // 3
	CompressionTypePiz          // 4
	CompressionTypePXR24        // 5
	CompresstionTypeB44         // 6
	CompressionTypeB44A         // 7
)

const (
	TileOneLevel = iota
	TileMipMapLevels
	TileRipMapLevels
)

const (
	TileRoundDown = iota
	TileRoundUp
)

const (
	LineOrderIncreasingY = iota
	LineOrderDecreasingY
	LineOrderRandomY
)

const EXRVersionSize = 8

type EXRVersion struct {
	version   int  // this must be 2
	tiled     bool // tile format image
	longName  bool // long name attribute
	nonImage  bool // deep image(EXR 2.0)
	multipart bool // multi-part(EXR 2.0)
}

func (v *EXRVersion) Packed() int32 {
	var version int32 = 2

	if v.tiled {
		// This needs checking, tiled images are usually denoted elsewhere
		version |= (1 << 9)
	}

	if v.longName {
		version |= (1 << 10)
	}

	if v.nonImage {
		version |= (1 << 11)
	}

	if v.multipart {
		version |= (1 << 12)
	}

	return version
}

func (v *EXRVersion) String() string {
	return fmt.Sprintf("EXRVersion: \ntiled: %v\nlongName: %v\nnonImage: %v\nmultipart: %v", v.tiled, v.longName, v.nonImage, v.multipart)
}

type EXRAttribute struct {
	name       string // name and type are up to 255 chars long.
	attribType string
	value      []byte // uint8_t*

}

func (v *EXRAttribute) String() string {
	return fmt.Sprintf("EXRAttribute: %v (%v), %v: %v", v.name, v.attribType, len(v.value), v.value)
}

type EXRChannelInfo struct {
	name      string // less than 255 bytes long
	pixelType int
	xSampling int
	ySampling int
	pLinear   bool
	reserved  [3]uint8
}

type EXRTile struct {
	offsetX int32
	offsetY int32
	levelX  int32
	levelY  int32

	width  int32 // actual width in a tile.
	height int32 // actual height int a tile.

	images **byte // image[channels][pixels]
}

type EXRHeader struct {
	pixelAspectRatio   float32
	lineOrder          int32
	dataWindow         [4]int32
	displayWindow      [4]int32
	screenWindowCenter [2]float32
	screenWindowWidth  float32

	chunkCount int32

	// Properties for tiled format(`tiledesc`).
	tiled            int32
	tileSizeX        int32
	tileSizeY        int32
	tileLevelMode    int32
	tileRoundingMode int32

	longName  int32
	nonImage  int32
	multipart int32
	headerLen uint32

	// Custom attributes(exludes required attributes(e.g. `channels`,
	// `compression`, etc)
	numCustomAttributes int32
	customAttributes    [MaxAttributes]EXRAttribute

	channels *EXRChannelInfo // [num_channels]

	pixelTypes *int32 // Loaded pixel type(TINYEXR_PIXELTYPE_*) of `images` for
	// each channel. This is overwritten with `requested_pixel_types` when
	// loading.
	numChannels int32

	compressionType     int32  // compression type(TINYEXR_COMPRESSIONTYPE_*)
	requestedPixelTypes *int32 // Filled initially by
	// ParseEXRHeaderFrom(Meomory|File), then users
	// can edit it(only valid for HALF pixel type
	// channel)

}

type EXRMultiPartHeader struct {
	numHeaders int32
	headers    *EXRHeader
}
