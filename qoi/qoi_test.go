package qoi

import (
	"testing"
	"image/png"
    "os"
    "bufio"
)

type VirtualWriter struct {
    d []byte
}


func (w *VirtualWriter) Write(p []byte) (n int, err error) {
    w.d = append(w.d, p...)
    return len(p), nil
}


func _TestDecoder(t *testing.T) {
	path := "../qoi_test_images/"
	images := []string{
        "dice",
		"kodim10",
		"kodim23",
		"qoi_logo",
		"testcard",
        "testcard_rgba",
		"wikipedia_008",
	}

	for _, loc := range images {
		loc_qoi := path + loc + ".qoi"
		loc_png := path + loc + ".png"
		test_decoder(loc_qoi, loc_png, *t)
	}
}

func TestEncoder(t *testing.T) {
	path := "../qoi_test_images/"
	images := []string{
        "dice",
        "kodim10",
        "kodim23",
        "qoi_logo",
		"testcard",
        "testcard_rgba",
        "wikipedia_008",
	}

	for _, loc := range images {
		loc_qoi := path + loc + ".qoi"
		loc_png := path + loc + ".png"
        if !test_encoder(loc_qoi, loc_png, *t) {
            t.FailNow()
        }
	}
}

func test_decoder(loc_qoi, loc_png string, t testing.T) bool {
	f, err := os.Open(loc_qoi)
	if err != nil {
		t.Fatalf("%v.", err)
	}
	im_qoi, err := Decode(f)
	if err != nil {
		t.Fatalf("%v.", err)
	}

	f, err = os.Open(loc_png)
	if err != nil {
		t.Fatalf("%v.", err)
	}

	im_png, err := png.Decode(f)
	if err != nil {
		t.Fatalf("%v.", err)
	}

	if im_png.Bounds() != im_qoi.Bounds() {
		t.Fatalf("Image bounds don't match up!\n")
	}

	passed := true

	for x := 0; x < im_qoi.Bounds().Max.X; x++ {
		for y := 0; y < im_qoi.Bounds().Max.Y; y++ {
			if im_png.At(x, y) != im_qoi.At(x, y) {
				t.Logf("Here: (%v, %v). Value QOI: %v, Value PNG: %v\n", x, y, im_qoi.At(x, y), im_png.At(x, y))
				passed = false
			}
		}
	}
	return passed
}

func test_encoder(loc_qoi, loc_png string, t testing.T) bool {
    p("Now testing on %v...\n", loc_png)
    f, err  := os.Open(loc_png)
    if err != nil {
        t.Fatalf("%v.", err)
    }

    im_png, err := png.Decode(f)
    if err != nil {
        t.Fatalf("%v.", err)
    }

    encoderOutput := &VirtualWriter{make([]byte, 0)}
    Encode(encoderOutput, im_png)

    p("Now decoding...\n")

    f, err = os.Open(loc_qoi)
    if err != nil {
        t.Fatalf("Opening file failed: %v.", err)
    }

    buff := bufio.NewReader(f)

    for _ = range encoderOutput.d[:14] {
        _, _ = buff.ReadByte()
    }

    for k, val := range encoderOutput.d[14:] {
        wanted, err := buff.ReadByte()
        //p("Got/wanted: %v == %v\n", val, wanted)
        if err != nil {
            p("Unable to read from %v.\n", loc_qoi)
            return false
        }
        if val != wanted {
            p("Incorrect value at position %v. Got: %v. Wanted: %v.\n", k, val, wanted)
            return false
        }
    }

    return true
}
