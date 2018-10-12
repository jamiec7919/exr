package exr

import (
	"bytes"
	"fmt"
	"testing"
)

// TestRLECompressDecompress will check that compression->decompression results in the same buffer.  This does
// not test the EXR specific RLE preprocess.
func TestRLECompressDecompress(t *testing.T) {
	testCases := [][]byte{
		[]byte("134234sdfsdfffffffsdfsdfsdfffeeee33445"),
		[]byte("1011011011011012022022022222222222222220000001011011013233233240000000003233232323322323323323323323323"),
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 100, 0, 100, 100, 0, 100, 100, 0, 100, 100, 0, 100, 100, 0, 100, 100, 0, 100, 100, 0, 100, 100, 0, 100},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", string(tc)), func(t *testing.T) {
			out := rleCompress(tc)

			//t.Logf("out: %v %v", out, string(out))

			t.Logf("Compression ratio: %v (%v/%v)", float32(len(out))/float32(len(tc)), len(out), len(tc))

			if len(out) == len(tc) {
				t.Logf("didn't compress, len(out) == len(tc)")
			} else {

				decomp := rleDecompress(out)
				//t.Logf("decomp: %v %v", decomp, string(decomp))

				if bytes.Compare(tc, decomp) != 0 {
					t.Fatalf("decompressed != compressed, expected %v %v, got %v %v", tc, string(tc), decomp, string(decomp))
				}
			}
		})
	}

}

// TestRLEEncodeDecode will check that compression->decompression results in the same buffer.  This DOES
// test the EXR specific RLE preprocess.
func TestRLEEncodeDecode(t *testing.T) {
	testCases := [][]byte{
		[]byte("134234sdfsdfffffffsdfsdfsdfffeeee33445"),
		[]byte("1011011011011012022022022222222222222220000001011011013233233240000000003233232323322323323323323323323"),
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 100, 0, 100, 100, 0, 100, 100, 0, 100, 100, 0, 100, 100, 0, 100, 100, 0, 100, 100, 0, 100, 100, 0, 100},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", string(tc)), func(t *testing.T) {
			out := rleEncode(tc)

			//t.Logf("out: %v %v", out, string(out))

			if len(out) == len(tc) {
				t.Logf("didn't compress, len(out) == len(tc)")
			} else {
				decomp := rleDecode(out)

				//t.Logf("decomp: %v %v", decomp, string(decomp))

				t.Logf("Compression ratio: %v (%v/%v)", float32(len(out))/float32(len(tc)), len(out), len(tc))

				if bytes.Compare(tc, decomp) != 0 {
					t.Fatalf("decompressed != compressed, expected %v %v, got %v %v", tc, string(tc), decomp, string(decomp))
				}
			}
		})
	}

}
