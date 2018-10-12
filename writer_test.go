package exr

import (
	"bufio"
	//"bytes"
	//"fmt"
	"math"
	"os"
	"testing"
)

func genImage() (r, g, b []float32) {
	r = make([]float32, 128*128)
	g = make([]float32, 128*128)
	b = make([]float32, 128*128)

	for y := 0; y < 128; y++ {
		ydeg := float32(y) / 128.0
		for x := 0; x < 128; x++ {
			xdeg := float32(x) / 128.0

			rv := float32(math.Abs(math.Sin(float64(xdeg)*2)*math.Cos(float64(ydeg)*4)) * 255.0)
			gv := float32(math.Abs(math.Sin(float64(xdeg))*math.Sin(float64(ydeg)*2)) * 255.0)
			bv := float32(math.Abs(math.Sin(float64(xdeg)*3)*math.Cos(float64(ydeg)*5)) * 255.0)

			r[x+(y*128)] = rv
			g[x+(y*128)] = gv
			b[x+(y*128)] = bv
		}
	}

	return
}

func TestWriterFullImage(t *testing.T) {
	r, g, b := genImage()

	hd := NewHeader(128, 128)
	hd.AddChannel(Channel{Name: "R", PixelType: PixelTypeHalf, XSampling: 1, YSampling: 1})
	hd.AddChannel(Channel{Name: "G", PixelType: PixelTypeHalf, XSampling: 1, YSampling: 1})
	hd.AddChannel(Channel{Name: "B", PixelType: PixelTypeHalf, XSampling: 1, YSampling: 1})

	f, err := os.Create("testdata/out2.exr")
	defer f.Close()

	if err != nil {
		t.Fatalf("error creating testdata: %v", err)
	}

	of := NewOutputFile(f, hd)
	fb := Framebuffer{}
	fb.Insert("R", Pixels{PixelTypeFloat, r, 0, 1, 128, 1, 1, 0})
	fb.Insert("G", Pixels{PixelTypeFloat, g, 0, 1, 128, 1, 1, 0})
	fb.Insert("B", Pixels{PixelTypeFloat, b, 0, 1, 128, 1, 1, 0})
	of.SetFramebuffer(fb)

	err = of.WritePixels(128)

	if err != nil {
		t.Fatalf("error writing scanlines: %v", err)
	}
}

func TestWriter(t *testing.T) {
	f, err := os.Create("testdata/out1.exr")
	defer f.Close()

	if err != nil {
		t.Fatalf("error creating testdata: %v", err)
	}

	w := bufio.NewWriter(f)

	version := EXRVersion{}

	err = WriteVersion(&version, w)

	if err != nil {
		t.Fatalf("error writing testdata: %v", err)
	}

	w.Flush()

	/*
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
	}*/
}
