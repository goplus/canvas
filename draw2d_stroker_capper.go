package canvas

import "math"

func (c LineCap) Capper() CapperFunc {
	switch c {
	case LineCapButt:
		return nil
	case LineCapRound:
		return fnCapperRound
	case LineCapSquare:
		return fnCapperSquare
	}
	return nil
}

var (
	fnCapperRound  = CapperFunc(capRound)
	fnCapperSquare = CapperFunc(capSquare)
)

func capRound(vertices []float64, rewind []float64, halfLineWidth float64) ([]float64, []float64) {
	x1, y1 := vertices[0], vertices[1]
	x2, y2 := rewind[0], rewind[1]
	pts1 := traceCapRound(x1, y1, x2, y2, halfLineWidth)
	//end
	x2, y2 = vertices[len(vertices)-2], vertices[len(vertices)-1]
	x1, y1 = rewind[len(rewind)-2], rewind[len(rewind)-1]
	pts2 := traceCapRound(x1, y1, x2, y2, halfLineWidth)

	return pts1, pts2
}

func capSquare(vertices []float64, rewind []float64, halfLineWidth float64) ([]float64, []float64) {
	x1, y1 := vertices[0], vertices[1]
	x2, y2 := rewind[0], rewind[1]
	pts1 := traceCapSquare(x1, y1, x2, y2, halfLineWidth)
	//end
	x2, y2 = vertices[len(vertices)-2], vertices[len(vertices)-1]
	x1, y1 = rewind[len(rewind)-2], rewind[len(rewind)-1]
	pts2 := traceCapSquare(x1, y1, x2, y2, halfLineWidth)
	return pts1, pts2
}

func traceCapRound(x1, y1, x2, y2, halfLineWidth float64) []float64 {
	x := x2 + (x1-x2)/2
	y := y1 + (y2-y1)/2
	angle := callAngle(x1, y1, x2, y2)
	p := NewPath()
	TraceArc(p, x, y, halfLineWidth, halfLineWidth, angle, math.Pi, 1)
	return p.Points
}

func traceCapSquare(x1, y1, x2, y2, halfLineWidth float64) []float64 {
	angle := math.Asin((x1 - x2) / (2 * halfLineWidth))
	xoff, yoff := halfLineWidth*math.Cos(angle), halfLineWidth*math.Sin(angle)
	if y2 < y1 {
		xoff *= -1
	}
	p := NewPath()
	p.LineTo(x2-xoff, y2-yoff)
	p.LineTo(x1-xoff, y1-yoff)
	return p.Points
}
