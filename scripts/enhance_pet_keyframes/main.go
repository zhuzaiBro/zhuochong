// 统一桌宠关键帧质量：高质量缩放到固定画布，并可从边缘抠掉白底。
// 运行示例：go run ./scripts/enhance_pet_keyframes -root lucheng-sprites -size 1248 -key-white
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/image/draw"
)

type point struct {
	x, y int
}

type bgReport struct {
	path          string
	borderPixels  int
	borderOpaque  int
	cornerOpaque  int
	edgeConnected int
}

func main() {
	root := flag.String("root", "lucheng-sprites", "素材根目录")
	size := flag.Int("size", 1248, "输出正方形画布尺寸")
	padding := flag.Int("padding", 0, "输出画布四周保留的透明安全边")
	keyWhite := flag.Bool("key-white", true, "从画布边缘移除近白背景")
	cleanAlpha := flag.Bool("clean-alpha", false, "清理角色内部半透明脏像素")
	alphaCutoff := flag.Int("alpha-cutoff", 24, "clean-alpha: 低于该 alpha 的像素清为透明")
	alphaSolid := flag.Int("alpha-solid", 240, "clean-alpha: 高于等于该 alpha 的像素扶正为不透明")
	alphaRadius := flag.Int("alpha-radius", 2, "clean-alpha: 判断内部像素的邻域半径")
	despeckle := flag.Bool("despeckle", false, "清理孤立黑白噪点")
	despeckleIters := flag.Int("despeckle-iters", 1, "despeckle: 去斑点迭代次数")
	checkOnly := flag.Bool("check-only", false, "仅检查背景透明度，不改写图片")
	threshold := flag.Int("white-threshold", 238, "白底阈值，数值越低抠得越多")
	flag.Parse()

	if *size < 1 {
		log.Fatal("-size 必须大于 0")
	}
	if *threshold < 0 || *threshold > 255 {
		log.Fatal("-white-threshold 必须在 0..255")
	}
	if *padding < 0 || *padding*2 >= *size {
		log.Fatal("-padding 必须大于等于 0，且小于输出尺寸的一半")
	}
	if *alphaCutoff < 0 || *alphaCutoff > 255 || *alphaSolid < 0 || *alphaSolid > 255 {
		log.Fatal("-alpha-cutoff 和 -alpha-solid 必须在 0..255")
	}
	if *alphaRadius < 0 {
		log.Fatal("-alpha-radius 必须大于等于 0")
	}
	if *despeckleIters < 0 {
		log.Fatal("-despeckle-iters 必须大于等于 0")
	}

	var count int
	var reports []bgReport
	err := filepath.WalkDir(*root, func(path string, ent os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ent.IsDir() || strings.ToLower(filepath.Ext(path)) != ".png" {
			return nil
		}
		switch filepath.Base(path) {
		case "reference.png", "contact-sheet.png":
			return nil
		}
		img, err := loadPNG(path)
		if err != nil {
			return err
		}
		if *checkOnly {
			if report := checkBackground(path, img); report.borderOpaque > 0 {
				reports = append(reports, report)
			}
			count++
			return nil
		}
		out := normalize(img, *size, *padding)
		if *keyWhite {
			removeEdgeWhite(out, uint8(*threshold))
		}
		if *cleanAlpha {
			cleanInteriorAlpha(out, uint8(*alphaCutoff), uint8(*alphaSolid), *alphaRadius)
		}
		if *despeckle {
			despeckleImage(out, *despeckleIters)
		}
		if err := writePNG(path, out); err != nil {
			return err
		}
		count++
		log.Printf("增强 %s", path)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	if *checkOnly {
		for _, report := range reports {
			log.Printf(
				"背景疑似不透明 %s: border=%d/%d corner=%d edge-connected=%d",
				report.path,
				report.borderOpaque,
				report.borderPixels,
				report.cornerOpaque,
				report.edgeConnected,
			)
		}
		if len(reports) > 0 {
			log.Fatalf("检查完成：%d/%d 张 PNG 边缘存在不透明像素", len(reports), count)
		}
		log.Printf("检查完成：%d 张 PNG 背景边缘均为透明", count)
		return
	}
	log.Printf("完成：共处理 %d 张 PNG，输出尺寸 %dx%d", count, *size, *size)
}

func loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
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

func normalize(src image.Image, size, padding int) *image.NRGBA {
	sr := src.Bounds()
	sw, sh := sr.Dx(), sr.Dy()
	target := size - 2*padding
	scale := float64(target) / float64(max(sw, sh))
	dw := int(float64(sw)*scale + 0.5)
	dh := int(float64(sh)*scale + 0.5)
	if dw < 1 {
		dw = 1
	}
	if dh < 1 {
		dh = 1
	}

	scaled := image.NewNRGBA(image.Rect(0, 0, dw, dh))
	draw.CatmullRom.Scale(scaled, scaled.Bounds(), src, sr, draw.Src, nil)

	out := image.NewNRGBA(image.Rect(0, 0, size, size))
	pt := image.Pt((size-dw)/2, (size-dh)/2)
	draw.Draw(out, image.Rectangle{Min: pt, Max: pt.Add(scaled.Bounds().Size())}, scaled, image.Point{}, draw.Src)
	return out
}

func removeEdgeWhite(img *image.NRGBA, threshold uint8) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w == 0 || h == 0 {
		return
	}

	seen := make([]bool, w*h)
	queue := make([]point, 0, 2*w+2*h)
	push := func(x, y int) {
		if x < 0 || x >= w || y < 0 || y >= h {
			return
		}
		i := y*w + x
		if seen[i] || !isEdgeBackground(img.NRGBAAt(x, y), threshold) {
			return
		}
		seen[i] = true
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
		i := img.PixOffset(p.x, p.y)
		img.Pix[i+0] = 0
		img.Pix[i+1] = 0
		img.Pix[i+2] = 0
		img.Pix[i+3] = 0
		push(p.x+1, p.y)
		push(p.x-1, p.y)
		push(p.x, p.y+1)
		push(p.x, p.y-1)
	}
}

func cleanInteriorAlpha(img *image.NRGBA, cutoff, solid uint8, radius int) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	mask := make([]bool, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := img.PixOffset(x, y)
			a := img.Pix[i+3]
			if a < cutoff {
				img.Pix[i+0] = 0
				img.Pix[i+1] = 0
				img.Pix[i+2] = 0
				img.Pix[i+3] = 0
				continue
			}
			mask[y*w+x] = true
		}
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := img.PixOffset(x, y)
			a := img.Pix[i+3]
			if a == 0 {
				continue
			}
			interior := isInterior(mask, w, h, x, y, radius)
			if a >= solid || interior {
				if a < solid && interior {
					r, g, b, ok := localOpaqueColor(img, x, y, radius+3, solid)
					if ok {
						img.Pix[i+0] = r
						img.Pix[i+1] = g
						img.Pix[i+2] = b
						img.Pix[i+3] = 255
						continue
					}
				}
				img.Pix[i+0] = compositeOverWhite(img.Pix[i+0], a)
				img.Pix[i+1] = compositeOverWhite(img.Pix[i+1], a)
				img.Pix[i+2] = compositeOverWhite(img.Pix[i+2], a)
				img.Pix[i+3] = 255
			}
		}
	}
}

