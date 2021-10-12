package canvas

import (
	"math"
)

func (join LineJoin) Joiner() JoinerFunc {
	switch join {
	case LineJoinBevel:
		return nil
	case LineJoinRound:
		return fnRoundJoiner
	case LineJoinMiter:
		return fnMiterJoiner
	}
	return nil
}

var (
	fnRoundJoiner = JoinerFunc(joinRound)
	fnMiterJoiner = JoinerFunc(joinMiter)
)

func traceJoinRound(x0, y0, x1, y1, x2, y2 float64, ar []float64, clockWise bool) []float64 {
	ab := math.Hypot(ar[0]-x0, ar[1]-y0)
	ac := math.Hypot(x0-x1, y0-y1)
	a_acp := math.Asin(ab / 2 / ac)
	ap := ac * math.Tan(a_acp)
	cp := ac / math.Cos(a_acp)
	cd := math.Hypot(x1-x2, y1-y2)
	px := x2 + (cd-cp)/cd*(x1-x2)
	py := y2 + (cd-cp)/cd*(y1-y2)
	a_start := callAngle(px, py, x0, y0)
	a_angle := (90*math.Pi/180 - a_acp) * 2
	p := NewPath()
	if clockWise {
		TraceArc(p, px, py, ap, ap, a_start, a_angle, 1)
	} else {
		TraceArc(p, px, py, ap, ap, a_start, -a_angle, 1)
	}
	return p.Points
}

func joinRound(vertices []float64, rewind []float64, halfLineWidth float64, miterLimitCheck float64, clockWise bool, ar1 []float64, ar2 []float64) ([]float64, []float64) {
	if len(ar1) == 2 {
		x1, y1, b1 := callVerticesIntersect(vertices, vertices[len(vertices)-4:])
		x2, y2, b2 := callVerticesIntersect(rewind, rewind[len(rewind)-4:])
		if b1 && b2 {
			if clockWise {
				x0, y0 := vertices[len(vertices)-2], vertices[len(vertices)-1]
				ar := traceJoinRound(x0, y0, x1, y1, x2, y2, ar1, clockWise)
				return append(ar, ar1...), ar2
			} else {
				x0, y0 := rewind[len(rewind)-2], rewind[len(rewind)-1]
				ar := traceJoinRound(x0, y0, x2, y2, x1, y1, ar2, clockWise)
				return ar1, append(ar, ar2...)
			}
		}
		return ar1, ar2
	}
	x1, y1, b1 := callVerticesIntersect(vertices[len(vertices)-4:], ar1)
	x2, y2, b2 := callVerticesIntersect(rewind[len(rewind)-4:], ar2)
	if b1 && b2 {
		if clockWise {
			x0, y0 := vertices[len(vertices)-2], vertices[len(vertices)-1]
			ar := traceJoinRound(x0, y0, x1, y1, x2, y2, ar1, clockWise)
			return append(ar, ar1...), ar2
		} else {
			x0, y0 := rewind[len(rewind)-2], rewind[len(rewind)-1]
			ar := traceJoinRound(x0, y0, x2, y2, x1, y1, ar2, clockWise)
			return ar1, append(ar, ar2...)
		}
	}
	return ar1, ar2
}

// 计算直线斜率 K
func getLineK(x1, y1, x2, y2 float64) float64 {
	if x1 == x2 {
		return 0
	}
	return (y2 - y1) / (x2 - x1)
}

// 计算两条直线夹角
func GetLineArctan(x1, y1, x2, y2, x3, y3, x4, y4 float64) (float64, bool) {
	k1 := getLineK(x1, y1, x2, y2)
	k2 := getLineK(x3, y3, x4, y4)
	if k1 == k2 {
		return 0, false
	}
	return math.Atan((k1 - k2) / (1 + k2*k1)), true
}

// 判断点与线关系: 0 在直线上; <0 顺时针; >0 逆时针
func pointLineRelationship(x1, y1, x2, y2, x, y float64) float64 {
	return (y2-y1)*x + (x1-x2)*y + (x2*y1 - x1*y2)
}

// Ax+By+C=0
func GeneralEquation(x1, y1, x2, y2 float64) (float64, float64, float64) {
	return y2 - y1, x1 - x2, x2*y1 - x1*y2
}

// https://baike.baidu.com/item/直线的一般式方程/11052424?fr=aladdin
// 计算两条直线交点
func GetIntersectPointofLines(x1, y1, x2, y2, x3, y3, x4, y4 float64) (float64, float64, bool) {
	a1, b1, c1 := GeneralEquation(x1, y1, x2, y2)
	a2, b2, c2 := GeneralEquation(x3, y3, x4, y4)
	m := a1*b2 - a2*b1
	if m == 0 {
		return 0, 0, false
	}
	return (c2*b1 - c1*b2) / m, (c1*a2 - c2*a1) / m, true
}

func callVerticesIntersect(ar1 []float64, ar2 []float64) (float64, float64, bool) {
	return GetIntersectPointofLines(ar1[0], ar1[1], ar1[2], ar1[3], ar2[0], ar2[1], ar2[2], ar2[3])
}

// 顺时针画线时 ar 前两点为输出的直线，与之前的 vertices 计算交点,更改 ar 第一点为交点
func joinMiter(vertices []float64, rewind []float64, halfLineWidth float64, miterLimitCheck float64, clockWise bool, ar1 []float64, ar2 []float64) ([]float64, []float64) {
	//close path
	if len(ar1) == 2 {
		x1, y1, b1 := callVerticesIntersect(vertices, vertices[len(vertices)-4:])
		x2, y2, b2 := callVerticesIntersect(rewind, rewind[len(rewind)-4:])
		if b1 && b2 {
			if math.Hypot(x1-x2, y1-y2) > miterLimitCheck {
				return ar1, ar2
			}
			if clockWise {
				return []float64{x1, y1}, []float64{vertices[0], vertices[1]}
			} else {
				return []float64{rewind[0], rewind[1]}, []float64{x2, y2}
			}
		}
		return ar1, ar2
	}
	x1, y1, b1 := callVerticesIntersect(vertices[len(vertices)-4:], ar1)
	x2, y2, b2 := callVerticesIntersect(rewind[len(rewind)-4:], ar2)
	if b1 && b2 {
		if math.Hypot(x1-x2, y1-y2) > miterLimitCheck {
			return ar1, ar2
		}
		if clockWise {
			//ar[0], ar[1] = x1, y1
			return []float64{x1, y1, ar1[2], ar1[3]}, ar2
		} else {
			//ar[4], ar[5] = x2, y2
			return ar1, []float64{x2, y2, ar2[2], ar2[3]}
		}
	}
	return ar1, ar2
}

func callAngle(x1, y1, x2, y2 float64) (angle float64) {
	xx := x2 - x1
	yy := y2 - y1

	if xx == 0.0 {
		angle = math.Pi / 2.0
	} else {
		angle = math.Atan(math.Abs(yy / xx))
	}

	if (xx < 0.0) && (yy >= 0.0) {
		angle = math.Pi - angle
	} else if (xx < 0.0) && (yy < 0.0) {
		angle = math.Pi + angle
	} else if (xx >= 0.0) && (yy < 0.0) {
		angle = math.Pi*2.0 - angle
	}
	return
}
