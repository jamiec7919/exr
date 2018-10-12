package exr

import (
	"bytes"
)

// rleEncode will apply an EXR specific preprocess and then byte-level RLE compress the buffer.
// It will return either the compressed buffer or the original buffer depending on which is smaller.
func rleEncode(buf []byte) []byte {
	tmpBuf := make([]byte, len(buf))

	// Apply EXR-specific preprocess.  From OpenEXR's ImfRleCompressor.cpp
	// And tinyexr

	// 1) Reorder the pixel data
	t1 := 0
	t2 := (len(tmpBuf) + 1) / 2
	src := 0

	for {
		if src < len(buf) {
			tmpBuf[t1] = buf[src]
			src++
			t1++
		} else {
			break
		}

		if src < len(buf) {
			tmpBuf[t2] = buf[src]
			src++
			t2++
		} else {
			break
		}

	}

	// 2) Predictor
	p := tmpBuf[0]

	for t := 1; t < len(buf); t++ {
		d := int(tmpBuf[t]) - int(p) + (128 + 256)
		p = tmpBuf[t]
		tmpBuf[t] = byte(d)
	}
	// Now perform rle encode on tmpBuf
	compressed := rleCompress(tmpBuf)

	if len(compressed) < len(buf) {
		return compressed
	}

	return buf
}

const (
	minRunLength = 3
	maxRunLength = 127
)

// rleCompress will byte-level RLE compress the given buffer.
func rleCompress(in []byte) []byte {
	out := bytes.Buffer{}

	runStart := 0
	runEnd := runStart + 1

	for runStart < len(in) {
		for runEnd < len(in) && in[runStart] == in[runEnd] && runEnd-runStart-1 < maxRunLength {
			runEnd++
		}

		if runEnd-runStart >= minRunLength {
			// Compressable run
			out.WriteByte(byte(runEnd-runStart) - 1)
			out.WriteByte(in[runStart])
			runStart = runEnd
		} else {
			// Not compressable
			for runEnd < len(in) &&
				((runEnd+1 >= len(in) || in[runEnd] != in[runEnd+1]) ||
					(runEnd+2 >= len(in) || in[runEnd+1] != in[runEnd+2])) &&
				runEnd-runStart < maxRunLength {
				runEnd++
			}

			out.WriteByte(byte(runStart - runEnd))

			for runStart < runEnd {
				out.WriteByte(in[runStart])
				runStart++
			}
		}
	}

	return out.Bytes()
}

func rleDecode(buf []byte) []byte {
	tmpBuf := rleDecompress(buf)

	// Predictor
	t := 1
	stop := len(tmpBuf)

	for t < stop {
		d := int(tmpBuf[t-1]) + int(tmpBuf[t]) - 128
		tmpBuf[t] = byte(d)
		t++
	}

	// Reorder pixel data
	t1 := 0
	t2 := (len(tmpBuf) + 1) / 2

	out := bytes.Buffer{}

	for {
		if out.Len() < len(tmpBuf) {
			out.WriteByte(tmpBuf[t1])
			t1++
		} else {
			break
		}

		if out.Len() < len(tmpBuf) {
			out.WriteByte(tmpBuf[t2])
			t2++
		} else {
			break
		}
	}

	return out.Bytes()
}

func rleDecompress(buf []byte) []byte {
	out := bytes.Buffer{}

	in := 0

	for in < len(buf) {

		count := int8(buf[in])
		in++

		if count < 0 {
			count := -count
			out.Write(buf[in : in+int(count)])

			in += int(count)
		} else {
			val := buf[in]
			in++

			for i := 0; i < int(count)+1; i++ {
				out.WriteByte(val)
			}

		}
	}

	return out.Bytes()
}

/*

const int MIN_RUN_LENGTH = 3;
const int MAX_RUN_LENGTH = 127;

//
// Compress an array of bytes, using run-length encoding,
// and return the length of the compressed data.
//

static int rleCompress(int inLength, const char in[], signed char out[]) {
  const char *inEnd = in + inLength;
  const char *runStart = in;
  const char *runEnd = in + 1;
  signed char *outWrite = out;

  while (runStart < inEnd) {
    while (runEnd < inEnd && *runStart == *runEnd &&
           runEnd - runStart - 1 < MAX_RUN_LENGTH) {
      ++runEnd;
    }

    if (runEnd - runStart >= MIN_RUN_LENGTH) {
      //
      // Compressable run
      //

      *outWrite++ = static_cast<char>(runEnd - runStart) - 1;
      *outWrite++ = *(reinterpret_cast<const signed char *>(runStart));
      runStart = runEnd;
    } else {
      //
      // Uncompressable run
      //

      while (runEnd < inEnd &&
             ((runEnd + 1 >= inEnd || *runEnd != *(runEnd + 1)) ||
              (runEnd + 2 >= inEnd || *(runEnd + 1) != *(runEnd + 2))) &&
             runEnd - runStart < MAX_RUN_LENGTH) {
        ++runEnd;
      }

      *outWrite++ = static_cast<char>(runStart - runEnd);

      while (runStart < runEnd) {
        *outWrite++ = *(reinterpret_cast<const signed char *>(runStart++));
      }
    }

    ++runEnd;
  }

  return static_cast<int>(outWrite - out);
}


//
  // Apply EXR-specific? postprocess. Grabbed from OpenEXR's
  // ImfRleCompressor.cpp
  //

  //
  // Reorder the pixel data.
  //

  const char *srcPtr = reinterpret_cast<const char *>(src);

  {
    char *t1 = reinterpret_cast<char *>(&tmpBuf.at(0));
    char *t2 = reinterpret_cast<char *>(&tmpBuf.at(0)) + (src_size + 1) / 2;
    const char *stop = srcPtr + src_size;

    for (;;) {
      if (srcPtr < stop)
        *(t1++) = *(srcPtr++);
      else
        break;

      if (srcPtr < stop)
        *(t2++) = *(srcPtr++);
      else
        break;
    }
  }

  //
  // Predictor.
  //

  {
    unsigned char *t = &tmpBuf.at(0) + 1;
    unsigned char *stop = &tmpBuf.at(0) + src_size;
    int p = t[-1];

    while (t < stop) {
      int d = int(t[0]) - p + (128 + 256);
      p = t[0];
      t[0] = static_cast<unsigned char>(d);
      ++t;
    }
  }

  // outSize will be (srcSiz * 3) / 2 at max.
  int outSize = rleCompress(static_cast<int>(src_size),
                            reinterpret_cast<const char *>(&tmpBuf.at(0)),
                            reinterpret_cast<signed char *>(dst));
  assert(outSize > 0);

  compressedSize = static_cast<tinyexr::tinyexr_uint64>(outSize);

  // Use uncompressed data when compressed data is larger than uncompressed.
  // (Issue 40)
  if (compressedSize >= src_size) {
    compressedSize = src_size;
    memcpy(dst, src, src_size);
}
*/
