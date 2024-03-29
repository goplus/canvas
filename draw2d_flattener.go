// Copyright 2010 The draw2d Authors. All rights reserved.
// created: 06/12/2010 by Laurent Le Goff

package canvas

import (
	"errors"
	"log"
	"math"

	"golang.org/x/image/math/fixed"

	"github.com/golang/freetype/raster"
)

type FtLineBuilder struct {
	Adder raster.Adder
}

func (liner FtLineBuilder) MoveTo(x, y float64) {
	liner.Adder.Start(fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)})
}

func (liner FtLineBuilder) LineTo(x, y float64) {
	liner.Adder.Add1(fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)})
}

func (liner FtLineBuilder) QuadCurveTo(cpx float64, cpy float64, x float64, y float64) {
	liner.Adder.Add2(fixed.Point26_6{X: fixed.Int26_6(cpx * 64), Y: fixed.Int26_6(cpy * 64)}, fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)})
}

func (liner FtLineBuilder) LineJoin() {
}

func (liner FtLineBuilder) Close() {
}

func (liner FtLineBuilder) End() {

}

// Liner receive segment definition
type Liner interface {
	// LineTo Draw a line from the current position to the point (x, y)
	LineTo(x, y float64)
}

// Flattener receive segment definition
type Flattener interface {
	// MoveTo Start a New line from the point (x, y)
	MoveTo(x, y float64)
	// LineTo Draw a line from the current position to the point (x, y)
	LineTo(x, y float64)
	// LineJoin use Round, Bevel or miter to join points
	LineJoin()
	// Close add the most recent starting point to close the path to create a polygon
	Close()
	// End mark the current line as finished so we can draw caps
	End()
}

// Flatten convert curves into straight segments keeping join segments info
func Flatten(path *Path, flattener Flattener, scale float64) {
	// First Point
	var startX, startY float64 = 0, 0
	// Current Point
	var x, y float64 = 0, 0
	i := 0
	for _, cmp := range path.Components {
		switch cmp {
		case MoveToCmp:
			x, y = path.Points[i], path.Points[i+1]
			startX, startY = x, y
			if i != 0 {
				flattener.End()
			}
			flattener.MoveTo(x, y)
			i += 2
		case LineToCmp:
			x, y = path.Points[i], path.Points[i+1]
			flattener.LineTo(x, y)
			flattener.LineJoin()
			i += 2
		case QuadCurveToCmp:
			TraceQuad(flattener, path.Points[i-2:], 0.5)
			x, y = path.Points[i+2], path.Points[i+3]
			flattener.LineTo(x, y)
			i += 4
		case CubicCurveToCmp:
			TraceCubic(flattener, path.Points[i-2:], 0.5)
			x, y = path.Points[i+4], path.Points[i+5]
			flattener.LineTo(x, y)
			i += 6
		case ArcAngleCmp:
			x, y = TraceArc(flattener, path.Points[i], path.Points[i+1], path.Points[i+2], path.Points[i+3], path.Points[i+4], path.Points[i+5], scale)
			flattener.LineTo(x, y)
			i += 6
		case CloseCmp:
			flattener.LineTo(startX, startY)
			flattener.Close()
		}
	}
	flattener.End()
}

type DemuxFlattener struct {
	Flatteners []Flattener
}

func (dc DemuxFlattener) MoveTo(x, y float64) {
	for _, flattener := range dc.Flatteners {
		flattener.MoveTo(x, y)
	}
}

func (dc DemuxFlattener) LineTo(x, y float64) {
	for _, flattener := range dc.Flatteners {
		flattener.LineTo(x, y)
	}
}

func (dc DemuxFlattener) LineJoin() {
	for _, flattener := range dc.Flatteners {
		flattener.LineJoin()
	}
}

func (dc DemuxFlattener) Close() {
	for _, flattener := range dc.Flatteners {
		flattener.Close()
	}
}