func localOpaqueColor(img *image.NRGBA, x, y, radius int, solid uint8) (uint8, uint8, uint8, bool) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	var total, rs, gs, bs int
	for yy := y - radius; yy <= y+radius; yy++ {
		if yy < 0 || yy >= h {
			continue
		}
		for xx := x - radius; xx <= x+radius; xx++ {
			if xx < 0 || xx >= w || (xx == x && yy == y) {
				continue
			}
			i := img.PixOffset(xx, yy)
			if img.Pix[i+3] < solid {
				continue
			}
			total++
			rs += int(img.Pix[i+0])
			gs += int(img.Pix[i+1])
			bs += int(img.Pix[i+2])
		}
	}
	if total < 4 {
		return 0, 0, 0, false
	}
	return uint8(rs / total), uint8(gs / total), uint8(bs / total), true
}

func despeckleImage(img *image.NRGBA, iterations int) {
	for range iterations {
		src := cloneNRGBA(img)
		b := img.Bounds()
		w, h := b.Dx(), b.Dy()
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				i := src.PixOffset(x, y)
				if src.Pix[i+3] == 0 {
					continue
				}
				r, g, bb, ok := medianNeighborColor(src, x, y, 2)
				if !ok {
					continue
				}
				dist := colorDistance(src.Pix[i+0], src.Pix[i+1], src.Pix[i+2], r, g, bb)
				if dist < 95 {
					continue
				}
				if similarNeighborCount(src, x, y, 2, src.Pix[i+0], src.Pix[i+1], src.Pix[i+2]) >= 5 {
					continue
				}
				di := img.PixOffset(x, y)
				img.Pix[di+0] = r
				img.Pix[di+1] = g
				img.Pix[di+2] = bb
			}
		}
	}
}

func cloneNRGBA(src *image.NRGBA) *image.NRGBA {
	dst := image.NewNRGBA(src.Bounds())
	copy(dst.Pix, src.Pix)
	return dst
}

