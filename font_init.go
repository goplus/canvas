// +build !nofont !wx

package canvas

import "golang.org/x/image/font/gofont/goregular"

func init() {
	defaultFontDatebase.LoadCollectData(goregular.TTF)
	SetDefaultFont("Go", 10)
}
