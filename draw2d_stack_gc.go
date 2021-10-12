// Copyright 2010 The draw2d Authors. All rights reserved.
// created: 21/11/2010 by Laurent Le Goff

package canvas

import (
	"image"
	"image/color"
	"math"
)

type FillRule int

const (
	// FillRuleEvenOdd determines the "insideness" of a point in the shape
	// by drawing a ray from that point to infinity in any direction
	// and counting the number of path segments from the given shape that the ray crosses.
	// If this number is odd, the point is inside; if even, the point is outside.
	FillRuleEvenOdd FillRule = iota
	// FillRuleWinding determines the "insideness" of a point in the shape
	// by drawing a ray from that point to infinity in any direction
	// and then examining the places where a segment of the shape crosses the ray.
	// Starting with a count of zero, add one each time a path segment crosses
	// the ray from left to right and subtract one each time
	// a path segment crosses the ray from right to left. After counting the crossings,
	// if the result is zero then the point is outside the path. Otherwise, it is inside.
	FillRuleWinding
)

type StackGraphicContext struct {
	Current *ContextStack
}

type ContextStack struct {
	Tr                       Matrix
	mask                     *image.Alpha
	Path                     *Path
	LineWidth                float64
	GlobalAlpha              float64
	Dash                     []float64
	DashOffset               float64
	MiterLimit               float64
	StrokePattern            Pattern
	FillPattern              Pattern
	FillRule                 FillRule
	Cap                      LineCap
	Join                     LineJoin
	GlobalCompositeOperation CompositeOperation
	ShadowOffsetX            float64
	ShadowOffsetY            float64
	ShadowBlur               float64
	ShadowColor              color.Color
	Font                     *Font
	TextAlign                TextAlign
	TextBaseline             TextBaseline
	Previous                 *ContextStack
}

/**
 * Create a new Graphic context from an image
 */
func NewStackGraphicContext() *StackGraphicContext {
	gc := &StackGraphicContext{}
	gc.Current = new(ContextStack)
	gc.Current.Tr = NewIdentityMatrix()
	gc.Current.Path = NewPath()
	gc.Current.LineWidth = 1.0
	gc.Current.GlobalAlpha = 1.0
	gc.Current.StrokePattern = NewSolidPattern(color.Black)
	gc.Current.FillPattern = NewSolidPattern(color.Black)
	gc.Current.Cap = LineCapButt
	gc.Current.FillRule = FillRuleEvenOdd
	gc.Current.Join = LineJoinMiter
	gc.Current.MiterLimit = 10
	gc.Current.GlobalCompositeOperation = SourceOver
	gc.Current.Font = defaultFont
	gc.Current.TextAlign = AlignLeft
	gc.Current.TextBaseline = AlignAlphabetic
	gc.Current.ShadowBlur = 0
	gc.Current.ShadowOffsetX = 0
	gc.Current.ShadowOffsetY = 0
	gc.Current.ShadowColor = color.Transparent
	return gc
}

func (gc *StackGraphicContext) Transform(a, b, c, d, e, f float64) {
	gc.Current.Tr.Compose(Matrix([6]float64{a, b, c, d, e, f}))
}

func (gc *StackGraphicContext) SetTransform(a, b, c, d, e, f float64) {
	gc.Current.Tr = Matrix([6]float64{a, b, c, d, e, f})
}

func (gc *StackGraphicContext) SetMatrixTransform(tr Matrix) {
	gc.Current.Tr = tr
}

func (gc *StackGraphicContext) ResetTransform() {
	gc.Current.Tr = NewIdentityMatrix()
}

func (gc *StackGraphicContext) MatrixTransform() Matrix {
	return gc.Current.Tr
}

func (gc *StackGraphicContext) ComposeMatrixTransform(Tr Matrix) {
	gc.Current.Tr.Compose(Tr)
}

func (gc *StackGraphicContext) Rotate(angle float64) {
	gc.Current.Tr.Rotate(angle)
}

func (gc *StackGraphicContext) RotateAbout(angle, x, y float64) {
	gc.Translate(x, y)
	gc.Rotate(angle)
	gc.Translate(-x, -y)
}

func (gc *StackGraphicContext) Translate(tx, ty float64) {
	gc.Current.Tr.Translate(tx, ty)
}

func (gc *StackGraphicContext) Scale(sx, sy float64) {
	gc.Current.Tr.Scale(sx, sy)
}

