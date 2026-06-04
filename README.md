# 📺 抖音弹幕监听器

## 😎介绍及配置

### 介绍

基于系统代理抓包打造的抖音弹幕服务推送程序，它能够获取浏览器直播间上抖音弹幕数据，它可以监听**弹幕**，**点赞**，**送礼**，**进入直播间**，等系列消息，你可使用它做自己的直播间数据分析，以及弹幕互动游戏，语音播报等。


### 推送数据格式

弹幕数据由WebSocket服务进行分发，使用Json格式进行推送，可前往[ws在线测试](http://wstool.jackxiang.com/)网站，连接 ws://127.0.0.1:8709/ws 进行测试

### 监听控制
通过HTTP服务API进行控制，POST请求 http://127.0.0.1:8709/api
参数form-data
- roomId:42643374680 (web房间号)
- command:close （connect表示监听、close表示取消监听）

### 桌面宠物（直播悬浮窗）

使用 [Ebitengine](https://ebitengine.org/)，行为与素材布局参考 [shark](https://github.com/nhanb/shark)（溜达 / 拖拽 / 右键眨眼 / 长时间未喂食进入疲倦态，右键喂食恢复）。立绘为仓库内 **`lucheng-sprites/`**（PNG 透明底，嵌入可执行文件，不依赖浏览器页面）。

无边框、透明背景、置顶，通过 **`/ws`** 接收弹幕 / 礼物 / 点赞等并播报。与 **`room` 内语音队列**同时开时请将 `DOUYIN_MONITOR_SERVER_TTS=0`，仅由精灵侧 `EnqueuePetSpeech` 走本机 `say`，避免重复播报。

macOS 上可为精灵单独指定 `say -v` 音色：环境变量 **`DOUYIN_MONITOR_PET_SAY_VOICE`**；与 **`DOUYIN_MONITOR_HACHIWARE_SAY_VOICE`**（小八 / 哈奇主题，二选一即可，等价于 PET）优先于通用的 **`DOUYIN_MONITOR_SAY_VOICE`**。本机可用声音列表：`say -v '?'`

```bash
DOUYIN_MONITOR_SERVER_TTS=0 go run -tags ebitenpet .
```

仅开精灵、连已在跑的监控服务：

```bash
DOUYIN_MONITOR_SERVER_TTS=0 go run -tags ebitenpet . -no-server -ws=ws://127.0.0.1:8709/ws
```

可选参数：`-size` 缩放，`-x` / `-y` 初始窗口坐标，`-addr` HTTP 监听地址，`-hungry` 秒数（`0` 关闭饥饿态），`-walk` / `-stop` 溜达概率（百分比；**默认 `-walk 0`** 关闭横向自动移动）。**无互动睡觉 / 弹幕 talking / 礼物 happy**：`-sleep-idle`（秒，`0` 关）、`-talking-sec`、`-gift-sec`。弹幕会在立绘右侧显示气泡（需本机中文字体；可设 **`DOUYIN_MONITOR_BUBBLE_FONT`** 指向 `.ttf` / `.otf` / `.ttc`）。

### 桌宠素材增强

`lucheng-sprites/` 已统一为 `1248x1248` RGBA 透明 PNG。若替换或新增关键帧后需要重新统一质量，可运行：

```bash
go run ./scripts/enhance_pet_keyframes -root lucheng-sprites -size 1248 -key-white
```

脚本会用 Catmull-Rom 高质量重采样，并只从画布边缘移除近白背景，避免误抠角色衣服等内部白色区域。可用 `-check-only` 检查背景边缘是否透明；若角色贴到画布边缘，可给单个动作目录加透明安全边，例如：

```bash
go run ./scripts/enhance_pet_keyframes -root lucheng-sprites/headpat -size 1248 -padding 64 -key-white
```

若前 6 张低清关键帧出现灰色脏遮罩，可从每组动作的干净高分辨率帧重新衍生：

```bash
go run ./scripts/derive_lucheng_keyframes
go run ./scripts/enhance_pet_keyframes -root lucheng-sprites -size 1248 -key-white -clean-alpha -alpha-cutoff 24 -alpha-radius 0
```


## ⚖️免责声明

+ 本程序仅供学习参考，不得用于商业用途，不得用于恶意搜集他人直播间用户信息!
