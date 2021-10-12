package canvas

import (
	"math"
)

func NewPath() *Path {
	return &Path{}
}

func (p *Path) AddPath(o *Path) {
	p.Components = append(p.Components, o.Components...)
	p.Points = append(p.Points, o.Points...)
}

func (p *Path) Translate(dx, dy float64) {
	for i := 0; i < len(p.Points); i += 2 {
		p.Points[i] += dx
		p.Points[i+1] += dy
	}
}

func (p *Path) Transfrom(tr Matrix) {
	tr.Transform(p.Points[:])
}

func (p *Path) Translated(dx, dy float64) *Path {
	o := p.Copy()
	o.Translate(dx, dy)
	return o
}

type Point struct {
	X float64
	Y float64
}

func (p *Path) AddLines(pts ...float64) {
	size := len(pts)
	if size < 4 {
		return
	}
	if size%2 != 0 {
		return
	}
	p.MoveTo(pts[0], pts[1])
	for i := 2; i < size; i += 2 {
		p.LineTo(pts[i], pts[i+1])
	}
}

func (p *Path) AddPolygon(pts ...float64) {
	size := len(pts)
	if size < 4 {
		return
	}
	if size%2 != 0 {
		return
	}
	p.MoveTo(pts[0], pts[1])
	for i := 2; i < size; i += 2 {
		p.LineTo(pts[i], pts[i+1])
	}
	p.Close()
}

// Ellipse draws an ellipse using a path with center (cx,cy) and radius (rx,ry)
func (p *Path) AddEllipse(cx, cy, rx, ry float64) {
	p.MoveTo(cx+rx, cy)
	p.ArcAngle(cx, cy, rx, ry, 0, -math.Pi*2)
	p.Close()
}

// Circle draws a circle using a path with center (cx,cy) and radius
func (p *Path) AddCircle(cx, cy, radius float64) {
	p.MoveTo(cx+radius, cy)
	p.ArcAngle(cx, cy, radius, radius, 0, -math.Pi*2)
	p.Close()
}

func (p *Path) AddRectangle(x, y, width, height float64) {
	x2, y2 := x+width, y+height
	p.MoveTo(x, y)
	p.LineTo(x2, y)
	p.LineTo(x2, y2)
	p.LineTo(x, y2)
	p.Close()
}

// RoundedRectangle draws a rectangle using a path between (x1,y1) and (x2,y2)
func (p *Path) AddRoundedRectangle(x, y, width, height, arcWidth, arcHeight float64) {
	x2, y2 := x+width, y+height
	arcWidth = arcWidth / 2
	arcHeight = arcHeight / 2
	p.MoveTo(x, y+arcHeight)
	p.QuadraticCurveTo(x, y, x+arcWidth, y)
	p.LineTo(x2-arcWidth, y)
	p.QuadraticCurveTo(x2, y, x2, y+arcHeight)
	p.LineTo(x2, y2-arcHeight)
	p.QuadraticCurveTo(x2, y2, x2-arcWidth, y2)
	p.LineTo(x+arcWidth, y2)
	p.QuadraticCurveTo(x, y2, x, y2-arcHeight)
	p.Close()
}
