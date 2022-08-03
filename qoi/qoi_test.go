package qoi

import (
	"testing"
	"image/png"
    "os"
    "bufio"
    "path/filepath"
    "strings"
    "image/color"
)

type VirtualWriter struct {
    d []byte
}


func (w *VirtualWriter) Write(p []byte) (n int, err error) {
    w.d = append(w.d, p...)
    return len(p), nil
}


func TestDecoder(t *testing.T) {
    files, err := filepath.Glob("testdata/qoi_test_images/*.qoi")
    if err != nil {
        t.Fatalf("%v", err)
    }

    for _, path := range files {
        t.Logf("Now decoding: %v\n", path)
        f, err := os.Open(path)
        if err != nil {
            t.Fatalf("%v", err)
        }
        defer f.Close()

        given, err := Decode(f)
        if err != nil {
            t.Fatalf("%v", err)
        }

        f_, err := os.Open(strings.TrimSuffix(path, filepath.Ext(path)) + ".png")
        if err != nil {
            t.Fatalf("%v", err)
        }
        defer f_.Close()

        wanted, err := png.Decode(f_)
        if err != nil {
            t.Fatalf("%v", err)
        }

        if given.Bounds() != wanted.Bounds() {
            t.Fatalf("Image bounds don't match up!\n")
        }

        for x := 0; x < given.Bounds().Max.X; x++ {
            for y := 0; y < given.Bounds().Max.Y; y++ {
                if color.NRGBAModel.Convert(wanted.At(x, y)) != given.At(x, y) {
                    t.Errorf("Here: (%v, %v). Value QOI: %v, Value PNG: %v\n", x, y, given.At(x, y), wanted.At(x, y))
                }
            }
        }
    }
}

func TestEncoder(t *testing.T) {
    files, err := filepath.Glob("testdata/qoi_test_images/*.png")
    if err != nil {
        t.Fatalf("%v", err)
    }

    for _, path := range files {
        t.Logf("Now encoding: %v\n", path)
        f, err := os.Open(path)
        if err != nil {
            t.Fatalf("%v", err)
        }
        defer f.Close()

        im_png, err := png.Decode(f)
        if err != nil {
            t.Fatalf("%v", err)
        }

        given := &VirtualWriter{make([]byte, 0)}
        Encode(given, im_png)

        f_, err := os.Open(strings.TrimSuffix(path, filepath.Ext(path)) + ".qoi")
        if err != nil {
            t.Fatalf("%v", err)
        }
        defer f_.Close()


        wanted := bufio.NewReader(f_)

        for k, val := range given.d[:12] {
            wanted, err := wanted.ReadByte()
            if err != nil {
                t.Fatalf("%v", err)
            }
            if val != wanted {
                t.Fatalf("Incorrect value in header at position %v. Got: %v. Wanted: %v.\n", k, val, wanted)
            }
        }

        wanted.ReadByte()
        wanted.ReadByte()
        // Considering the official specification, and other implementations, the channel and colorspace values of the header don't need to match exactly.
        if given.d[12] != 3 && given.d[12] != 4 {
            t.Fatalf("Incorrect channel number in header. Got: %v. Wanted: 3 or 4.\n", given.d[12])
        }
        if given.d[13] > 1 {
            t.Fatalf("Incorrect colorspace value in header. Got: %v. Wanted: 0 or 1.\n", given.d[13])
        }

        for k, val := range given.d[14:] {
            ref, err := wanted.ReadByte()
            if err != nil {
                t.Fatalf("%v", err)
            }
            if val != ref {
                t.Errorf("Incorrect value at position %v. Got: %v. Wanted: %v.\n", k, val, wanted)
            }
        }
    }
}


