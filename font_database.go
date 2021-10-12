// +build !nofont !wx

package canvas

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

//https://www.w3.org/html/ig/zh/wiki/CSS3字体模块
//https://www.cnblogs.com/starof/p/4562514.html

type cacheInfo struct {
	Family string
	Weight font.Weight
	Style  font.Style
}

type fontDatabase struct {
	fontMap       map[string]*FontFamily
	fontCache     map[cacheInfo]*rawFont
	fontLookupDir []string
}

func (db *fontDatabase) SetLookupDirs(paths ...string) {
	db.fontLookupDir = append(db.fontLookupDir, paths...)
}

func (db *fontDatabase) FamilyNames() []string {
	var names []string
	for k, _ := range db.fontMap {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func (db *fontDatabase) Family(name string) *FontFamily {
	if ff, ok := db.fontMap[name]; ok {
		return ff
	}
	for _, v := range db.fontMap {
		if v.FileName == name {
			return v
		}
	}
	return nil
}

func (db *fontDatabase) MetricsFont(f *Font) (*font.Metrics, error) {
	raw := db.LoadRawFont(f)
	var b sfnt.Buffer
	m, err := raw.Font.Metrics(&b, fixed.I(f.PointSize), font.HintingNone)
	return &m, err
}

func (db *fontDatabase) LoadRawFont(f *Font) *RawFont {
	if f == nil {
		f = defaultFont
	}
	if f.Family == "" {
		f.Family = defaultFontFamily.Family
	}
	cache := cacheInfo{Family: f.Family, Weight: f.Weight, Style: f.Style}
	if raw, ok := db.fontCache[cache]; ok {
		return &RawFont{raw, f.PointSize}
	}
	var ff *FontFamily
	for _, family := range strings.Split(f.Family, ",") {
		family = strings.TrimSpace(family)
		ff = db.fontMap[family]
		if ff != nil {
			break
		}
		for _, v := range db.fontMap {
			if v.Family == family || v.FileName == family {
				ff = v
				break
			}
		}
	}

	if ff == nil {
		ff = defaultFontFamily
	}
	if ff == nil {
		return nil
	}
	raw := ff.LoadRawFont(f.Style, f.Weight)
	db.fontCache[cache] = raw

	return &RawFont{raw, f.PointSize}
}

var (
	defaultFontDatebase = &fontDatabase{
		fontMap:   make(map[string]*FontFamily),
		fontCache: make(map[cacheInfo]*rawFont)}
	defaultFontFamily *FontFamily
	defaultRawFont    *RawFont
	fallbackRawFont   *RawFont
)

func SetDefaultFont(family string, pointSize int) {
	defaultFont = &Font{Family: family, PointSize: pointSize}
	defaultFontFamily = defaultFontDatebase.Family(family)
	defaultRawFont = defaultFontDatebase.LoadRawFont(defaultFont)
}

func PreloadFont(family string, fpath ...string) (err error) {
	return defaultFontDatebase.PreloadFont(family, fpath...)
}

func SetFontPaths(paths ...string) {
	defaultFontDatebase.SetLookupDirs(paths...)
}

func FontDatabase() *fontDatabase {
	return defaultFontDatebase
}

type rawFont struct {
	Path      string
	FullName  string
	Family    string
	SubFamily string
	Weight    font.Weight
	Style     font.Style
	Stretch   font.Stretch
	Font      *sfnt.Font
}

type RawFont struct {
	*rawFont
	PointSize int
}

func NewRawFont(r *rawFont, pointSize int) *RawFont {
	return &RawFont{r, pointSize}
}

func (r *rawFont) TestSupportText(text string) bool {
	var b sfnt.Buffer
	for _, t := range text {
		index, err := r.Font.GlyphIndex(&b, t)
		if err != nil {
			return false
		}
		if index == 0 {
			return false
		}
	}
	return true
}

func (r *rawFont) LoadFont(fnt *sfnt.Font) error {
	r.Font = fnt
	var b sfnt.Buffer
	family, err := fnt.Name(&b, sfnt.NameIDFamily)
	if err != nil {
		return err
	}
	r.Family = family
	subfamily, err := fnt.Name(&b, sfnt.NameIDSubfamily)
	if err != nil {
		return err
	}
	r.SubFamily = subfamily
	r.parserSubfamily(subfamily)
	fullname, err := fnt.Name(&b, sfnt.NameIDFull)
	if err == nil {
		r.FullName = fullname
	} else if r.FullName == "" {
		r.FullName = r.Family + " " + r.SubFamily
	}
	return nil
}

func IsLatin1(text string) bool {
	for _, v := range text {
		if v > unicode.MaxLatin1 {
			return false
		}
	}
	return true
}

func (r *rawFont) LoadData(data []byte) error {
	fnt, err := sfnt.Parse(data)
	if err != nil {
		return err
	}
	return r.LoadFont(fnt)
}

func (r *rawFont) LoadPath(fpath string) error {
	r.Path = fpath
	name := filepath.Base(fpath)
	ext := filepath.Ext(name)
	r.FullName = name[:len(name)-len(ext)]
	r.parserFamily(r.FullName)
	return nil
}

func (r *rawFont) ParseFont() error {
	if r.Font != nil {
		return nil
	}
	if r.Path == "" {
		return os.ErrInvalid
	}
	read, err := os.Open(r.Path)
	if err != nil {
		return err
	}
	fnt, err := sfnt.ParseReaderAt(read)
	if err != nil {
		return err
	}
	r.LoadFont(fnt)
	return nil
}

func (r *rawFont) parserFamily(name string) {
	pos := strings.Index(name, "-")
	if pos != -1 {
		r.Family = name[:pos]
		r.SubFamily = name[pos+1:]
		r.parserSubfamily(r.SubFamily)
		return
	}
	checkList := []string{
		" Bold Italic",
		" Bold",
		" Italic",
		"Bol",
		"Reg",
		"BolIta",
		"Italic",
	}
	r.Family = name
	r.SubFamily = ""
	for _, check := range checkList {
		if strings.HasSuffix(name, check) {
			r.Family = name[:len(name)-len(check)]
			r.SubFamily = strings.TrimSpace(check)
			r.parserSubfamily(r.SubFamily)
			return
		}
	}
}

func (r *rawFont) parserSubfamily(sub string) {
	var ar []string
	if strings.Contains(sub, " ") {
		ar = strings.Split(sub, " ")
	} else {
		ar = splitUpper(sub, []string{"Extra", "Ultra"})
	}
	for _, a := range ar {
		switch a {
		case "Regular", "Book", "Roman", "Normal", "Reg", "Plain", "Text":
			r.Weight = font.WeightNormal
		case "Bold", "Bol", "Bd":
			r.Weight = font.WeightBold
		case "Thin":
			r.Weight = font.WeightThin
		case "Light":
			r.Weight = font.WeightLight
		case "ExtraLight", "UltraLight", "Ultralight":
			r.Weight = font.WeightExtraLight
		case "Medium":
			r.Weight = font.WeightMedium
		case "Semibold", "Semi", "SemiBold":
			r.Weight = font.WeightSemiBold
		case "ExtraBold", "ExtraBlack", "UltraBold":
			r.Weight = font.WeightExtraBold
		case "Black", "Heavy":
			r.Weight = font.WeightBlack
		case "Italic", "Ita", "It", "Slanted":
			r.Style = font.StyleItalic
		case "Oblique", "Obl", "Inclined":
			r.Style = font.StyleOblique
		case "Cond", "Condensed", "Cn":
			r.Stretch = font.StretchCondensed
		case "Ext":
			r.Stretch = font.StretchExtraExpanded
			// default:
			//	log.Printf("Unknown Subfamily: %v (%v)\n", a, r.Family)
		}
	}
}

type FontFamily struct {
	FileName   string
	Family     string
	Collect    string
	RawFontMap map[string]*rawFont
}

func NewFontFamily(filename string, family string) *FontFamily {
	return &FontFamily{FileName: filename, Family: family, RawFontMap: make(map[string]*rawFont)}
}

func (ff *FontFamily) loadCollect(fpath string) error {
	r, err := os.Open(fpath)
	if err != nil {
		return err
	}
	c, err := sfnt.ParseCollectionReaderAt(r)
	if err != nil {
		return err
	}
	ff.RawFontMap = make(map[string]*rawFont)
	for i := 0; i < c.NumFonts(); i++ {
		fnt, err := c.Font(i)
		if err != nil {
			continue
		}
		raw := &rawFont{}
		err = raw.LoadFont(fnt)
		if err != nil {
			log.Printf("LoadFont: %v\n", err)
			continue
		}
		ff.RawFontMap[raw.FullName] = raw
		ff.Family = raw.Family
	}
	return nil
}

type checkFont struct {
	weight font.Weight
	style  font.Style
}

func (ff *FontFamily) checkRawFontList(chks ...*checkFont) *rawFont {
	for _, chk := range chks {
		raw := ff.checkRawFont(chk)
		if raw != nil {
			return raw
		}
	}
	return nil
}

func (ff *FontFamily) checkRawFont(chk *checkFont) *rawFont {
	for _, raw := range ff.RawFontMap {
		if raw.Weight == chk.weight && raw.Style == chk.style {
			return raw
		}
	}
	return nil
}

func (ff *FontFamily) LoadRawFont(style font.Style, weight font.Weight) *rawFont {
	if ff.RawFontMap == nil {
		if ff.Collect == "" {
			return nil
		}
		err := ff.loadCollect(ff.Collect)
		if err != nil {
			log.Printf("LoadCollect: %v\n", err)
			return nil
		}
	}
	var raw *rawFont
	if style == font.StyleItalic {

	}
	if style != font.StyleNormal && weight != font.WeightNormal {
		raw = ff.checkRawFontList(
			&checkFont{weight, style},
			&checkFont{font.WeightBold, font.StyleItalic},
			&checkFont{font.WeightBold, font.StyleOblique},
			&checkFont{font.WeightSemiBold, font.StyleItalic},
			&checkFont{font.WeightSemiBold, font.StyleOblique},
			&checkFont{font.WeightBold, font.StyleNormal},
			&checkFont{font.WeightBlack, font.StyleNormal},
			&checkFont{font.WeightNormal, font.StyleNormal},
		)
	} else if weight != font.WeightNormal {
		raw = ff.checkRawFontList(
			&checkFont{weight, font.StyleNormal},
			&checkFont{font.WeightBold, font.StyleNormal},
			&checkFont{font.WeightSemiBold, font.StyleNormal},
			&checkFont{font.WeightBlack, font.StyleNormal},
			&checkFont{font.WeightNormal, font.StyleNormal},
		)
	} else if style != font.StyleNormal {
		raw = ff.checkRawFontList(
			&checkFont{weight, font.StyleNormal},
			&checkFont{font.WeightNormal, font.StyleItalic},
			&checkFont{font.WeightNormal, font.StyleOblique},
			&checkFont{font.WeightBold, font.StyleItalic},
			&checkFont{font.WeightNormal, font.StyleNormal},
		)
	} else {
		raw = ff.checkRawFontList(
			&checkFont{font.WeightNormal, font.StyleNormal},
			&checkFont{font.WeightBold, font.StyleNormal},
			&checkFont{font.WeightMedium, font.StyleNormal},
			&checkFont{font.WeightLight, font.StyleNormal},
		)
	}
	if raw == nil {
		var names []string
		for k, _ := range ff.RawFontMap {
			names = append(names, k)
		}
		if len(names) == 0 {
			return nil
		}
		sort.Strings(names)
		raw = ff.RawFontMap[names[0]]
	}
	err := raw.ParseFont()
	if err != nil {
		return nil
	}
	return raw
}

func contains(s string, ar []string) bool {
	for _, v := range ar {
		if s == v {
			return true
		}
	}
	return false
}

func splitUpper(s string, skips []string) []string {
	var ar []string
	start := -1
	for n, v := range s {
		if unicode.IsUpper(v) {
			if start != -1 {
				sub := s[start:n]
				if contains(sub, skips) {
					continue
				}
				ar = append(ar, s[start:n])
			}
			start = n
		}
	}
	if start != -1 {
		ar = append(ar, s[start:])
	} else {
		ar = append(ar, s)
	}
	return ar
}

func (db *fontDatabase) LoadCollectFile(fpath string) error {
	r, err := os.Open(fpath)
	if err != nil {
		return err
	}
	c, err := sfnt.ParseCollectionReaderAt(r)
	if err != nil {
		return err
	}
	return db.loadCollect(c)
}

func (db *fontDatabase) loadCollect(c *sfnt.Collection) error {
	for i := 0; i < c.NumFonts(); i++ {
		fnt, err := c.Font(i)
		if err != nil {
			continue
		}
		raw := &rawFont{}
		err = raw.LoadFont(fnt)
		if err != nil {
			log.Printf("LoadFont: %v\n", err)
			continue
		}
		f, ok := db.fontMap[raw.Family]
		if !ok {
			f = NewFontFamily(raw.Family, raw.Family)
			db.fontMap[raw.Family] = f
		}
		f.RawFontMap[raw.FullName] = raw
	}
	return nil
}

func (db *fontDatabase) LoadCollectData(data []byte) error {
	c, err := sfnt.ParseCollection(data)
	if err != nil {
		return err
	}
	return db.loadCollect(c)
}

func (db *fontDatabase) LoadFontData(data []byte) error {
	fnt, err := sfnt.Parse(data)
	if err != nil {
		return err
	}
	raw := &rawFont{}
	err = raw.LoadFont(fnt)
	if err != nil {
		return err
	}
	f, ok := db.fontMap[raw.Family]
	if !ok {
		f = NewFontFamily(raw.Family, raw.Family)
		db.fontMap[raw.Family] = f
	}
	f.RawFontMap[raw.FullName] = raw
	return nil
}

func (db *fontDatabase) PreloadFont(family string, fpath ...string) (err error) {
	for _, f := range fpath {
		if filepath.IsAbs(f) {
			err = db.loadFontFile(family, f, true)
			if err == nil {
				return
			}
		} else {
			for _, dir := range db.fontLookupDir {
				err = db.loadFontFile(family, filepath.Join(dir, f), true)
				if err == nil {
					return
				}
			}
		}
	}
	return fmt.Errorf("not find font %q", family)
}

func (db *fontDatabase) PreloadFonts(fpath ...string) {
	for _, f := range fpath {
		if filepath.IsAbs(f) {
			db.loadFontFile("", f, true)
		} else {
			for _, dir := range db.fontLookupDir {
				if db.loadFontFile("", filepath.Join(dir, f), true) == nil {
					break
				}
			}
		}
	}
}

func (db *fontDatabase) LoadFontFile(path string) error {
	return db.loadFontFile("", path, true)
}

func (db *fontDatabase) loadFontFile(fname string, path string, parse bool) error {
	name := filepath.Base(path)
	ext := filepath.Ext(name)
	raw := &rawFont{}
	raw.LoadPath(path)
	if parse {
		err := raw.ParseFont()
		if err != nil {
			return fmt.Errorf("ParseFont: %v, %v", name, err)
		}
	}
	f, ok := db.fontMap[raw.Family]
	if !ok {
		f = NewFontFamily(name[:len(name)-len(ext)], raw.Family)
		db.fontMap[raw.Family] = f
	}
	f.RawFontMap[raw.FullName] = raw
	if fname != "" {
		db.fontMap[fname] = f
	}
	return nil
}

func (db *fontDatabase) LoadFontDir(root string, parse bool) error {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		name := filepath.Base(path)
		ext := filepath.Ext(name)
		switch ext {
		case ".ttc":
			if parse {
				err := db.LoadCollectFile(path)
				if err != nil {
					log.Println(err)
				}
			} else {
				name = name[:len(name)-4]
				fi := &FontFamily{FileName: name, Family: name, Collect: path}
				db.fontMap[fi.Family] = fi
			}
		case ".ttf", ".otf":
			err := db.loadFontFile("", path, parse)
			if err != nil {
				log.Println(err)
			}
		}
		return nil
	})
	return nil
}

