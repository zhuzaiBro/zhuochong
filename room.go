package main

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"

	_ "encoding/json"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	dyproto "io.github.javpower/douyin-monitor/protobuf"
	"io.github.javpower/douyin-monitor/wssign"
)

type Room struct {
	// 房间地址
	Url string

	Ttwid string

	RoomStore string

	RoomId string

	WebRoomId string

	RoomTitle string

	wsConnect *websocket.Conn

	callback string
}

func NewRoom(u string, callback string) (*Room, error) {
	var uu = "https://live.douyin.com/" + u
	h := map[string]string{
		"accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36",
		"cookie":     "__ac_nonce=0638733a400869171be51",
	}
	req, err := http.NewRequest("GET", uu, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	for k, v := range h {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer res.Body.Close()
	data := res.Cookies()
	var ttwid string
	for _, c := range data {
		if c.Name == "ttwid" {
			ttwid = c.Value
			break
		}
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	resText := string(body)
	re := regexp.MustCompile(`roomId\\":\\"(\d+)\\"`)
	match := re.FindStringSubmatch(resText)
	if match == nil || len(match) < 2 {
		log.Println("No match found")
		return nil, err
	}
	liveRoomId := match[1]
	return &Room{
		Url:       uu,
		Ttwid:     ttwid,
		RoomId:    liveRoomId,
		WebRoomId: u,
		callback:  callback,
	}, nil
}

func randomPushID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	n := binary.BigEndian.Uint64(b[:]) % 9_000_000_000_000_000_000
	n += 1_000_000_000_000_000_000
	return fmt.Sprintf("%d", n)
}

func (r *Room) buildDouyinWssURL() (string, error) {
	pushID := randomPushID()
	fetchMs := time.Now().UnixMilli()
	firstMs := fetchMs - 50
	var fhBuf [8]byte
	_, _ = rand.Read(fhBuf[:])
	fh := int64(binary.BigEndian.Uint64(fhBuf[:]) & ((1 << 62) - 1))
	var wrdsBuf [8]byte
	_, _ = rand.Read(wrdsBuf[:])
	wrdsV := int64(binary.BigEndian.Uint64(wrdsBuf[:]) & ((1 << 62) - 1))

	q := url.Values{}
	q.Set("app_name", "douyin_web")
	q.Set("version_code", "180800")
	q.Set("webcast_sdk_version", "1.0.14-beta.0")
	q.Set("update_version_code", "1.0.14-beta.0")
	q.Set("compress", "gzip")
	q.Set("device_platform", "web")
	q.Set("cookie_enabled", "true")
	q.Set("screen_width", "1536")
	q.Set("screen_height", "864")
	q.Set("browser_language", "zh-CN")
	q.Set("browser_platform", "Win32")
	q.Set("browser_name", "Mozilla")
	q.Set("browser_version", "5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	q.Set("browser_online", "true")
	q.Set("tz_name", "Asia/Shanghai")
	q.Set("cursor", fmt.Sprintf("d-1_u-1_fh-%d_t-%d_r-1", fh, fetchMs))
	q.Set("internal_ext", fmt.Sprintf(
		"internal_src:dim|wss_push_room_id:%s|wss_push_did:%s|first_req_ms:%d|fetch_time:%d|seq:1|wss_info:0-%d-0-0|wrds_v:%d",
		r.RoomId, pushID, firstMs, fetchMs, fetchMs, wrdsV,
	))
	q.Set("host", "https://live.douyin.com")
	q.Set("aid", "6383")
	q.Set("live_id", "1")
	q.Set("did_rule", "3")
	q.Set("endpoint", "live_pc")
	q.Set("support_wrds", "1")
	q.Set("user_unique_id", pushID)
	q.Set("im_path", "/webcast/im/fetch/")
	q.Set("identity", "audience")
	q.Set("need_persist_msg_count", "15")
	q.Set("insert_task_id", "")
	q.Set("live_reason", "")
	q.Set("room_id", r.RoomId)
	q.Set("heartbeatDuration", "0")

	base := "wss://webcast100-ws-web-lq.douyin.com/webcast/im/push/v2/?" + q.Encode()
	sig, err := wssign.SignatureForWssURL(base)
	if err != nil {
		return "", err
	}
	return base + "&signature=" + sig, nil
}

func (r *Room) Connect() error {
	wsUrl, err := r.buildDouyinWssURL()
	if err != nil {
		return err
	}
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
	h := http.Header{}
	h.Set("Cookie", "ttwid="+r.Ttwid)
	h.Set("User-Agent", ua)
	h.Set("Origin", "https://live.douyin.com")
	h.Set("Referer", "https://live.douyin.com/"+r.WebRoomId)

	wsConn, wsResp, err := websocket.DefaultDialer.Dial(wsUrl, h)
	if err != nil {
		if wsResp != nil {
			body, _ := io.ReadAll(io.LimitReader(wsResp.Body, 2048))
			_ = wsResp.Body.Close()
			log.Printf("douyin WSS handshake failed: HTTP %d body=%q", wsResp.StatusCode, string(body))
		}
		return err
	}
	log.Println(wsResp.StatusCode)
	r.wsConnect = wsConn
	go r.read()
	go r.send()
	return nil
}

func (r *Room) read() {
	for {
		_, data, err := r.wsConnect.ReadMessage()
		if err != nil {
			return
			//panic(err.Error())
		}
		var msgPack dyproto.PushFrame
		_ = proto.Unmarshal(data, &msgPack)
		// log.Println(msgPack.LogId)
		decompressed, _ := degzip(msgPack.Payload)
		var payloadPackage dyproto.Response
		_ = proto.Unmarshal(decompressed, &payloadPackage)
		if payloadPackage.NeedAck {
			r.sendAck(msgPack.LogId, payloadPackage.InternalExt)
		}
		// log.Println(len(payloadPackage.MessagesList))
		for _, msg := range payloadPackage.MessagesList {
			switch msg.Method {
			case "WebcastChatMessage":
				parseChatMsg(msg.Payload, r.WebRoomId, r.callback)
			case "WebcastGiftMessage":
				parseGiftMsg(msg.Payload, r.WebRoomId, r.callback)
			case "WebcastLikeMessage":
				parseLikeMsg(msg.Payload, r.WebRoomId, r.callback)
			case "WebcastMemberMessage":
				parseEnterMsg(msg.Payload, r.WebRoomId, r.callback)
			case "WebcastControlMessage":
				parseControlMsg(msg.Payload, r.WebRoomId, r.callback)
			}
		}
	}
}

func (r *Room) send() {
	for {
		pingPack := &dyproto.PushFrame{
			PayloadType: "bh",
		}
		data, _ := proto.Marshal(pingPack)
		err := r.wsConnect.WriteMessage(websocket.BinaryMessage, data)
		if err != nil {
			//panic(err.Error())
			return
		}
		// log.Println("发送心跳")
		time.Sleep(time.Second * 10)
	}
}

func (r *Room) sendAck(logId uint64, iExt string) {
	ackPack := &dyproto.PushFrame{
		LogId:       logId,
		PayloadType: iExt,
	}
	data, _ := proto.Marshal(ackPack)
	err := r.wsConnect.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		panic(err.Error())
	}
	// log.Println("发送 ack 包")
}

func degzip(data []byte) ([]byte, error) {
	b := bytes.NewReader(data)
	var out bytes.Buffer
	r, err := gzip.NewReader(b)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(&out, r)
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func parseChatMsg(msg []byte, webRoomId string, callback string) {
	var chatMsg dyproto.ChatMessage
	_ = proto.Unmarshal(msg, &chatMsg)
	log.Printf("[房间] %s : [弹幕] %s : %s\n", webRoomId, chatMsg.User.NickName, chatMsg.Content)
	EnqueueSpeech(fmt.Sprintf("%s说，%s", chatMsg.User.NickName, chatMsg.Content))
	data := map[string]interface{}{
		"webRoomId": webRoomId,
		"data":      chatMsg,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}
	Wss.BroadcastMessage(jsonData)
}

func parseGiftMsg(msg []byte, webRoomId string, callback string) {
	var giftMsg dyproto.GiftMessage
	_ = proto.Unmarshal(msg, &giftMsg)
	log.Printf("[房间] %s : [礼物] %s : %s * %d \n", webRoomId, giftMsg.User.NickName, giftMsg.Gift.Name, giftMsg.ComboCount)
	data := map[string]interface{}{
		"webRoomId": webRoomId,
		"remark":    "礼物",
		"data":      giftMsg,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}
	Wss.BroadcastMessage(jsonData)

}

func parseLikeMsg(msg []byte, webRoomId string, callback string) {
	var likeMsg dyproto.LikeMessage
	_ = proto.Unmarshal(msg, &likeMsg)
	log.Printf("[房间] %s : [点赞] %s 点赞 * %d \n", webRoomId, likeMsg.User.NickName, likeMsg.Count)
	data := map[string]interface{}{
		"webRoomId": webRoomId,
		"remark":    "点赞",
		"data":      likeMsg,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}
	Wss.BroadcastMessage(jsonData)
}

func parseEnterMsg(msg []byte, webRoomId string, callback string) {
	var enterMsg dyproto.MemberMessage
	_ = proto.Unmarshal(msg, &enterMsg)
	log.Printf("[房间] %s : [入场] %s 直播间\n", webRoomId, enterMsg.User.NickName)
	data := map[string]interface{}{
		"webRoomId": webRoomId,
		"remark":    "入场",
		"data":      enterMsg,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}
	Wss.BroadcastMessage(jsonData)
}
func parseControlMsg(msg []byte, webRoomId string, callback string) {
	var enterMsg dyproto.MemberMessage
	_ = proto.Unmarshal(msg, &enterMsg)
	log.Printf("[房间] %s : [直播间状态变更] %s 直播间\n", webRoomId, enterMsg.Action)
	if room, ok := rooms[webRoomId]; ok {
		room.Close()
		delete(rooms, webRoomId)
	}
	data := map[string]interface{}{
		"webRoomId": webRoomId,
		"remark":    "直播间状态变更",
		"data":      enterMsg,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}
	Wss.BroadcastMessage(jsonData)
}
