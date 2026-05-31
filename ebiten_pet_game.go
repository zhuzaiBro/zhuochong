//go:build ebitenpet

package main

import (
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// moodSleepIdle 等由 ebiten_pet_main 在解析 flag 后写入。
var (
	moodSleepIdle  time.Duration
	moodTalkingDur time.Duration
	moodHappyDur   time.Duration
)

type livePetGame struct {
	sm *StateMachine

	mu sync.Mutex

	subtitle     string
	wsOK         bool
	wsErr        string
	excitedUntil time.Time

	lastInteract time.Time
	talkingUntil time.Time
	happyUntil   time.Time

	moodTick int64

	bubbleLine     string
	bubbleImg      *ebiten.Image
	bubbleCacheKey string

	initPos    bool
	winW, winH int
}

func newLivePetGame(size float64, initX, initY int) *livePetGame {
	initEmbeddedSprites()

	bw, bh := 192, 192
	if len(embeddedSprites.Idle.Frames) > 0 {
		im := embeddedSprites.Idle.Frames[0]
		bw = im.Bounds().Dx()
		bh = im.Bounds().Dy()
	}
	ww := int(math.Round(float64(bw) * size))
	hh := int(math.Round(float64(bh) * size))

	g := &livePetGame{
		sm:           newStateMachine(ww, hh),
		lastInteract: time.Now(),
	}
	g.winW = g.sm.winW
	g.winH = g.sm.winH
	g.placeInitialWindow(initX, initY)
	ensureBubbleFace()
	return g
}

func (g *livePetGame) placeInitialWindow(initX, initY int) {
	sw, sh := ebiten.Monitor().Size()
	if initX >= 0 && initY >= 0 {
		ebiten.SetWindowPosition(initX, initY)
		return
	}
	x := sw - g.winW - 40
	if x < 0 {
		x = 40
	}
	y := sh - g.winH - 120
	if y < 0 {
		y = 40
	}
	ebiten.SetWindowPosition(x, y)
}

func defaultEbitenRunOptions() *ebiten.RunGameOptions {
	return &ebiten.RunGameOptions{
		ScreenTransparent: true,
		InitUnfocused:     false,
	}
}

func runEbitenPet(game *livePetGame, opts *ebiten.RunGameOptions) error {
	w0, h0 := game.outerLayout()
	ebiten.SetWindowSize(w0, h0)
	ebiten.SetWindowTitle("抖音直播精灵")
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowFloating(true)
	ebiten.SetTPS(60)
	return ebiten.RunGameWithOptions(game, opts)
}

func (g *livePetGame) setWSState(ok bool, errStr string) {
	g.mu.Lock()
	g.wsOK = ok
	g.wsErr = errStr
	g.mu.Unlock()
}

func (g *livePetGame) setSubtitle(s string) {
	s = strings.TrimSpace(s)
	g.mu.Lock()
	g.subtitle = s
	g.mu.Unlock()
	ebiten.SetWindowTitle(truncWindowTitle(s))
}

func (g *livePetGame) triggerExcited(d time.Duration) {
	g.mu.Lock()
	g.excitedUntil = time.Now().Add(d)
	g.mu.Unlock()
}

func (g *livePetGame) readExcited() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return time.Now().Before(g.excitedUntil)
}

func (g *livePetGame) bumpInteraction() {
	g.mu.Lock()
	g.lastInteract = time.Now()
	g.mu.Unlock()
}

func (g *livePetGame) startTalking() {
	g.mu.Lock()
	now := time.Now()
	g.lastInteract = now
	g.happyUntil = time.Time{}
	g.talkingUntil = now.Add(moodTalkingDur)
	g.mu.Unlock()
}

func (g *livePetGame) startHappyGift() {
	g.mu.Lock()
	now := time.Now()
	g.lastInteract = now
	g.talkingUntil = time.Time{}
	g.happyUntil = now.Add(moodHappyDur)
	g.bubbleLine = ""
	g.bubbleCacheKey = ""
	if g.bubbleImg != nil {
		g.bubbleImg.Dispose()
		g.bubbleImg = nil
	}
	g.mu.Unlock()
}

