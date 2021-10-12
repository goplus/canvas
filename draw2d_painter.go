package canvas

import (
	"crypto/md5"
	"image"
	"image/color"

	stackblur "github.com/esimov/stackblur-go"
	"github.com/golang/freetype/raster"
)

type AlphaOverPainter struct {
	Image *image.Alpha
	rect  image.Rectangle
}

// Paint satisfies the Painter interface.
func (r *AlphaOverPainter) Paint(ss []raster.Span, done bool) {
	b := r.Image.Bounds()
	for _, s := range ss {
		if s.Y < b.Min.Y {
			continue
		}
		if s.Y >= b.Max.Y {
			return
		}
		if s.X0 < b.Min.X {
			s.X0 = b.Min.X
		}
		if s.X1 > b.Max.X {
			s.X1 = b.Max.X
		}
		if s.X0 >= s.X1 {
			continue
		}
		if r.rect.Min.X > s.X0 {
			r.rect.Min.X = s.X0
		}
		if r.rect.Min.Y > s.Y {
			r.rect.Min.Y = s.Y
		}
		if r.rect.Max.Y < s.Y {
			r.rect.Max.Y = s.Y
		}
		if r.rect.Max.X < s.X1 {
			r.rect.Max.X = s.X1
		}

		base := (s.Y-r.Image.Rect.Min.Y)*r.Image.Stride - r.Image.Rect.Min.X
		p := r.Image.Pix[base+s.X0 : base+s.X1]
		a := int(s.Alpha >> 8)
		for i, c := range p {
			v := int(c)
			p[i] = uint8((v*255 + (255-v)*a) / 255)
		}
	}
}

// NewAlphaOverPainter creates a new AlphaOverPainter for the given image.
func NewAlphaOverPainter(m *image.Alpha) *AlphaOverPainter {
	return &AlphaOverPainter{m, image.Rectangle{image.Point{1e9, 1e9}, image.Point{-1e9, -1e9}}}
}

type RGBAPainter struct {
	Image         *image.RGBA
	Mask          *image.Alpha
	Op            CompositeOperation
	empty         *image.RGBA
	canvas        *image.RGBA
	canvasOp      CompositeOperation
	shadow        *image.RGBA
	spanRect      image.Rectangle
	pattern       Pattern
	solid         bool
	solidColor    color.Color
	canvasAlpha   float64
	shadowOffsetX float64
	shadowOffsetY float64
	shadowBlur    float64
	shadowColor   color.Color
	hasShadow     bool
}

func (r *RGBAPainter) SetGlobalAlpha(alpha float64) {
	r.canvasAlpha = alpha
}

func (r *RGBAPainter) SetShadow(offsetX float64, offsetY float64, blur float64, clr color.Color) {
	r.shadowOffsetX = offsetX
	r.shadowOffsetY = offsetY
	r.shadowBlur = blur
	r.shadowColor = clr
	_, _, _, a := r.shadowColor.RGBA()
	if a != 0 && (offsetX != 0 || offsetY != 0 || blur != 0) {
		r.hasShadow = true
	} else {
		r.hasShadow = false
	}
}

// Paint satisfies the canvas interface.
func (r *RGBAPainter) Paint(ss []raster.Span, done bool) {
	if r.solid {
		if r.Mask != nil && !r.hasShadow {
			r.paintSolidMask(ss, done)
		} else {
			r.paintSolid(ss, done)
		}
	} else {
		if r.Mask != nil && !r.hasShadow {
			r.paintPatternMask(ss, done)
		} else {
			r.paintPattern(ss, done)
		}
	}
}

func (r *RGBAPainter) SetColor(c color.Color) {
	r.solidColor = c
	r.solid = true
}

