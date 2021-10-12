// +build !js

package canvas

func init() {
	SetFontPaths("/Library/Fonts", "/System/Library/Fonts")
	err := PreloadFont("Arial Unicode MS", "Arial Unicode.ttf")
	if err == nil {
		fallbackRawFont = defaultFontDatebase.LoadRawFont(&Font{Family: "Arial Unicode MS", PointSize: 10})
	}
}
