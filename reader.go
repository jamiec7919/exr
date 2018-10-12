package exr

import (
	"bufio"
	"fmt"
)

func ReadVersion(r *bufio.Reader) (*EXRVersion, error) {
	var magic [4]byte

	n, err := r.Read(magic[:])

	if err != nil {
		return nil, fmt.Errorf("error reading magic: %v", err)
	}

	if n != 4 {
		return nil, err
	}

	if magic[0] != 0x76 ||
		magic[1] != 0x2f ||
		magic[2] != 0x31 ||
		magic[3] != 0x01 {
		return nil, fmt.Errorf("incorrect magic: %v (%v)", magic, string(magic[:]))
	}

	var versionBuf [4]byte

	n, err = r.Read(versionBuf[:])

	if err != nil {
		return nil, err
	}

	if n != 4 {
		return nil, err
	}

	version := (int)(versionBuf[3])<<24 | (int)(versionBuf[2])<<16 | (int)(versionBuf[1])<<8 | (int)(versionBuf[0])

	if version&0xff != 2 {
		return nil, fmt.Errorf("version != 2 (%v)", version)
	}

	v := &EXRVersion{}

	v.version = 2

	if version&(1<<9) != 0 {
		// Old version 'tiled' bit
		if version&((1<<11)|(1<<12)) != 0 {
			// error
			return nil, fmt.Errorf("incompatible tiled bit (%v)", version)
		}

		v.tiled = true
	}

	if version&(1<<10) != 0 {
		v.longName = true
	}

	if version&(1<<11) != 0 {
		v.nonImage = true
	}

	if version&(1<<12) != 0 {
		v.multipart = true
	}

	return v, nil

}

func ReadAttrib(r *bufio.Reader) (*EXRAttribute, error) {
	buf, err := r.ReadBytes(0x00)

	if err != nil {
		return nil, err
	}

	if len(buf) == 1 {
		// We've only read the single null byte so this is the last attribute in header
		return nil, nil
	}

	name := string(buf[:len(buf)-1])

	buf, err = r.ReadBytes(0x00)

	if err != nil {
		return nil, err
	}

	attribType := string(buf[:len(buf)-1])

	var size [4]byte

	n, err := r.Read(size[:])

	if err != nil {
		return nil, err
	}

	if n != 4 {
		return nil, err
	}

	dataSize := (int)(size[3])<<24 | (int)(size[2])<<16 | (int)(size[1])<<8 | (int)(size[0])
	data := make([]byte, dataSize)

	n, err = r.Read(data[:])

	if err != nil {
		return nil, err
	}

	if n != dataSize {
		return nil, err
	}

	return &EXRAttribute{
		name:       name,
		attribType: attribType,
		value:      data}, nil
}

func ReadChlist(r *bufio.Reader) ([]*EXRChannelInfo, error) {
	var channels []*EXRChannelInfo

	for {
		buf, err := r.ReadBytes(0x00)

		if err != nil {
			return nil, err
		}

		if len(buf) == 1 {
			// We've only read the single null byte so this is the last channel in attribute
			return channels, nil
		}

		name := string(buf[:len(buf)-1])

		var intbuf [4]byte

		n, err := r.Read(intbuf[:])

		if err != nil {
			return nil, err
		}

		if n != 4 {
			return nil, err
		}

		pixelType := int(intbuf[0])

		n, err = r.Read(intbuf[:])

		if err != nil {
			return nil, err
		}

		if n != 4 {
			return nil, err
		}

		pLinear := int(intbuf[0]) != 0

		n, err = r.Read(intbuf[:])

		if err != nil {
			return nil, err
		}

		if n != 4 {
			return nil, err
		}

		xSampling := (int)(intbuf[3])<<24 | (int)(intbuf[2])<<16 | (int)(intbuf[1])<<8 | (int)(intbuf[0])

		n, err = r.Read(intbuf[:])

		if err != nil {
			return nil, err
		}

		if n != 4 {
			return nil, err
		}

		ySampling := (int)(intbuf[3])<<24 | (int)(intbuf[2])<<16 | (int)(intbuf[1])<<8 | (int)(intbuf[0])

		channels = append(channels, &EXRChannelInfo{
			name:      name,
			pixelType: pixelType,
			pLinear:   pLinear,
			xSampling: xSampling,
			ySampling: ySampling,
		})
	}
}