func (r *RGBAPainter) paintSolid(ss []raster.Span, done bool) {
	cr, cg, cb, ca := r.solidColor.RGBA()
	b := r.canvas.Bounds()
	for _, s := range ss {
		if s.Y < b.Min.Y {
			continue
		}
		if s.Y >= b.Max.Y {
			return
		}
		if s.X0 < b.Min.X {
			s.X0 = b.Min.X
		}
		if s.X1 > b.Max.X {
			s.X1 = b.Max.X
		}
		if s.X0 >= s.X1 {
			continue
		}
		if r.spanRect.Min.X > s.X0 {
			r.spanRect.Min.X = s.X0
		}
		if r.spanRect.Min.Y > s.Y {
			r.spanRect.Min.Y = s.Y
		}
		if r.spanRect.Max.Y < s.Y {
			r.spanRect.Max.Y = s.Y
		}
		if r.spanRect.Max.X < s.X1 {
			r.spanRect.Max.X = s.X1
		}
		ma := s.Alpha
		ma = uint32(float64(ma) * r.canvasAlpha)
		const m = 1<<16 - 1
		i0 := (s.Y-r.canvas.Rect.Min.Y)*r.canvas.Stride + (s.X0-r.canvas.Rect.Min.X)*4
		i1 := i0 + (s.X1-s.X0)*4
		if r.canvasOp == SourceOver {
			for i := i0; i < i1; i += 4 {
				dr := uint32(r.canvas.Pix[i+0])
				dg := uint32(r.canvas.Pix[i+1])
				db := uint32(r.canvas.Pix[i+2])
				da := uint32(r.canvas.Pix[i+3])
				a := (m - (ca * ma / m)) * 0x101
				r.canvas.Pix[i+0] = uint8((dr*a + cr*ma) / m >> 8)
				r.canvas.Pix[i+1] = uint8((dg*a + cg*ma) / m >> 8)
				r.canvas.Pix[i+2] = uint8((db*a + cb*ma) / m >> 8)
				r.canvas.Pix[i+3] = uint8((da*a + ca*ma) / m >> 8)
			}
		} else {
			for i := i0; i < i1; i += 4 {
				r.canvas.Pix[i+0] = uint8(cr * ma / m >> 8)
				r.canvas.Pix[i+1] = uint8(cg * ma / m >> 8)
				r.canvas.Pix[i+2] = uint8(cb * ma / m >> 8)
				r.canvas.Pix[i+3] = uint8(ca * ma / m >> 8)
			}
		}
	}
}

func (r *RGBAPainter) paintSolidMask(ss []raster.Span, done bool) {
	cr, cg, cb, ca := r.solidColor.RGBA()
	b := r.canvas.Bounds()
	for _, s := range ss {
		if s.Y < b.Min.Y {
			continue
		}
		if s.Y >= b.Max.Y {
			return
		}
		if s.X0 < b.Min.X {
			s.X0 = b.Min.X
		}
		if s.X1 > b.Max.X {
			s.X1 = b.Max.X
		}
		if s.X0 >= s.X1 {
			continue
		}
		if r.spanRect.Min.X > s.X0 {
			r.spanRect.Min.X = s.X0
		}
		if r.spanRect.Min.Y > s.Y {
			r.spanRect.Min.Y = s.Y
		}
		if r.spanRect.Max.Y < s.Y {
			r.spanRect.Max.Y = s.Y
		}
		if r.spanRect.Max.X < s.X1 {
			r.spanRect.Max.X = s.X1
		}
		const m = 1<<16 - 1
		y := s.Y - r.canvas.Rect.Min.Y
		x0 := s.X0 - r.canvas.Rect.Min.X
		i0 := (s.Y-r.canvas.Rect.Min.Y)*r.canvas.Stride + (s.X0-r.canvas.Rect.Min.X)*4
		i1 := i0 + (s.X1-s.X0)*4
		if r.canvasOp == SourceOver {
			for i, x := i0, x0; i < i1; i, x = i+4, x+1 {
				ma := s.Alpha
				ma = uint32(float64(ma) * r.canvasAlpha)
				ma = ma * uint32(r.Mask.AlphaAt(x, y).A) / 255
				if ma == 0 {
					continue
				}
				dr := uint32(r.canvas.Pix[i+0])
				dg := uint32(r.canvas.Pix[i+1])
				db := uint32(r.canvas.Pix[i+2])
				da := uint32(r.canvas.Pix[i+3])
				a := (m - (ca * ma / m)) * 0x101
				r.canvas.Pix[i+0] = uint8((dr*a + cr*ma) / m >> 8)
				r.canvas.Pix[i+1] = uint8((dg*a + cg*ma) / m >> 8)
				r.canvas.Pix[i+2] = uint8((db*a + cb*ma) / m >> 8)
				r.canvas.Pix[i+3] = uint8((da*a + ca*ma) / m >> 8)
			}
		} else {
			for i, x := i0, x0; i < i1; i, x = i+4, x+1 {
				ma := s.Alpha
				ma = uint32(float64(ma) * r.canvasAlpha)
				ma = ma * uint32(r.Mask.AlphaAt(x, y).A) / 255
				if ma == 0 {
					continue
				}
				r.canvas.Pix[i+0] = uint8(cr * ma / m >> 8)
				r.canvas.Pix[i+1] = uint8(cg * ma / m >> 8)
				r.canvas.Pix[i+2] = uint8(cb * ma / m >> 8)
				r.canvas.Pix[i+3] = uint8(ca * ma / m >> 8)
			}
		}
	}
}

