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

const (
	qoi_OP_RGB   byte = 0b1111_1110
	qoi_OP_RGBA  byte = 0b1111_1111
	qoi_OP_INDEX byte = 0b0000_0000
	qoi_OP_DIFF  byte = 0b0100_0000
	qoi_OP_LUMA  byte = 0b1000_0000
	qoi_OP_RUN   byte = 0b1100_0000

	qoi_MASK byte = 0b1100_0000
)

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

type chunkData struct {
	px  color.NRGBA
	run int
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

func (d *Decoder) nextChunk() (chunkData, error) {
	b, err := d.buff.ReadByte()
	if err != nil {
		return chunkData{}, err
	}

	px := d.prev
	run := 1

	switch b {
	case qoi_OP_RGB:
		vals := []byte{0, 0, 0}
		_, err = io.ReadFull(&d.buff, vals)
		if err != nil {
			return chunkData{}, err
		} // err != nil iff less than three bytes were read in

		px.R = vals[0]
		px.G = vals[1]
		px.B = vals[2]

	case qoi_OP_RGBA:
		vals := []byte{0, 0, 0, 0}
		_, err = io.ReadFull(&d.buff, vals)
		if err != nil {
			return chunkData{}, err
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
				return chunkData{}, err
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
	return chunkData{px, run}, nil

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

// TODO: Implement a function which shows which of a pictures pixels were encoded in which way
func showHowEncoded() {}


// TODO: Implement encoder

