package exr

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
)

type Writer struct {
}

// WriteVersion writes the magix and version ints.
func WriteVersion(version *EXRVersion, w *bufio.Writer) error {
	magic := [4]byte{0x76, 0x2f, 0x31, 0x01}

	n, err := w.Write(magic[:])

	if err != nil {
		return fmt.Errorf("error writing magic: %v", err)
	}

	if n != 4 {
		return fmt.Errorf("error writing magic: incorrect number of bytes written")
	}

	versionBuf := version.Packed()

	err = binary.Write(w, binary.LittleEndian, versionBuf)

	//	n, err = r.Write(versionBuf[:])

	if err != nil {
		return fmt.Errorf("error writing version: %v", err)
	}

	return nil

}

func attribType(v interface{}) string {

	switch v.(type) {
	case int32:
		return "int"
	case uint32:
		return "int"
	case float32:
		return "float"
	case float64:
		return "double"
	case Compression:
		return "compression"
	case LineOrder:
		return "lineOrder"
	case V2f:
		return "v2f"
	case Box2i:
		return "box2i"
	case Chlist:
		return "chlist"
	case string:
		return "string"
	case String:
		return "string"
	case Stringvector:
		return "stringvector"
	}

	return "<unknown>"
}

func WriteAttrib(attrib *EXRAttribute, w *bufio.Writer) error {
	if attrib == nil {
		w.WriteByte(0x00)
		return nil
	}

	w.WriteString(attrib.name)
	w.WriteByte(0x00)
	w.WriteString(attrib.attribType)
	w.WriteByte(0x00)

	buf := &bytes.Buffer{}

	binary.Write(buf, binary.LittleEndian, attrib.value)

	size := int32(buf.Len())
	binary.Write(w, binary.LittleEndian, size)
	w.Write(buf.Bytes()) // Since we want to use the marshalling this should be binary.Write but need to know the size. ?
	// Maybe but attrib.value is []byte which we can marshall to/from
	return nil
}

func WriteChlist(chlist []*EXRChannelInfo, w *bufio.Writer) error {
	return nil
}
