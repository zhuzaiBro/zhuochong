// 从每组动作里的干净高分辨率帧重新衍生前 6 张关键帧，避开低清脏底素材。
// 运行：go run ./scripts/derive_lucheng_keyframes
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/image/draw"
)

type keyframe struct {
	sx, sy float64
	dx, dy int
}

var sources = map[string]string{
	"coffee":         "frame7.png",
	"happy":          "frame7.png",
	"headpat":        "frame7.png",
	"idle":           "frame10.png",
	"talking":        "frame7.png",
	"water_reminder": "frame7.png",
}

var presets = []keyframe{
	{1.000, 1.000, 0, 0},
	{1.006, 0.996, 0, 1},
	{0.996, 1.006, 0, -1},
	{1.008, 1.006, 0, -2},
	{1.002, 1.002, 0, -1},
	{1.000, 1.000, 0, 0},
}

func main() {
	root := flag.String("root", "lucheng-sprites", "素材根目录")
	flag.Parse()

	for action, srcName := range sources {
		dir := filepath.Join(*root, action)
		srcPath := filepath.Join(dir, srcName)
		src, err := loadPNG(srcPath)
		if err != nil {
			log.Fatalf("%s: %v", srcPath, err)
		}
		for i, k := range presets {
			dst := render(src, k)
			outPath := filepath.Join(dir, fmt.Sprintf("frame%d.png", i+1))
			if err := writePNG(outPath, dst); err != nil {
				log.Fatal(err)
			}
			log.Printf("写 %s <- %s", outPath, srcPath)
		}
	}
	log.Println("完成")
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

func render(src image.Image, k keyframe) *image.NRGBA {
	sr := src.Bounds()
	w, h := sr.Dx(), sr.Dy()
	dw := int(float64(w)*k.sx + 0.5)
	dh := int(float64(h)*k.sy + 0.5)
	scaled := image.NewNRGBA(image.Rect(0, 0, dw, dh))
	draw.CatmullRom.Scale(scaled, scaled.Bounds(), src, sr, draw.Over, nil)

	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	px := (w-dw)/2 + k.dx
	py := (h-dh)/2 + k.dy
	draw.Draw(out, image.Rect(px, py, px+dw, py+dh), scaled, image.Point{}, draw.Over)
	return out
}
