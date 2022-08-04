package qoi

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type encodingAnalysis struct {
	_ops       [6]byte // Purely to ensure that we always iterate over these in the same order.
	op         map[byte]string
	num_pixels map[byte]int // Pixels encoded by a given tool
	num_bytes  map[byte]int // Bytes used to encode certain amount of pixels
	color      map[byte]color.NRGBA
}

func newEncodingAnalysis() encodingAnalysis {
	d := encodingAnalysis{
		_ops:       [6]byte{},
		op:         map[byte]string{},
		num_pixels: map[byte]int{},
		num_bytes:  map[byte]int{},
		color:      map[byte]color.NRGBA{},
	}

	d._ops[0] = qoi_OP_RGB
	d._ops[1] = qoi_OP_RGBA
	d._ops[2] = qoi_OP_INDEX
	d._ops[3] = qoi_OP_DIFF
	d._ops[4] = qoi_OP_LUMA
	d._ops[5] = qoi_OP_RUN

	d.op[qoi_OP_RGB] = "qoi_OP_RGB"
	d.op[qoi_OP_RGBA] = "qoi_OP_RGBA"
	d.op[qoi_OP_INDEX] = "qoi_OP_INDEX"
	d.op[qoi_OP_DIFF] = "qoi_OP_DIFF"
	d.op[qoi_OP_LUMA] = "qoi_OP_LUMA"
	d.op[qoi_OP_RUN] = "qoi_OP_RUN"

	d.num_pixels[qoi_OP_RGB] = 0
	d.num_pixels[qoi_OP_RGBA] = 0
	d.num_pixels[qoi_OP_INDEX] = 0
	d.num_pixels[qoi_OP_DIFF] = 0
	d.num_pixels[qoi_OP_LUMA] = 0
	d.num_pixels[qoi_OP_RUN] = 0

	d.num_bytes[qoi_OP_RGB] = 0
	d.num_bytes[qoi_OP_RGBA] = 0
	d.num_bytes[qoi_OP_INDEX] = 0
	d.num_bytes[qoi_OP_DIFF] = 0
	d.num_bytes[qoi_OP_LUMA] = 0
	d.num_bytes[qoi_OP_RUN] = 0

	d.color[qoi_OP_RGB] = color.NRGBA{255, 0, 0, 255}
	d.color[qoi_OP_RGBA] = color.NRGBA{255, 0, 255, 255}
	d.color[qoi_OP_INDEX] = color.NRGBA{0, 255, 0, 255}
	d.color[qoi_OP_DIFF] = color.NRGBA{255, 255, 0, 255}
	d.color[qoi_OP_LUMA] = color.NRGBA{0, 0, 255, 255}
	d.color[qoi_OP_RUN] = color.NRGBA{0, 255, 255, 255}

	return d
}

func showEncoding(r io.Reader) (*image.NRGBA, encodingAnalysis, error) {
	header, err := DecodeHeader(r)
	if err != nil {
		return &image.NRGBA{}, encodingAnalysis{}, err
	}

	d := newEncodingAnalysis()

	im := image.NewNRGBA(image.Rect(0, 0, int(header.Width), int(header.Height)))

	buff := *bufio.NewReader(r)
	for pos := 0; pos < len(im.Pix); {
		b, err := buff.ReadByte()
		if err != nil {
			return im, d, err
		}

		run := 1
		px := color.NRGBA{}

		switch b {
		case qoi_OP_RGB:
			px = d.color[b]
			_, _ = buff.ReadByte()
			_, _ = buff.ReadByte()
			_, _ = buff.ReadByte()
			d.num_pixels[b]++
			d.num_bytes[b] += 3

		case qoi_OP_RGBA:
			px = d.color[b]
			_, _ = buff.ReadByte()
			_, _ = buff.ReadByte()
			_, _ = buff.ReadByte()
			_, _ = buff.ReadByte()
			d.num_pixels[b]++
			d.num_bytes[b] += 4

		default:
			switch b & qoi_MASK {
			case qoi_OP_INDEX:
				px = d.color[b&qoi_MASK]
				d.num_pixels[b&qoi_MASK]++
				d.num_bytes[b&qoi_MASK]++

			case qoi_OP_DIFF:
				px = d.color[b&qoi_MASK]
				d.num_pixels[b&qoi_MASK]++
				d.num_bytes[b&qoi_MASK]++

			case qoi_OP_LUMA:
				px = d.color[b&qoi_MASK]
				_, _ = buff.ReadByte()
				d.num_pixels[b&qoi_MASK]++
				d.num_bytes[b&qoi_MASK] += 2

			case qoi_OP_RUN:
				px = d.color[b&qoi_MASK]
				run = int(1 + (b & ^qoi_MASK))
				d.num_pixels[b&qoi_MASK] += run
				d.num_bytes[b&qoi_MASK]++
			}
		}
		for k := 0; k < run; k++ {
			im.Pix[pos] = px.R
			im.Pix[pos+1] = px.G
			im.Pix[pos+2] = px.B
			im.Pix[pos+3] = px.A
			pos += 4
		}
	}
	return im, d, nil
}

// For each .qoi file in folder, create a .PNG file in folder/analysis/ which shows which compression was used for the individual pixels. Also creates .txt files containing some basic statistical information.
func AnalyzeEncodedImagesInFolder(folder string) {
	files, err := filepath.Glob(folder + "/*.qoi")
	if err != nil {
		p("%v", err)
	}

	for _, path := range files {
		f, err := os.Open(path)
		if err != nil {
			p("%v\n", err)
		}
		defer f.Close()

		new_im, d, err := showEncoding(f)
		if err != nil {
			p("%v\n", err)
		}

		target_folder := folder + "/analysis/"
		os.Mkdir(target_folder, os.ModeDir|0b111000000)
		out_im := folder + "/analysis/" + strings.TrimSuffix(filepath.Base(path), ".qoi") + ".png"
		out_txt := folder + "/analysis/" + strings.TrimSuffix(filepath.Base(path), ".qoi") + ".txt"

		w, err := os.Create(out_im)
		if err != nil {
			p("%v\n", err)
		}
		defer w.Close()

		err = png.Encode(w, new_im)

		if err != nil {
			p("%v", err)
		}

		w_txt, err := os.Create(out_txt)
		if err != nil {
			p("%v\n", err)
		}
		defer w_txt.Close()

		txt_out := bufio.NewWriter(w_txt)

		total_bytes := 0
		for _, v := range d.num_bytes {
			total_bytes += v
		}

		for _, k := range d._ops {
			v := d.op[k]
			num_pixels := d.num_pixels[k]
			num_bytes := d.num_bytes[k]

			txt_out.WriteString(fmt.Sprintf("%12v was used to encode %10v pixels (%6.3f%%), using %11v bytes (%6.3f%%) total. ", v, num_pixels, float32(400*num_pixels)/float32(len(new_im.Pix)), num_bytes, float32(100*num_bytes)/float32(total_bytes)))
			txt_out.WriteString(fmt.Sprintf("Bytes per pixel ratio: 1 Byte ~ %6.3f Pixels. Color: %v\n", float32(num_pixels)/float32(num_bytes), d.color[k]))
		}
		txt_out.WriteString(fmt.Sprintf("Overall compression ratio: %v%%\n", float32(100*total_bytes)/float32(len(new_im.Pix))))

		txt_out.Flush()
	}
}