type FontProvider interface {
	GlyphIndex(b *sfnt.Buffer, r rune) (sfnt.GlyphIndex, *RawFont, error)
}

type TableFontProvider struct {
	Tables  []*unicode.RangeTable
	Backup  *RawFont
	FontMap map[*unicode.RangeTable]*RawFont
}

func NewTableFontProvider(tabs ...*unicode.RangeTable) *TableFontProvider {
	return &TableFontProvider{tabs, defaultRawFont, make(map[*unicode.RangeTable]*RawFont)}
}

func (p *TableFontProvider) SetBackupRawFont(raw *RawFont) {
	p.Backup = raw
}

func (p *TableFontProvider) SetBackupFont(f *Font) {
	raw := defaultFontDatebase.LoadRawFont(f)
	if raw == nil {
		return
	}
	p.SetBackupRawFont(raw)
}

func (p *TableFontProvider) SetFontRange(t *unicode.RangeTable, f *Font) {
	raw := defaultFontDatebase.LoadRawFont(f)
	if raw == nil {
		return
	}
	p.FontMap[t] = raw
}

func (p *TableFontProvider) SetRawFontRange(t *unicode.RangeTable, raw *RawFont) {
	p.FontMap[t] = raw
}

func (p *TableFontProvider) GlyphIndex(b *sfnt.Buffer, r rune) (sfnt.GlyphIndex, *RawFont, error) {
	for _, tab := range p.Tables {
		if unicode.In(r, tab) {
			if f, ok := p.FontMap[tab]; ok {
				i, err := f.Font.GlyphIndex(b, r)
				if i != 0 && err == nil {
					return i, f, nil
				}
			}
			break
		}
	}
	i, err := p.Backup.Font.GlyphIndex(b, r)
	return i, p.Backup, err
}

type CJKFontProvider struct {
	Latin  *RawFont
	Han    *RawFont
	Hangul *RawFont
	Backup *RawFont
}

func (p *CJKFontProvider) GlyphIndex(b *sfnt.Buffer, r rune) (sfnt.GlyphIndex, *RawFont, error) {
	if r <= unicode.MaxLatin1 {
		i, err := p.Latin.Font.GlyphIndex(b, r)
		if i != 0 && err == nil {
			return i, p.Latin, nil
		}
	} else {
		i, err := p.Han.Font.GlyphIndex(b, r)
		if i != 0 && err == nil {
			return i, p.Han, nil
		}
		if unicode.In(r, unicode.Hangul) {
			i, err := p.Hangul.Font.GlyphIndex(b, r)
			if i != 0 && err == nil {
				return i, p.Hangul, nil
			}
		}
	}
	i, err := p.Backup.Font.GlyphIndex(b, r)
	return i, p.Backup, err
}
