package canvas

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
)

type TextAlign int

const (
	AlignLeft TextAlign = iota
	AlignCenter
	AlignRight
)

func (a TextAlign) String() string {
	switch a {
	case AlignLeft:
		return "left"
	case AlignCenter:
		return "center"
	case AlignRight:
		return "right"
	}
	return ""
}

func ParserTextAlign(x string) TextAlign {
	if x == "left" || x == "start" {
		return AlignLeft
	} else if x == "right" || x == "end" {
		return AlignRight
	} else if x == "center" {
		return AlignCenter
	}
	return AlignLeft
}

type TextBaseline int

const (
	AlignAlphabetic TextBaseline = iota
	AlignTop
	AlignHanging
	AlignMiddle
	AlignIdeographic
	AlignBottom
)

func (b TextBaseline) String() string {
	switch b {
	case AlignAlphabetic:
		return "alphabetic"
	case AlignTop:
		return "top"
	case AlignHanging:
		return "hanging"
	case AlignMiddle:
		return "middle"
	case AlignIdeographic:
		return "ideographic"
	case AlignBottom:
		return "bottom"
	}
	return ""
}

func ParserTextBaseline(x string) TextBaseline {
	switch x {
	case "alphabetic":
		return AlignAlphabetic
	case "top":
		return AlignTop
	case "hanging":
		return AlignHanging
	case "middle":
		return AlignMiddle
	case "ideographic":
		return AlignIdeographic
	case "bottom":
		return AlignBottom
	}
	return AlignAlphabetic
}

type LineCap int

const (
	LineCapButt LineCap = iota
	LineCapRound
	LineCapSquare
)

func (c LineCap) String() string {
	switch c {
	case LineCapButt:
		return "butt"
	case LineCapRound:
		return "round"
	case LineCapSquare:
		return "square"
	}
	return ""
}

func ParserLineCap(x string) LineCap {
	if x == "round" {
		return LineCapRound
	} else if x == "square" {
		return LineCapSquare
	}
	return LineCapButt
}

type LineJoin int

const (
	LineJoinMiter LineJoin = iota
	LineJoinRound
	LineJoinBevel
)

func (join LineJoin) String() string {
	switch join {
	case LineJoinMiter:
		return "miter"
	case LineJoinRound:
		return "round"
	case LineJoinBevel:
		return "bevel"
	}
	return ""
}

func ParserLineJoin(x string) LineJoin {
	if x == "round" {
		return LineJoinRound
	} else if x == "bevel" {
		return LineJoinBevel
	}
	return LineJoinMiter
}

type CompositeOperation int

const (
	SourceAtop CompositeOperation = iota
	SourceIn
	SourceOut
	SourceOver
	DestinationAtop
	DestinationIn
	DestinationOut
	DestinationOver
	Lighter
	Copy
	Xor
)

func (op CompositeOperation) String() string {
	switch op {
	case SourceOver:
		return "source-over"
	case SourceAtop:
		return "source-atop"
	case SourceIn:
		return "source-in"
	case SourceOut:
		return "source-out"
	case DestinationOver:
		return "destination-over"
	case DestinationAtop:
		return "destination-atop"
	case DestinationIn:
		return "destination-in"
	case DestinationOut:
		return "destination-out"
	case Lighter:
		return "lighter"
	case Copy:
		return "copy"
	case Xor:
		return "xor"
	}
	return ""
}

type FRGBA struct {
	R, G, B, A float64
}

func (c FRGBA) Mul(alpha float64) FRGBA {
	return FRGBA{c.R * alpha, c.G * alpha, c.B * alpha, c.A * alpha}
}

func (c FRGBA) Add(x FRGBA) FRGBA {
	r := c.R + x.R
	g := c.G + x.G
	b := c.B + x.B
	a := c.A + x.A
	if r > 1 {
		r = 1
	}
	if g > 1 {
		g = 1
	}
	if b > 1 {
		b = 1
	}
	if a > 1 {
		a = 1
	}
	return FRGBA{r, g, b, a}
}

func (c FRGBA) RGBA() (r, g, b, a uint32) {
	r = uint32(0xffff * c.R)
	g = uint32(0xffff * c.G)
	b = uint32(0xffff * c.B)
	a = uint32(0xffff * c.A)
	return
}

func ToFRGBA(c color.Color) FRGBA {
	if c, ok := c.(FRGBA); ok {
		return c
	}
	r, g, b, a := c.RGBA()
	return FRGBA{
		float64(r) / 0xffff,
		float64(g) / 0xffff,
		float64(b) / 0xffff,
		float64(a) / 0xffff,
	}
}

