//go:build ebitenpet

package main

import (
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// 逻辑参考 https://github.com/nhanb/shark 的状态机；素材为 chiikawa-sprites。

var walkChancePct, stopChancePct int
var durationTillHungry time.Duration

type Vector struct{ x, y int }

func createVector(x, y int) Vector { return Vector{x, y} }

func (a Vector) add(b Vector) Vector { return Vector{a.x + b.x, a.y + b.y} }

func (a Vector) sub(b Vector) Vector { return Vector{a.x - b.x, a.y - b.y} }

func globalCursorPosition() Vector {
	cx, cy := ebiten.CursorPosition()
	wx, wy := ebiten.WindowPosition()
	return Vector{cx + wx, cy + wy}
}

func randBool(chance int) bool {
	if chance <= 0 {
		return false
	}
	if chance >= 100 {
		return true
	}
	return rand.Intn(100) < chance
}

type StateMachine struct {
	state State
	anim  *Anim

	frameIdx         int
	animFrameCount   int
	ticks            int
	ticksPerAnimTick int // 与 shark 一致：每 N tick 推进一帧（多帧时）
	cycleAsSingle    bool
	ticksPerCycle    int // 单帧动画时一整轮逻辑的 tick 长度

	lastFed time.Time
	x, y    int

	isXYInit bool

	winW, winH int
}

func newStateMachine(baseW, baseH int) *StateMachine {
	sm := &StateMachine{
		winW:             baseW,
		winH:             baseH,
		ticksPerAnimTick: petAnimFrameHold,
		lastFed:          time.Now(),
	}
	sm.SetState(&stateIdle{})
	return sm
}

func (sm *StateMachine) SetAnim(a *Anim) {
	sm.anim = a
	sm.animFrameCount = len(a.Frames)
	sm.frameIdx = 0
	sm.ticks = 0
}

func (sm *StateMachine) frameImg() *ebiten.Image {
	return sm.anim.Frames[sm.frameIdx]
}

func (sm *StateMachine) SetState(s State) {
	sm.state = s
	s.enter(sm)
}

func (sm *StateMachine) Update() error {
	if !sm.isXYInit {
		sm.x, sm.y = ebiten.WindowPosition()
		sm.isXYInit = true
	}

	sm.state.update(sm)
	// 不按单块 Monitor 裁剪坐标，否则多显示器（副屏在左侧或为负坐标）时无法拖过去。
	ebiten.SetWindowPosition(sm.x, sm.y)

	// 动画帧推进（多 PNG 时沿用 shark；每帧仅 1 张图时走 ticksPerCycle）
	if sm.cycleAsSingle {
		sm.ticks++
		if sm.ticks < sm.ticksPerCycle {
			return nil
		}
		sm.ticks = 0
		sm.state.endAnimHook(sm)
		return nil
	}

	sm.ticks++
	if sm.ticks < sm.ticksPerAnimTick {
		return nil
	}
	sm.ticks = 0
	if sm.frameIdx < sm.animFrameCount-1 {
		sm.frameIdx++
		return nil
	}
	sm.frameIdx = 0
	sm.state.endAnimHook(sm)
	return nil
}

func (sm *StateMachine) showHappyOverlay() bool {
	switch sm.state.(type) {
	case *stateIdle, *stateWalk:
		return true
	default:
		return false
	}
}

// mustDrawStateOnly 为 true 时不叠直播间情绪层（睡觉/talking/happy），仅用状态机贴图。
func (sm *StateMachine) mustDrawStateOnly() bool {
	switch sm.state.(type) {
	case *stateDrag, *stateRClick, *stateHungry, *stateFeed, *stateWalk:
		return true
	default:
		return false
	}
}

func syncWinPos(sm *StateMachine) {
	ax, ay := ebiten.WindowPosition()
	if ax != sm.x || ay != sm.y {
		sm.x, sm.y = ax, ay
	}
}

type State interface {
	enter(sm *StateMachine)
	update(sm *StateMachine)
	endAnimHook(sm *StateMachine)
}

type stateIdle struct{}

func (s *stateIdle) enter(sm *StateMachine) {
	sm.SetAnim(embeddedSprites.Idle)
	sm.cycleAsSingle = len(embeddedSprites.Idle.Frames) == 1
	if sm.cycleAsSingle {
		sm.ticksPerCycle = 200
	} else {
		sm.ticksPerCycle = 0
	}
}

func (s *stateIdle) update(sm *StateMachine) {
	if checkHunger(sm) {
		return
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		sm.SetState(&stateDrag{})
		return
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		sm.SetState(&stateRClick{})
		return
	}
}

func (s *stateIdle) endAnimHook(sm *StateMachine) {
	if randBool(walkChancePct) {
		sm.SetState(&stateWalk{left: randBool(50)})
	}
}

type stateDrag struct {
	prevMouse    Vector
	winStart     Vector
	mouseStart   Vector
}

func (s *stateDrag) enter(sm *StateMachine) {
	sm.SetAnim(embeddedSprites.Idle)
	sm.cycleAsSingle = true
	sm.ticksPerCycle = 10
	s.prevMouse = globalCursorPosition()
	s.winStart = createVector(sm.x, sm.y)
	s.mouseStart = globalCursorPosition()
}

func (s *stateDrag) update(sm *StateMachine) {
	if checkHunger(sm) {
		return
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		sm.SetState(&stateIdle{})
		return
	}
	mousePos := globalCursorPosition()
	if mousePos != s.prevMouse {
		wp := s.winStart.add(mousePos.sub(s.mouseStart))
		sm.x, sm.y = wp.x, wp.y
	}
	s.prevMouse = mousePos
}

func (s *stateDrag) endAnimHook(sm *StateMachine) {
	syncWinPos(sm)
}

type stateRClick struct{}

func (s *stateRClick) enter(sm *StateMachine) {
	sm.SetAnim(embeddedSprites.Wink)
	sm.cycleAsSingle = len(embeddedSprites.Wink.Frames) == 1
	if sm.cycleAsSingle {
		sm.ticksPerCycle = 28
	}
}

func (s *stateRClick) update(sm *StateMachine) {
	if checkHunger(sm) {
		return
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		sm.SetState(&stateDrag{})
		return
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		sm.SetState(&stateRClick{})
		return
	}
}

func (s *stateRClick) endAnimHook(sm *StateMachine) {
	sm.SetState(&stateIdle{})
}

type stateHungry struct{}

func (s *stateHungry) enter(sm *StateMachine) {
	sm.SetAnim(embeddedSprites.Tired)
	sm.cycleAsSingle = len(embeddedSprites.Tired.Frames) == 1
	if sm.cycleAsSingle {
		sm.ticksPerCycle = 40
	}
}

func (s *stateHungry) update(sm *StateMachine) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		sm.SetState(&stateFeed{})
		return
	}
}

