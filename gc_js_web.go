//go:build js && !wegame && !weapp
// +build js,!wegame,!weapp

package canvas

import (
	"image"

	"github.com/goplus/canvas/jsutil"
)

func NewWebContext2D(width, height int) Context2D {
	canvas := document.Call("createElement", "canvas")
	canvas.Set("width", width)
	canvas.Set("height", height)
	ctx2d := canvas.Call("getContext", "2d")
	return &WebContext2D{canvas, ctx2d, width, height}
}

func (c *WebContext2D) SetImage(img image.Image, dx float64, dy float64) {
	nrgba := imageToNRGBA(img)
	data := jsutil.SliceToTypedArray(nrgba.Pix)
	imdata := c.ctx2d.Call("createImageData", nrgba.Rect.Dx(), nrgba.Rect.Dy())
	imdata.Get("data").Call("set", data)
	c.ctx2d.Call("putImageData", imdata, dx, dy)
}

func (c *WebContext2D) DrawImage(img image.Image, dx float64, dy float64) {
	// ctx := NewWebContext2DForImage(img).(*WebContext2D)
	// c.ctx2d.Call("drawImage", ctx.canvas, dx, dy)
	c.SetImage(img, dx, dy)
}

func (c *WebContext2D) DrawImageEx(img image.Image, sx, sy, sw, sh, dx, dy, dw, dh float64) {
	ctx := NewWebContext2DForImage(img).(*WebContext2D)
	c.ctx2d.Call("drawImage", ctx.canvas, sx, sy, sw, sh, dx, dy, dw, dh)
}
