// 生成陆沉关键帧草案，不覆盖运行时 lucheng-sprites。
// 运行：go run ./scripts/build_lucheng_keyframe_draft
package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"

	xdraw "golang.org/x/image/draw"
)

const (
	root = "lucheng-sprites"
	out  = "lucheng-keyframes-draft"
)

type draft struct {
	file string
	src  string
	edit func(*image.NRGBA)
}

func main() {
	if err := os.RemoveAll(out); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		log.Fatal(err)
	}

	drafts := []draft{
		{"1-1_office_executive.png", "idle/frame9.png", decorateExecutiveOffice},
		{"2-1_noon_sleep.png", "coffee/frame7.png", decorateSleep},
		{"2-2_neutral.png", "idle/frame1.png", nil},
		{"2-3_talk_weather.png", "talking/frame7.png", decorateWeatherReminder},
		{"2-4_aggrieved.png", "headpat/frame7.png", decorateAggrieved},
		{"2-5_walk_weather.png", "happy/frame8.png", decorateWeatherReminder},
		{"3-1_talk.png", "talking/frame8.png", nil},
		{"3-2_cheer.png", "happy/frame7.png", nil},
		{"3-3_closed_eye_smile.png", "coffee/frame7.png", nil},
		{"3-4_walk.png", "happy/frame8.png", nil},
		{"3-5_thinking.png", "idle/frame10.png", decorateThinking},
		{"3-6_weather_reminder.png", "talking/frame7.png", decorateWeatherReminder},
	}

	for _, d := range drafts {
		img, err := load(filepath.Join(root, d.src))
		if err != nil {
			log.Fatal(err)
		}
		if d.edit != nil {
			d.edit(img)
		}
		path := filepath.Join(out, d.file)
		if err := save(path, img); err != nil {
			log.Fatal(err)
		}
		log.Printf("写 %s", path)
	}

	sheet, err := contactSheet(filepath.Join(out, "contact-sheet.png"))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("写 %s", sheet)
}

func load(path string) (*image.NRGBA, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	b := img.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), img, b.Min, draw.Src)
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

