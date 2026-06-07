// 精确编辑 1-1 办公关键帧：桌子改深玫粉，电脑盾牌改粉色小兔子。
// 运行：go run ./scripts/edit_office_keyframe
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
)

const target = "lucheng-keyframes-draft/1-1_office_executive.png"

func main() {
	img, err := load(target)
	if err != nil {
		log.Fatal(err)
	}
	recolorDesk(img)
	drawLaptopBunny(img)
	if err := save(target, img); err != nil {
		log.Fatal(err)
	}
	log.Printf("updated %s", target)
}

func load(path string) (*image.NRGBA, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	src, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	b := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			dst.Set(x, y, src.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return dst, nil
}

func save(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	return enc.Encode(f, img)
}

func recolorDesk(img *image.NRGBA) {
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			if !inDeskArea(x, y) {
				continue
			}
			i := img.PixOffset(x, y)
			a := img.Pix[i+3]
			if a == 0 {
				continue
			}
			r, g, b := img.Pix[i+0], img.Pix[i+1], img.Pix[i+2]
			if !isDeskWood(r, g, b) {
				continue
			}
			nr, ng, nb := rosewood(r, g, b)
			img.Pix[i+0] = nr
			img.Pix[i+1] = ng
			img.Pix[i+2] = nb
		}
	}
}

func inDeskArea(x, y int) bool {
	if y < 625 || y > 1190 {
		return false
	}
	areas := [][]image.Point{
		{{58, 630}, {1198, 630}, {1228, 820}, {28, 820}},
		{{72, 790}, {1190, 790}, {1140, 1158}, {88, 1158}},
		{{84, 1075}, {225, 1075}, {225, 1188}, {76, 1188}},
		{{1018, 1075}, {1170, 1075}, {1178, 1188}, {1010, 1188}},
	}
	p := image.Pt(x, y)
	for _, area := range areas {
		if pointInPoly(p, area) {
			return true
		}
	}
	return false
}

func isDeskWood(r, g, b uint8) bool {
	rr, gg, bb := float64(r), float64(g), float64(b)
	maxc := math.Max(rr, math.Max(gg, bb))
	minc := math.Min(rr, math.Min(gg, bb))
	if maxc < 18 || maxc > 205 {
		return false
	}
	if maxc-minc < 10 {
		return false
	}
	// Keep brass/gold trim and laptop/black accessories.
	if rr > 125 && gg > 82 && bb < 80 && rr-gg < 95 {
		return false
	}
	h, s, l := rgbToHsl(rr/255, gg/255, bb/255)
	return l < 0.56 && s > 0.16 && (h <= 44 || h >= 350)
}

func rosewood(r, g, b uint8) (uint8, uint8, uint8) {
	_, s, l := rgbToHsl(float64(r)/255, float64(g)/255, float64(b)/255)
	l = clamp(l*0.92+0.035, 0.08, 0.54)
	s = clamp(math.Max(s*0.78, 0.42), 0.35, 0.68)
	return hslToRgb(342, s, l)
}

func drawLaptopBunny(img *image.NRGBA) {
	// Cover the old shield with a laptop-colored patch, then draw a tiny pink bunny emblem.
	fillEllipse(img, 748, 633, 24, 18, rgba(47, 47, 57, 235))
	fillEllipse(img, 740, 629, 6, 16, rgba(255, 151, 190, 245))
	fillEllipse(img, 756, 629, 6, 16, rgba(255, 151, 190, 245))
	fillEllipse(img, 748, 641, 15, 13, rgba(255, 151, 190, 245))
	fillEllipse(img, 743, 638, 3, 3, rgba(74, 42, 58, 230))
	fillEllipse(img, 753, 638, 3, 3, rgba(74, 42, 58, 230))
	fillEllipse(img, 748, 645, 3, 2, rgba(74, 42, 58, 230))
	strokeEllipse(img, 748, 641, 15, 13, rgba(255, 212, 228, 210), 1)
}

func pointInPoly(p image.Point, points []image.Point) bool {
	inside := false
	j := len(points) - 1
	for i := range points {
		pi, pj := points[i], points[j]
		if (pi.Y > p.Y) != (pj.Y > p.Y) &&
			p.X < (pj.X-pi.X)*(p.Y-pi.Y)/(pj.Y-pi.Y)+pi.X {
			inside = !inside
		}
		j = i
	}
	return inside
}

func fillEllipse(img *image.NRGBA, cx, cy, rx, ry int, c color.NRGBA) {
	for y := cy - ry; y <= cy+ry; y++ {
		for x := cx - rx; x <= cx+rx; x++ {
			dx := float64(x-cx) / float64(rx)
			dy := float64(y-cy) / float64(ry)
			if dx*dx+dy*dy <= 1 {
				blend(img, x, y, c)
			}
		}
	}
}

func strokeEllipse(img *image.NRGBA, cx, cy, rx, ry int, c color.NRGBA, width int) {
	for y := cy - ry - width; y <= cy+ry+width; y++ {
		for x := cx - rx - width; x <= cx+rx+width; x++ {
			dx := float64(x-cx) / float64(rx)
			dy := float64(y-cy) / float64(ry)
			v := dx*dx + dy*dy
			if v >= 1-float64(width)/float64(max(rx, ry)) && v <= 1+float64(width)/float64(max(rx, ry)) {
				blend(img, x, y, c)
			}
		}
	}
}

func blend(img *image.NRGBA, x, y int, c color.NRGBA) {
	if !image.Pt(x, y).In(img.Bounds()) || c.A == 0 {
		return
	}
	i := img.PixOffset(x, y)
	a := int(c.A)
	ia := 255 - a
	img.Pix[i+0] = uint8((int(c.R)*a + int(img.Pix[i+0])*ia) / 255)
	img.Pix[i+1] = uint8((int(c.G)*a + int(img.Pix[i+1])*ia) / 255)
	img.Pix[i+2] = uint8((int(c.B)*a + int(img.Pix[i+2])*ia) / 255)
	img.Pix[i+3] = uint8(min(255, a+int(img.Pix[i+3])*ia/255))
}

func rgba(r, g, b, a uint8) color.NRGBA {
	return color.NRGBA{R: r, G: g, B: b, A: a}
}

func rgbToHsl(r, g, b float64) (float64, float64, float64) {
	maxc := math.Max(r, math.Max(g, b))
	minc := math.Min(r, math.Min(g, b))
	l := (maxc + minc) / 2
	if maxc == minc {
		return 0, 0, l
	}
	d := maxc - minc
	s := d / (1 - math.Abs(2*l-1))
	var h float64
	switch maxc {
	case r:
		h = math.Mod((g-b)/d, 6)
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	h *= 60
	if h < 0 {
		h += 360
	}
	return h, s, l
}

func hslToRgb(h, s, l float64) (uint8, uint8, uint8) {
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2
	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	return uint8(clamp((r+m)*255, 0, 255)), uint8(clamp((g+m)*255, 0, 255)), uint8(clamp((b+m)*255, 0, 255))
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