func (r *RGBAPainter) paintPattern(ss []raster.Span, done bool) {
	b := r.canvas.Bounds()
	for _, s := range ss {
		if s.Y < b.Min.Y {
			continue
		}
		if s.Y >= b.Max.Y {
			return
		}
		if s.X0 < b.Min.X {
			s.X0 = b.Min.X
		}
		if s.X1 > b.Max.X {
			s.X1 = b.Max.X
		}
		if s.X0 >= s.X1 {
			continue
		}
		if r.spanRect.Min.X > s.X0 {
			r.spanRect.Min.X = s.X0
		}
		if r.spanRect.Min.Y > s.Y {
			r.spanRect.Min.Y = s.Y
		}
		if r.spanRect.Max.Y < s.Y {
			r.spanRect.Max.Y = s.Y
		}
		if r.spanRect.Max.X < s.X1 {
			r.spanRect.Max.X = s.X1
		}
		ma := s.Alpha
		ma = uint32(float64(ma) * r.canvasAlpha)
		const m = 1<<16 - 1
		y := s.Y - r.canvas.Rect.Min.Y
		x0 := s.X0 - r.canvas.Rect.Min.X
		// RGBAcanvas.Paint() in $GOPATH/src/github.com/golang/freetype/raster/paint.go
		i0 := (s.Y-r.canvas.Rect.Min.Y)*r.canvas.Stride + (s.X0-r.canvas.Rect.Min.X)*4
		i1 := i0 + (s.X1-s.X0)*4
		if r.canvasOp == SourceOver {
			for i, x := i0, x0; i < i1; i, x = i+4, x+1 {
				c := r.pattern.ColorAt(x, y)
				cr, cg, cb, ca := c.RGBA()
				dr := uint32(r.canvas.Pix[i+0])
				dg := uint32(r.canvas.Pix[i+1])
				db := uint32(r.canvas.Pix[i+2])
				da := uint32(r.canvas.Pix[i+3])
				a := (m - (ca * ma / m)) * 0x101
				r.canvas.Pix[i+0] = uint8((dr*a + cr*ma) / m >> 8)
				r.canvas.Pix[i+1] = uint8((dg*a + cg*ma) / m >> 8)
				r.canvas.Pix[i+2] = uint8((db*a + cb*ma) / m >> 8)
				r.canvas.Pix[i+3] = uint8((da*a + ca*ma) / m >> 8)
			}
		} else {
			for i, x := i0, x0; i < i1; i, x = i+4, x+1 {
				c := r.pattern.ColorAt(x, y)
				cr, cg, cb, ca := c.RGBA()
				r.canvas.Pix[i+0] = uint8(cr * ma / m >> 8)
				r.canvas.Pix[i+1] = uint8(cg * ma / m >> 8)
				r.canvas.Pix[i+2] = uint8(cb * ma / m >> 8)
				r.canvas.Pix[i+3] = uint8(ca * ma / m >> 8)
			}
		}
	}
}

