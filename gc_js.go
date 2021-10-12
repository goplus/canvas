//go:build js
// +build js

package canvas

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strconv"
	"strings"
	"syscall/js"

	"github.com/goplus/canvas/jsutil"
)

var (
	window   = js.Global().Get("window")
	document = js.Global().Get("document")
	wx       js.Value
)

var jsCopyArray js.Value = js.Global().Call("eval", `(
	function(a, b) {
		for (let i = 0; i < b.length; i++) {
			a[i] = b[i]
		}
	}
)`)

var jsCanvasToImage js.Value = js.Global().Call("eval", `(
	function(canvas) {
		var image = new Image();
		image.src = canvas.toDataURL("image/png");
		return image;
	}
)`)

type WebContext2D struct {
	canvas js.Value
	ctx2d  js.Value
	width  int
	height int
}

func NewWebContext2DForContext(canvas js.Value, ctx2d js.Value) Context2D {
	width := canvas.Get("width").Int()
	height := canvas.Get("height").Int()
	return &WebContext2D{canvas, ctx2d, width, height}
}

func NewWebContext2DForCanvas(canvas js.Value) Context2D {
	width := canvas.Get("width").Int()
	height := canvas.Get("height").Int()
	ctx2d := canvas.Call("getContext", "2d")
	return &WebContext2D{canvas, ctx2d, width, height}
}

func NewWebContext2DForImage(img image.Image) Context2D {
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	gc := NewWebContext2D(width, height).(*WebContext2D)
	gc.SetImage(img, 0, 0)
	return gc
}

func (gc *WebContext2D) Width() float64 {
	return float64(gc.width)
}

func (gc *WebContext2D) Height() float64 {
	return float64(gc.height)
}

func (r *WebContext2D) Draw() {
	r.ctx2d.Call("draw")
}

func (r *WebContext2D) Save() {
	r.ctx2d.Call("save")
}

func (r *WebContext2D) Restore() {
	r.ctx2d.Call("restore")
}

func (r *WebContext2D) Scale(x float64, y float64) {
	r.ctx2d.Call("scale", x, y)
}

func (r *WebContext2D) Rotate(angle float64) {
	r.ctx2d.Call("rotate", angle)
}

func (r *WebContext2D) ScaleAbout(sx, sy, x, y float64) {
	r.ctx2d.Call("translate", x, y)
	r.ctx2d.Call("scale", sx, sy)
	r.ctx2d.Call("translate", -x, -y)
}

func (r *WebContext2D) RotateAbout(angle, x, y float64) {
	r.ctx2d.Call("translate", x, y)
	r.ctx2d.Call("rotate", angle)
	r.ctx2d.Call("translate", -x, -y)
}

func (r *WebContext2D) Translate(x float64, y float64) {
	r.ctx2d.Call("translate", x, y)
}

func (r *WebContext2D) Transform(a, b, c, d, e, f float64) {
	r.ctx2d.Call("transform", a, b, c, d, e, f)
}

func (r *WebContext2D) SetTransform(a, b, c, d, e, f float64) {
	r.ctx2d.Call("setTransform", a, b, c, d, e, f)
}

func (r *WebContext2D) ResetTransform() {
	r.ctx2d.Call("setTransform", 1, 0, 0, 1, 0, 0)
}

func (r *WebContext2D) SetMatrixTransform(tr Matrix) {
	r.ctx2d.Call("setTransform", tr[0], tr[1], tr[2], tr[3], tr[4], tr[5])
}

func (r *WebContext2D) TransformMatrix() Matrix {
	ar := r.ctx2d.Call("transform")
	return Matrix([6]float64{
		ar.Index(0).Float(),
		ar.Index(1).Float(),
		ar.Index(2).Float(),
		ar.Index(3).Float(),
		ar.Index(4).Float(),
		ar.Index(5).Float()})
}

func (r *WebContext2D) ClearRect(x, y, width, height float64) {
	r.ctx2d.Call("clearRect", x, y, width, height)
}

