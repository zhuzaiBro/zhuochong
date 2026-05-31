package wssign

import (
	"crypto/md5"
	_ "embed"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/dop251/goja"
)

//go:embed sign.js
var signJSSource string

// Sign.js 生成逻辑来自 DouyinLiveWebFetcher（MIT），用于 frontier signature。

var (
	initOnce sync.Once
	vm       *goja.Runtime
	initErr  error
	vmMu     sync.Mutex
)

func initVM() {
	initOnce.Do(func() {
		vm = goja.New()
		_, initErr = vm.RunString(signJSSource)
	})
}

func md5FrontierInput(wss string) (string, error) {
	u, err := url.Parse(wss)
	if err != nil {
		return "", err
	}
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", err
	}
	get := func(k string) string {
		if v := q[k]; len(v) > 0 {
			return v[0]
		}
		return ""
	}
	// 与 DouyinLiveWebFetcher liveMan.py generateSignature 中 params 顺序一致
	keys := []string{
		"live_id", "aid", "version_code", "webcast_sdk_version",
		"room_id", "sub_room_id", "sub_channel_id", "did_rule",
		"user_unique_id", "device_platform", "device_type", "ac",
		"identity",
	}
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, get(k)))
	}
	param := strings.Join(parts, ",")
	sum := md5.Sum([]byte(param))
	return hex.EncodeToString(sum[:]), nil
}

// SignatureForWssURL 根据完整 WSS URL（不含 signature 参数）计算 signature 查询值。
func SignatureForWssURL(wss string) (string, error) {
	initVM()
	if initErr != nil {
		return "", initErr
	}
	md5hex, err := md5FrontierInput(wss)
	if err != nil {
		return "", err
	}
	vmMu.Lock()
	defer vmMu.Unlock()
	fn, ok := goja.AssertFunction(vm.Get("get_sign"))
	if !ok {
		return "", fmt.Errorf("get_sign is not a function")
	}
	v, err := fn(goja.Undefined(), vm.ToValue(md5hex))
	if err != nil {
		return "", err
	}
	return v.String(), nil
}
