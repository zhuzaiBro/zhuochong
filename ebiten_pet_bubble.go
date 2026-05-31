//go:build ebitenpet

package main

import (
	"errors"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

// 右侧固定留出弹幕列宽度（与抖音弹幕长度上限相匹配），窗口总宽始终为 winW+bubbleColumnW，立绘区域恒为 winW×winH，不因气泡伸缩。
const (
	// 右侧弹幕列总宽：立绘右缘到窗口右缘（含间距、最大气泡宽、右内边距）。
	bubbleColumnW        = 236
	bubbleColumnRightPad = 8

	bubblePadV     = 10
	bubblePadH     = 12
	bubbleTextMaxW = 176 // 单列文字最大宽度（像素）

	bubbleFramePad   = 6
	bubbleShadowDX   = float32(2.5)
	bubbleShadowDY   = float32(3.5)
	bubbleCornerR    = float32(14)
	bubbleStrokeW    = float32(1.2)

	bubbleFontPoints = 12
	bubbleFontDPI    = 96
)

var (
	bubbleFace     font.Face
	bubbleFaceErr  error
	bubbleFaceOnce sync.Once
	bubbleTriWhite *ebiten.Image // 1×1 白贴图，供 vector 三角剖分上色
)

func init() {
	bubbleTriWhite = ebiten.NewImage(3, 3)
	b := bubbleTriWhite.Bounds()
	pix := make([]byte, 4*b.Dx()*b.Dy())
	for i := range pix {
		pix[i] = 0xff
	}
	bubbleTriWhite.WritePixels(pix)
}

func ensureBubbleFace() {
	bubbleFaceOnce.Do(func() {
		bubbleFace, bubbleFaceErr = loadBubbleFont()
		if bubbleFaceErr != nil {
			log.Println("弹幕气泡：未加载到字体（需中文字体），将不绘制气泡。", bubbleFaceErr)
		}
	})
}

func loadBubbleFont() (font.Face, error) {
	candidates := []string{
		strings.TrimSpace(os.Getenv("DOUYIN_MONITOR_BUBBLE_FONT")),
		"/System/Library/Fonts/STHeiti Light.ttc",
		"/Library/Fonts/Arial Unicode.ttf",
		"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/truetype/noto/NotoSansCJK-Regular.ttc",
	}
	for _, p := range candidates {
		if p == "" {
			continue
		}
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		low := strings.ToLower(filepath.Ext(p))
		switch low {
		case ".ttc":
			coll, err := opentype.ParseCollection(data)
			if err != nil {
				continue
			}
			n := coll.NumFonts()
			for i := 0; i < n; i++ {
				f, err := coll.Font(i)
				if err != nil {
					continue
				}
				face, err := opentype.NewFace(f, &opentype.FaceOptions{
					Size:    bubbleFontPoints,
					DPI:     bubbleFontDPI,
					Hinting: font.HintingNone,
				})
				if err != nil {
					continue
				}
				if _, _, ok := face.GlyphBounds('测'); ok {
					return face, nil
				}
				if i == n-1 {
					return face, nil
				}
				_ = face.Close()
			}
		case ".ttf", ".otf":
			f, err := opentype.Parse(data)
			if err != nil {
				continue
			}
			return opentype.NewFace(f, &opentype.FaceOptions{
				Size:    bubbleFontPoints,
				DPI:     bubbleFontDPI,
				Hinting: font.HintingNone,
			})
		default:
			continue
		}
	}
	return nil, errors.New("未找到支持中文的 .ttf/.otf/.ttc，可设置 DOUYIN_MONITOR_BUBBLE_FONT")
}

func bubbleWrap(face font.Face, s string, maxPx int) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var lines []string
	var cur strings.Builder
	for _, r := range s {
		test := cur.String() + string(r)
		w := font.MeasureString(face, test).Ceil()
		if w > maxPx && cur.Len() > 0 {
			lines = append(lines, cur.String())
			cur.Reset()
			cur.WriteRune(r)
			continue
		}
		cur.WriteRune(r)
	}
	if cur.Len() > 0 {
		lines = append(lines, cur.String())
	}
	return lines
}

func drawPathFill(dst *ebiten.Image, path *vector.Path, clr color.Color, antialias bool) {
	sub := bubbleTriWhite.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	r, g, b, a := clr.RGBA()
	rf := float32(r) / 0xffff
	gf := float32(g) / 0xffff
	bf := float32(b) / 0xffff
	af := float32(a) / 0xffff
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = rf * vs[i].ColorR
		vs[i].ColorG = gf * vs[i].ColorG
		vs[i].ColorB = bf * vs[i].ColorB
		vs[i].ColorA = af * vs[i].ColorA
	}
	op := &ebiten.DrawTrianglesOptions{
		ColorScaleMode: ebiten.ColorScaleModePremultipliedAlpha,
		AntiAlias:      antialias,
	}
	dst.DrawTriangles(vs, is, sub, op)
}