func (r *WebContext2D) Clear(clr color.Color) {
	if clr == nil {
		clr = color.Transparent
	}
	r.Save()
	r.ResetTransform()
	if clr == color.Transparent {
		r.ClearRect(0, 0, float64(r.width), float64(r.height))
	} else {
		r.SetFillColor(clr)
		r.FillRect(0, 0, float64(r.width), float64(r.height))
	}
	r.Restore()
}

func (r *WebContext2D) FillRect(x, y, width, height float64) {
	r.ctx2d.Call("fillRect", x, y, width, height)
}

func (r *WebContext2D) StrokeRect(x, y, width, height float64) {
	r.ctx2d.Call("strokeRect", x, y, width, height)
}

func (r *WebContext2D) BeginPath() {
	r.ctx2d.Call("beginPath")
}

func (r *WebContext2D) Stroke() {
	r.ctx2d.Call("stroke")
}

func (r *WebContext2D) Fill() {
	r.ctx2d.Call("fill")
}

func (r *WebContext2D) Clip() {
	r.ctx2d.Call("clip")
}

func (r *WebContext2D) ClosePath() {
	r.ctx2d.Call("closePath")
}

func (r *WebContext2D) SetLineWidth(width float64) {
	r.ctx2d.Set("lineWidth", width)
}

func (r *WebContext2D) LineWidth() float64 {
	return r.ctx2d.Get("lineWidth").Float()
}

func (r *WebContext2D) MoveTo(x, y float64) {
	r.ctx2d.Call("moveTo", x, y)
}

func (r *WebContext2D) LineTo(x, y float64) {
	r.ctx2d.Call("lineTo", x, y)
}

func (r *WebContext2D) QuadraticCurveTo(cpx float64, cpy float64, x float64, y float64) {
	r.ctx2d.Call("quadraticCurveTo", cpx, cpy, x, y)
}

func (r *WebContext2D) BezierCurveTo(cp1x float64, cp1y float64, cp2x float64, cp2y float64, x float64, y float64) {
	r.ctx2d.Call("bezierCurveTo", cp1x, cp1y, cp2x, cp2y, x, y)
}

func (r *WebContext2D) ArcTo(x1 float64, y1 float64, x2 float64, y2 float64, radius float64) {
	r.ctx2d.Call("arcTo", x1, y1, x2, y2, radius)
}

func (r *WebContext2D) Rect(x float64, y float64, w float64, h float64) {
	r.ctx2d.Call("rect", x, y, w, h)
}

func (r *WebContext2D) Arc(x float64, y float64, radius float64, startAngle float64, endAngle float64, counterclockwise bool) {
	r.ctx2d.Call("arc", x, y, radius, startAngle, endAngle, counterclockwise)
}

func (r *WebContext2D) StrokeText(s string, x, y float64) {
	r.ctx2d.Call("strokeText", s, x, y)
}

func (r *WebContext2D) FillText(s string, x, y float64) {
	r.ctx2d.Call("fillText", s, x, y)
}

func (r *WebContext2D) MeasureText(text string) float64 {
	m := r.ctx2d.Call("measureText", text)
	return m.Get("width").Float()
}

func (r *WebContext2D) SetFont(f *Font) {
	r.ctx2d.Set("font", f.String())
}

func (c *WebContext2D) DrawContext2D(cv Context2D, dx float64, dy float64) {
	switch p := cv.(type) {
	case *WebContext2D:
		c.ctx2d.Call("drawImage", p.canvas, dx, dy)
	default:
		ctx := NewWebContext2DForImage(cv.Image()).(*WebContext2D)
		c.ctx2d.Call("drawImage", ctx.canvas, dx, dy)
	}
}

func (c *WebContext2D) DrawContext2DEx(cv Context2D, sx, sy, sw, sh, dx, dy, dh, dw float64) {
	switch p := cv.(type) {
	case *WebContext2D:
		c.ctx2d.Call("drawImage", p.canvas, dx, dy)
	default:
		ctx := NewWebContext2DForImage(cv.Image()).(*WebContext2D)
		c.ctx2d.Call("drawImage", ctx.canvas, dx, dy)
	}
}

func (c *WebContext2D) Image() image.Image {
	imdata := c.ctx2d.Call("getImageData", 0, 0, c.width, c.height)
	data := imdata.Get("data")
	pix := jsutil.ArrayBufferToSlice(data)
	return &image.NRGBA{pix, c.width * 4, image.Rect(0, 0, c.width, c.height)}
}