func decorateExecutiveOffice(img *image.NRGBA) {
	// Executive desk: deep walnut body, black leather blotter, and restrained brass trim.
	fillRect(img, 250, 742, 1002, 808, rgba(35, 20, 16, 248))
	fillRect(img, 270, 804, 982, 996, rgba(70, 42, 33, 248))
	fillRect(img, 288, 826, 964, 974, rgba(86, 52, 41, 245))
	fillRect(img, 318, 858, 934, 944, rgba(52, 32, 28, 246))
	strokeRect(img, 258, 736, 994, 812, rgba(190, 142, 67, 245), 5)
	strokeRect(img, 302, 842, 950, 960, rgba(178, 128, 62, 235), 4)
	strokeRect(img, 340, 876, 912, 926, rgba(28, 18, 16, 230), 3)
	fillRect(img, 276, 1010, 326, 1190, rgba(46, 28, 23, 245))
	fillRect(img, 900, 1010, 950, 1190, rgba(46, 28, 23, 245))
	strokeRect(img, 276, 1010, 326, 1190, rgba(27, 16, 14, 230), 3)
	strokeRect(img, 900, 1010, 950, 1190, rgba(27, 16, 14, 230), 3)

	fillRect(img, 410, 712, 846, 782, rgba(24, 27, 34, 238))
	strokeRect(img, 410, 712, 846, 782, rgba(102, 92, 75, 230), 3)
	line(img, 430, 768, 826, 768, rgba(214, 170, 92, 180), 2)

	// Open laptop with a visible screen and keyboard deck.
	fillRect(img, 500, 616, 748, 744, rgba(21, 24, 30, 255))
	strokeRect(img, 500, 616, 748, 744, rgba(10, 12, 16, 255), 4)
	fillRect(img, 520, 638, 728, 718, rgba(58, 74, 94, 245))
	fillRect(img, 534, 654, 714, 704, rgba(76, 96, 120, 170))
	fillRect(img, 452, 744, 794, 778, rgba(206, 214, 222, 245))
	fillRect(img, 478, 754, 768, 768, rgba(124, 137, 151, 240))
	fillRect(img, 584, 760, 664, 772, rgba(84, 92, 104, 240))
	strokeRect(img, 452, 744, 794, 778, rgba(32, 36, 42, 255), 3)

	// Upright porcelain coffee cup on a saucer.
	fillEllipse(img, 858, 736, 82, 24, rgba(224, 226, 232, 245))
	strokeEllipse(img, 858, 736, 82, 24, rgba(74, 74, 82, 210), 2)
	fillRect(img, 818, 654, 898, 734, rgba(246, 247, 250, 255))
	fillEllipse(img, 858, 654, 42, 18, rgba(250, 250, 252, 255))
	fillEllipse(img, 858, 654, 31, 11, rgba(74, 42, 27, 255))
	strokeEllipse(img, 858, 654, 42, 18, rgba(36, 36, 42, 230), 2)
	strokeRect(img, 818, 654, 898, 734, rgba(224, 226, 232, 220), 2)
	strokeEllipse(img, 902, 690, 24, 30, rgba(246, 247, 250, 245), 6)

	// Desk accessories: document tray, name plate, and a slim brass lamp.
	fillRect(img, 330, 690, 458, 722, rgba(240, 233, 214, 245))
	strokeRect(img, 330, 690, 458, 722, rgba(166, 126, 64, 235), 2)
	fillRect(img, 350, 704, 436, 709, rgba(88, 70, 54, 185))
	fillRect(img, 332, 728, 476, 760, rgba(70, 45, 34, 235))
	line(img, 760, 664, 860, 598, rgba(196, 150, 78, 230), 6)
	fillEllipse(img, 884, 588, 58, 22, rgba(206, 164, 88, 235))
	fillEllipse(img, 742, 760, 34, 10, rgba(190, 142, 67, 225))
}

func decorateSleep(img *image.NRGBA) {
	fillEllipse(img, 860, 270, 56, 56, rgba(246, 232, 160, 245))
	fillEllipse(img, 880, 258, 42, 52, rgba(0, 0, 0, 0))
	drawZ(img, 900, 340, rgba(80, 80, 92, 220), 6)
	drawZ(img, 954, 292, rgba(80, 80, 92, 180), 4)
}

func decorateAggrieved(img *image.NRGBA) {
	fillEllipse(img, 455, 490, 28, 44, rgba(112, 184, 230, 210))
	fillEllipse(img, 795, 490, 28, 44, rgba(112, 184, 230, 210))
	fillEllipse(img, 450, 482, 12, 18, rgba(238, 250, 255, 230))
	fillEllipse(img, 790, 482, 12, 18, rgba(238, 250, 255, 230))
}

func decorateWeatherReminder(img *image.NRGBA) {
	fillEllipse(img, 872, 300, 58, 58, rgba(255, 205, 82, 245))
	for _, p := range [][4]int{{872, 220, 872, 252}, {872, 348, 872, 382}, {790, 300, 824, 300}, {920, 300, 956, 300}, {815, 244, 838, 266}, {928, 244, 905, 266}, {815, 356, 838, 334}, {928, 356, 905, 334}} {
		line(img, p[0], p[1], p[2], p[3], rgba(255, 205, 82, 230), 5)
	}
	fillEllipse(img, 938, 374, 76, 44, rgba(232, 240, 248, 245))
	fillEllipse(img, 890, 382, 58, 38, rgba(232, 240, 248, 245))
	fillEllipse(img, 970, 386, 64, 40, rgba(232, 240, 248, 245))
	strokeEllipse(img, 938, 374, 76, 44, rgba(112, 126, 142, 210), 3)
	line(img, 904, 448, 884, 496, rgba(96, 150, 232, 230), 5)
	line(img, 960, 448, 940, 500, rgba(96, 150, 232, 230), 5)
}

