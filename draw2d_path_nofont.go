//go:build nofont
// +build nofont

package canvas

import (
	"fmt"

	"golang.org/x/image/font"
)

func (p *Path) AddText(text string, x, y float64, fnt *Font) float64 {
	return 0
}

func (p *Path) MeasureText(text string, fnt *Font) float64 {
	return 0
}

func (p *Path) MetricsFont(f *Font) (*font.Metrics, error) {
	return nil, fmt.Errorf("not support font")
}
