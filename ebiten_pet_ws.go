//go:build ebitenpet

package main

import (
	"bytes"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	dyproto "io.github.javpower/douyin-monitor/protobuf"
)

type wsPush struct {
	Remark string          `json:"remark"`
	Data   json.RawMessage `json:"data"`
}

func runLivePetWS(url string, game *livePetGame) {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	for {
		c, _, err := dialer.Dial(url, nil)
		if err != nil {
			log.Println("ebiten 精灵 WebSocket:", err)
			game.setWSState(false, err.Error())
			time.Sleep(2 * time.Second)
			continue
		}
		game.setWSState(true, "")
		game.setSubtitle("已连接 · 等待弹幕…")
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("ebiten 精灵 WebSocket 读:", err)
				break
			}
			if strings.TrimSpace(string(msg)) == "OK" {
				continue
			}
			handlePetPush(msg, game)
		}
		_ = c.Close()
		game.setWSState(false, "连接断开")
		game.setSubtitle("重连中…")
		time.Sleep(time.Second)
	}
}

func handlePetPush(raw []byte, game *livePetGame) {
	var env wsPush
	if err := json.Unmarshal(raw, &env); err != nil {
		return
	}
	switch env.Remark {
	case "":
		if len(bytes.TrimSpace(env.Data)) == 0 {
			return
		}
		var chat dyproto.ChatMessage
		if err := json.Unmarshal(env.Data, &chat); err != nil {
			return
		}
		text := strings.TrimSpace(chat.GetContent())
		if text == "" {
			return
		}
		nick := nickOrAnon(chat.GetUser())
		line := "「" + nick + "」：" + text
		game.startTalking()
		game.setBubble(line)
		game.setSubtitle(line)
		EnqueuePetSpeech(line)

	case "礼物":
		var gm dyproto.GiftMessage
		if err := json.Unmarshal(env.Data, &gm); err != nil {
			return
		}
		nick := nickOrAnon(gm.GetUser())
		gname := ""
		if gi := gm.GetGift(); gi != nil {
			gname = strings.TrimSpace(gi.GetName())
		}
		line := nick + "送出礼物"
		if gname != "" {
			line += " " + gname
		}
		if n := gm.GetComboCount(); n > 1 {
			line += " ×" + strconv.FormatUint(uint64(n), 10)
		}
		game.startHappyGift()
		game.setSubtitle(line)
		EnqueuePetSpeech(line)

	case "点赞":
		// 不语音播报、不刷窗口标题；仅保留轻微互动反馈。
		game.bumpInteraction()
		game.triggerExcited(500 * time.Millisecond)

	case "入场":
		// 入场不播报、不刷屏标题；room 仍会向 WebSocket 推送，其它客户端可用。

	case "直播间状态变更":
		game.bumpInteraction()
		game.triggerExcited(300 * time.Millisecond)
		game.setSubtitle("直播间状态变更")
	}
}

func nickOrAnon(u *dyproto.User) string {
	if u == nil {
		return "观众"
	}
	if n := strings.TrimSpace(u.GetNickName()); n != "" {
		return n
	}
	return "观众"
}
