package qoi

import (
	"testing"
	"image/png"
	"os"
)


func TestDecoder(t *testing.T) {
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
		test_img(loc_qoi, loc_png, *t)
	}
}

func test_img(loc_qoi, loc_png string, t testing.T) bool {
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
