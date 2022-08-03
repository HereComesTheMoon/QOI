package qoi

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
)


var p = log.Printf


const (
	qoi_OP_RGB   byte = 0b1111_1110
	qoi_OP_RGBA  byte = 0b1111_1111
	qoi_OP_INDEX byte = 0b0000_0000
	qoi_OP_DIFF  byte = 0b0100_0000
	qoi_OP_LUMA  byte = 0b1000_0000
	qoi_OP_RUN   byte = 0b1100_0000

	qoi_MASK byte = 0b1100_0000
)

var endianness binary.ByteOrder = binary.BigEndian

type qoiHeader struct {
	Magic      [4]byte
	Width      uint32
	Height     uint32
	Channels   uint8
	Colorspace uint8
}

type Decoder struct {
	buff bufio.Reader
	prev color.NRGBA
	seen [64]color.NRGBA
}

type decoderData struct {
	px  color.NRGBA
	run int
}

type Encoder struct {
    //im image.Image
	prev color.NRGBA
	seen [64]color.NRGBA
}


var ErrInvalidHeader = errors.New("Invalid QOIF header.")

func DecodeHeader(r io.Reader) (qoiHeader, error) {
	header := qoiHeader{}

	err := binary.Read(r, binary.BigEndian, &header)

	if err != nil {
		return header, fmt.Errorf("Reading input data failed: %w", err)
	}

	if header.Magic != [4]byte{'q', 'o', 'i', 'f'} {
		return header, fmt.Errorf("%w: Missing 'qoif' in header. Found: %v.", ErrInvalidHeader, header.Magic)
	}

	if header.Channels != 0 && header.Channels != 3 && header.Channels != 4 {
		return header, fmt.Errorf("%w: Invalid number of channels: %v.", ErrInvalidHeader, header.Channels)
	}

	if header.Colorspace > 1 {
		return header, fmt.Errorf("%w: Invalid colorspace: %v.", ErrInvalidHeader, header.Colorspace)
	}

	return header, nil
}

func (d *Decoder) nextChunk() (decoderData, error) {
	b, err := d.buff.ReadByte()
	if err != nil {
		return decoderData{}, err
	}

	px := d.prev
	run := 1

	switch b {
	case qoi_OP_RGB:
		vals := []byte{0, 0, 0}
		_, err = io.ReadFull(&d.buff, vals)
		if err != nil {
			return decoderData{}, err
		} // err != nil iff less than three bytes were read in

		px.R = vals[0]
		px.G = vals[1]
		px.B = vals[2]

	case qoi_OP_RGBA:
		vals := []byte{0, 0, 0, 0}
		_, err = io.ReadFull(&d.buff, vals)
		if err != nil {
			return decoderData{}, err
		} // err != nil iff less than four bytes were read in

		px.R = vals[0]
		px.G = vals[1]
		px.B = vals[2]
		px.A = vals[3]

	default:
		switch b & qoi_MASK {
		case qoi_OP_INDEX:
			px = d.seen[b]

		case qoi_OP_DIFF:
			px.R = d.prev.R + (b >> 4 & 0b0000_0011) - 2
			px.G = d.prev.G + (b >> 2 & 0b0000_0011) - 2
			px.B = d.prev.B + (b >> 0 & 0b0000_0011) - 2
			px.A = d.prev.A

		case qoi_OP_LUMA:
			dg := (b & ^qoi_MASK) - 32

			s, err := d.buff.ReadByte()
			if err != nil {
				return decoderData{}, err
			}

			px.R = d.prev.R + dg + (s >> 4 & 0b0000_1111) - 8
			px.G = d.prev.G + dg
			px.B = d.prev.B + dg + (s >> 0 & 0b0000_1111) - 8
			px.A = d.prev.A

		case qoi_OP_RUN:
			run = int(1 + (b & ^qoi_MASK))

		default:
			log.Fatalf("Should not happen: %b & %b == %b ... %b \n", b, qoi_MASK, b&qoi_MASK, qoi_OP_DIFF)
		}
	}

	d.seen[indexHash(px)] = px
	d.prev = px
	return decoderData{px, run}, nil

}

func Decode(r io.Reader) (image.Image, error) {
	header, err := DecodeHeader(r)
	if err != nil {
		return nil, err
	}

	// TODO: Consider to add a warning when surpassing a maximum size in here. The original implementation uses 20000x20000 pixels max, but this is not part of the spec.
	im := image.NewNRGBA(image.Rect(0, 0, int(header.Width), int(header.Height)))

	decoder := Decoder{
		buff: *bufio.NewReader(r),
		prev: color.NRGBA{0, 0, 0, 255},
		seen: [64]color.NRGBA{},
	}

	for pos := 0; pos < len(im.Pix); {
		c, err := decoder.nextChunk()
		if err != nil {
			return im, fmt.Errorf("Invalid chunk data: %w", err)
		}

		for k := 0; k < c.run; k++ {
			im.Pix[pos] = c.px.R
			im.Pix[pos+1] = c.px.G
			im.Pix[pos+2] = c.px.B
			im.Pix[pos+3] = c.px.A
			pos += 4
		}
	}

	return im, nil
}