func medianNeighborColor(img *image.NRGBA, x, y, radius int) (uint8, uint8, uint8, bool) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	rs := make([]int, 0, (2*radius+1)*(2*radius+1))
	gs := make([]int, 0, cap(rs))
	bs := make([]int, 0, cap(rs))
	for yy := y - radius; yy <= y+radius; yy++ {
		if yy < 0 || yy >= h {
			continue
		}
		for xx := x - radius; xx <= x+radius; xx++ {
			if xx < 0 || xx >= w || (xx == x && yy == y) {
				continue
			}
			i := img.PixOffset(xx, yy)
			if img.Pix[i+3] == 0 {
				continue
			}
			rs = append(rs, int(img.Pix[i+0]))
			gs = append(gs, int(img.Pix[i+1]))
			bs = append(bs, int(img.Pix[i+2]))
		}
	}
	if len(rs) < 8 {
		return 0, 0, 0, false
	}
	sort.Ints(rs)
	sort.Ints(gs)
	sort.Ints(bs)
	mid := len(rs) / 2
	return uint8(rs[mid]), uint8(gs[mid]), uint8(bs[mid]), true
}

func similarNeighborCount(img *image.NRGBA, x, y, radius int, r, g, b uint8) int {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	var count int
	for yy := y - radius; yy <= y+radius; yy++ {
		if yy < 0 || yy >= h {
			continue
		}
		for xx := x - radius; xx <= x+radius; xx++ {
			if xx < 0 || xx >= w || (xx == x && yy == y) {
				continue
			}
			i := img.PixOffset(xx, yy)
			if img.Pix[i+3] == 0 {
				continue
			}
			if colorDistance(img.Pix[i+0], img.Pix[i+1], img.Pix[i+2], r, g, b) < 55 {
				count++
			}
		}
	}
	return count
}

func colorDistance(r1, g1, b1, r2, g2, b2 uint8) int {
	return abs(int(r1)-int(r2)) + abs(int(g1)-int(g2)) + abs(int(b1)-int(b2))
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func isInterior(mask []bool, w, h, x, y, radius int) bool {
	if radius == 0 {
		return true
	}
	for yy := y - radius; yy <= y+radius; yy++ {
		if yy < 0 || yy >= h {
			return false
		}
		for xx := x - radius; xx <= x+radius; xx++ {
			if xx < 0 || xx >= w || !mask[yy*w+xx] {
				return false
			}
		}
	}
	return true
}

func compositeOverWhite(v, a uint8) uint8 {
	return uint8((int(v)*int(a) + 255*(255-int(a)) + 127) / 255)
}

func checkBackground(path string, img image.Image) bgReport {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	report := bgReport{path: path}
	if w == 0 || h == 0 {
		return report
	}
	report.borderPixels = 2*w + 2*h - 4
	isOpaque := func(x, y int) bool {
		_, _, _, a := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
		return a != 0
	}
	for x := 0; x < w; x++ {
		if isOpaque(x, 0) {
			report.borderOpaque++
		}
		if h > 1 && isOpaque(x, h-1) {
			report.borderOpaque++
		}
	}
	for y := 1; y < h-1; y++ {
		if isOpaque(0, y) {
			report.borderOpaque++
		}
		if w > 1 && isOpaque(w-1, y) {
			report.borderOpaque++
		}
	}
	corners := []point{{0, 0}, {w - 1, 0}, {0, h - 1}, {w - 1, h - 1}}
	for _, c := range corners {
		if isOpaque(c.x, c.y) {
			report.cornerOpaque++
		}
	}
	report.edgeConnected = countEdgeConnectedOpaque(img)
	return report
}

func countEdgeConnectedOpaque(img image.Image) int {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	seen := make([]bool, w*h)
	queue := make([]point, 0, 2*w+2*h)
	isOpaque := func(x, y int) bool {
		_, _, _, a := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
		return a != 0
	}
	push := func(x, y int) {
		if x < 0 || x >= w || y < 0 || y >= h {
			return
		}
		i := y*w + x
		if seen[i] || !isOpaque(x, y) {
			return
		}
		seen[i] = true
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
	return len(queue)
}

func isEdgeBackground(c color.NRGBA, threshold uint8) bool {
	if c.A == 0 {
		return true
	}
	return c.A == 255 && c.R >= threshold && c.G >= threshold && c.B >= threshold
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func init() {
	log.SetFlags(0)
	log.SetPrefix(fmt.Sprintf("%s: ", filepath.Base(os.Args[0])))
}
