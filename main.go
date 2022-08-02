package main

import (
	"log"
    "os"
    "qoi/qoi"
    "image/png"
	//"io"
)

func main() {
    //f, err := os.Open("./qoi_test_images/testcard.qoi")
    //if err != nil {
        //log.Fatalf("No\n.")
    //}
    ////header, err := qoi.DecodeHeader(f)
    ////log.Printf("Header: %v\n", header)
    ////if err != nil {
        ////log.Fatalf("%v.", err)
    ////}
    ////os.Exit(1)

    //im, err := qoi.Decode(f)
    //if err != nil {
        //log.Fatalf("%v.", err)
    //}

    //out, err := os.Create("./own_test_images/testcard.png")
    //if err != nil {
        //log.Fatalf("%v.", err)
    //}
    //err = png.Encode(out, im)
    //if err != nil {
        //log.Fatalf("%v.", err)
    //}

    if !test_img("./qoi_test_images/testcard.qoi", "./qoi_test_images/testcard.png") {
        log.Printf("Nooo")
    }
    
}

func test_img(loc_qoi, loc_png string) bool {
    f, err := os.Open(loc_qoi)
    if err != nil {
        log.Fatalf("%v.", err)
    }
    im_qoi, err := qoi.Decode(f)
    if err != nil {
        log.Fatalf("%v.", err)
    }

    f, err  = os.Open(loc_png)
    if err != nil {
        log.Fatalf("%v.", err)
    }

    im_png, err := png.Decode(f)
    if err != nil {
        log.Fatalf("%v.", err)
    }

    if im_png.Bounds() != im_qoi.Bounds() {
        log.Fatalf("Image bounds don't match up!\n")
    }

    for x := 0; x < im_qoi.Bounds().Max.X ; x++ {
        for y := 0; y < im_qoi.Bounds().Max.Y ; y++ {
            if im_png.At(x, y) != im_qoi.At(x, y) {
                log.Printf("Here: (%v, %v). Value QOI: %v, Value PNG: %v\n", x, y, im_qoi.At(x,y), im_png.At(x,y))
                //return false
            }
        }
    }
    return true
}