func (gc *StackGraphicContext) ScaleAbout(sx, sy, x, y float64) {
	gc.Translate(x, y)
	gc.Scale(sx, sy)
	gc.Translate(-x, -y)
}

func (gc *StackGraphicContext) SetStrokeStyle(pattern Pattern) {
	gc.Current.StrokePattern = pattern
}

func (gc *StackGraphicContext) StrokeStyle() Pattern {
	return gc.Current.StrokePattern
}

func (gc *StackGraphicContext) SetStrokeColor(c color.Color) {
	gc.Current.StrokePattern = NewSolidPattern(c)
}

func (gc *StackGraphicContext) StrokeColor() color.Color {
	if pattern, ok := gc.Current.StrokePattern.(*SolidPattern); ok {
		return pattern.color
	}
	return nil
}

func (gc *StackGraphicContext) SetFillStyle(pattern Pattern) {
	gc.Current.FillPattern = pattern
}

func (gc *StackGraphicContext) FillStyle() Pattern {
	return gc.Current.FillPattern
}

func (gc *StackGraphicContext) SetFillColor(c color.Color) {
	gc.Current.FillPattern = NewSolidPattern(c)
}

func (gc *StackGraphicContext) FillColor() color.Color {
	if pattern, ok := gc.Current.FillPattern.(*SolidPattern); ok {
		return pattern.color
	}
	return nil
}
func (gc *StackGraphicContext) SetFillRule(f FillRule) {
	gc.Current.FillRule = f
}

func (gc *StackGraphicContext) SetLineWidth(lineWidth float64) {
	gc.Current.LineWidth = lineWidth
}

func (gc *StackGraphicContext) LineWidth() float64 {
	return gc.Current.LineWidth
}

func (gc *StackGraphicContext) SetLineCap(cap LineCap) {
	gc.Current.Cap = cap
}

func (gc *StackGraphicContext) SetLineJoin(join LineJoin) {
	gc.Current.Join = join
}

func (gc *StackGraphicContext) LineCap() LineCap {
	return gc.Current.Cap
}

func (gc *StackGraphicContext) LineJoin() LineJoin {
	return gc.Current.Join
}

func (gc *StackGraphicContext) SetMiterLimit(limit float64) {
	gc.Current.MiterLimit = limit
}

func (gc *StackGraphicContext) MiterLimit() float64 {
	return gc.Current.MiterLimit
}

func (gc *StackGraphicContext) SetLineDash(dash []float64) {
	gc.Current.Dash = dash
	if len(dash)%2 != 0 {
		gc.Current.Dash = append(gc.Current.Dash, dash...)
	}
}

func (gc *StackGraphicContext) LineDash() []float64 {
	return gc.Current.Dash
}

func (gc *StackGraphicContext) SetLineDashOffset(offset float64) {
	gc.Current.DashOffset = offset
}

func (gc *StackGraphicContext) LineDashOffset() float64 {
	return gc.Current.DashOffset
}

func (gc *StackGraphicContext) SetFont(f *Font) {
	gc.Current.Font = f
}

func (gc *StackGraphicContext) GetFont() *Font {
	return gc.Current.Font
}

func (gc *StackGraphicContext) SetTextAlign(align TextAlign) {
	gc.Current.TextAlign = align
}

func (gc *StackGraphicContext) TextAlign() TextAlign {
	return gc.Current.TextAlign
}

func (gc *StackGraphicContext) SetTextBaseline(base TextBaseline) {
	gc.Current.TextBaseline = base
}

func (gc *StackGraphicContext) TextBaseline() TextBaseline {
	return gc.Current.TextBaseline
}

func (gc *StackGraphicContext) BeginPath() {
	gc.Current.Path.Clear()
}

func (gc *StackGraphicContext) GetPath() *Path {
	return gc.Current.Path
}

func (gc *StackGraphicContext) IsEmpty() bool {
	return gc.Current.Path.IsEmpty()
}

func (gc *StackGraphicContext) LastPoint() (float64, float64) {
	return gc.Current.Path.LastPoint()
}

func (gc *StackGraphicContext) MoveTo(x, y float64) {
	x, y = gc.TransformPoint(x, y)
	gc.Current.Path.MoveTo(x, y)
}

func (gc *StackGraphicContext) LineTo(x, y float64) {
	x, y = gc.TransformPoint(x, y)
	gc.Current.Path.LineTo(x, y)
}