func makeRGBA(r, g, b, a uint32) color.RGBA {
	if r > 0xff {
		r = 0xff
	}
	if g > 0xff {
		g = 0xff
	}
	if b > 0xff {
		b = 0xff
	}
	if a > 0xff {
		a = 0xff
	}
	return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
}
func (op CompositeOperation) ComposeRGBA3(a, b, c color.RGBA) color.RGBA {
	ma := uint32(0xff - a.A)
	mb := uint32(0xff - b.A)
	return makeRGBA(
		uint32(b.R)+(uint32(a.R)+uint32(c.R)*ma/0xff)*mb/0xff,
		uint32(b.G)+(uint32(a.G)+uint32(c.G)*ma/0xff)*mb/0xff,
		uint32(b.B)+(uint32(a.B)+uint32(c.B)*ma/0xff)*mb/0xff,
		uint32(b.A)+(uint32(a.A)+uint32(c.A)*ma/0xff)*mb/0xff,
	)
}

func (op CompositeOperation) ComposeRGBA(a, b color.RGBA) color.RGBA {
	switch op {
	case SourceOver:
		// fa, fb = 1, 1-a.A
		mb := uint32(0xff - a.A)
		// return makeRGBA(
		// 	(uint32(a.R) + uint32(b.R)*mb/0xff),
		// 	(uint32(a.G) + uint32(b.G)*mb/0xff),
		// 	(uint32(a.B) + uint32(b.B)*mb/0xff),
		// 	(uint32(a.A) + uint32(b.A)*mb/0xff),
		// )
		return color.RGBA{
			uint8(uint32(a.R) + uint32(b.R)*mb/0xff),
			uint8(uint32(a.G) + uint32(b.G)*mb/0xff),
			uint8(uint32(a.B) + uint32(b.B)*mb/0xff),
			uint8(uint32(a.A) + uint32(b.A)*mb/0xff),
		}
	case SourceIn:
		// fa, fb = b.A, 0
		return color.RGBA{
			uint8(uint32(a.R) * uint32(b.A) / 0xff),
			uint8(uint32(a.G) * uint32(b.A) / 0xff),
			uint8(uint32(a.B) * uint32(b.A) / 0xff),
			uint8(uint32(a.A) * uint32(b.A) / 0xff),
		}
	case SourceOut:
		// fa, fb = 1-b.A, 0
		ma := uint32(0xff - b.A)
		return color.RGBA{
			uint8(uint32(a.R) * ma / 0xff),
			uint8(uint32(a.G) * ma / 0xff),
			uint8(uint32(a.B) * ma / 0xff),
			uint8(uint32(a.A) * ma / 0xff),
		}
	case SourceAtop:
		// fa, fb = b.A, 1-a.A
		mb := uint32(0xff - a.A)
		return color.RGBA{
			uint8(uint32(a.R)*uint32(b.A)/0xff + uint32(b.R)*mb/0xff),
			uint8(uint32(a.G)*uint32(b.A)/0xff + uint32(b.G)*mb/0xff),
			uint8(uint32(a.B)*uint32(b.A)/0xff + uint32(b.B)*mb/0xff),
			uint8(uint32(a.A)*uint32(b.A)/0xff + uint32(b.A)*mb/0xff),
		}
	case DestinationOver:
		// fa, fb = 1-b.A, 1
		ma := uint32(0xff - b.A)
		return color.RGBA{
			uint8(uint32(a.R)*ma/0xff + uint32(b.R)),
			uint8(uint32(a.G)*ma/0xff + uint32(b.G)),
			uint8(uint32(a.B)*ma/0xff + uint32(b.B)),
			uint8(uint32(a.A)*ma/0xff + uint32(b.A)),
		}
	case DestinationIn:
		// fa, fb = 0, a.A
		return color.RGBA{
			uint8(uint32(b.R) * uint32(a.A) / 0xff),
			uint8(uint32(b.G) * uint32(a.A) / 0xff),
			uint8(uint32(b.B) * uint32(a.A) / 0xff),
			uint8(uint32(b.A) * uint32(a.A) / 0xff),
		}
	case DestinationOut:
		// fa, fb = 0, 1-a.A
		mb := uint32(0xff - a.A)
		return color.RGBA{
			uint8(uint32(b.R) * mb / 0xff),
			uint8(uint32(b.G) * mb / 0xff),
			uint8(uint32(b.B) * mb / 0xff),
			uint8(uint32(b.A) * mb / 0xff),
		}
	case DestinationAtop:
		// fa, fb = 1-b.A, a.A
		ma := uint32(0xff - b.A)
		return color.RGBA{
			uint8(uint32(a.R)*ma/0xff + uint32(b.R)*uint32(a.A)/0xff),
			uint8(uint32(a.G)*ma/0xff + uint32(b.G)*uint32(a.A)/0xff),
			uint8(uint32(a.B)*ma/0xff + uint32(b.B)*uint32(a.A)/0xff),
			uint8(uint32(a.A)*ma/0xff + uint32(b.A)*uint32(a.A)/0xff),
		}
	case Lighter:
		// fa, fb = 1, 1
		return makeRGBA(uint32(a.R)+uint32(b.R),
			uint32(a.G)+uint32(b.G),
			uint32(a.B)+uint32(b.B),
			uint32(a.A)+uint32(b.A))
	case Copy:
		// fa, fb = 1, 0
		return a
	case Xor:
		// fa, fb = 1-b.A, 1-a.A
		ma := uint32(0xff - b.A)
		mb := uint32(0xff - a.A)
		return makeRGBA(
			(uint32(a.R)*ma/0xff + uint32(b.R)*mb/0xff),
			(uint32(a.G)*ma/0xff + uint32(b.G)*mb/0xff),
			(uint32(a.B)*ma/0xff + uint32(b.B)*mb/0xff),
			(uint32(a.A)*ma/0xff + uint32(b.A)*mb/0xff))

	}
	return color.RGBA{}
}

