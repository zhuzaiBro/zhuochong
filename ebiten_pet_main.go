//go:build ebitenpet

package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// 参考 https://github.com/nhanb/shark：Ebitengine 无边框透明悬浮窗 + 本仓库 /ws 实时互动。

func main() {
	addr := flag.String("addr", ":8709", "HTTP 监听地址（与默认弹幕服务一致）；-no-server 时忽略")
	wsURL := flag.String("ws", "", "WebSocket URL，默认由 -addr 推导为 ws://127.0.0.1:8709/ws")
	noServer := flag.Bool("no-server", false, "不启动 HTTP+WS 服务，仅连接已有进程的 -ws")
	x := flag.Int("x", -1, "窗口初始 X（默认：Dock/任务栏右上）")
	y := flag.Int("y", -1, "窗口初始 Y（默认：Dock/任务栏右上）")
	size := flag.Float64("size", 0.1, "立绘缩放（相对 lucheng-sprites PNG 像素尺寸）")
	hungrySec := flag.Int64("hungry", 3600, "多少秒后进入「疲倦/饥饿」状态；0 关闭")
	walkPct := flag.Int("walk", 0, "自动溜达概率 %；0 为关闭（默认：精灵不自己横向移动）")
	stopPct := flag.Int("stop", 40, "溜达时每周期停下的概率 %")
	officeIdle := flag.Float64("office-idle", 120, "无互动多少秒后进入办公状态（秒，弹幕/礼物/点赞会刷新计时）；0 关闭")
	sleepIdle := flag.Float64("sleep-idle", 0, "已废弃，请用 -office-idle；若 >0 则覆盖 -office-idle")
	talkingSec := flag.Float64("talking-sec", 3.5, "收到弹幕后 talking 动画持续时长（秒）")
	giftSec := flag.Float64("gift-sec", 2.8, "收到礼物后 happy 动画持续时长（秒）")
	flag.Parse()

	idleSec := *officeIdle
	if *sleepIdle > 0 {
		idleSec = *sleepIdle
	}
	if idleSec <= 0 {
		moodOfficeIdle = 0
	} else {
		moodOfficeIdle = time.Duration(idleSec * float64(time.Second))
	}
	moodTalkingDur = time.Duration(*talkingSec * float64(time.Second))
	moodHappyDur = time.Duration(*giftSec * float64(time.Second))

	if *hungrySec <= 0 {
		durationTillHungry = 0
	} else {
		durationTillHungry = time.Duration(*hungrySec) * time.Second
	}
	walkChancePct = *walkPct
	stopChancePct = *stopPct

	var finalWS string
	if strings.TrimSpace(*wsURL) != "" {
		finalWS = strings.TrimSpace(*wsURL)
	} else if *noServer {
		log.Fatal("使用 -no-server 时必须指定 -ws，例如 -ws=ws://127.0.0.1:8709/ws")
	} else {
		finalWS = deriveWSURL(*addr)
	}

	if !*noServer {
		MountHTTPHandlers()
		go func() {
			log.Println("Ebitengine 桌面精灵：HTTP+WS", *addr, "（控制直播间 API 不变）")
			if err := http.ListenAndServe(*addr, nil); err != nil {
				log.Fatal(err)
			}
		}()
		time.Sleep(120 * time.Millisecond)
	}

	log.Println("悬浮窗：Ebitengine / 透明背景 / 置顶；连接", finalWS)
	log.Println("建议 DOUYIN_MONITOR_SERVER_TTS=0，仅由本进程 EnqueuePetSpeech 播报，避免与 room 重复")
	if serverTTSEnabled() {
		if !*noServer {
			log.Println("语音：未关闭 room 侧 say 时，同一条弹幕会播两次（先/后听到默认与「小八」两种音色）。请执行 export DOUYIN_MONITOR_SERVER_TTS=0 后重新启动本程序。")
		} else {
			log.Println("语音：若另开终端用默认入口监听房间，请在那一侧也设 DOUYIN_MONITOR_SERVER_TTS=0，否则弹幕会在那里与精灵各播一遍。")
		}
	}

	game := newLivePetGame(*size, *x, *y)
	go runLivePetWS(finalWS, game)

	opts := defaultEbitenRunOptions()
	if err := runEbitenPet(game, opts); err != nil {
		log.Fatal(err)
	}
}

func deriveWSURL(listenAddr string) string {
	host := "127.0.0.1"
	port := "8709"
	if strings.HasPrefix(listenAddr, ":") {
		port = listenAddr[1:]
	} else {
		h, p, err := net.SplitHostPort(listenAddr)
		if err == nil {
			if h != "" && h != "0.0.0.0" && h != "::" {
				host = h
			}
			port = p
		}
	}
	return "ws://" + host + ":" + port + "/ws"
}
