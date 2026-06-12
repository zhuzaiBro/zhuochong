// 生成陆沉关键帧草案，不覆盖运行时 lucheng-sprites。
// 运行：go run ./scripts/build_lucheng_keyframe_draft
package main

import (
	"flag"
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
	contactOnly := flag.Bool("contact-sheet-only", false, "只刷新 contact-sheet.png，不重新生成草案帧")
	flag.Parse()

	if *contactOnly {
		sheet, err := contactSheet(filepath.Join(out, "contact-sheet.png"))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("写 %s", sheet)
		return
	}

	if err := os.RemoveAll(out); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		log.Fatal(err)
	}

	drafts := []draft{
		{"1-1_office_executive.png", "office/action_typing.png", nil},
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
	// High-end executive desk in one consistent 3/4 perspective.
	top := []image.Point{{232, 724}, {1016, 724}, {966, 824}, {282, 824}}
	front := []image.Point{{282, 824}, {966, 824}, {928, 1042}, {318, 1042}}
	leftSide := []image.Point{{232, 724}, {282, 824}, {318, 1042}, {266, 1010}}
	rightSide := []image.Point{{1016, 724}, {966, 824}, {928, 1042}, {984, 1008}}
	leather := []image.Point{{404, 746}, {846, 746}, {820, 802}, {430, 802}}
	frontInset := []image.Point{{330, 858}, {920, 858}, {894, 976}, {354, 976}}
	frontCore := []image.Point{{370, 890}, {884, 890}, {866, 946}, {388, 946}}

	fillPoly(img, top, rgba(44, 25, 18, 255))
	fillPoly(img, leftSide, rgba(40, 22, 17, 255))
	fillPoly(img, rightSide, rgba(34, 19, 15, 255))
	fillPoly(img, front, rgba(82, 48, 34, 255))
	fillPoly(img, frontInset, rgba(128, 80, 48, 245))
	fillPoly(img, frontCore, rgba(48, 27, 22, 250))
	strokePoly(img, top, rgba(208, 154, 70, 245), 5)
	strokePoly(img, front, rgba(112, 70, 42, 235), 5)
	strokePoly(img, frontInset, rgba(202, 146, 66, 240), 4)
	strokePoly(img, frontCore, rgba(30, 18, 16, 230), 3)
	fillPoly(img, leather, rgba(20, 24, 31, 245))
	strokePoly(img, leather, rgba(110, 96, 72, 230), 3)
	line(img, 430, 792, 820, 792, rgba(214, 170, 92, 170), 2)

	// Block legs and side plinths aligned to the desk body.
	fillPoly(img, []image.Point{{278, 1010}, {330, 1038}, {330, 1184}, {278, 1184}}, rgba(42, 24, 18, 255))
	fillPoly(img, []image.Point{{902, 1038}, {956, 1010}, {956, 1184}, {902, 1184}}, rgba(42, 24, 18, 255))
	strokePoly(img, []image.Point{{278, 1010}, {330, 1038}, {330, 1184}, {278, 1184}}, rgba(26, 16, 13, 230), 3)
	strokePoly(img, []image.Point{{902, 1038}, {956, 1010}, {956, 1184}, {902, 1184}}, rgba(26, 16, 13, 230), 3)

	// Open laptop with perspective base.
	fillRect(img, 506, 620, 746, 744, rgba(20, 23, 29, 255))
	strokeRect(img, 506, 620, 746, 744, rgba(8, 10, 14, 255), 4)
	fillRect(img, 526, 640, 726, 714, rgba(58, 74, 94, 245))
	fillRect(img, 540, 656, 712, 700, rgba(80, 100, 122, 180))
	laptopBase := []image.Point{{456, 744}, {794, 744}, {830, 782}, {424, 782}}
	fillPoly(img, laptopBase, rgba(204, 212, 220, 248))
	strokePoly(img, laptopBase, rgba(30, 34, 40, 255), 3)
	fillPoly(img, []image.Point{{488, 756}, {770, 756}, {786, 770}, {472, 770}}, rgba(124, 137, 151, 240))
	fillPoly(img, []image.Point{{588, 760}, {668, 760}, {672, 772}, {584, 772}}, rgba(82, 90, 102, 240))

	// Smaller porcelain coffee cup and saucer placed on the same top plane.
	fillEllipse(img, 874, 752, 56, 16, rgba(224, 226, 232, 245))
	strokeEllipse(img, 874, 752, 56, 16, rgba(74, 74, 82, 210), 2)
	fillRect(img, 848, 690, 900, 744, rgba(246, 247, 250, 255))
	fillEllipse(img, 874, 690, 27, 11, rgba(250, 250, 252, 255))
	fillEllipse(img, 874, 690, 20, 7, rgba(74, 42, 27, 255))
	strokeEllipse(img, 874, 690, 27, 11, rgba(36, 36, 42, 230), 2)
	strokeRect(img, 848, 690, 900, 744, rgba(224, 226, 232, 220), 2)
	strokeEllipse(img, 904, 716, 16, 20, rgba(246, 247, 250, 245), 5)

	// Quiet premium desk accessories.
	fillPoly(img, []image.Point{{330, 692}, {454, 692}, {474, 724}, {312, 724}}, rgba(240, 233, 214, 245))
	strokePoly(img, []image.Point{{330, 692}, {454, 692}, {474, 724}, {312, 724}}, rgba(166, 126, 64, 235), 2)
	line(img, 350, 706, 438, 706, rgba(88, 70, 54, 185), 4)
	fillPoly(img, []image.Point{{326, 730}, {472, 730}, {492, 762}, {306, 762}}, rgba(70, 45, 34, 235))
	fillEllipse(img, 742, 766, 34, 10, rgba(190, 142, 67, 225))
	fillRect(img, 926, 700, 952, 748, rgba(42, 42, 48, 235))
	strokeRect(img, 926, 700, 952, 748, rgba(184, 138, 66, 230), 2)
	line(img, 932, 696, 916, 650, rgba(206, 164, 88, 230), 3)
	line(img, 942, 696, 946, 644, rgba(206, 164, 88, 230), 3)
	line(img, 950, 698, 970, 658, rgba(206, 164, 88, 230), 3)
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

func fillPoly(img *image.NRGBA, points []image.Point, c color.NRGBA) {
	if len(points) < 3 {
		return
	}
	minX, maxX := points[0].X, points[0].X
	minY, maxY := points[0].Y, points[0].Y
	for _, p := range points[1:] {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if pointInPoly(image.Pt(x, y), points) {
				blend(img, x, y, c)
			}
		}
	}
}

func strokePoly(img *image.NRGBA, points []image.Point, c color.NRGBA, width int) {
	if len(points) < 2 {
		return
	}
	for i := range points {
		a := points[i]
		b := points[(i+1)%len(points)]
		line(img, a.X, a.Y, b.X, b.Y, c, width)
	}
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
