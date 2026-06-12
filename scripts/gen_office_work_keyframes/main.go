// 从办公状态主图生成轻量工作循环关键帧。
// 运行：go run ./scripts/gen_office_work_keyframes
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
)

type motion struct {
	headDX, headDY int
	handDX, handDY int
}

var workLoop = []motion{
	{},
	{headDY: -1, handDX: -1},
	{headDY: -2, handDX: 1},
	{headDX: 1, headDY: -1, handDX: -1},
	{},
	{headDY: 1, handDX: 1},
	{headDY: 2, handDX: -1},
	{headDX: -1, headDY: 1, handDX: 1},
}

func main() {
	dir := flag.String("dir", "lucheng-sprites/office", "办公状态素材目录")
	source := flag.String("source", "frame1.png", "源图文件名")
	frames := flag.Int("frames", 8, "输出帧数")
	flag.Parse()

	if *frames < 2 {
		log.Fatal("-frames 必须大于等于 2")
	}
	srcPath := filepath.Join(*dir, *source)
	src, err := loadPNG(srcPath)
	if err != nil {
		log.Fatal(err)
	}
	for i := 1; i <= *frames; i++ {
		img := clone(src)
		m := workLoop[(i-1)%len(workLoop)]
		movePerson(src, img, m)
		phase := float64(i-1) / float64(*frames)
		addCoffeeSteam(img, phase, i)
		outPath := filepath.Join(*dir, fmt.Sprintf("frame%d.png", i))
		if err := writePNG(outPath, img); err != nil {
			log.Fatal(err)
		}
		log.Printf("写 %s", outPath)
	}
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
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			dst.Set(x, y, src.At(b.Min.X+x, b.Min.Y+y))
		}
	}
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

func clone(src *image.NRGBA) *image.NRGBA {
	dst := image.NewNRGBA(src.Bounds())
	copy(dst.Pix, src.Pix)
	return dst
}

func movePerson(src, dst *image.NRGBA, m motion) {
	if m == (motion{}) {
		return
	}
	moveMasked(src, dst, isHeadPixel, m.headDX, m.headDY)
	moveMasked(src, dst, isTypingHandPixel, m.handDX, m.handDY)
}

func moveMasked(src, dst *image.NRGBA, mask func(int, int, color.NRGBA) bool, dx, dy int) {
	if dx == 0 && dy == 0 {
		return
	}
	b := src.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := src.NRGBAAt(x, y)
			if !mask(x, y, c) {
				continue
			}
			c.A = uint8(int(c.A) * 92 / 255)
			blend(dst, x+dx, y+dy, c)
		}
	}
}

func isHeadPixel(x, y int, c color.NRGBA) bool {
	if c.A < 24 || y < 28 || y > 468 {
		return false
	}
	return inEllipse(x, y, 612, 254, 248, 230)
}

func isTypingHandPixel(x, y int, c color.NRGBA) bool {
	if c.A < 24 || y < 595 || y > 695 || x < 520 || x > 635 {
		return false
	}
	return int(c.R)+int(c.G)+int(c.B) < 165 && inEllipse(x, y, 575, 646, 58, 46)
}

func inEllipse(x, y, cx, cy, rx, ry int) bool {
	dx := float64(x-cx) / float64(rx)
	dy := float64(y-cy) / float64(ry)
	return dx*dx+dy*dy <= 1
}

func addCoffeeSteam(img *image.NRGBA, phase float64, frame int) {
	if frame == 1 {
		return
	}
	alpha := uint8(46 + 28*math.Sin(phase*math.Pi))
	offset := -int(phase*10 + 0.5)
	drawWisp(img, 920, 618+offset, 10, 28, phase*math.Pi*2, rgba(255, 248, 237, alpha))
	drawWisp(img, 941, 612+offset/2, 8, 24, phase*math.Pi*2+1.2, rgba(255, 248, 237, alpha-12))
}

func drawWisp(img *image.NRGBA, cx, cy, amp, height int, phase float64, c color.NRGBA) {
	steps := height * 3
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := float64(cx) + math.Sin(t*math.Pi*1.6+phase)*float64(amp)*(1-t)
		y := float64(cy) - t*float64(height)
		r := 1.25 - t*0.55
		fillCircle(img, x, y, r, c)
	}
}

func fillCircle(img *image.NRGBA, cx, cy, r float64, c color.NRGBA) {
	minX, maxX := int(cx-r-1), int(cx+r+1)
	minY, maxY := int(cy-r-1), int(cy+r+1)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			d := math.Hypot(float64(x)+0.5-cx, float64(y)+0.5-cy)
			if d > r+0.8 {
				continue
			}
			cc := c
			if d > r {
				cc.A = uint8(float64(cc.A) * (1 - (d-r)/0.8))
			}
			blend(img, x, y, cc)
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
