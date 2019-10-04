package main

import (
	"github.com/qeesung/image2ascii/ascii"

	tb "github.com/nsf/termbox-go"
)

type View struct {
	width int
	heigh int
	cells [][]tb.Cell
}

// Inspired from github.com/aybabtme/rgbterm
func colorise(r, g, b uint8) uint16 {
	// if all colors are equal, it might be in the grayscale range
	if r == g && g == b {
		color, ok := grayscale(r)
		if ok {
			return color
		}
	}

	// the general case approximates RGB by using the closest color.
	r6 := ((uint16(r) * 5) / 255)
	g6 := ((uint16(g) * 5) / 255)
	b6 := ((uint16(b) * 5) / 255)
	i := 36*r6 + 6*g6 + b6
	return 17 + i
}

func grayscale(scale uint8) (uint16, bool) {

	switch scale {
	case 0x08:
		return 232, true
	case 0x12:
		return 233, true
	case 0x1c:
		return 234, true
	case 0x26:
		return 235, true
	case 0x30:
		return 236, true
	case 0x3a:
		return 237, true
	case 0x44:
		return 238, true
	case 0x4e:
		return 239, true
	case 0x58:
		return 240, true
	case 0x62:
		return 241, true
	case 0x6c:
		return 242, true
	case 0x76:
		return 243, true
	case 0x80:
		return 244, true
	case 0x8a:
		return 245, true
	case 0x94:
		return 246, true
	case 0x9e:
		return 247, true
	case 0xa8:
		return 248, true
	case 0xb2:
		return 249, true
	case 0xbc:
		return 250, true
	case 0xc6:
		return 251, true
	case 0xd0:
		return 252, true
	case 0xda:
		return 253, true
	case 0xe4:
		return 254, true
	case 0xee:
		return 255, true
	}
	return 0, false
}

func NewView(width, heigh int, pixels [][]ascii.CharPixel) *View {
	view := &View{
		width: width,
		heigh: heigh,
	}
	view.cells = make([][]tb.Cell, heigh)
	for line := range view.cells {
		view.cells[line] = make([]tb.Cell, width)
	}

	for bx := 0; bx < width; bx++ {
		for by := 0; by < heigh; by++ {
			pix := pixels[by][bx]
			col := colorise(pix.R, pix.G, pix.B)

			view.cells[by][bx] = tb.Cell{
				rune(pixels[by][bx].Char),
				tb.Attribute(col),
				tb.Attribute(0),
			}
		}
	}
	return view
}

func (self *View) Print() {

	for bx := 0; bx < self.width; bx++ {
		for by := 0; by < self.heigh; by++ {
			c := self.cells[by][bx]
			tb.SetCell(bx, by, c.Ch, c.Fg, c.Bg)
		}
	}
	tb.Flush()
}
