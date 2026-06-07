.PHONY: start run package package-darwin package-windows package-all

# 桌面直播精灵。追加参数示例：
#   make start ARGS='-size 0.3 -office-idle 120 -talking-sec 3.5 -gift-sec 2.8'
# 仅连已有监控、不开本进程服务：
#   make start ARGS='-no-server -ws=ws://127.0.0.1:8709/ws'
ARGS ?=

DIST ?= dist
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

start run:
	DOUYIN_MONITOR_SERVER_TTS=0 go run -tags ebitenpet . -size 0.2 $(ARGS)

# 打包当前平台（macOS 打 .app + 可执行文件，Windows 打 .exe）
package:
	@chmod +x scripts/package_pet.sh
	@case "$$(uname -s)" in \
		Darwin) DIST=$(DIST) VERSION=$(VERSION) ./scripts/package_pet.sh darwin ;; \
		MINGW*|MSYS*|CYGWIN*) DIST=$(DIST) VERSION=$(VERSION) ./scripts/package_pet.sh windows ;; \
		*) echo "请使用 make package-darwin 或 make package-windows"; exit 1 ;; \
	esac

package-darwin:
	@chmod +x scripts/package_pet.sh
	DIST=$(DIST) VERSION=$(VERSION) ./scripts/package_pet.sh darwin

package-windows:
	@chmod +x scripts/package_pet.sh
	DIST=$(DIST) VERSION=$(VERSION) ./scripts/package_pet.sh windows

# macOS 上可同时打 darwin + windows；Windows/Linux 上请分别执行 package-darwin / package-windows
package-all:
	@chmod +x scripts/package_pet.sh
	DIST=$(DIST) VERSION=$(VERSION) ./scripts/package_pet.sh all
