package main

import (
	//"image/png"
	"log"
	//"os"
    "qoi/qoi"
    //"path/filepath"
    //"strings"
	//"io"
)

var p = log.Printf

func main() {
    qoi.AnalyzeEncodedImagesInFolder("./qoi/testdata/qoi_test_images/")
}

