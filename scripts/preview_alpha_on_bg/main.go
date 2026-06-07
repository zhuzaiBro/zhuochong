// 把透明 PNG 合成到指定底色上，方便检查抠图边缘。
package main

import (
	"flag"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"strconv"
	"strings"

	"golang.org/x/image/draw"
)

func main() {
	in := flag.String("in", "", "输入 PNG")
	out := flag.String("out", "", "输出 PNG")
	bg := flag.String("bg", "ffffff", "背景色，格式 RRGGBB")
	flag.Parse()
	if *in == "" || *out == "" {
		log.Fatal("必须提供 -in 和 -out")
	}
	bgColor, err := parseHex(*bg)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Open(*in)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	src, err := png.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	b := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{C: bgColor}, image.Point{}, draw.Src)
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Over)
	if err := writePNG(*out, dst); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s", *out)
}

func parseHex(s string) (color.NRGBA, error) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(s) != 6 {
		return color.NRGBA{}, strconv.ErrSyntax
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return color.NRGBA{}, err
	}
	return color.NRGBA{
		R: uint8(v >> 16),
		G: uint8(v >> 8),
		B: uint8(v),
		A: 255,
	}, nil
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
