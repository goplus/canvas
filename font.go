package canvas

import (
	"fmt"
	"log"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var (
	defaultFont = &Font{Family: "Go", PointSize: 10}
)

// font family example
// font family: Helvetica, Verdana, sans-serif
// If Helvetica is available it will be used when rendering.
// If neither Helvetica nor Verdana is present,
// then the user-agent-defined sans serif font will be used.

type Font struct {
	Family    string       // font name list split by ,
	PointSize int          // font point size
	Style     font.Style   // default font.StyleNormal
	Weight    font.Weight  // default font.WeightNormal
	Stretch   font.Stretch // default font.StretchNormal
}

func (f Font) String() string {
	var ar []string
	if f.Style == font.StyleItalic {
		ar = append(ar, "italic")
	} else if f.Style == font.StyleOblique {
		ar = append(ar, "oblique")
	}
	if f.Weight == font.WeightNormal {
		ar = append(ar, "normal")
	} else if f.Weight == font.WeightBold {
		ar = append(ar, "bold")
	} else {
		ar = append(ar, fmt.Sprintf("%v", f.Weight*100+400))
	}
	if f.PointSize != 0 {
		ar = append(ar, fmt.Sprintf("%vpx", f.PointSize))
	}
	if f.Family != "" {
		ar = append(ar, f.Family)
	}
	return strings.Join(ar, " ")
}

func NewFont(family string, size int) *Font {
	f := &Font{Family: family, PointSize: size}
	return f
}

func (f *Font) SetSize(size int) {
	f.PointSize = size
}

func (f *Font) SetBold(bold bool) {
	if bold {
		f.Weight = font.WeightBold
	} else {
		f.Weight = font.WeightNormal
	}
}

func (f *Font) IsBold() bool {
	return f.Weight >= font.WeightBold
}

func (f *Font) SetItalic(italic bool) {
	if italic {
		f.Style = font.StyleItalic
	} else {
		f.Style = font.StyleNormal
	}
}

func (f *Font) IsItalic() bool {
	return f.Style != font.StyleNormal
}

func fUnitsToFloat64(x fixed.Int26_6) float64 {
	scaled := x << 2
	return float64(scaled/256) + float64(scaled%256)/256.0
}

func saneSize(oldSize, newSize int) int {
	if newSize < 5 || newSize > 144 {
		log.Printf("font.saneSize(): ignored invalid size '%d'", newSize)
		return oldSize
	}
	return newSize
}
