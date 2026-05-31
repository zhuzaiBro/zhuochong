package main

import (
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"unicode/utf8"
)

const (
	speechMaxQueued = 128
	speechMaxRunes  = 400
)

var speechQ = newSpeechQueue()

func init() {
	go speechQ.run()
}

type speechQueue struct {
	mu    sync.Mutex
	cond  *sync.Cond
	items []speechItem
}

type speechItem struct {
	text string
	pet  bool // true：来自 EnqueuePetSpeech，可走专用音色
}

func newSpeechQueue() *speechQueue {
	s := &speechQueue{items: make([]speechItem, 0, 32)}
	s.cond = sync.NewCond(&s.mu)
	return s
}

// serverTTSEnabled 为 false 时跳过 room 内 EnqueueSpeech 的本机 say。
// 环境变量 DOUYIN_MONITOR_SERVER_TTS：设为 0 / false / off / no 则关闭。
func serverTTSEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("DOUYIN_MONITOR_SERVER_TTS")))
	if v == "" {
		return true
	}
	switch v {
	case "0", "false", "off", "no":
		return false
	default:
		return true
	}
}

// EnqueueSpeech 将一行待播报文案加入队列尾部（FIFO）。队列满时丢弃最旧条目以免积压。
// 若设置 DOUYIN_MONITOR_SERVER_TTS=0 则不入队（便于与浏览器 Web Speech 二选一）。
func EnqueueSpeech(line string) {
	enqueueSpeechMaybe(line, true, false)
}

// EnqueuePetSpeech 供 Ebitengine 桌面精灵使用：不受 DOUYIN_MONITOR_SERVER_TTS 关闭影响。
// 与 room 内 EnqueueSpeech 同时开启会重复播报，运行 ebiten 模式时请设 DOUYIN_MONITOR_SERVER_TTS=0。
func EnqueuePetSpeech(line string) {
	enqueueSpeechMaybe(line, false, true)
}

func enqueueSpeechMaybe(line string, respectServerTTSDisable bool, pet bool) {
	if respectServerTTSDisable && !serverTTSEnabled() {
		return
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	if n := utf8.RuneCountInString(line); n > speechMaxRunes {
		line = string([]rune(line)[:speechMaxRunes]) + "……"
	}

	speechQ.mu.Lock()
	for len(speechQ.items) >= speechMaxQueued {
		speechQ.items = speechQ.items[1:]
	}
	speechQ.items = append(speechQ.items, speechItem{text: line, pet: pet})
	speechQ.mu.Unlock()
	speechQ.cond.Signal()
}

func (s *speechQueue) run() {
	for {
		s.mu.Lock()
		for len(s.items) == 0 {
			s.cond.Wait()
		}
		item := s.items[0]
		s.items = s.items[1:]
		s.mu.Unlock()

		if err := speakLine(item.text, item.pet); err != nil {
			log.Printf("语音播报失败: %v", err)
		}
	}
}

var speechBackendWarn sync.Once

// darwinSayVoice 返回 macOS say -v 的声音名；空表示系统默认。
//
// 桌面精灵可用 DOUYIN_MONITOR_PET_SAY_VOICE（或别名 DOUYIN_MONITOR_HACHIWARE_SAY_VOICE「小八」主题）
// 指定与 room 弹幕不同的音色；未设置时宠物侧会回退到 DOUYIN_MONITOR_SAY_VOICE。
func darwinSayVoice(pet bool) string {
	if pet {
		if v := strings.TrimSpace(os.Getenv("DOUYIN_MONITOR_PET_SAY_VOICE")); v != "" {
			return v
		}
		if v := strings.TrimSpace(os.Getenv("DOUYIN_MONITOR_HACHIWARE_SAY_VOICE")); v != "" {
			return v
		}
	}
	return strings.TrimSpace(os.Getenv("DOUYIN_MONITOR_SAY_VOICE"))
}

func speakLine(text string, pet bool) error {
	switch runtime.GOOS {
	case "darwin":
		args := []string{}
		if v := darwinSayVoice(pet); v != "" {
			args = append(args, "-v", v)
		}
		args = append(args, text)
		return exec.Command("say", args...).Run()
	default:
		if path, err := exec.LookPath("espeak"); err == nil {
			return exec.Command(path, text).Run()
		}
		speechBackendWarn.Do(func() {
			log.Println("语音播报：非 macOS 且未找到 espeak，后续弹幕将不再尝试播报")
		})
		return nil
	}
}
