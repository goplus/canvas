//go:build weapp
// +build weapp

package canvas

import (
	"image"
	"syscall/js"

	"github.com/goplus/canvas/jsutil"
)

func NewWebContext2D(width, height int) Context2D {
	return nil
}

func (c *WebContext2D) SetImage(img image.Image, dx float64, dy float64) {
	rgba := imageToNRGBA(img)
	data := jsutil.SliceToTypedArray(rgba.Pix)
	//p2 := jsId.Invoke(rgba.Pix)
	//data := js.Global.Get("Uint8ClampedArray").New(4 * img.Bounds().Dx() * img.Bounds().Dy())
	//jsCopy.Invoke(data, p2)
	obj := js.Global().Get("Object").New()
	obj.Set("canvasId", "myCanvas")
	obj.Set("data", data)
	obj.Set("x", dx)
	obj.Set("y", dy)
	obj.Set("width", img.Bounds().Dx())
	obj.Set("height", img.Bounds().Dy())
	js.Global().Get("wx").Call("canvasPutImageData", obj)
}

func (c *WebContext2D) DrawImage(img image.Image, dx float64, dy float64) {
	c.SetImage(img, 0, 0)
}

func (c *WebContext2D) DrawImageEx(img image.Image, sx, sy, sw, sh, dx, dy, dw, dh float64) {
	//	ctx := NewWXGameContext2D().(*WebContext2D)
	//	ctx.SetImage(img, 0, 0)
	//	c.ctx2d.Call("drawImage", ctx.canvas, sx, sy, sw, sh, dx, dy, dw, dh)
}