func (c *WebContext2D) GetImageData(x, y, width, height int) image.Image {
	imdata := c.ctx2d.Call("getImageData", x, y, width, height)
	data := imdata.Get("data")
	pix := jsutil.ArrayBufferToSlice(data)
	return &image.NRGBA{pix, width * 4, image.Rect(x, y, width, height)}
}

func (c *WebContext2D) SetStrokeColor(clr color.Color) {
	c.ctx2d.Set("strokeStyle", color2html(clr))
}

func (c *WebContext2D) StrokeColor() color.Color {
	v := c.ctx2d.Get("strokeStyle")
	if v.Type() == js.TypeString {
		return html2color(v.String())
	}
	return nil
}

func (c *WebContext2D) SetFillColor(clr color.Color) {
	c.ctx2d.Set("fillStyle", color2html(clr))
}

func (c *WebContext2D) FillColor() color.Color {
	v := c.ctx2d.Get("fillStyle")
	if v.Type() == js.TypeString {
		return html2color(v.String())
	}
	return nil
}

type jsPattern struct {
	v js.Value
}

func (c *jsPattern) ColorAt(x, y int) color.Color {
	return color.Transparent
}

type jsGradient struct {
	v js.Value
}

func (c *jsGradient) AddColorStop(offset float64, color color.Color) {
	c.v.Call("addColorStop", offset, color2html(color))
}

func (c *jsGradient) ColorAt(x, y int) color.Color {
	return color.Transparent
}

func (c *WebContext2D) CreateLinearGradient(x0, y0, x1, y1 float64) Gradient {
	p := c.ctx2d.Call("createLinearGradient", x0, y0, x1, y1)
	return &jsGradient{p}
}

func (c *WebContext2D) CreateRadialGradient(x0, y0, r0, x1, y1, r1 float64) Gradient {
	p := c.ctx2d.Call("createRadialGradient", x0, y0, r0, x1, y1, r1)
	return &jsGradient{p}
}

func (c *WebContext2D) CreatePattern(img image.Image, op RepeatOp) Pattern {
	dc := NewWebContext2DForImage(img).(*WebContext2D)
	p := c.ctx2d.Call("createPattern", dc.canvas, op.String())
	return &jsPattern{p}
}

func (c *WebContext2D) SetStrokeStyle(p Pattern) {
	switch t := p.(type) {
	case *SolidPattern:
		c.SetStrokeColor(t.color)
	case *jsGradient:
		c.ctx2d.Set("strokeStyle", t.v)
	case *jsPattern:
		c.ctx2d.Set("strokeStyle", t.v)
	}
}

func (c *WebContext2D) StrokeStyle() Pattern {
	v := c.ctx2d.Get("strokeStyle")
	if v.Type() == js.TypeString {
		if clr := html2color(v.String()); clr != nil {
			return &SolidPattern{clr}
		}
	}
	if strings.Contains(v.String(), "Gradient") {
		return &jsGradient{v}
	}
	return &jsPattern{v}
}

func (c *WebContext2D) SetFillStyle(p Pattern) {
	switch t := p.(type) {
	case *SolidPattern:
		c.SetFillColor(t.color)
	case *jsGradient:
		c.ctx2d.Set("fillStyle", t.v)
	case *jsPattern:
		c.ctx2d.Set("fillStyle", t.v)
	}
}

func (c *WebContext2D) FillStyle() Pattern {
	v := c.ctx2d.Get("fillStyle")
	if v.Type() == js.TypeString {
		if clr := html2color(v.String()); clr != nil {
			return &SolidPattern{clr}
		}
	}
	if strings.Contains(v.String(), "Gradient") {
		return &jsGradient{v}
	}
	return &jsPattern{v}
}

func (c *WebContext2D) SetLineCap(lineCap LineCap) {
	c.ctx2d.Set("lineCap", lineCap.String())
}

func (c *WebContext2D) SetLineJoin(lineJoin LineJoin) {
	c.ctx2d.Set("lineJoin", lineJoin.String())
}