// func (gc *StackGraphicContext) QuadCurveTo(cx, cy, x, y float64) {
// 	x, y = gc.TransformPoint(x, y)
// 	cx, cy = gc.TransformPoint(cx, cy)
// 	gc.Current.Path.QuadCurveTo(cx, cy, x, y)
// }

// func (gc *StackGraphicContext) CubicCurveTo(cx1, cy1, cx2, cy2, x, y float64) {
// 	cx1, cy1 = gc.TransformPoint(cx1, cy1)
// 	cx2, cy2 = gc.TransformPoint(cx2, cy2)
// 	x, y = gc.TransformPoint(x, y)
// 	gc.Current.Path.CubicCurveTo(cx1, cy1, cx2, cy2, x, y)
// }

func (gc *StackGraphicContext) QuadraticCurveTo(cx, cy, x, y float64) {
	x, y = gc.TransformPoint(x, y)
	cx, cy = gc.TransformPoint(cx, cy)
	gc.Current.Path.QuadraticCurveTo(cx, cy, x, y)
}

func (gc *StackGraphicContext) BezierCurveTo(cx1, cy1, cx2, cy2, x, y float64) {
	cx1, cy1 = gc.TransformPoint(cx1, cy1)
	cx2, cy2 = gc.TransformPoint(cx2, cy2)
	x, y = gc.TransformPoint(x, y)
	gc.Current.Path.BezierCurveTo(cx1, cy1, cx2, cy2, x, y)
}

// func (gc *StackGraphicContext) ArcTo(cx, cy, rx, ry, startAngle, angle float64) {
// 	// cx, cy = gc.TransformPoint(cx, cy)
// 	// rx, ry = gc.TransformScale(rx, ry)
// 	// gc.Current.Path.ArcTo(cx, cy, rx, ry, startAngle, angle)
// 	gc.arcTo(cx, cy, rx, ry, startAngle, angle)
// }

func normAngle(angle float64) float64 {
	angle = math.Mod(angle, math.Pi*2)
	if angle < 0 {
		angle += math.Pi * 2
	}
	return angle
}

func (gc *StackGraphicContext) Arc(cx, cy, radius, startAngle, endAngle float64, counterclockwise bool) {
	// check circle
	if !counterclockwise && endAngle-startAngle >= 2*math.Pi {
		gc.ArcAngle(cx, cy, radius, radius, startAngle, 2*math.Pi)
		return
	} else if counterclockwise && startAngle-endAngle >= 2*math.Pi {
		gc.ArcAngle(cx, cy, radius, radius, startAngle, -2*math.Pi)
		return
	}
	// norm angle
	startAngle = normAngle(startAngle)
	endAngle = normAngle(endAngle)
	if !counterclockwise && endAngle <= startAngle {
		endAngle += math.Pi * 2
	} else if counterclockwise && endAngle >= startAngle {
		endAngle -= math.Pi * 2
	}
	gc.ArcAngle(cx, cy, radius, radius, startAngle, endAngle-startAngle)
}

func (gc *StackGraphicContext) ArcTo(x1, y1, x2, y2, radius float64) {
	if gc.Current.Path.IsEmpty() {
		return
	}
	p0 := Vec{gc.Current.Path.x, gc.Current.Path.y}
	p1 := Vec{x1, y1}
	p2 := Vec{x2, y2}
	v0, v1 := p0.Sub(p1).Norm(), p2.Sub(p1).Norm()
	angle := math.Acos(v0.Dot(v1))
	// should be in the range [0-pi]. if parallel, use a straight line
	if angle <= 0 || angle >= math.Pi {
		gc.LineTo(x2, y2)
		return
	}
	// cv are the vectors orthogonal to the lines that point to the center of the circle
	cv0 := Vec{-v0.Y, v0.X}
	cv1 := Vec{v1.Y, -v1.X}
	x := cv1.Sub(cv0).Div(v0.Sub(v1)).X * radius
	if x < 0 {
		cv0 = cv0.Mulf(-1)
		cv1 = cv1.Mulf(-1)
	}
	center := p1.Add(v0.Mulf(math.Abs(x))).Add(cv0.Mulf(radius))
	a0, a1 := cv0.Mulf(-1).Atan2(), cv1.Mulf(-1).Atan2()
	gc.Arc(center.X, center.Y, radius, a0, a1, x > 0)
}

