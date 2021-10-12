//go:build wegame
// +build wegame

package canvas

import (
	"image"

	"syscall/js"
)

func NewWebContext2D(width, height int) Context2D {
	canvas := js.Global().Get("wx").Call("createCanvas")
	canvas.Set("width", width)
	canvas.Set("height", height)
	ctx2d := canvas.Call("getContext", "2d")
	return &WebContext2D{canvas, ctx2d, width, height}
}

func NewWXGameContext2D() Context2D {
	canvas := js.Global().Get("wx").Call("createCanvas")
	width := canvas.Get("width").Int()
	height := canvas.Get("height").Int()
	ctx2d := canvas.Call("getContext", "2d")
	return &WebContext2D{canvas, ctx2d, width, height}
}

var (
	bkctx *WebContext2D
)

func (c *WebContext2D) SetImage(img image.Image, dx float64, dy float64) {
	nrgba := imageToNRGBA(img)
	data := js.TypedArrayOf(nrgba.Pix)
	imdata := c.ctx2d.Call("createImageData", nrgba.Rect.Dx(), nrgba.Rect.Dy())
	imdata.Get("data").Call("set", data)
	c.ctx2d.Call("putImageData", imdata, dx, dy)
}

func (c *WebContext2D) DrawImage(img image.Image, dx float64, dy float64) {
	c.SetImage(img, dx, dy)
}

func (c *WebContext2D) DrawImageEx(img image.Image, sx, sy, sw, sh, dx, dy, dw, dh float64) {
	ctx := NewWXGameContext2D().(*WebContext2D)
	ctx.SetImage(img, 0, 0)
	c.ctx2d.Call("drawImage", ctx.canvas, sx, sy, sw, sh, dx, dy, dw, dh)
}
