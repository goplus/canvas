package canvas

import "math"

type Vec struct {
	X, Y float64
}

func (v Vec) Add(u Vec) Vec {
	return Vec{
		v.X + u.X,
		v.Y + u.Y,
	}
}

func (v Vec) Sub(u Vec) Vec {
	return Vec{
		v.X - u.X,
		v.Y - u.Y,
	}
}

func (v Vec) Mul(u Vec) Vec {
	return Vec{v.X * u.X, v.Y * u.Y}
}

func (v Vec) Mulf(f float64) Vec {
	return Vec{v.X * f, v.Y * f}
}

func (v Vec) Div(u Vec) Vec {
	return Vec{v.X / u.X, v.Y / u.Y}
}

func (v Vec) Divf(f float64) Vec {
	return Vec{v.X / f, v.Y / f}
}

func (v Vec) Len() float64 {
	return math.Hypot(v.X, v.Y)
}

func (v Vec) Angle() float64 {
	return math.Atan2(v.Y, v.X)
}

func (v Vec) Rotated(angle float64) Vec {
	sin, cos := math.Sincos(angle)
	return Vec{
		v.X*cos - v.Y*sin,
		v.X*sin + v.Y*cos,
	}
}

func (v Vec) Normal() Vec {
	return Vec{-v.Y, v.X}
}

func (v Vec) Norm() Vec {
	return v.Mulf(1.0 / v.Len())
}

func (v Vec) Dot(u Vec) float64 {
	return v.X*u.X + v.Y*u.Y
}

func (v Vec) Scaledf(c float64) Vec {
	return Vec{v.X * c, v.Y * c}
}

func (v Vec) Unit() Vec {
	if v.X == 0 && v.Y == 0 {
		return Vec{1, 0}
	}
	return v.Scaledf(1 / v.Len())
}

func (v Vec) Cross(u Vec) float64 {
	return v.X*u.Y - u.X*v.Y
}

func (v Vec) Project(u Vec) Vec {
	len := v.Dot(u) / u.Len()
	return u.Unit().Scaledf(len)
}

func (v Vec) Atan2() float64 {
	return math.Atan2(v.Y, v.X)
}

func Lerp(a, b Vec, t float64) Vec {
	return a.Scaledf(1 - t).Add(b.Scaledf(t))
}

// vector angle
func VectorAngle(a, b Vec) (angle float64) {
	xx := b.X - a.X
	yy := b.Y - a.Y

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