func (gc *StackGraphicContext) traceArc(x, y, rx, ry, start, angle, scale float64) (lastX, lastY float64) {
	end := start + angle
	clockWise := true
	if angle < 0 {
		clockWise = false
	}
	ra := (math.Abs(rx) + math.Abs(ry)) / 2
	da := math.Acos(ra/(ra+0.125/scale)) * 2
	//normalize
	if !clockWise {
		da = -da
	}
	angle = start + da
	var curX, curY float64
	for {
		if (angle < end-da/4) != clockWise {
			curX = x + math.Cos(end)*rx
			curY = y + math.Sin(end)*ry
			return curX, curY
		}
		curX = x + math.Cos(angle)*rx
		curY = y + math.Sin(angle)*ry

		angle += da
		gc.LineTo(curX, curY)
	}
}

func (gc *StackGraphicContext) ArcAngle(x, y, rx, ry, startAngle, sweepAngle float64) {
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
			gc.LineTo(x0, y0)
		}
		gc.QuadraticCurveTo(cx, cy, x2, y2)
	}
}

func (gc *StackGraphicContext) ClosePath() {
	gc.Current.Path.Close()
}

func (gc *StackGraphicContext) Save() {
	context := new(ContextStack)
	context.mask = gc.Current.mask
	context.LineWidth = gc.Current.LineWidth
	context.GlobalAlpha = gc.Current.GlobalAlpha
	context.StrokePattern = gc.Current.StrokePattern
	context.FillPattern = gc.Current.FillPattern
	context.FillRule = gc.Current.FillRule
	context.Dash = gc.Current.Dash
	context.DashOffset = gc.Current.DashOffset
	context.Cap = gc.Current.Cap
	context.Join = gc.Current.Join
	context.MiterLimit = gc.Current.MiterLimit
	context.GlobalCompositeOperation = gc.Current.GlobalCompositeOperation
	context.Path = gc.Current.Path.Copy()
	context.Font = gc.Current.Font
	context.TextAlign = gc.Current.TextAlign
	context.TextBaseline = gc.Current.TextBaseline
	context.ShadowOffsetX = gc.Current.ShadowOffsetX
	context.ShadowOffsetY = gc.Current.ShadowOffsetY
	context.ShadowBlur = gc.Current.ShadowBlur
	context.ShadowColor = gc.Current.ShadowColor
	//copy(context.Tr[:], gc.Current.Tr[:])
	context.Tr = gc.Current.Tr.Copy()
	context.Previous = gc.Current
	gc.Current = context
}

func (gc *StackGraphicContext) Restore() {
	if gc.Current.Previous != nil {
		oldContext := gc.Current
		gc.Current = gc.Current.Previous
		oldContext.Previous = nil
	}
}

func (gc *StackGraphicContext) TransformPoint(x, y float64) (float64, float64) {
	return gc.Current.Tr.TransformPoint(x, y)
}

func (gc *StackGraphicContext) TransformScale(w, h float64) (float64, float64) {
	x, y := gc.Current.Tr.GetScaling()
	return w * x, h * y
}

func (gc *StackGraphicContext) SetGlobalAlpha(alpha float64) {
	if alpha < 0 || alpha > 1 {
		return
	}
	gc.Current.GlobalAlpha = alpha
}

func (gc *StackGraphicContext) GlobalAlpha() float64 {
	return gc.Current.GlobalAlpha
}

func (gc *StackGraphicContext) SetGlobalCompositeOperation(op CompositeOperation) {
	gc.Current.GlobalCompositeOperation = op
}

func (gc *StackGraphicContext) GlobalCompositeOperation() CompositeOperation {
	return gc.Current.GlobalCompositeOperation
}

func (gc *StackGraphicContext) SetShadowOffset(x float64, y float64) {
	gc.Current.ShadowOffsetX = x
	gc.Current.ShadowOffsetY = y
}

func (gc *StackGraphicContext) ShadowOffset() (float64, float64) {
	return gc.Current.ShadowOffsetX, gc.Current.ShadowOffsetY
}

func (gc *StackGraphicContext) SetShadowBlur(blur float64) {
	gc.Current.ShadowBlur = blur
}

func (gc *StackGraphicContext) ShadowBlur() float64 {
	return gc.Current.ShadowBlur
}

func (gc *StackGraphicContext) SetShadowColor(clr color.Color) {
	gc.Current.ShadowColor = clr
}

func (gc *StackGraphicContext) ShadowColor() color.Color {
	return gc.Current.ShadowColor
}