func (dc DemuxFlattener) End() {
	for _, flattener := range dc.Flatteners {
		flattener.End()
	}
}

// Transformer apply the Matrix transformation tr
type Transformer struct {
	Tr        Matrix
	Flattener Flattener
}

func (t Transformer) MoveTo(x, y float64) {
	u := x*t.Tr[0] + y*t.Tr[2] + t.Tr[4]
	v := x*t.Tr[1] + y*t.Tr[3] + t.Tr[5]
	t.Flattener.MoveTo(u, v)
}

func (t Transformer) LineTo(x, y float64) {
	u := x*t.Tr[0] + y*t.Tr[2] + t.Tr[4]
	v := x*t.Tr[1] + y*t.Tr[3] + t.Tr[5]
	t.Flattener.LineTo(u, v)
}

func (t Transformer) LineJoin() {
	t.Flattener.LineJoin()
}

func (t Transformer) Close() {
	t.Flattener.Close()
}

func (t Transformer) End() {
	t.Flattener.End()
}

const (
	// CurveRecursionLimit represents the maximum recursion that is really necessary to subsivide a curve into straight lines
	CurveRecursionLimit = 32
)

// Cubic
//	x1, y1, cpx1, cpy1, cpx2, cpy2, x2, y2 float64

// SubdivideCubic a Bezier cubic curve in 2 equivalents Bezier cubic curves.
// c1 and c2 parameters are the resulting curves
// length of c, c1 and c2 must be 8 otherwise it panics.
func SubdivideCubic(c, c1, c2 []float64) {
	// First point of c is the first point of c1
	c1[0], c1[1] = c[0], c[1]
	// Last point of c is the last point of c2
	c2[6], c2[7] = c[6], c[7]

	// Subdivide segment using midpoints
	c1[2] = (c[0] + c[2]) / 2
	c1[3] = (c[1] + c[3]) / 2

	midX := (c[2] + c[4]) / 2
	midY := (c[3] + c[5]) / 2

	c2[4] = (c[4] + c[6]) / 2
	c2[5] = (c[5] + c[7]) / 2

	c1[4] = (c1[2] + midX) / 2
	c1[5] = (c1[3] + midY) / 2

	c2[2] = (midX + c2[4]) / 2
	c2[3] = (midY + c2[5]) / 2

	c1[6] = (c1[4] + c2[2]) / 2
	c1[7] = (c1[5] + c2[3]) / 2

	// Last Point of c1 is equal to the first point of c2
	c2[0], c2[1] = c1[6], c1[7]
}

// TraceCubic generate lines subdividing the cubic curve using a Liner
// flattening_threshold helps determines the flattening expectation of the curve
func TraceCubic(t Liner, cubic []float64, flatteningThreshold float64) error {
	if len(cubic) < 8 {
		return errors.New("cubic length must be >= 8")
	}
	// Allocation curves
	var curves [CurveRecursionLimit * 8]float64
	copy(curves[0:8], cubic[0:8])
	i := 0

	// current curve
	var c []float64

	var dx, dy, d2, d3 float64

	for i >= 0 {
		c = curves[i:]
		dx = c[6] - c[0]
		dy = c[7] - c[1]

		d2 = math.Abs((c[2]-c[6])*dy - (c[3]-c[7])*dx)
		d3 = math.Abs((c[4]-c[6])*dy - (c[5]-c[7])*dx)

		// if it's flat then trace a line
		if (d2+d3)*(d2+d3) <= flatteningThreshold*(dx*dx+dy*dy) || i == len(curves)-8 {
			t.LineTo(c[6], c[7])
			i -= 8
		} else {
			// second half of bezier go lower onto the stack
			SubdivideCubic(c, curves[i+8:], curves[i:])
			i += 8
		}
	}
	return nil
}

// Quad
// x1, y1, cpx1, cpy2, x2, y2 float64

