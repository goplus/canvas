// Copyright 2010 The draw2d Authors. All rights reserved.
// created: 21/11/2010 by Laurent Le Goff

package canvas

import (
	"image"
	"image/color"
	"math"

	"github.com/golang/freetype/raster"

	"golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

// GraphicContext is the implementation of draw2d.GraphicContext for a raster image
type GraphicContext2D struct {
	*StackGraphicContext
	width, height    int
	img              *image.RGBA
	fillRasterizer   *raster.Rasterizer
	strokeRasterizer *raster.Rasterizer
	painter          *RGBAPainter
	DPI              int
}

func (gc *GraphicContext2D) Image() image.Image {
	return gc.img
}

func (gc *GraphicContext2D) SubImage(r image.Rectangle) image.Image {
	return gc.img.SubImage(r)
}

func (gc *GraphicContext2D) GetImageData(x, y, width, height int) image.Image {
	return gc.img.SubImage(image.Rect(x, y, x+width, y+height))
}

// ImageFilter defines the type of filter to use
type ImageFilter int

const (
	// LinearFilter defines a linear filter
	LinearFilter ImageFilter = iota
	// BilinearFilter defines a bilinear filter
	BilinearFilter
	// BicubicFilter defines a bicubic filter
	BicubicFilter
)

func NewGraphicContext2D(width, height int) *GraphicContext2D {
	return NewGraphicContext2DForImage(image.NewRGBA(image.Rect(0, 0, width, height)))
}

// NewContextForImage copies the specified image into a new image.RGBA
// and prepares a context for rendering onto that image.
func NewGraphicContext2DForImage(im image.Image) *GraphicContext2D {
	img := imageToRGBA(im)
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	gc := &GraphicContext2D{
		NewStackGraphicContext(),
		width,
		height,
		img,
		raster.NewRasterizer(width, height),
		raster.NewRasterizer(width, height),
		NewRGBAPainter(img),
		72,
	}
	return gc
}

func (gc *GraphicContext2D) Width() float64 {
	return float64(gc.width)
}

func (gc *GraphicContext2D) Height() float64 {
	return float64(gc.height)
}

// GetDPI returns the resolution of the Image GraphicContext
func (gc *GraphicContext2D) GetDPI() int {
	return gc.DPI
}

// Clear fills the current canvas with a default transparent color
func (gc *GraphicContext2D) Clear(clr color.Color) {
	if clr == nil {
		clr = color.Transparent
	}
	src := image.NewUniform(clr)
	draw.Draw(gc.img, gc.img.Bounds(), src, image.ZP, draw.Src)
	gc.Current.Path.Clear()
}

func (gc *GraphicContext2D) ClearRect(x, y, w, h float64) {
	gc.Save()
	gc.SetFillColor(color.Transparent)
	p := NewPath()
	x, y = gc.TransformPoint(x, y)
	w, h = gc.TransformScale(w, h)
	p.AddRectangle(x, y, w, h)
	gc.fill(p)
	gc.Restore()
}

func (gc *GraphicContext2D) FillRect(x, y, w, h float64) {
	p := NewPath()
	p.AddRectangle(x, y, w, h)
	p.Transfrom(gc.Current.Tr)
	gc.fill(p)
}

func (gc *GraphicContext2D) StrokeRect(x, y, w, h float64) {
	p := NewPath()
	p.AddRectangle(x, y, w, h)
	p.Transfrom(gc.Current.Tr)
	gc.stroke(p)
}

// TODO AlignTop and AlignBottom
func (gc *GraphicContext2D) CreateTextPath(text string, x float64, y float64) *Path {
	p := NewPath()
	if gc.Current.TextBaseline != AlignAlphabetic {
		m, err := p.MetricsFont(gc.Current.Font)
		if m != nil && err == nil {
			switch gc.Current.TextBaseline {
			case AlignTop:
				y += fUnitsToFloat64(m.Ascent)
			case AlignHanging:
				y += fUnitsToFloat64(m.Ascent)
			case AlignMiddle:
				y += fUnitsToFloat64(m.Ascent)/2 - fUnitsToFloat64(m.Descent)/2
			case AlignIdeographic:
				y -= fUnitsToFloat64(m.Descent)
			case AlignBottom:
				y -= fUnitsToFloat64(m.Descent)
			}
		}
	}
	size := p.AddText(text, x, y, gc.Current.Font)
	if gc.Current.TextAlign == AlignRight {
		p.Translate(-size, 0)
	} else if gc.Current.TextAlign == AlignCenter {
		size = size / 2
		p.Translate(-size, 0)
	}
	p.Transfrom(gc.Current.Tr)
	return p
}

func (gc *GraphicContext2D) FillText(text string, x float64, y float64) {
	p := gc.CreateTextPath(text, x, y)
	gc.fill(p)
}

func (gc *GraphicContext2D) StrokeText(text string, x float64, y float64) {
	p := gc.CreateTextPath(text, x, y)
	gc.stroke(p)
}

func (gc *GraphicContext2D) MeasureText(text string) float64 {
	p := NewPath()
	return p.MeasureText(text, gc.Current.Font)
}

// DrawImage draws an image into dest using an affine transformation matrix, an op and a filter
func DrawImage(src image.Image, mask *image.Alpha, dst draw.Image, tr Matrix, op draw.Op, filter ImageFilter) {
	var transformer draw.Transformer
	switch filter {
	case LinearFilter:
		transformer = draw.NearestNeighbor
	case BilinearFilter:
		transformer = draw.BiLinear
	case BicubicFilter:
		transformer = draw.CatmullRom
	}
	if mask == nil {
		transformer.Transform(dst, f64.Aff3{tr[0], tr[2], tr[4], tr[1], tr[3], tr[5]}, src, src.Bounds(), op, nil)
	} else {
		opts := &draw.Options{
			DstMask:  mask,
			DstMaskP: image.ZP,
		}
		transformer.Transform(dst, f64.Aff3{tr[0], tr[2], tr[4], tr[1], tr[3], tr[5]}, src, src.Bounds(), op, opts)
	}
}

// DrawImage draws the raster image in the current canvas
func (gc *GraphicContext2D) DrawImage(img image.Image, dx float64, dy float64) {
	if gc.Current.GlobalAlpha <= 0 {
		return
	}
	tr := gc.Current.Tr.Copy()
	tr.Translate(dx, dy)
	src := CopyToNRGBA(img)
	if gc.Current.GlobalAlpha < 1 {
		for i := 0; i < len(src.Pix); i += 4 {
			src.Pix[i+3] = uint8(float64(src.Pix[i+3]) * gc.Current.GlobalAlpha)
		}
	}
	DrawImage(src, gc.Current.mask, gc.img, tr, draw.Over, BilinearFilter)
}

func toRect(x, y, w, h float64) image.Rectangle {
	return image.Rect(int(x), int(y), int(x+w), int(y+h))
}

func (gc *GraphicContext2D) DrawImageEx(img image.Image, sx, sy, sw, sh, dx, dy, dw, dh float64) {
	if gc.Current.GlobalAlpha <= 0 {
		return
	}
	tr := gc.Current.Tr.Copy()
	tr.Translate(dx, dy)
	tr.Scale(dw/sw, dh/sh)
	//tr.Translate(dx*dw/sw-dx, dy*dh/sh-dy)
	src := CopyToNRGBA(img).SubImage(toRect(sx, sy, sw, sh)).(*image.NRGBA)
	src.Rect = image.Rect(0, 0, src.Rect.Dx(), src.Rect.Dy())
	if gc.Current.GlobalAlpha < 1 {
		for i := 0; i < len(src.Pix); i += 4 {
			src.Pix[i+3] = uint8(float64(src.Pix[i+3]) * gc.Current.GlobalAlpha)
		}
	}
	DrawImage(src, gc.Current.mask, gc.img, tr, draw.Over, BilinearFilter)
}

func (gc *GraphicContext2D) DrawContext2D(cv Context2D, dx float64, dy float64) {
	gc.DrawImage(cv.Image(), dx, dy)
}

func (gc *GraphicContext2D) DrawContext2DEx(cv Context2D, sx, sy, sw, sh, dx, dy, dw, dh float64) {
	gc.DrawImageEx(cv.Image(), sx, sy, sw, sh, dx, dy, dw, dh)
}

// Stroke strokes the paths with the color specified by SetStrokeColor
func (gc *GraphicContext2D) Stroke() {
	gc.stroke(gc.Current.Path)
}

func (gc *GraphicContext2D) StrokePath(p *Path) {
	gc.Save()
	var j int
	for _, cmd := range p.Components {
		switch cmd {
		case MoveToCmp:
			gc.MoveTo(p.Points[j], p.Points[j+1])
			j = j + 2
		case LineToCmp:
			gc.LineTo(p.Points[j], p.Points[j+1])
			j = j + 2
		case QuadCurveToCmp:
			gc.QuadraticCurveTo(p.Points[j], p.Points[j+1], p.Points[j+2], p.Points[j+3])
			j = j + 4
		case CubicCurveToCmp:
			gc.BezierCurveTo(p.Points[j], p.Points[j+1], p.Points[j+2], p.Points[j+3], p.Points[j+4], p.Points[j+5])
			j = j + 6
		case ArcAngleCmp:
			gc.ArcAngle(p.Points[j], p.Points[j+1], p.Points[j+2], p.Points[j+3], p.Points[j+4], p.Points[j+5])
			j = j + 6
		case CloseCmp:
			gc.ClosePath()
		}
	}
	gc.Stroke()
	gc.Restore()
}

func (gc *GraphicContext2D) FillPath(p *Path) {
	gc.Save()
	gc.BeginPath()
	var j int
	for _, cmd := range p.Components {
		switch cmd {
		case MoveToCmp:
			gc.MoveTo(p.Points[j], p.Points[j+1])
			j = j + 2
		case LineToCmp:
			gc.LineTo(p.Points[j], p.Points[j+1])
			j = j + 2
		case QuadCurveToCmp:
			gc.QuadraticCurveTo(p.Points[j], p.Points[j+1], p.Points[j+2], p.Points[j+3])
			j = j + 4
		case CubicCurveToCmp:
			gc.BezierCurveTo(p.Points[j], p.Points[j+1], p.Points[j+2], p.Points[j+3], p.Points[j+4], p.Points[j+5])
			j = j + 6
		case ArcAngleCmp:
			gc.ArcAngle(p.Points[j], p.Points[j+1], p.Points[j+2], p.Points[j+3], p.Points[j+4], p.Points[j+5])
			j = j + 6
		case CloseCmp:
			gc.ClosePath()
		}
	}
	gc.Fill()
	gc.Restore()
}

func (gc *GraphicContext2D) stroke(paths ...*Path) {
	gc.strokeRasterizer.UseNonZeroWinding = true

	stroker := &LineStroker{}
	stroker.Cap = gc.Current.Cap
	stroker.Join = gc.Current.Join
	stroker.Flattener = &FtLineBuilder{Adder: gc.strokeRasterizer}
	stroker.HalfLineWidth = gc.Current.LineWidth / 2
	stroker.MiterLimitCheck = gc.Current.MiterLimit * gc.Current.LineWidth

	var liner Flattener
	if gc.Current.Dash != nil && len(gc.Current.Dash) > 0 {
		liner = NewDashConverter(gc.Current.Dash, gc.Current.DashOffset, stroker)
	} else {
		liner = stroker
	}
	hasShadow := gc.HasShadow()
	var offsetx float64
	var offsety float64
	gc.strokeRasterizer.Dx = 0
	gc.strokeRasterizer.Dy = 0
	if hasShadow {
		w, h := int(math.Abs(gc.Current.ShadowOffsetX)), int(math.Abs(gc.Current.ShadowOffsetY))
		if w != 0 || h != 0 {
			gc.strokeRasterizer.SetBounds(gc.width+w, gc.height+h)
		}
		if gc.Current.ShadowOffsetX > 0 {
			offsetx = gc.Current.ShadowOffsetX
			gc.strokeRasterizer.Dx = int(-offsetx)
		}
		if gc.Current.ShadowOffsetY > 0 {
			offsety = gc.Current.ShadowOffsetY
			gc.strokeRasterizer.Dy = int(-offsety)
		}
	}
	for _, p := range paths {
		if offsetx != 0 || offsety != 0 {
			p = p.Translated(offsetx, offsety)
		}
		Flatten(p, liner, 1)
	}
	gc.painter.SetMask(gc.Current.mask)
	gc.painter.SetGlobalAlpha(gc.Current.GlobalAlpha)
	gc.painter.SetPattern(gc.Current.StrokePattern, gc.Current.Tr)
	gc.painter.SetCompositeOperation(gc.Current.GlobalCompositeOperation)
	gc.painter.SetShadow(gc.Current.ShadowOffsetX, gc.Current.ShadowOffsetY, gc.Current.ShadowBlur, gc.Current.ShadowColor)
	gc.painter.Begin()
	gc.strokeRasterizer.Rasterize(gc.painter)
	gc.strokeRasterizer.Clear()
	gc.painter.End()
}

// Fill fills the paths with the color specified by SetFillColor
func (gc *GraphicContext2D) Fill() {
	gc.fill(gc.Current.Path)
}

func (gc *GraphicContext2D) HasShadow() bool {
	_, _, _, a := gc.Current.ShadowColor.RGBA()
	return a != 0 && (gc.Current.ShadowOffsetX != 0 || gc.Current.ShadowOffsetY != 0 || gc.Current.ShadowBlur != 0)
}

func (gc *GraphicContext2D) fill(paths ...*Path) {
	gc.fillRasterizer.UseNonZeroWinding = gc.Current.FillRule == FillRuleWinding

	/**** first method ****/
	//flattener := draw2dbase.Transformer{Tr: gc.Current.Tr, Flattener: FtLineBuilder{Adder: gc.fillRasterizer}}
	flattener := FtLineBuilder{Adder: gc.fillRasterizer}
	hasShadow := gc.HasShadow()
	var offsetx float64
	var offsety float64
	gc.fillRasterizer.Dx = 0
	gc.fillRasterizer.Dy = 0
	if hasShadow {
		w, h := int(math.Abs(gc.Current.ShadowOffsetX)), int(math.Abs(gc.Current.ShadowOffsetY))
		if w != 0 || h != 0 {
			gc.fillRasterizer.SetBounds(gc.width+w, gc.height+h)
		}
		if gc.Current.ShadowOffsetX > 0 {
			offsetx = gc.Current.ShadowOffsetX
			gc.fillRasterizer.Dx = int(-offsetx)
		}
		if gc.Current.ShadowOffsetY > 0 {
			offsety = gc.Current.ShadowOffsetY
			gc.fillRasterizer.Dy = int(-offsety)
		}
	}

	for _, p := range paths {
		if !p.IsClosed() {
			p = p.Copy()
			p.Close()
		}
		if offsetx != 0 || offsety != 0 {
			p = p.Translated(offsetx, offsety)
		}
		Flatten(p, flattener, gc.Current.Tr.GetScale())
	}
	gc.painter.SetMask(gc.Current.mask)
	gc.painter.SetGlobalAlpha(gc.Current.GlobalAlpha)
	gc.painter.SetPattern(gc.Current.FillPattern, gc.Current.Tr)
	gc.painter.SetCompositeOperation(gc.Current.GlobalCompositeOperation)
	gc.painter.SetShadow(gc.Current.ShadowOffsetX, gc.Current.ShadowOffsetY, gc.Current.ShadowBlur, gc.Current.ShadowColor)
	gc.painter.Begin()
	gc.fillRasterizer.Rasterize(gc.painter)
	gc.fillRasterizer.Clear()
	gc.painter.End()
}

func (gc *GraphicContext2D) fillClip(painter raster.Painter, paths ...*Path) {
	gc.fillRasterizer.UseNonZeroWinding = gc.Current.FillRule == FillRuleWinding

	/**** first method ****/
	//flattener := draw2dbase.Transformer{Tr: gc.Current.Tr, Flattener: FtLineBuilder{Adder: gc.fillRasterizer}}
	flattener := FtLineBuilder{Adder: gc.fillRasterizer}
	gc.fillRasterizer.Dx = 0
	gc.fillRasterizer.Dy = 0
	for _, p := range paths {
		if !p.IsClosed() {
			p = p.Copy()
			p.Close()
		}
		Flatten(p, flattener, gc.Current.Tr.GetScale())
	}

	gc.fillRasterizer.Rasterize(painter)
	gc.fillRasterizer.Clear()
}

// FillStroke first fills the paths and than strokes them
func (gc *GraphicContext2D) FillStroke() {
	gc.fillStroke(gc.Current.Path)
}

func (gc *GraphicContext2D) fillStroke(paths ...*Path) {
	gc.fill(paths...)
	gc.stroke(paths...)
	// gc.fillRasterizer.UseNonZeroWinding = gc.Current.FillRule == FillRuleWinding
	// gc.strokeRasterizer.UseNonZeroWinding = true

	// //flattener := draw2dbase.Transformer{Tr: gc.Current.Tr, Flattener: FtLineBuilder{Adder: gc.fillRasterizer}}
	// flattener := FtLineBuilder{Adder: gc.fillRasterizer}

	// //stroker := draw2dbase.NewLineStroker(gc.Current.Cap, gc.Current.Join, draw2dbase.Transformer{Tr: gc.Current.Tr, Flattener: FtLineBuilder{Adder: gc.strokeRasterizer}})
	// stroker := &LineStroker{}
	// stroker.Cap = gc.Current.Cap
	// stroker.Join = gc.Current.Join
	// stroker.Flattener = &FtLineBuilder{Adder: gc.strokeRasterizer}
	// stroker.HalfLineWidth = gc.Current.LineWidth / 2
	// stroker.MiterLimitCheck = gc.Current.MiterLimit * gc.Current.LineWidth

	// var liner Flattener
	// if gc.Current.Dash != nil && len(gc.Current.Dash) > 0 {
	// 	liner = NewDashConverter(gc.Current.Dash, gc.Current.DashOffset, stroker)
	// } else {
	// 	liner = stroker
	// }

	// demux := DemuxFlattener{Flatteners: []Flattener{flattener, liner}}
	// for _, p := range paths {
	// 	Flatten(p, demux, gc.Current.Tr.GetScale())
	// }
	// // Fill
	// gc.paint(gc.fillRasterizer, gc.Current.FillColor, gc.Current.FillPattern)
	// // Stroke
	// gc.paint(gc.strokeRasterizer, gc.Current.StrokeColor, gc.Current.StrokePattern)
}

func (gc *GraphicContext2D) CreateLinearGradient(x0, y0, x1, y1 float64) Gradient {
	return newLinearGradient(x0, y0, x1, y1)
}

func (gc *GraphicContext2D) CreateRadialGradient(x0, y0, r0, x1, y1, r1 float64) Gradient {
	x0, y0 = gc.TransformPoint(x0, y0)
	x1, y1 = gc.TransformPoint(x1, y1)
	s := gc.Current.Tr.GetScale()
	r0 *= s
	r1 *= s
	return newRadialGradient(x0, y0, r0, x1, y1, r1)
}

func (gc *GraphicContext2D) CreatePattern(img image.Image, op RepeatOp) Pattern {
	return newSurfacePattern(img, op)
}

func (gc *GraphicContext2D) Clip() {
	clip := image.NewAlpha(image.Rect(0, 0, gc.width, gc.height))
	painter := NewAlphaOverPainter(clip)
	gc.fillClip(painter, gc.Current.Path)
	if gc.Current.mask == nil {
		gc.Current.mask = clip
	} else {
		mask := image.NewAlpha(image.Rect(0, 0, gc.width, gc.height))
		draw.DrawMask(mask, mask.Bounds(), clip, image.ZP, gc.Current.mask, image.ZP, draw.Over)
		gc.Current.mask = mask
	}
}

func (gc *GraphicContext2D) Rect(x, y, width, height float64) {
	gc.MoveTo(x, y)
	gc.LineTo(x+width, y)
	gc.LineTo(x+width, y+height)
	gc.LineTo(x, y+height)
	gc.ClosePath()
}

func (gc *GraphicContext2D) RoundedRect(x float64, y float64, width float64, height float64, arcWidth float64, arcHeight float64) {
	x2, y2 := x+width, y+height
	arcWidth = arcWidth / 2
	arcHeight = arcHeight / 2
	gc.MoveTo(x, y+arcHeight)
	gc.QuadraticCurveTo(x, y, x+arcWidth, y)
	gc.LineTo(x2-arcWidth, y)
	gc.QuadraticCurveTo(x2, y, x2, y+arcHeight)
	gc.LineTo(x2, y2-arcHeight)
	gc.QuadraticCurveTo(x2, y2, x2-arcWidth, y2)
	gc.LineTo(x+arcWidth, y2)
	gc.QuadraticCurveTo(x, y2, x, y2-arcHeight)
	gc.ClosePath()
}

func (gc *GraphicContext2D) Ellipse(cx, cy, rx, ry float64) {
	gc.ArcAngle(cx, cy, rx, ry, 0, -math.Pi*2)
	gc.ClosePath()
}

func (gc *GraphicContext2D) Circle(cx, cy, radius float64) {
	gc.ArcAngle(cx, cy, radius, radius, 0, -math.Pi*2)
	gc.ClosePath()
}
