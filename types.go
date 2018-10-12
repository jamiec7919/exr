package exr

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type attrib struct {
	name string
	val  interface{}
}

// String represents a string in an attribute (non null terminated, has length prepended)
type String string

func (b String) MarshalBinary() ([]byte, error) {
	buf := bytes.Buffer{}
	binary.Write(&buf, binary.LittleEndian, int32(len(b)))
	buf.WriteString(string(b))
	return buf.Bytes(), nil
}

// NullString represents a null-terminated string in an attribute
type NullString string

func (b NullString) MarshalBinary() ([]byte, error) {
	b1 := []byte(b)
	b1 = append(b1, 0)
	return b1, nil
}

type Box2i struct {
	xMin, yMin int32
	xMax, yMax int32
}

type Box2f struct {
	xMin, yMin float32
	xMax, yMax float32
}

type Chromaticities struct {
	redX, redY     float32
	greenX, greenY float32
	blueX, blueY   float32
	whiteX, whiteY float32
}

type Channel struct {
	Name                 string
	PixelType            int32
	PLinear              uint8
	XSampling, YSampling int32
}

func (b Channel) MarshalBinary() (data []byte, err error) {
	buf := bytes.Buffer{}

	s, err := NullString(b.Name).MarshalBinary()

	if err != nil {
		return nil, err
	}
	buf.Write(s)
	binary.Write(&buf, binary.LittleEndian, b.PixelType)
	binary.Write(&buf, binary.LittleEndian, b.PLinear)
	binary.Write(&buf, binary.LittleEndian, [3]byte{})
	binary.Write(&buf, binary.LittleEndian, b.XSampling)
	binary.Write(&buf, binary.LittleEndian, b.YSampling)

	fmt.Printf("****MB: %v\n", buf.Bytes())
	return buf.Bytes(), nil
}

type Chlist []Channel

func (b Chlist) MarshalBinary() ([]byte, error) {
	buf := bytes.Buffer{}
	for _, ch := range b {
		fmt.Printf("**WRCH: %v\n", ch)
		b, err := ch.MarshalBinary()

		if err != nil {
			return nil, fmt.Errorf("Chlist.MarshalBinary: %v", err)
		}

		buf.Write(b)
	}

	buf.WriteByte(0)

	return buf.Bytes(), nil
}

type Compression uint8
type LineOrder uint8

type Keycode struct {
	FlimMfcCode, FilmType, Prefix            int32
	Count                                    int32
	perfOffset, perfsPerFrame, perfsPerCount int32
}

type M33f [9]float32

type M44f [16]float32

type Preview struct {
	Width, Height uint32
	Data          []byte
}

func (b *Preview) MarshalBinary() ([]byte, error) {
	return nil, nil
}

type Rational struct {
	Num   int32
	Denom uint32
}

type Stringvector []string

func (b Stringvector) MarshalBinary() ([]byte, error) {
	buf := bytes.Buffer{}
	for s := range b {
		v, err := String(s).MarshalBinary()

		if err != nil {
			return nil, fmt.Errorf("\"%v\" MarshalBinary: %v", s, err)
		}

		buf.Write(v)
	}
	return buf.Bytes(), nil
}

type Timecode struct {
	TimeAndFlags uint32
	UserData     uint32
}

type V2i [2]int32
type V2f [2]float32
type V3i [3]int32
type V3f [3]float32

type Half uint16

type TileDesc struct {
	XSize, YSize uint32
	Mode         uint8
}