func (c *WebContext2D) SetMiterLimit(limit float64) {
	c.ctx2d.Set("miterLimit", limit)
}

func (c *WebContext2D) MiterLimit() float64 {
	return c.ctx2d.Get("miterLimit").Float()
}

func (c *WebContext2D) LineCap() LineCap {
	x := c.ctx2d.Get("lineCap").String()
	return ParserLineCap(x)
}

func (c *WebContext2D) LineJoin() LineJoin {
	x := c.ctx2d.Get("lineJoin").String()
	return ParserLineJoin(x)
}

func (c *WebContext2D) SetLineDash(dash []float64) {
	ar, _ := jsutil.SliceToTypedArray(dash)
	c.ctx2d.Call("setLineDash", ar)
}

func (c *WebContext2D) LineDash() []float64 {
	ar := c.ctx2d.Call("getLineDash")
	size := ar.Length()
	if size == 0 {
		return nil
	}
	v := make([]float64, size, size)
	for i := 0; i < size; i++ {
		v[i] = ar.Index(i).Float()
	}
	return v
}

func (c *WebContext2D) SetLineDashOffset(offset float64) {
	c.ctx2d.Set("lineDashOffset", offset)
}

func (c *WebContext2D) LineDashOffset() float64 {
	return c.ctx2d.Get("lineDashOffset").Float()
}

func (c *WebContext2D) SetTextAlign(align TextAlign) {
	c.ctx2d.Set("textAlign", align.String())
}

func (c *WebContext2D) TextAlign() TextAlign {
	x := c.ctx2d.Get("textAlign").String()
	return ParserTextAlign(x)
}

func (c *WebContext2D) SetTextBaseline(base TextBaseline) {
	c.ctx2d.Set("textBaseline", base.String())
}

func (c *WebContext2D) TextBaseline() TextBaseline {
	x := c.ctx2d.Get("textBaseline").String()
	return ParserTextBaseline(x)
}

func (c *WebContext2D) SetGlobalAlpha(alpha float64) {
	c.ctx2d.Set("globalAlpha", alpha)
}

func (c *WebContext2D) GlobalAlpha() float64 {
	return c.ctx2d.Get("globalAlpha").Float()
}

func (c *WebContext2D) SetGlobalCompositeOperation(op CompositeOperation) {
	c.ctx2d.Set("globalCompositeOperation", op.String())
}

func (c *WebContext2D) GlobalCompositeOperation() CompositeOperation {
	attr := c.ctx2d.Get("globalCompositeOperation").String()
	return parserCompositeOperation(attr)
}

func (c *WebContext2D) SetShadowOffset(x float64, y float64) {
	c.ctx2d.Set("shadowOffsetX", x)
	c.ctx2d.Set("shadowOffsetY", y)
}

func (c *WebContext2D) ShadowOffset() (float64, float64) {
	return c.ctx2d.Get("shadowOffsetX").Float(), c.ctx2d.Get("shadowOffsetY").Float()
}

func (c *WebContext2D) SetShadowBlur(blur float64) {
	c.ctx2d.Set("shadowBlur", blur)
}

func (c *WebContext2D) ShadowBlur() float64 {
	return c.ctx2d.Get("shadowBlur").Float()
}

func (c *WebContext2D) SetShadowColor(clr color.Color) {
	c.ctx2d.Set("shadowColor", color2html(clr))
}

func (c *WebContext2D) ShadowColor() color.Color {
	x := c.ctx2d.Get("shadowColor").String()
	return html2color(x)
}

func (c *WebContext2D) JSContext2D() js.Value {
	return c.ctx2d
}

func (c *WebContext2D) ArcAngle(x, y, rx, ry, startAngle, sweepAngle float64) {
	const n = 16
	for i := 0; i < n; i++ {
		p1 := float64(i+0) / n
		p2 := float64(i+1) / n
		a1 := startAngle + sweepAngle*p1
		a2 := startAngle + sweepAngle*p2
		x0 := x + rx*math.Cos(a1)
		y0 := y + ry*math.Sin(a1)
		x1 := x + rx*math.Cos(a1+(a2-a1)/2)
		y1 := y + ry*math.Sin(a1+(a2-a1)/2)
		x2 := x + rx*math.Cos(a2)
		y2 := y + ry*math.Sin(a2)
		cx := 2*x1 - x0/2 - x2/2
		cy := 2*y1 - y0/2 - y2/2
		if i == 0 {
			c.LineTo(x0, y0)
		}
		c.QuadraticCurveTo(cx, cy, x2, y2)
	}
}

