// Copyright 2010 The draw2d Authors. All rights reserved.
// created: 21/11/2010 by Laurent Le Goff

package canvas

import (
	"math"
)

// return start and end cappter array
type CapperFunc func(vertices []float64, rewind []float64, halfLineWidth float64) ([]float64, []float64)

// update input vertex and return to new vertex
type JoinerFunc func(vertices []float64, rewind []float64, halfLineWidth float64, miterLimitCheck float64, clockWise bool, inputVertex []float64, inputRewin []float64) ([]float64, []float64)

type LineStroker struct {
	Flattener        Flattener
	HalfLineWidth    float64
	MiterLimitCheck  float64
	Cap              LineCap
	Join             LineJoin
	vertices         []float64
	rewind           []float64
	x, y, nx, ny     float64
	secondx, secondy float64
	lastx, lasty     float64
	join             int
	closepath        bool
}

func (l *LineStroker) MoveTo(x, y float64) {
	l.x, l.y = x, y
	l.lastx, l.lasty = x, y
	l.join = 0
	l.closepath = false
}

func (l *LineStroker) LineTo(x, y float64) {
	l.line(l.x, l.y, x, y)
}

func (l *LineStroker) LineCap() {
	if l.closepath {
		return
	}
	fn := l.Cap.Capper()
	if fn != nil {
		pts1, pts2 := fn(l.vertices, l.rewind, l.HalfLineWidth)
		l.vertices = append(pts1, l.vertices...)
		l.vertices = append(l.vertices, pts2...)
	}
}

func (l *LineStroker) LineJoin() {
	return
}

func (l *LineStroker) line(x1, y1, x2, y2 float64) {
	dx := (x2 - x1)
	dy := (y2 - y1)
	d := vectorDistance(dx, dy)
	if d != 0 {
		l.join++
		nx := dy * l.HalfLineWidth / d
		ny := -(dx * l.HalfLineWidth / d)
		chk := pointLineRelationship(l.lastx, l.lasty, x1, y1, x2, y2)
		ar1 := []float64{x1 + nx, y1 + ny, x2 + nx, y2 + ny}
		ar2 := []float64{x1 - nx, y1 - ny, x2 - nx, y2 - ny}
		if chk != 0 {
			fn := l.Join.Joiner()
			if fn != nil {
				ar1, ar2 = fn(l.vertices, l.rewind, l.HalfLineWidth, l.MiterLimitCheck, chk < 0, ar1, ar2)
			}
		}
		//l.appendVertex(ar...)
		l.vertices = append(l.vertices, ar1...)
		l.rewind = append(l.rewind, ar2...)
		l.lastx, l.lasty, l.x, l.y, l.nx, l.ny = l.x, l.y, x2, y2, nx, ny
		if l.join == 1 {
			l.secondx, l.secondy = x2, y2
		}
	}
}

func (l *LineStroker) Close() {
	if len(l.vertices) > 1 {
		ar1 := []float64{l.vertices[0], l.vertices[1]}
		ar2 := []float64{l.rewind[0], l.rewind[1]}
		chk := pointLineRelationship(l.lastx, l.lasty, l.x, l.y, l.secondx, l.secondy)
		if chk != 0 {
			fn := l.Join.Joiner()
			if fn != nil {
				ar1, ar2 = fn(l.vertices, l.rewind, l.HalfLineWidth, l.MiterLimitCheck, chk < 0, ar1, ar2)
			}
		}
		l.vertices = append(l.vertices, ar1...)
		l.rewind = append(l.rewind, ar2...)
		//l.appendVertex(ar...)
	}
	l.closepath = true
}

func (l *LineStroker) End() {
	if len(l.vertices) > 1 {
		l.LineCap()
		l.Flattener.MoveTo(l.vertices[0], l.vertices[1])
		for i, j := 2, 3; j < len(l.vertices); i, j = i+2, j+2 {
			l.Flattener.LineTo(l.vertices[i], l.vertices[j])
		}
	}
	for i, j := len(l.rewind)-2, len(l.rewind)-1; j > 0; i, j = i-2, j-2 {
		l.Flattener.LineTo(l.rewind[i], l.rewind[j])
	}
	if len(l.vertices) > 1 {
		l.Flattener.LineTo(l.vertices[0], l.vertices[1])
	}
	l.Flattener.End()
	// reinit vertices
	l.vertices = l.vertices[0:0]
	l.rewind = l.rewind[0:0]
	l.join, l.lastx, l.lasty, l.secondx, l.secondy, l.x, l.y, l.nx, l.ny = 0, 0, 0, 0, 0, 0, 0, 0, 0
}

func (l *LineStroker) appendVertex(vertices ...float64) {
	s := len(vertices) / 2
	l.vertices = append(l.vertices, vertices[:s]...)
	l.rewind = append(l.rewind, vertices[s:]...)
}

func vectorDistance(dx, dy float64) float64 {
	return float64(math.Sqrt(dx*dx + dy*dy))
}
