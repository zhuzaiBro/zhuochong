.PHONY: start run

# 桌面直播精灵。追加参数示例：
#   make start ARGS='-size 0.3 -sleep-idle 10 -talking-sec 3.5 -gift-sec 2.8'
# 仅连已有监控、不开本进程服务：
#   make start ARGS='-no-server -ws=ws://127.0.0.1:8709/ws'
ARGS ?=

start run:
	DOUYIN_MONITOR_SERVER_TTS=0 go run -tags ebitenpet . -size 0.3 $(ARGS)
