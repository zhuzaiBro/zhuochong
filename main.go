//go:build !ebitenpet

package main

import (
	"log"
	"net/http"
)

func main() {
	MountHTTPHandlers()
	log.Println("Starting server on http://localhost:8709")
	log.Println("桌面精灵（Ebitengine + chiikawa-sprites）: go run -tags ebitenpet .")
	log.Println("数字人播报建议 DOUYIN_MONITOR_SERVER_TTS=0，避免与本机 say 重叠")
	log.Fatal(http.ListenAndServe(":8709", nil))
}
