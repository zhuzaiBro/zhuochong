// 从各动作 folder 下的 frame01.png 生成 frame02…frame07（轻度缩放/位移），便于逐帧动画更自然。
// 运行：在仓库根目录执行 go run ./scripts/gen_chiikawa_keyframes
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
)

type kf struct {
	sx, sy float64
	dx, dy int
}

// 每组 6 个中间关键帧，对应 frame02.png … frame07.png（frame01 为手搓原图）
var presets = map[string][]kf{
	"idle": {
		// 相对原呼吸幅度，缩放变化量减半（仅 idle）
		{1.0075, 1.005, 0, 0},
		{1.0125, 1.01, 0, -1},
		{1.0075, 1.005, 0, 0},
		{1.0, 1.0, 0, 0},
		{0.995, 0.995, 0, 0},
		{1.004, 1.003, 0, 0},
	},
	"talking": {
		{1.04, 0.96, 0, 0},
		{0.96, 1.03, 1, 0},
		{1.03, 0.97, -1, 0},
		{0.98, 1.02, 0, 0},
		{1.02, 0.99, 0, 0},
		{1.0, 1.0, 0, 0},
	},
	"happy": {
		{1.0, 1.0, 0, 0},
		{1.04, 1.04, 0, -2},
		{1.07, 1.07, 0, -4},
		{1.05, 1.05, 0, -2},
		{1.02, 1.02, 0, 0},
		{1.0, 1.0, 0, 0},
	},
	"confused": {
		{1.0, 1.0, -2, 0},
		{1.01, 0.99, 2, 0},
		{0.99, 1.01, -1, 1},
		{1.02, 0.98, 3, 0},
		{1.0, 1.0, -2, 0},
		{1.0, 1.0, 0, 0},
	},
	"sleeping": {
		{1.0, 1.0, 0, 0},
		{1.008, 1.006, 0, 0},
		{1.012, 1.01, 0, 1},
		{1.008, 1.006, 0, 0},
		{1.0, 1.0, 0, 0},
		{0.998, 0.998, 0, -1},
	},
	"angry": {
		{1.02, 1.02, -3, 0},
		{1.03, 0.98, 4, 0},
		{0.98, 1.02, -4, 1},
		{1.04, 1.0, 3, -1},
		{1.0, 1.03, -2, 0},
		{1.0, 1.0, 0, 0},
	},
	"shy": {
		{0.98, 0.98, 2, 2},
		{0.96, 0.97, 3, 3},
		{0.95, 0.96, 4, 4},
		{0.97, 0.98, 2, 2},
		{0.99, 0.99, 0, 0},
		{1.0, 1.0, 0, 0},
	},
	"surprised": {
		{1.03, 1.03, 0, -2},
		{1.08, 1.08, 0, -4},
		{1.06, 1.06, 0, -2},
		{1.04, 1.04, 0, 0},
		{1.01, 1.01, 0, 0},
		{1.0, 1.0, 0, 0},
	},
	"tired": {
		{1.0, 0.99, 0, 1},
		{1.0, 0.98, 0, 2},
		{1.0, 0.97, 0, 3},
		{1.0, 0.98, 0, 2},
		{1.0, 0.99, 0, 1},
		{1.0, 1.0, 0, 0},
	},
	"wink": {
		{1.0, 0.97, 0, 1},
		{1.0, 0.93, 0, 2},
		{1.0, 0.9, 0, 3},
		{1.0, 0.95, 0, 1},
		{1.0, 1.0, 0, 0},
		{1.0, 1.0, 0, 0},
	},
}

func main() {
	root := flag.String("root", "chiikawa-sprites", "素材根目录（相对 cwd）")
	size := flag.Int("size", 0, "输出正方形画布尺寸；0 表示沿用 frame01 尺寸")
	flag.Parse()

	entries, err := os.ReadDir(*root)
	if err != nil {
		log.Fatal(err)
	}
	for _, ent := range entries {
		if !ent.IsDir() || strings.HasPrefix(ent.Name(), ".") {
			continue
		}
		name := ent.Name()
		preset, ok := presets[name]
		if !ok {
			log.Printf("跳过未知动作 %q（无预设）", name)
			continue
		}
		dir := filepath.Join(*root, name)
		srcPath := filepath.Join(dir, "frame01.png")
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			legacy := filepath.Join(dir, "frame1.png")
			if _, err2 := os.Stat(legacy); err2 == nil {
				srcPath = legacy
			} else {
				log.Printf("跳过 %q：无 frame01.png", dir)
				continue
			}
		}
		src, err := loadPNG(srcPath)
		if err != nil {
			log.Fatal(err)
		}
		for i, k := range preset {
			frameIdx := i + 2
			outPath := filepath.Join(dir, fmt.Sprintf("frame%02d.png", frameIdx))
			dst := renderKeyframe(src, k.sx, k.sy, k.dx, k.dy, *size)
			if err := writePNG(outPath, dst); err != nil {
				log.Fatal(err)
			}
			log.Println("写", outPath)
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

// renderKeyframe 在固定画布上居中绘制缩放后的角色（保留透明底）。
func renderKeyframe(src image.Image, sx, sy float64, dx, dy int, canvasSize int) *image.NRGBA {
	sr := src.Bounds()
	sw, sh := sr.Dx(), sr.Dy()
	outW, outH := sw, sh
	if canvasSize > 0 {
		outW, outH = canvasSize, canvasSize
	}
	baseScale := float64(outW) / float64(sw)
	if by := float64(outH) / float64(sh); by < baseScale {
		baseScale = by
	}
	dw := int(float64(sw)*baseScale*sx + 0.5)
	dh := int(float64(sh)*baseScale*sy + 0.5)
	if dw < 1 {
		dw = 1
	}
	if dh < 1 {
		dh = 1
	}
	scaled := image.NewNRGBA(image.Rect(0, 0, dw, dh))
	draw.CatmullRom.Scale(scaled, scaled.Bounds(), src, sr, draw.Over, nil)

	out := image.NewNRGBA(image.Rect(0, 0, outW, outH))
	// 透明画布
	for i := range out.Pix {
		out.Pix[i] = 0
	}
	px := (outW-dw)/2 + int(float64(dx)*baseScale+0.5)
	py := (outH-dh)/2 + int(float64(dy)*baseScale+0.5)
	draw.Draw(out, image.Rect(px, py, px+dw, py+dh), scaled, image.Point{}, draw.Over)
	return out
}
