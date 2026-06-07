// 从纯黑/近黑背景的 PNG 中抠出透明背景。
// 只从画布边缘连通区域取背景，避免误删人物衣服、头发和描边里的黑色。
package main

import (
	"flag"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"golang.org/x/image/draw"
)

type point struct {
	x, y int
}

func main() {
	in := flag.String("in", "", "输入 PNG")
	out := flag.String("out", "", "输出 PNG")
	size := flag.Int("size", 0, "可选：输出正方形尺寸，0 表示保持原尺寸")
	bgThreshold := flag.Int("bg-threshold", 18, "边缘黑底阈值，越大抠得越多")
	featherThreshold := flag.Int("feather-threshold", 42, "边缘柔化阈值，越大越容易去黑边")
	edgeRadius := flag.Int("edge-radius", 1, "边缘柔化半径")
	flag.Parse()

	if *in == "" || *out == "" {
		log.Fatal("必须提供 -in 和 -out")
	}
	if *bgThreshold < 0 || *bgThreshold > 255 || *featherThreshold < 0 || *featherThreshold > 255 {
		log.Fatal("阈值必须在 0..255")
	}
	if *edgeRadius < 0 || *edgeRadius > 4 {
		log.Fatal("-edge-radius 必须在 0..4")
	}

	img, err := loadPNG(*in)
	if err != nil {
		log.Fatal(err)
	}
	cutout(img, uint8(*bgThreshold), uint8(*featherThreshold), *edgeRadius)
	if *size > 0 {
		img = fitSquare(img, *size)
	}
	if err := writePNG(*out, img); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s", *out)
}

func loadPNG(path string) (*image.NRGBA, error) {
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
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
	return dst, nil
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	return enc.Encode(f, img)
}

func cutout(img *image.NRGBA, bgThreshold, featherThreshold uint8, edgeRadius int) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	bg := make([]bool, w*h)
	queue := make([]point, 0, w*2+h*2)

	push := func(x, y int) {
		if x < 0 || x >= w || y < 0 || y >= h {
			return
		}
		idx := y*w + x
		if bg[idx] || !isNearBlack(img.NRGBAAt(x, y), bgThreshold) {
			return
		}
		bg[idx] = true
		queue = append(queue, point{x: x, y: y})
	}

	for x := 0; x < w; x++ {
		push(x, 0)
		push(x, h-1)
	}
	for y := 0; y < h; y++ {
		push(0, y)
		push(w-1, y)
	}

	for head := 0; head < len(queue); head++ {
		p := queue[head]
		push(p.x+1, p.y)
		push(p.x-1, p.y)
		push(p.x, p.y+1)
		push(p.x, p.y-1)
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if bg[y*w+x] {
				setTransparent(img, x, y)
			}
		}
	}

	if edgeRadius == 0 {
		return
	}

	next := cloneNRGBA(img)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if bg[y*w+x] || !touchesBackground(bg, w, h, x, y, edgeRadius) {
				continue
			}
			c := img.NRGBAAt(x, y)
			if c.A == 0 || !isNearBlack(c, featherThreshold) {
				continue
			}
			m := max3(c.R, c.G, c.B)
			span := int(featherThreshold) - int(bgThreshold)
			if span <= 0 {
				span = 1
			}
			alphaScale := clampFloat(float64(int(m)-int(bgThreshold))/float64(span), 0.18, 1)
			i := next.PixOffset(x, y)
			next.Pix[i+3] = uint8(float64(next.Pix[i+3]) * alphaScale)
		}
	}
	*img = *next
}

func touchesBackground(bg []bool, w, h, x, y, radius int) bool {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			nx, ny := x+dx, y+dy
			if nx < 0 || nx >= w || ny < 0 || ny >= h {
				continue
			}
			if bg[ny*w+nx] {
				return true
			}
		}
	}
	return false
}

func isNearBlack(c color.NRGBA, threshold uint8) bool {
	if c.A == 0 {
		return true
	}
	m := max3(c.R, c.G, c.B)
	if m > threshold {
		return false
	}
	return int(max3(c.R, c.G, c.B))-int(min3(c.R, c.G, c.B)) <= max(8, int(threshold)/2)
}

func setTransparent(img *image.NRGBA, x, y int) {
	i := img.PixOffset(x, y)
	img.Pix[i+0] = 0
	img.Pix[i+1] = 0
	img.Pix[i+2] = 0
	img.Pix[i+3] = 0
}

func fitSquare(src *image.NRGBA, size int) *image.NRGBA {
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	scale := float64(size) / float64(max(sw, sh))
	dw := max(1, int(float64(sw)*scale+0.5))
	dh := max(1, int(float64(sh)*scale+0.5))
	scaled := image.NewNRGBA(image.Rect(0, 0, dw, dh))
	draw.CatmullRom.Scale(scaled, scaled.Bounds(), src, b, draw.Src, nil)

	out := image.NewNRGBA(image.Rect(0, 0, size, size))
	pt := image.Pt((size-dw)/2, (size-dh)/2)
	draw.Draw(out, image.Rectangle{Min: pt, Max: pt.Add(scaled.Bounds().Size())}, scaled, image.Point{}, draw.Src)
	return out
}

func cloneNRGBA(src *image.NRGBA) *image.NRGBA {
	dst := image.NewNRGBA(src.Bounds())
	copy(dst.Pix, src.Pix)
	return dst
}

func max3(a, b, c uint8) uint8 {
	if a < b {
		a = b
	}
	if a < c {
		a = c
	}
	return a
}

func min3(a, b, c uint8) uint8 {
	if a > b {
		a = b
	}
	if a > c {
		a = c
	}
	return a
}

func clampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