func (op CompositeOperation) Compose(a, b FRGBA) FRGBA {
	var fa, fb float64

	switch op {
	case SourceOver:
		fa, fb = 1, 1-a.A
	case SourceIn:
		fa, fb = b.A, 0
	case SourceOut:
		fa, fb = 1-b.A, 0
	case SourceAtop:
		fa, fb = b.A, 1-a.A
	case DestinationOver:
		fa, fb = 1-b.A, 1
	case DestinationIn:
		fa, fb = 0, a.A
	case DestinationOut:
		fa, fb = 0, 1-a.A
	case DestinationAtop:
		fa, fb = 1-b.A, a.A
	case Lighter:
		fa, fb = 1, 1
	case Copy:
		fa, fb = 1, 0
	case Xor:
		fa, fb = 1-b.A, 1-a.A
	default:
		panic(fmt.Errorf("CompositeOperation invalid value %v", op))
	}

	return a.Mul(fa).Add(b.Mul(fb))
}

func parserCompositeOperation(x string) CompositeOperation {
	switch x {
	case "source-over":
		return SourceOver
	case "source-atop":
		return SourceAtop
	case "source-in":
		return SourceIn
	case "source-out":
		return SourceOut
	case "destination-over":
		return DestinationOver
	case "destination-atop":
		return DestinationAtop
	case "destination-in":
		return DestinationIn
	case "destination-out":
		return DestinationOut
	case "lighter":
		return Lighter
	case "copy":
		return Copy
	case "xor":
		return Xor
	}
	return SourceOver
}

