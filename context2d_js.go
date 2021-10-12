// +build js

package canvas

import (
	"image"
)

func NewContext2D(width int, height int) Context2D {
	return NewWebContext2D(width, height)
}

func NewContext2DForImage(img image.Image) Context2D {
	return NewWebContext2DForImage(img)
}
