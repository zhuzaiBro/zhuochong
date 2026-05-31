package main

import (
	"net/http"

	websocket2 "io.github.javpower/douyin-monitor/websocket"
)

var Wss *websocket2.WebSocketServer

// MountHTTPHandlers 注册 HTTP 与 WS（默认 ServeMux）。桌面精灵与纯服务端共用。
func MountHTTPHandlers() {
	http.HandleFunc("/api", handleAPI)
	Wss = websocket2.NewWebSocketServer()
	http.HandleFunc("/ws", Wss.HandleWebSocket)
}