func decorateThinking(img *image.NRGBA) {
	fillEllipse(img, 830, 330, 36, 36, rgba(245, 245, 248, 240))
	fillEllipse(img, 894, 286, 52, 52, rgba(245, 245, 248, 235))
	fillEllipse(img, 970, 244, 68, 68, rgba(245, 245, 248, 225))
	strokeEllipse(img, 830, 330, 36, 36, rgba(92, 92, 104, 180), 2)
	strokeEllipse(img, 894, 286, 52, 52, rgba(92, 92, 104, 170), 2)
	strokeEllipse(img, 970, 244, 68, 68, rgba(92, 92, 104, 160), 2)
}

func contactSheet(outPath string) (string, error) {
	files, err := filepath.Glob(filepath.Join(out, "*.png"))
	if err != nil {
		return "", err
	}
	const cell = 312
	sheet := image.NewNRGBA(image.Rect(0, 0, 4*cell, 3*cell))
	fillRect(sheet, 0, 0, 4*cell, 3*cell, rgba(18, 18, 20, 255))
	for i, file := range files {
		if filepath.Base(file) == "contact-sheet.png" {
			continue
		}
		img, err := load(file)
		if err != nil {
			return "", err
		}
		x := (i % 4) * cell
		y := (i / 4) * cell
		drawScaled(sheet, img, image.Rect(x+16, y+16, x+cell-16, y+cell-16))
	}
	return outPath, save(outPath, sheet)
}

func drawScaled(dst *image.NRGBA, src image.Image, r image.Rectangle) {
	sb := src.Bounds()
	dw, dh := r.Dx(), r.Dy()
	sw, sh := sb.Dx(), sb.Dy()
	scale := float64(dw) / float64(sw)
	if s := float64(dh) / float64(sh); s < scale {
		scale = s
	}
	w := int(float64(sw) * scale)
	h := int(float64(sh) * scale)
	rr := image.Rect(r.Min.X+(dw-w)/2, r.Min.Y+(dh-h)/2, r.Min.X+(dw-w)/2+w, r.Min.Y+(dh-h)/2+h)
	xdraw.CatmullRom.Scale(dst, rr, src, sb, draw.Over, nil)
}

func rgba(r, g, b, a uint8) color.NRGBA { return color.NRGBA{R: r, G: g, B: b, A: a} }

func fillRect(img *image.NRGBA, x0, y0, x1, y1 int, c color.NRGBA) {
	draw.Draw(img, image.Rect(x0, y0, x1, y1), &image.Uniform{C: c}, image.Point{}, draw.Over)
}

func strokeRect(img *image.NRGBA, x0, y0, x1, y1 int, c color.NRGBA, w int) {
	fillRect(img, x0, y0, x1, y0+w, c)
	fillRect(img, x0, y1-w, x1, y1, c)
	fillRect(img, x0, y0, x0+w, y1, c)
	fillRect(img, x1-w, y0, x1, y1, c)
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

func drawZ(img *image.NRGBA, x, y int, c color.NRGBA, w int) {
	line(img, x, y, x+52, y, c, w)
	line(img, x+52, y, x, y+44, c, w)
	line(img, x, y+44, x+52, y+44, c, w)
}

func line(img *image.NRGBA, x0, y0, x1, y1 int, c color.NRGBA, width int) {
	dx := abs(x1 - x0)
	dy := -abs(y1 - y0)
	sx, sy := -1, -1
	if x0 < x1 {
		sx = 1
	}
	if y0 < y1 {
		sy = 1
	}
	err := dx + dy
	for {
		fillEllipse(img, x0, y0, width, width, c)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
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

func abs(v int) int {
	if v < 0 {
		return -v
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