func indexHash(px color.NRGBA) uint8 {
	return (3*px.R + 5*px.G + 7*px.B + 11*px.A) % 64
}

func getPixel(im image.Image, pos int) color.NRGBA {
    r := im.Bounds()
    //p("Getting pixel at: %v, %v\n", r.Min.X + pos % r.Dx(), r.Min.Y + pos / r.Dx())
    cl := im.At(r.Min.X + pos % r.Dx(), r.Min.Y + pos / r.Dx())
    // Conver from premultiplied alpha to non-premultiplied
    return color.NRGBAModel.Convert(cl).(color.NRGBA)
    //cr, cg, cb, ca := color.NRGBAModel.Convert(cl).RGBA()
    //return color.NRGBA{byte(cr >> 8), byte(cg >> 8), byte(cb >> 8), byte(ca >> 8)}
}

func (e *Encoder) nextPixel(px color.NRGBA) []byte {
    hash := indexHash(px)
    res := []byte{}

    dr := px.R - e.prev.R + 2
    dg := px.G - e.prev.G + 2
    db := px.B - e.prev.B + 2

    if px == e.prev {
        res = []byte{qoi_OP_RUN}
        goto DEFER
    }
    
    if e.seen[hash] == px {
        res = []byte{hash}
        goto DEFER
    }

    //Check if the RGB values of the current and previous pixel have a difference somewhere in -2,-1,0,1. If yes, qoi_OP_DIFF
    if dr < 4 && dg < 4 && db < 4 && e.prev.A == px.A {
        res = []byte{qoi_OP_DIFF | dr << 4 | dg << 2 | db}
        //if res[0] == 0x55 {
            //p("%v, %v, %v, %v,\n", qoi_OP_DIFF, dr << 4, dg << 2, db)
        //}
        goto DEFER
    }

    //Check qoi_OP_LUMA
    dr = dr - dg + 8
    db = db - dg + 8
    dg = dg + 30
    if dg < 64 && dr < 16 && db < 16 && e.prev.A == px.A {
        res = []byte{
            qoi_OP_LUMA | dg,
            dr << 4 | db,
        }
        //p("LUMA: %8b|%8b\n", res[0], res[1])
        goto DEFER
    }

    if px.A == e.prev.A {
        res = []byte{qoi_OP_RGB, px.R, px.G, px.B}
        goto DEFER
    }

    res = []byte{qoi_OP_RGBA, px.R, px.G, px.B, px.A}

    DEFER:
    //if res[0] == 0x55 {
        //log.Fatalf("Something went wrong! %v, %v", res, px)
    //}
    e.prev = px
    e.seen[hash] = px
    return res
}
// TODO: Implement encoder
func Encode(w io.Writer, im image.Image) error {
    buff := bufio.NewWriter(w)
    // Write header
    header := make([]byte, 14)
    header[0] = 'q'
    header[1] = 'o'
    header[2] = 'i'
    header[3] = 'f'
    binary.BigEndian.PutUint32(header[4:8], uint32(im.Bounds().Dx()))
    binary.BigEndian.PutUint32(header[8:12], uint32(im.Bounds().Dy()))
    header[12] = 4
    header[13] = 0

    binary.Write(buff, binary.BigEndian, header)


    number_pixels := im.Bounds().Dx() * im.Bounds().Dy()

    encoder := Encoder{
        prev: color.NRGBA{0, 0, 0, 255},
        seen: [64]color.NRGBA{},
    }

    for pos := 0; pos < number_pixels; pos++ {
        px := getPixel(im, pos)

        //p("%v\n", px)
        nextChunk := encoder.nextPixel(px)

        //p("Pixel: %v\n", px)
        //if nextChunk[0] == 0x6b {
            //log.Fatalf("Something went wrong! %v, prev: %v, %v\n", nextChunk, px, pos)
        //}
        if nextChunk[0] == qoi_OP_RUN {
            var run byte = 0
            for run < 61 {
                runPos := pos + int(run) + 1
                if runPos >= number_pixels {
                    break
                }
                runPx := getPixel(im, runPos)
                if runPx != px {
                    break
                }
                run++
            }
            nextChunk[0] = qoi_OP_RUN | run
            pos += int(run)
        }

        binary.Write(buff, binary.BigEndian, nextChunk)
    }
    binary.Write(buff, binary.BigEndian, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01})
    buff.Flush()
    return nil
}


// TODO: Implement a function which shows which of a pictures pixels were encoded in which way
func showHowEncoded() {}



