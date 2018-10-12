package exr

import (
	"bufio"
	"bytes"
	"os"
	"testing"
)

func TestLoader(t *testing.T) {
	testCases := []string{
		"testdata/transparent.exr",
		"testdata/asakusa.exr",
		"testdata/out2.exr",
	}

	for _, tc := range testCases {

		t.Run(tc, func(t *testing.T) {
			f, err := os.Open(tc)
			defer f.Close()

			if err != nil {
				t.Fatalf("Error loading testdata: %v", err)
			}

			r := bufio.NewReader(f)

			v, err := ReadVersion(r)

			if err != nil {
				t.Fatalf("Error loading testdata: %v", err)
			}

			if v == nil {
				t.Fatalf("Error loading testdata: %v", "Incompatible version")

			}

			t.Log(v)

			// Read first header
			var header []*EXRAttribute
			for {
				attrib, err := ReadAttrib(r)

				if err != nil {
					t.Fatalf("Error reading attribute: %v", err)
				}

				if attrib == nil {
					// end of header
					break
				}

				header = append(header, attrib)

				if attrib.name == "channels" && attrib.attribType == "chlist" {
					cr := bufio.NewReader(bytes.NewReader(attrib.value))
					channels, err := ReadChlist(cr)

					if err != nil {
						t.Fatalf("Error parsing chlist: %v", err)
					}
					for k := range channels {
						t.Logf("%v: %v %v %v %v", channels[k].name, channels[k].pixelType, channels[k].pLinear, channels[k].xSampling, channels[k].ySampling)
					}
				}
				t.Log(attrib)
			}

			t.Logf("type: %v", FindStringAttribute(header, "type"))

			if v.multipart {
				for {
					var header []*EXRAttribute
					t.Logf("Next header")
					for {
						attrib, err := ReadAttrib(r)

						if err != nil {
							t.Fatalf("Error reading attribute: %v", err)
						}

						if attrib == nil {
							// end of header
							break
						}

						header = append(header, attrib)
						t.Log(attrib)
					}

					if header == nil {
						break
					}

					header = nil
				}
			}

		})
	}
}