func drawPathStroke(dst *ebiten.Image, path *vector.Path, width float32, clr color.Color, antialias bool) {
	sub := bubbleTriWhite.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
	so := &vector.StrokeOptions{}
	so.Width = width
	so.MiterLimit = 4
	vs, is := path.AppendVerticesAndIndicesForStroke(nil, nil, so)
	r, g, b, a := clr.RGBA()
	rf := float32(r) / 0xffff
	gf := float32(g) / 0xffff
	bf := float32(b) / 0xffff
	af := float32(a) / 0xffff
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = rf * vs[i].ColorR
		vs[i].ColorG = gf * vs[i].ColorG
		vs[i].ColorB = bf * vs[i].ColorB
		vs[i].ColorA = af * vs[i].ColorA
	}
	op := &ebiten.DrawTrianglesOptions{
		ColorScaleMode: ebiten.ColorScaleModePremultipliedAlpha,
		AntiAlias:      antialias,
	}
	dst.DrawTriangles(vs, is, sub, op)
}

func appendRoundRect(p *vector.Path, x, y, w, h, r float32) {
	if w <= 0 || h <= 0 {
		return
	}
	if r <= 0 {
		p.MoveTo(x, y)
		p.LineTo(x+w, y)
		p.LineTo(x+w, y+h)
		p.LineTo(x, y+h)
		p.Close()
		return
	}
	if r > w/2 {
		r = w / 2
	}
	if r > h/2 {
		r = h / 2
	}
	p.MoveTo(x+r, y)
	p.LineTo(x+w-r, y)
	p.Arc(x+w-r, y+r, r, -float32(math.Pi)/2, 0, vector.Clockwise)
	p.LineTo(x+w, y+h-r)
	p.Arc(x+w-r, y+h-r, r, 0, float32(math.Pi)/2, vector.Clockwise)
	p.LineTo(x+r, y+h)
	p.Arc(x+r, y+h-r, r, float32(math.Pi)/2, float32(math.Pi), vector.Clockwise)
	p.LineTo(x, y+r)
	p.Arc(x+r, y+r, r, float32(math.Pi), 3*float32(math.Pi)/2, vector.Clockwise)
	p.Close()
}

// renderBubbleImage 圆角 + 阴影 + 柔和描边 + 正文。
func renderBubbleImage(face font.Face, line string) *ebiten.Image {
	lines := bubbleWrap(face, line, bubbleTextMaxW)
	if len(lines) == 0 {
		return nil
	}
	m := face.Metrics()
	lineH := m.Height.Ceil()
	if lineH < 14 {
		lineH = 14
	}
	maxLineW := 0
	for _, ln := range lines {
		xw := font.MeasureString(face, ln).Ceil()
		if xw > maxLineW {
			maxLineW = xw
		}
	}
	tw := maxLineW
	if tw > bubbleTextMaxW {
		tw = bubbleTextMaxW
	}
	textBlockW := tw
	textBlockH := len(lines)*lineH + (len(lines)-1)*2

	bodyW := float32(textBlockW + 2*bubblePadH)
	bodyH := float32(textBlockH + 2*bubblePadV)
	fp := float32(bubbleFramePad)
	imgW := int(math.Ceil(float64(bodyW + 2*fp + bubbleShadowDX)))
	imgH := int(math.Ceil(float64(bodyH + 2*fp + bubbleShadowDY)))

	dst := ebiten.NewImage(imgW, imgH)
	dst.Fill(color.NRGBA{A: 0x00})

	bx := fp
	by := fp

	// 阴影（略偏移的圆角块）
	var shadowPath vector.Path
	appendRoundRect(&shadowPath, bx+bubbleShadowDX, by+bubbleShadowDY, bodyW, bodyH, bubbleCornerR)
	drawPathFill(dst, &shadowPath, color.NRGBA{R: 0x12, G: 0x1a, B: 0x2e, A: 0x34}, true)

	// 渐变感底色：略偏暖白
	var fillPath vector.Path
	appendRoundRect(&fillPath, bx, by, bodyW, bodyH, bubbleCornerR)
	drawPathFill(dst, &fillPath, color.NRGBA{R: 0xfc, G: 0xfd, B: 0xff, A: 0xf5}, true)

	// 内沿细亮线（上半部略亮）
	var innerPath vector.Path
	inset := float32(1.25)
	appendRoundRect(&innerPath, bx+inset, by+inset, bodyW-2*inset, bodyH-2*inset, max(bubbleCornerR-inset, float32(0)))
	drawPathStroke(dst, &innerPath, 0.75, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x55}, true)

	// 外描边
	var strokePath vector.Path
	appendRoundRect(&strokePath, bx+0.5, by+0.5, bodyW-1, bodyH-1, max(bubbleCornerR-float32(0.5), float32(0)))
	drawPathStroke(dst, &strokePath, bubbleStrokeW, color.NRGBA{R: 0x7a, G: 0x8e, B: 0xba, A: 0xf0}, true)

	txtCol := color.NRGBA{R: 0x2c, G: 0x36, B: 0x48, A: 0xff}
	ascent := m.Ascent.Ceil()
	tx := int(bx) + bubblePadH
	y := int(by) + bubblePadV + ascent
	for _, ln := range lines {
		text.Draw(dst, ln, face, tx, y, txtCol)
		y += lineH + 2
	}
	return dst
}

func clampBubbleLine(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return s
	}
	n := utf8.RuneCountInString(s)
	if n <= maxRunes {
		return s
	}
	r := []rune(s)
	return string(r[:maxRunes]) + "…"
}