func (c *WebContext2D) addPath(p *Path) {
	var j int
	for _, cmd := range p.Components {
		switch cmd {
		case MoveToCmp:
			c.MoveTo(p.Points[j], p.Points[j+1])
			j = j + 2
		case LineToCmp:
			c.LineTo(p.Points[j], p.Points[j+1])
			j = j + 2
		case QuadCurveToCmp:
			c.QuadraticCurveTo(p.Points[j], p.Points[j+1], p.Points[j+2], p.Points[j+3])
			j = j + 4
		case CubicCurveToCmp:
			c.BezierCurveTo(p.Points[j], p.Points[j+1], p.Points[j+2], p.Points[j+3], p.Points[j+4], p.Points[j+5])
			j = j + 6
		case ArcAngleCmp:
			c.ArcAngle(p.Points[j], p.Points[j+1], p.Points[j+2], p.Points[j+3], p.Points[j+4], p.Points[j+5])
			j = j + 6
		case CloseCmp:
			c.ClosePath()
		}
	}
}

func (c *WebContext2D) RoundedRect(x float64, y float64, width float64, height float64, arcWidth float64, arcHeight float64) {
	x2, y2 := x+width, y+height
	arcWidth = arcWidth / 2
	arcHeight = arcHeight / 2
	c.MoveTo(x, y+arcHeight)
	c.QuadraticCurveTo(x, y, x+arcWidth, y)
	c.LineTo(x2-arcWidth, y)
	c.QuadraticCurveTo(x2, y, x2, y+arcHeight)
	c.LineTo(x2, y2-arcHeight)
	c.QuadraticCurveTo(x2, y2, x2-arcWidth, y2)
	c.LineTo(x+arcWidth, y2)
	c.QuadraticCurveTo(x, y2, x, y2-arcHeight)
	c.ClosePath()
}

func (c *WebContext2D) Ellipse(cx, cy, rx, ry float64) {
	c.ArcAngle(cx, cy, rx, ry, 0, -math.Pi*2)
	c.ClosePath()
}

func (c *WebContext2D) Circle(cx, cy, radius float64) {
	c.ArcAngle(cx, cy, radius, radius, 0, -math.Pi*2)
	c.ClosePath()
}

func (c *WebContext2D) StrokePath(p *Path) {
	c.Save()
	c.BeginPath()
	c.addPath(p)
	c.Stroke()
	c.Restore()
}

func (c *WebContext2D) FillPath(p *Path) {
	c.Save()
	c.BeginPath()
	c.addPath(p)
	c.Fill()
	c.Restore()
}

func color2html(clr color.Color) string {
	rgba := color.NRGBAModel.Convert(clr).(color.NRGBA)
	if rgba.A == 0xff {
		return fmt.Sprintf("#%02x%02x%02x", rgba.R, rgba.G, rgba.B)
	}
	return fmt.Sprintf("rgba(%v,%v,%v,%v)",
		rgba.R, rgba.G, rgba.B, float64(rgba.A)/0xff)
}

func html2color(x string) color.Color {
	if len(x) == 7 && x[0] == '#' {
		r, e1 := strconv.ParseUint(x[1:3], 16, 0)
		g, e2 := strconv.ParseUint(x[3:5], 16, 0)
		b, e3 := strconv.ParseUint(x[5:7], 16, 0)
		if e1 == nil && e2 == nil && e3 == nil {
			return color.NRGBA{uint8(r), uint8(g), uint8(b), 0xff}
		}
	} else if strings.HasPrefix(x, "rgba") {
		var r, g, b int
		var a float64
		n, err := fmt.Sscanf(x, "rgba(%d,%d,%d,%f)", &r, &g, &b, &a)
		if err == nil && n == 4 {
			return color.NRGBA{uint8(r), uint8(g), uint8(b), uint8(a * 0xff)}
		}
	}
	return nil
}