func (r *RGBAPainter) paintPatternMask(ss []raster.Span, done bool) {
	b := r.canvas.Bounds()
	for _, s := range ss {
		if s.Y < b.Min.Y {
			continue
		}
		if s.Y >= b.Max.Y {
			return
		}
		if s.X0 < b.Min.X {
			s.X0 = b.Min.X
		}
		if s.X1 > b.Max.X {
			s.X1 = b.Max.X
		}
		if s.X0 >= s.X1 {
			continue
		}
		if r.spanRect.Min.X > s.X0 {
			r.spanRect.Min.X = s.X0
		}
		if r.spanRect.Min.Y > s.Y {
			r.spanRect.Min.Y = s.Y
		}
		if r.spanRect.Max.Y < s.Y {
			r.spanRect.Max.Y = s.Y
		}
		if r.spanRect.Max.X < s.X1 {
			r.spanRect.Max.X = s.X1
		}
		const m = 1<<16 - 1
		y := s.Y - r.canvas.Rect.Min.Y
		x0 := s.X0 - r.canvas.Rect.Min.X
		// RGBAcanvas.Paint() in $GOPATH/src/github.com/golang/freetype/raster/paint.go
		i0 := (s.Y-r.canvas.Rect.Min.Y)*r.canvas.Stride + (s.X0-r.canvas.Rect.Min.X)*4
		i1 := i0 + (s.X1-s.X0)*4
		if r.canvasOp == SourceOver {
			for i, x := i0, x0; i < i1; i, x = i+4, x+1 {
				ma := s.Alpha
				ma = uint32(float64(ma) * r.canvasAlpha)
				ma = ma * uint32(r.Mask.AlphaAt(x, y).A) / 255
				if ma == 0 {
					continue
				}
				c := r.pattern.ColorAt(x, y)
				cr, cg, cb, ca := c.RGBA()
				dr := uint32(r.canvas.Pix[i+0])
				dg := uint32(r.canvas.Pix[i+1])
				db := uint32(r.canvas.Pix[i+2])
				da := uint32(r.canvas.Pix[i+3])
				a := (m - (ca * ma / m)) * 0x101
				r.canvas.Pix[i+0] = uint8((dr*a + cr*ma) / m >> 8)
				r.canvas.Pix[i+1] = uint8((dg*a + cg*ma) / m >> 8)
				r.canvas.Pix[i+2] = uint8((db*a + cb*ma) / m >> 8)
				r.canvas.Pix[i+3] = uint8((da*a + ca*ma) / m >> 8)
			}
		} else {
			for i, x := i0, x0; i < i1; i, x = i+4, x+1 {
				ma := s.Alpha
				ma = uint32(float64(ma) * r.canvasAlpha)
				ma = ma * uint32(r.Mask.AlphaAt(x, y).A) / 255
				if ma == 0 {
					continue
				}
				c := r.pattern.ColorAt(x, y)
				cr, cg, cb, ca := c.RGBA()
				r.canvas.Pix[i+0] = uint8((cr * ma) / m >> 8)
				r.canvas.Pix[i+1] = uint8((cg * ma) / m >> 8)
				r.canvas.Pix[i+2] = uint8((cb * ma) / m >> 8)
				r.canvas.Pix[i+3] = uint8((ca * ma) / m >> 8)
			}
		}
	}
}

func NewRGBAPainter(img *image.RGBA) *RGBAPainter {
	r := &RGBAPainter{Image: img, canvasAlpha: 1.0, Op: SourceOver}
	r.empty = image.NewRGBA(img.Rect)
	r.canvas = image.NewRGBA(img.Rect)
	return r
}

func (r *RGBAPainter) SetMask(mask *image.Alpha) {
	r.Mask = mask
}

func (r *RGBAPainter) SetPattern(p Pattern, m Matrix) {
	if solid, ok := p.(*SolidPattern); ok {
		r.solid = true
		r.SetColor(solid.color)
	} else {
		r.solid = false
		if m.IsIdentity() {
			r.pattern = p
		} else {
			r.pattern = &tranPattern{p, m}
		}
	}
}

func (r *RGBAPainter) SetCompositeOperation(op CompositeOperation) {
	r.Op = op
	r.canvasOp = op
}

func (r *RGBAPainter) Begin() {
	r.spanRect.Min.X = 1e9
	r.spanRect.Min.Y = 1e9
	r.spanRect.Max.X = -1e9
	r.spanRect.Max.Y = -1e9
	if !r.hasShadow {
		switch r.Op {
		case SourceOver:
			r.canvas = r.Image
			r.canvasOp = SourceOver
			return
		}
	}
	rect := r.Image.Rect
	if r.hasShadow {
		if r.shadowOffsetX > 0 {
			rect.Min.X -= int(r.shadowOffsetX)
		} else if r.shadowOffsetX < 0 {
			rect.Max.X -= int(r.shadowOffsetX)
		}
		if r.shadowOffsetY > 0 {
			rect.Min.Y -= int(r.shadowOffsetY)
		} else if r.shadowOffsetY < 0 {
			rect.Max.Y -= int(r.shadowOffsetY)
		}
	}
	r.canvas = image.NewRGBA(rect)
	r.canvasOp = Copy
}