func (g *livePetGame) setBubble(s string) {
	ensureBubbleFace()
	s = clampBubbleLine(s, 220)
	g.mu.Lock()
	defer g.mu.Unlock()
	g.bubbleLine = s
	g.bubbleCacheKey = ""
}

func (g *livePetGame) outerLayout() (int, int) {
	return g.winW + bubbleColumnW, g.winH
}

func (g *livePetGame) overlayFrame(a *Anim) *ebiten.Image {
	if a == nil || len(a.Frames) == 0 {
		return g.sm.frameImg()
	}
	idx := int((g.moodTick / petAnimFrameHold) % int64(len(a.Frames)))
	return a.Frames[idx]
}

func (g *livePetGame) pickDrawImage() *ebiten.Image {
	if g.sm.mustDrawStateOnly() {
		return g.sm.frameImg()
	}
	now := time.Now()
	g.mu.Lock()
	hu := g.happyUntil
	tu := g.talkingUntil
	li := g.lastInteract
	g.mu.Unlock()

	if !hu.IsZero() && now.Before(hu) && embeddedSprites.Happy != nil && len(embeddedSprites.Happy.Frames) > 0 {
		return g.overlayFrame(embeddedSprites.Happy)
	}
	if !tu.IsZero() && now.Before(tu) && embeddedSprites.Talking != nil && len(embeddedSprites.Talking.Frames) > 0 {
		return g.overlayFrame(embeddedSprites.Talking)
	}
	if moodSleepIdle > 0 && now.Sub(li) >= moodSleepIdle && embeddedSprites.Sleeping != nil && len(embeddedSprites.Sleeping.Frames) > 0 {
		return g.overlayFrame(embeddedSprites.Sleeping)
	}
	if g.readExcited() && g.sm.showHappyOverlay() && len(embeddedSprites.Happy.Frames) > 0 {
		return embeddedSprites.Happy.Frames[0]
	}
	return g.sm.frameImg()
}

func truncWindowTitle(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return "抖音直播精灵"
	}
	if len(r) <= 80 {
		return string(r)
	}
	return string(r[:79]) + "…"
}

func (g *livePetGame) Update() error {
	if !g.initPos {
		g.initPos = true
	}
	g.moodTick++
	return g.sm.Update()
}

func (g *livePetGame) Draw(screen *ebiten.Image) {
	screen.Clear()

	img := g.pickDrawImage()

	sw, sh := img.Bounds().Dx(), img.Bounds().Dy()
	if sw == 0 || sh == 0 {
		log.Println("ebiten 精灵：空贴图")
		return
	}
	sx := float64(g.winW) / float64(sw)
	sy := float64(g.winH) / float64(sh)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(sx, sy)
	screen.DrawImage(img, op)
	g.drawBubbleIfAny(screen)
}

func (g *livePetGame) drawBubbleIfAny(screen *ebiten.Image) {
	ensureBubbleFace()
	if bubbleFace == nil {
		return
	}
	g.mu.Lock()
	line := g.bubbleLine
	tu := g.talkingUntil
	if line == "" || tu.IsZero() || !time.Now().Before(tu) {
		g.mu.Unlock()
		return
	}
	if g.bubbleCacheKey != line {
		if g.bubbleImg != nil {
			g.bubbleImg.Dispose()
		}
		g.bubbleImg = renderBubbleImage(bubbleFace, line)
		g.bubbleCacheKey = line
	}
	bimg := g.bubbleImg
	g.mu.Unlock()
	if bimg == nil {
		return
	}
	bw := bimg.Bounds().Dx()
	bh := bimg.Bounds().Dy()
	y := 0
	if g.winH > bh {
		y = (g.winH - bh) / 2
	}
	op := &ebiten.DrawImageOptions{}
	x := g.winW + bubbleColumnW - bubbleColumnRightPad - bw
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(bimg, op)
}

func (g *livePetGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.outerLayout()
}
