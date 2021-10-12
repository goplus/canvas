// this code copy from https://godoc.org/github.com/fogleman/gg

package canvas

import (
	"fmt"
	"image"
	"image/color"
	"strings"
)

type RepeatOp int

const (
	Repeat RepeatOp = iota
	RepeatX
	RepeatY
	RepeatNone
)

func (p RepeatOp) String() string {
	switch p {
	case Repeat:
		return "repeat"
	case RepeatX:
		return "repeat-x"
	case RepeatY:
		return "repeat-y"
	case RepeatNone:
		return "no-repeat"
	}
	return ""
}

type Pattern interface {
	ColorAt(x, y int) color.Color
}

// Solid Pattern
type SolidPattern struct {
	color color.Color
}

func (p *SolidPattern) ColorAt(x, y int) color.Color {
	return p.color
}

func NewSolidPattern(color color.Color) Pattern {
	return &SolidPattern{color: color}
}

// Surface Pattern
type surfacePattern struct {
	im image.Image
	op RepeatOp
}

func (p *surfacePattern) ColorAt(x, y int) color.Color {
	b := p.im.Bounds()
	switch p.op {
	case RepeatX:
		x = x % b.Dx()
		return p.im.At(x, y)
	case RepeatY:
		y = y % b.Dy()
	case Repeat:
		x = x % b.Dx()
		y = y % b.Dy()
	}
	pt := image.Point{x, y}
	if !pt.In(b) {
		return color.Transparent
	}
	return p.im.At(x, y)
}

func newSurfacePattern(im image.Image, op RepeatOp) Pattern {
	return &surfacePattern{im: im, op: op}
}

func HexColor(x string) color.Color {
	r, g, b, a := parseHexColor(x)
	return color.NRGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
}

func parseHexColor(x string) (r, g, b, a int) {
	x = strings.TrimPrefix(x, "#")
	a = 255
	if len(x) == 3 {
		format := "%1x%1x%1x"
		fmt.Sscanf(x, format, &r, &g, &b)
		r |= r << 4
		g |= g << 4
		b |= b << 4
	}
	if len(x) == 6 {
		format := "%02x%02x%02x"
		fmt.Sscanf(x, format, &r, &g, &b)
	}
	if len(x) == 8 {
		format := "%02x%02x%02x%02x"
		fmt.Sscanf(x, format, &r, &g, &b, &a)
	}
	return
}
