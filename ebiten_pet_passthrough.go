//go:build ebitenpet

package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

const hitAlphaThreshold = 32

func (g *livePetGame) updateMousePassthrough() {
	want := !g.cursorHitsPet()
	if g.mousePassthrough == want {
		return
	}
	g.mousePassthrough = want
	ebiten.SetWindowMousePassthrough(want)
}

// cursorHitsPet 判断光标是否落在立绘不透明像素上（或正在与之交互）。
func (g *livePetGame) cursorHitsPet() bool {
	switch g.sm.state.(type) {
	case *stateDrag, *stateRClick, *stateHungry, *stateFeed:
		return true
	}

	lx, ly := g.cursorLocal()
	if lx < 0 || ly < 0 || lx >= g.winW || ly >= g.winH {
		return false
	}

	img := g.pickDrawImage()
	if img == nil {
		return false
	}
	b := img.Bounds()
	sw, sh := b.Dx(), b.Dy()
	if sw == 0 || sh == 0 {
		return false
	}

	ix := int(float64(lx) * float64(sw) / float64(g.winW))
	iy := int(float64(ly) * float64(sh) / float64(g.winH))
	if ix < 0 || iy < 0 || ix >= sw || iy >= sh {
		return false
	}

	_, _, _, a := img.At(b.Min.X+ix, b.Min.Y+iy).RGBA()
	return uint8(a>>8) >= hitAlphaThreshold
}

func (g *livePetGame) cursorLocal() (int, int) {
	wx, wy := ebiten.WindowPosition()
	if sx, sy, ok := screenCursorPos(); ok {
		return sx - wx, sy - wy
	}
	gp := globalCursorPosition()
	return gp.x - wx, gp.y - wy
}