// SubdivideQuad a Bezier quad curve in 2 equivalents Bezier quad curves.
// c1 and c2 parameters are the resulting curves
// length of c, c1 and c2 must be 6 otherwise it panics.
func SubdivideQuad(c, c1, c2 []float64) {
	// First point of c is the first point of c1
	c1[0], c1[1] = c[0], c[1]
	// Last point of c is the last point of c2
	c2[4], c2[5] = c[4], c[5]

	// Subdivide segment using midpoints
	c1[2] = (c[0] + c[2]) / 2
	c1[3] = (c[1] + c[3]) / 2
	c2[2] = (c[2] + c[4]) / 2
	c2[3] = (c[3] + c[5]) / 2
	c1[4] = (c1[2] + c2[2]) / 2
	c1[5] = (c1[3] + c2[3]) / 2
	c2[0], c2[1] = c1[4], c1[5]
	return
}

// TraceQuad generate lines subdividing the curve using a Liner
// flattening_threshold helps determines the flattening expectation of the curve
func TraceQuad(t Liner, quad []float64, flatteningThreshold float64) error {
	if len(quad) < 6 {
		return errors.New("quad length must be >= 6")
	}
	// Allocates curves stack
	var curves [CurveRecursionLimit * 6]float64
	copy(curves[0:6], quad[0:6])
	i := 0
	// current curve
	var c []float64
	var dx, dy, d float64

	for i >= 0 {
		c = curves[i:]
		dx = c[4] - c[0]
		dy = c[5] - c[1]

		d = math.Abs(((c[2]-c[4])*dy - (c[3]-c[5])*dx))

		// if it's flat then trace a line
		if (d*d) <= flatteningThreshold*(dx*dx+dy*dy) || i == len(curves)-6 {
			t.LineTo(c[4], c[5])
			i -= 6
		} else {
			// second half of bezier go lower onto the stack
			SubdivideQuad(c, curves[i+6:], curves[i:])
			i += 6
		}
	}
	return nil
}

// TraceArc trace an arc using a Liner
func TraceArc(t Liner, x, y, rx, ry, start, angle, scale float64) (lastX, lastY float64) {
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
		t.LineTo(curX, curY)
	}
}

func TraceArcEx(t Liner, x, y, rx, ry, startAngle, endAngle, scale float64, counterclockwise bool) (lastX, lastY float64) {
	var angle float64
	if !counterclockwise {
		if endAngle > startAngle {
			angle = endAngle - startAngle
		} else {
			endAngle += 2 * math.Pi
			angle = endAngle - startAngle
			log.Println(startAngle*180/3.14, endAngle*180/3.14, angle*180/3.14)
		}
	}
	//angle := endAngle - startAngle
	return TraceArc(t, x, y, rx, ry, startAngle, angle, scale)
	// ra := (math.Abs(rx) + math.Abs(ry)) / 2
	// da := math.Acos(ra/(ra+0.125/scale)) * 2
	// //normalize
	// if counterclockwise {
	// 	da = -da
	// }
	// angle := startAngle + da
	// var curX, curY float64
	// for {
	// 	if (angle < endAngle-da/4) == counterclockwise {
	// 		curX = x + math.Cos(endAngle)*rx
	// 		curY = y + math.Sin(endAngle)*ry
	// 		return curX, curY
	// 	}
	// 	curX = x + math.Cos(angle)*rx
	// 	curY = y + math.Sin(angle)*ry

	// 	angle += da
	// 	t.LineTo(curX, curY)
	// }
}

func fix(x float64) fixed.Int26_6 {
	return fixed.Int26_6(x * 64)
}

func unfix(x fixed.Int26_6) float64 {
	const shift, mask = 6, 1<<6 - 1
	if x >= 0 {
		return float64(x>>shift) + float64(x&mask)/64
	}
	x = -x
	if x >= 0 {
		return -(float64(x>>shift) + float64(x&mask)/64)
	}
	return 0
}

func fixp(x, y float64) fixed.Point26_6 {
	return fixed.Point26_6{fix(x), fix(y)}
}