func (s *stateHungry) endAnimHook(sm *StateMachine) {}

func checkHunger(sm *StateMachine) bool {
	if durationTillHungry <= 0 {
		return false
	}
	if time.Since(sm.lastFed) < durationTillHungry {
		return false
	}
	sm.SetState(&stateHungry{})
	return true
}

type stateFeed struct{}

func (s *stateFeed) enter(sm *StateMachine) {
	sm.SetAnim(embeddedSprites.Happy)
	sm.cycleAsSingle = len(embeddedSprites.Happy.Frames) == 1
	if sm.cycleAsSingle {
		sm.ticksPerCycle = 45
	}
}

func (s *stateFeed) update(sm *StateMachine) {}

func (s *stateFeed) endAnimHook(sm *StateMachine) {
	sm.SetState(&stateIdle{})
	sm.lastFed = time.Now()
}

type stateWalk struct{ left bool }

func (s *stateWalk) enter(sm *StateMachine) {
	sm.SetAnim(embeddedSprites.Idle)
	sm.cycleAsSingle = len(embeddedSprites.Idle.Frames) == 1
	if sm.cycleAsSingle {
		sm.ticksPerCycle = 15
	}
}

func (s *stateWalk) update(sm *StateMachine) {
	if checkHunger(sm) {
		return
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		sm.SetState(&stateDrag{})
		return
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		sm.SetState(&stateRClick{})
		return
	}
	if s.left {
		sm.x--
	} else {
		sm.x++
	}
}

func (s *stateWalk) endAnimHook(sm *StateMachine) {
	syncWinPos(sm)
	if randBool(stopChancePct) {
		sm.SetState(&stateIdle{})
	}
}
