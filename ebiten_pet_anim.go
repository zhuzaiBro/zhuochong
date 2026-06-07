//go:build ebitenpet

package main

import (
	"bytes"
	"embed"
	"image/png"
	"log"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed lucheng-sprites
var luchengSprites embed.FS

// petAnimFrameHold 每帧停留的游戏 tick 数（TPS=60；22 ≈ 367ms/帧，10 帧约 3.7s 一轮）。
const petAnimFrameHold = 22

type Anim struct {
	Frames []*ebiten.Image
}

type petSpriteSet struct {
	Idle    *Anim
	Wink    *Anim
	Tired   *Anim
	Happy   *Anim
	Talking *Anim
	Office  *Anim
}

var embeddedSprites *petSpriteSet

func frameNumber(name string) int {
	base := strings.TrimSuffix(name, path.Ext(name))
	if !strings.HasPrefix(base, "frame") {
		return 0
	}
	n, _ := strconv.Atoi(strings.TrimPrefix(base, "frame"))
	return n
}

func loadAnimSubdir(rel string) *Anim {
	entries, err := luchengSprites.ReadDir(rel)
	if err != nil {
		log.Fatalf("lucheng 素材目录 %q: %v", rel, err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if path.Ext(e.Name()) != ".png" {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Slice(names, func(i, j int) bool {
		ni, nj := frameNumber(names[i]), frameNumber(names[j])
		if ni != nj {
			return ni < nj
		}
		return names[i] < names[j]
	})
	if len(names) == 0 {
		log.Fatalf("lucheng 目录 %q 内没有 PNG", rel)
	}
	var frames []*ebiten.Image
	for _, n := range names {
		data, err := luchengSprites.ReadFile(path.Join(rel, n))
		if err != nil {
			log.Fatal(err)
		}
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			log.Fatal(err)
		}
		frames = append(frames, ebiten.NewImageFromImage(img))
	}
	return &Anim{Frames: frames}
}

func initEmbeddedSprites() {
	root := "lucheng-sprites"
	embeddedSprites = &petSpriteSet{
		Idle:    loadAnimSubdir(root + "/idle"),
		Wink:    loadAnimSubdir(root + "/headpat"), // 右键：摸头
		Tired:   loadAnimSubdir(root + "/coffee"),  // 饿了：想喝咖啡
		Happy:   loadAnimSubdir(root + "/happy"),
		Talking: loadAnimSubdir(root + "/talking"),
		Office:  loadAnimSubdir(root + "/office"), // 无互动一段时间后：办公
	}
}
