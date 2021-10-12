//go:build !nofont
// +build !nofont

package canvas

import (
	"log"
	"unicode"

	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// func (p *Path) AddText(x, y float64, fnt Font, text string) float64 {
// 	f := fnt.Font()
// 	scale := fnt.PointSize() * 64
// 	startx := x
// 	prev, hasPrev := truetype.Index(0), false
// 	for _, r := range text {
// 		index := f.Index(r)
// 		if hasPrev {
// 			x += fUnitsToFloat64(f.Kern(fixed.Int26_6(scale), prev, index))
// 		}
// 		err := p.drawGlyph(f, index, x, y, scale)
// 		if err != nil {
// 			log.Println(err)
// 			return startx - x
// 		}
// 		x += fUnitsToFloat64(f.HMetric(fixed.Int26_6(scale), index).AdvanceWidth)
// 		prev, hasPrev = index, true
// 	}
// 	return x - startx
// }

// func (p *Path) drawGlyph(f *truetype.Font, glyph truetype.Index, dx, dy float64, scale int) error {
// 	if err := p.glyphBuf.Load(f, fixed.Int26_6(scale), glyph, font.HintingNone); err != nil {
// 		return err
// 	}
// 	e0 := 0
// 	for _, e1 := range p.glyphBuf.Ends {
// 		DrawContour(p, p.glyphBuf.Points[e0:e1], dx, dy)
// 		e0 = e1
// 	}
// 	return nil
// }

func (p *Path) drawSegments(segments []sfnt.Segment, dx, dy float64) {
	for _, seg := range segments {
		// The divisions by 64 below is because the seg.Args values have type
		// fixed.Int26_6, a 26.6 fixed point number, and 1<<6 == 64.
		switch seg.Op {
		case sfnt.SegmentOpMoveTo:
			p.MoveTo(
				dx+float64(seg.Args[0].X)/64,
				dy+float64(seg.Args[0].Y)/64,
			)
		case sfnt.SegmentOpLineTo:
			p.LineTo(
				dx+float64(seg.Args[0].X)/64,
				dy+float64(seg.Args[0].Y)/64,
			)
		case sfnt.SegmentOpQuadTo:
			p.QuadraticCurveTo(
				dx+float64(seg.Args[0].X)/64,
				dy+float64(seg.Args[0].Y)/64,
				dx+float64(seg.Args[1].X)/64,
				dy+float64(seg.Args[1].Y)/64,
			)
		case sfnt.SegmentOpCubeTo:
			p.BezierCurveTo(
				dx+float64(seg.Args[0].X)/64,
				dy+float64(seg.Args[0].Y)/64,
				dx+float64(seg.Args[1].X)/64,
				dy+float64(seg.Args[1].Y)/64,
				dx+float64(seg.Args[2].X)/64,
				dy+float64(seg.Args[2].Y)/64,
			)
		}
	}
}

func (p *Path) AddText(text string, x, y float64, fnt *Font) float64 {
	raw := defaultFontDatebase.LoadRawFont(fnt)
	if raw == nil {
		return 0
	}
	return p.AddTextByRaw(text, x, y, raw)
}

func (p *Path) AddTextByRaw(text string, x, y float64, raw *RawFont) float64 {
	return p.AddTextByFont(text, x, y, raw.Font, raw.PointSize)
}

func (p *Path) AddTextByFont(text string, x, y float64, f *sfnt.Font, pointSize int) float64 {
	startx := x
	var b sfnt.Buffer
	for _, r := range text {
		fnt := f
		i, err := fnt.GlyphIndex(&b, r)
		var fallback bool
		if i == 0 && err == nil && fallbackRawFont != nil {
			fnt = fallbackRawFont.Font
			fallback = true
			i, err = fnt.GlyphIndex(&b, r)
		}
		if err != nil {
			log.Printf("GlyphIndex: %v", err)
			break
		}
		segments, err := fnt.LoadGlyph(&b, i, fixed.I(pointSize), nil)
		if err != nil {
			log.Printf("LoadGlyph: %v", err)
			break
		}
		p.drawSegments(segments, x, y)
		v, _ := fnt.GlyphAdvance(&b, i, fixed.I(pointSize), font.HintingFull)
		offset := fUnitsToFloat64(v)
		//TODO fix 汉字计算不准确如 "试"
		if fallback && unicode.Is(unicode.Han, r) {
			if offset < float64(pointSize) {
				offset = float64(pointSize)
			}
		}
		x += offset
	}
	return x - startx
}

func (p *Path) MeasureText(text string, fnt *Font) float64 {
	raw := defaultFontDatebase.LoadRawFont(fnt)
	if raw == nil {
		return 0
	}
	return p.MeasureTextByRawFont(text, raw)
}

func (p *Path) MeasureTextByRawFont(text string, raw *RawFont) float64 {
	return p.MeasureTextByFont(text, raw.Font, raw.PointSize)
}

func (p *Path) MeasureTextByFont(text string, f *sfnt.Font, pointSize int) float64 {
	var x float64
	var b sfnt.Buffer
	for _, r := range text {
		ff := f
		var fallback bool
		i, err := ff.GlyphIndex(&b, r)
		if i == 0 && err == nil && fallbackRawFont != nil {
			ff = fallbackRawFont.Font
			fallback = true
			i, err = ff.GlyphIndex(&b, r)
		}
		if err != nil {
			log.Printf("GlyphIndex: %v", err)
			break
		}
		v, _ := ff.GlyphAdvance(&b, i, fixed.I(pointSize), font.HintingNone)
		offset := fUnitsToFloat64(v)
		//TODO fix 汉字计算不准确如 "试"
		if fallback && unicode.Is(unicode.Han, r) {
			if offset < float64(pointSize) {
				offset = float64(pointSize)
			}
		}
		x += offset
	}
	return x
}

func (p *Path) AddTextByFontProvider(text string, x, y float64, fp FontProvider) float64 {
	startx := x
	var b sfnt.Buffer
	for _, r := range text {
		i, raw, err := fp.GlyphIndex(&b, r)
		if err != nil {
			log.Printf("GlyphIndex: %v", err)
			break
		}
		segments, err := raw.Font.LoadGlyph(&b, i, fixed.I(raw.PointSize), nil)
		if err != nil {
			log.Printf("LoadGlyph: %v", err)
			break
		}
		p.drawSegments(segments, x, y)
		v, _ := raw.Font.GlyphAdvance(&b, i, fixed.I(raw.PointSize), font.HintingNone)
		x += fUnitsToFloat64(v)
	}
	return x - startx
}

func (p *Path) MetricsFont(f *Font) (*font.Metrics, error) {
	return defaultFontDatebase.MetricsFont(f)
}