type Context2D interface {
	//
	Width() float64
	Height() float64
	// state
	Save()    // push state on state stack
	Restore() // pop state stack and restore state

	//
	Clear(clr color.Color)
	// transformations (default: transform is the identity matrix)
	Scale(x float64, y float64)
	Rotate(angle float64)
	Translate(x float64, y float64)
	Transform(a, b, c, d, e, f float64)
	SetTransform(a, b, c, d, e, f float64)

	ScaleAbout(sx, sy, x, y float64)
	RotateAbout(angle, x, y float64)

	ResetTransform()
	SetMatrixTransform(tr Matrix)

	// Transform(a float64, b float64, c float64, d float64, e float64, f float64)
	// SetTransformMatrix(a float64, b float64, c float64, d float64, e float64, f float64)

	// rects
	ClearRect(x float64, y float64, w float64, h float64)
	FillRect(x float64, y float64, w float64, h float64)
	StrokeRect(x float64, y float64, w float64, h float64)

	// path API (see also CanvasPathMethods)
	BeginPath()
	Fill()
	Stroke()
	Clip()

	StrokePath(path *Path)
	FillPath(path *Path)

	SetLineCap(lineCap LineCap)
	LineCap() LineCap
	SetLineJoin(lineJoin LineJoin)
	LineJoin() LineJoin

	SetLineWidth(width float64)
	LineWidth() float64

	SetLineDash(dash []float64)
	LineDash() []float64
	SetLineDashOffset(offset float64)
	LineDashOffset() float64

	//isPointInPath( x,  y);

	SetStrokeStyle(p Pattern)
	StrokeStyle() Pattern

	// SetStrokeColor => SetStrokeStyle(NewSolidPattern(color))
	SetStrokeColor(clr color.Color)
	StrokeColor() color.Color

	SetFillStyle(p Pattern)
	FillStyle() Pattern

	// SetFillColor => SetFillStyle(NewSolidPattern(color))
	SetFillColor(clr color.Color)
	FillColor() color.Color

	SetMiterLimit(limit float64)
	MiterLimit() float64

	SetGlobalAlpha(alpha float64)
	GlobalAlpha() float64

	SetGlobalCompositeOperation(op CompositeOperation)
	GlobalCompositeOperation() CompositeOperation

	CreateLinearGradient(x0, y0, x1, y1 float64) Gradient
	CreateRadialGradient(x0, y0, r0, x1, y1, r1 float64) Gradient
	CreatePattern(img image.Image, op RepeatOp) Pattern
	//
	SetFont(f *Font)
	SetTextAlign(align TextAlign)
	TextAlign() TextAlign
	SetTextBaseline(base TextBaseline)
	TextBaseline() TextBaseline

	// text (see also the CanvasDrawingStyles interface)
	FillText(text string, x float64, y float64)
	StrokeText(text string, x float64, y float64)
	MeasureText(text string) float64

	// drawing images
	DrawImage(img image.Image, dx float64, dy float64)
	DrawImageEx(img image.Image, sx, sy, sw, sh, dx, dy, dw, dh float64)
	// drawing Context2D
	DrawContext2D(cv Context2D, dx float64, dy float64)
	DrawContext2DEx(cv Context2D, sx, sy, sw, sh, dx, dy, dw, dh float64)

	SetShadowOffset(x float64, y float64)
	SetShadowBlur(blur float64)
	SetShadowColor(clr color.Color)
	ShadowOffset() (float64, float64)
	ShadowBlur() float64
	ShadowColor() color.Color

	ClosePath()
	MoveTo(x float64, y float64)
	LineTo(x float64, y float64)
	QuadraticCurveTo(cpx float64, cpy float64, x float64, y float64)
	BezierCurveTo(cp1x float64, cp1y float64, cp2x float64, cp2y float64, x float64, y float64)
	Arc(x float64, y float64, radius float64, startAngle float64, endAngle float64, counterclockwise bool)
	ArcTo(x1 float64, y1 float64, x2 float64, y2 float64, radius float64)
	//
	ArcAngle(x float64, y float64, rx float64, ry float64, startAngle float64, sweepAngle float64)
	Rect(x float64, y float64, width float64, height float64)
	RoundedRect(x float64, y float64, width float64, height float64, arcWidth float64, arcHeight float64)
	Circle(cx float64, cy float64, radius float64)
	Ellipse(cx float64, cy float64, rx float64, ry float64)
	//
	Image() image.Image
	GetImageData(x, y, width, height int) image.Image
}

// type CanvasPathMethods interface {
// 	// shared path API methods
// 	ClosePath()
// 	MoveTo(x float64, y float64)
// 	LineTo(x float64, y float64)
// 	QuadraticCurveTo(cpx float64, cpy float64, x float64, y float64)
// 	BezierCurveTo(cp1x float64, cp1y float64, cp2x float64, cp2y float64, x float64, y float64)
// 	ArcTo(x1 float64, y1 float64, x2 float64, y2 float64, radius float64)
// 	Rect(x float64, y float64, w float64, h float64)
// 	Arc(x float64, y float64, radius float64, startAngle float64, endAngle float64, counterclockwise bool)
// }

// type CanvasGradient interface {
// 	// opaque object
// 	AddColorStop(offset float64, color color.Color)
// }

func imageToRGBA(src image.Image) *image.RGBA {
	if rgba, ok := src.(*image.RGBA); ok {
		return rgba
	}
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, src, bounds.Min, draw.Src)
	return dst
}

func imageToNRGBA(src image.Image) *image.NRGBA {
	if rgba, ok := src.(*image.NRGBA); ok {
		return rgba
	}
	bounds := src.Bounds()
	dst := image.NewNRGBA(bounds)
	draw.Draw(dst, bounds, src, bounds.Min, draw.Src)
	return dst
}

func CopyToRGBA(src image.Image) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, src, bounds.Min, draw.Src)
	return dst
}

func CopyToNRGBA(src image.Image) *image.NRGBA {
	bounds := src.Bounds()
	dst := image.NewNRGBA(bounds)
	draw.Draw(dst, bounds, src, bounds.Min, draw.Src)
	return dst
}