type blur_info struct {
	offsetX int
	offsetY int
	blur    int
	sum     [16]byte
	rect    image.Rectangle
}

var (
	blurCache = make(map[blur_info]image.Image)
)

func (r *RGBAPainter) blur(src *image.RGBA) image.Image {
	sum := md5.Sum(src.Pix)
	info := blur_info{
		int(r.shadowOffsetX),
		int(r.shadowOffsetY),
		int(r.shadowBlur),
		sum,
		r.canvas.Rect,
	}
	if img, ok := blurCache[info]; ok {
		return img
	}
	rect := image.Rect(0, 0, r.canvas.Rect.Dx(), r.canvas.Rect.Dy())
	if r.shadowOffsetX < 0 {
		rect.Min.X += int(r.shadowOffsetX)
		rect.Max.X += int(r.shadowOffsetX)
	}
	if r.shadowOffsetY < 0 {
		rect.Min.Y += int(r.shadowOffsetY)
		rect.Max.Y += int(r.shadowOffsetY)
	}
	var img image.Image
	img = image.NewRGBA(rect)
	copy(img.(*image.RGBA).Pix, src.Pix)
	if r.shadowBlur > 0 {
		img = stackblur.Process(img, uint32(r.shadowBlur))
	}
	blurCache[info] = img
	return img
}

func (r *RGBAPainter) endShadow() {
	shadow := r.blur(r.canvas)
	dx, dy := r.Image.Rect.Dx(), r.Image.Rect.Dy()
	const m = 1<<16 - 1
	cr, cg, cb, ca := r.shadowColor.RGBA()
	var ma uint32
	switch r.Op {
	case SourceAtop, SourceOver, DestinationOver, DestinationOut, Lighter, Xor:
		overRect := r.spanRect.Add(image.Point{int(r.shadowOffsetX), int(r.shadowOffsetY)})
		overRect.Min.X -= int(r.shadowBlur)
		overRect.Min.Y -= int(r.shadowBlur)
		overRect.Max.X += int(r.shadowBlur)
		overRect.Max.Y += int(r.shadowBlur)
		overRect = overRect.Union(r.spanRect)
		for y := 0; y < dy; y++ {
			if y < overRect.Min.Y || y > overRect.Max.Y {
				continue
			}
			for x := 0; x < dx; x++ {
				if x < overRect.Min.X || x > overRect.Max.X {
					continue
				}
				if r.Mask != nil && r.Mask.AlphaAt(x, y).A == 0 {
					continue
				}
				if int(r.shadowBlur) == 0 {
					i0 := shadow.(*image.RGBA).PixOffset(x, y)
					ma = uint32(shadow.(*image.RGBA).Pix[i0+3]) << 8
				} else {
					i0 := shadow.(*image.NRGBA).PixOffset(x, y)
					ma = uint32(shadow.(*image.NRGBA).Pix[i0+3]) << 8
				}
				r0 := uint8((cr * ma) / m >> 8)
				g0 := uint8((cg * ma) / m >> 8)
				b0 := uint8((cb * ma) / m >> 8)
				a0 := uint8((ca * ma) / m >> 8)
				shc := color.RGBA{r0, g0, b0, a0}
				i := r.canvas.PixOffset(x, y)
				src := color.RGBA{r.canvas.Pix[i+0], r.canvas.Pix[i+1], r.canvas.Pix[i+2], r.canvas.Pix[i+3]}
				i = r.Image.PixOffset(x, y)
				dst := color.RGBA{r.Image.Pix[i+0], r.Image.Pix[i+1], r.Image.Pix[i+2], r.Image.Pix[i+3]} //r.Image.At(x, y)
				dst = r.Op.ComposeRGBA(shc, dst)
				rgba := r.Op.ComposeRGBA(src, dst)
				r.Image.Pix[i] = rgba.R
				r.Image.Pix[i+1] = rgba.G
				r.Image.Pix[i+2] = rgba.B
				r.Image.Pix[i+3] = rgba.A
			}
		}
	default:
		for y := 0; y < dy; y++ {
			for x := 0; x < dx; x++ {
				if r.Mask != nil && r.Mask.AlphaAt(x, y).A == 0 {
					continue
				}
				if int(r.shadowBlur) == 0 {
					i0 := shadow.(*image.RGBA).PixOffset(x, y)
					ma = uint32(shadow.(*image.RGBA).Pix[i0+3]) << 8
				} else {
					i0 := shadow.(*image.NRGBA).PixOffset(x, y)
					ma = uint32(shadow.(*image.NRGBA).Pix[i0+3]) << 8
				}
				r0 := uint8((cr * ma) / m >> 8)
				g0 := uint8((cg * ma) / m >> 8)
				b0 := uint8((cb * ma) / m >> 8)
				a0 := uint8((ca * ma) / m >> 8)
				shc := color.RGBA{r0, g0, b0, a0}
				i := r.canvas.PixOffset(x, y)
				src := color.RGBA{r.canvas.Pix[i+0], r.canvas.Pix[i+1], r.canvas.Pix[i+2], r.canvas.Pix[i+3]}
				i = r.Image.PixOffset(x, y)
				dst := color.RGBA{r.Image.Pix[i+0], r.Image.Pix[i+1], r.Image.Pix[i+2], r.Image.Pix[i+3]} //r.Image.At(x, y)
				dst = r.Op.ComposeRGBA(shc, dst)
				rgba := r.Op.ComposeRGBA(src, dst)
				r.Image.Pix[i] = rgba.R
				r.Image.Pix[i+1] = rgba.G
				r.Image.Pix[i+2] = rgba.B
				r.Image.Pix[i+3] = rgba.A
			}
		}
	}
}

func (r *RGBAPainter) End() {
	if r.hasShadow {
		r.endShadow()
		return
	}
	switch r.Op {
	case SourceOver:
		return
	case SourceAtop, DestinationOver, DestinationOut, Lighter, Xor:
		dx, dy := r.Image.Rect.Dx(), r.Image.Rect.Dy()
		for y := 0; y < dy; y++ {
			if y < r.spanRect.Min.Y || y > r.spanRect.Max.Y {
				continue
			}
			for x := 0; x < dx; x++ {
				if x < r.spanRect.Min.X || x > r.spanRect.Max.X {
					continue
				}
				i := r.canvas.PixOffset(x, y)
				src := color.RGBA{r.canvas.Pix[i+0], r.canvas.Pix[i+1], r.canvas.Pix[i+2], r.canvas.Pix[i+3]}
				dst := color.RGBA{r.Image.Pix[i+0], r.Image.Pix[i+1], r.Image.Pix[i+2], r.Image.Pix[i+3]} //r.Image.At(x, y)
				rgba := r.Op.ComposeRGBA(src, dst)
				r.Image.Pix[i] = rgba.R
				r.Image.Pix[i+1] = rgba.G
				r.Image.Pix[i+2] = rgba.B
				r.Image.Pix[i+3] = rgba.A
			}
		}
	default:
		dx, dy := r.Image.Rect.Dx(), r.Image.Rect.Dy()
		for y := 0; y < dy; y++ {
			for x := 0; x < dx; x++ {
				i := r.canvas.PixOffset(x, y)
				src := color.RGBA{r.canvas.Pix[i+0], r.canvas.Pix[i+1], r.canvas.Pix[i+2], r.canvas.Pix[i+3]}
				dst := color.RGBA{r.Image.Pix[i+0], r.Image.Pix[i+1], r.Image.Pix[i+2], r.Image.Pix[i+3]} //r.Image.At(x, y)
				rgba := r.Op.ComposeRGBA(src, dst)
				r.Image.Pix[i] = rgba.R
				r.Image.Pix[i+1] = rgba.G
				r.Image.Pix[i+2] = rgba.B
				r.Image.Pix[i+3] = rgba.A
			}
		}
	}
}

type tranPattern struct {
	p Pattern
	m Matrix
}

func (sp *tranPattern) ColorAt(x, y int) color.Color {
	rx, ry := sp.m.InverseTransformPoint(float64(x), float64(y))
	return sp.p.ColorAt(int(rx), int(ry))
}
