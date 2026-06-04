# 陆沉桌宠 - 切图资源包

## 目录结构

```
lucheng-sprites/
├── idle/              # 正常办公状态（待机）
│   ├── frame1.png     # 坐姿，正视前方，表情平静
│   ├── frame2.png     # 轻微眨眼，呼吸效果帧
│   └── frame3.png     # 回到 frame1 姿势，形成循环
├── headpat/           # 摸头反应状态
│   ├── frame1.png     # 被摸头前，正常表情
│   ├── frame2.png     # 被摸头中，眼睛微闭，害羞表情
│   └── frame3.png     # 摸头后，微笑，脸微红
├── happy/             # 开心表情状态
│   ├── frame1.png     # 开心微笑，眼睛弯弯
│   ├── frame2.png     # 跳起，挥手
│   └── frame3.png     # 落地，仍然开心
├── talking/           # 对话状态
│   ├── frame1.png     # 嘴巴张开，说话状态
│   ├── frame2.png     # 嘴巴微闭，眨眼
│   └── frame3.png     # 嘴巴再次张开，继续说话
├── water_reminder/    # 喝水提醒状态
│   ├── frame1.png     # 举起水杯，提醒动作
│   ├── frame2.png     # 指向水杯，表情认真
│   └── frame3.png     # 做出喝水动作示范
└── coffee/            # 喝咖啡状态
    ├── frame1.png     # 拿起咖啡杯，准备喝
    ├── frame2.png     # 喝咖啡中，表情满足
    └── frame3.png     # 喝完，满足表情，竖起大拇指
```

## 角色特征

陆沉（Lucheng）Q版纸片人形象：
- **发型**：深棕色蓬松短发，发丝飘逸
- **眼睛**：红色眼睛，眉毛较粗，表情冷淡
- **上身**：白色宽松长袖衬衫 + 黑色背心/马甲 + 红色领带
- **下身**：酒红色/深红色长裤
- **配件**：黑色手套、黑色靴子
- **风格**：Q版 Chibi 风格，厚黑线条，扁平着色

## 图像规格

| 属性 | 规格 |
|------|------|
| 格式 | PNG（RGBA，透明背景） |
| 分辨率 | 1248 × 1248 像素 |
| 色彩模式 | RGBA |
| 背景 | 完全透明 |

## 使用方法

### 直接使用 PNG

每个状态文件夹中的 PNG 文件可直接用于：
- 游戏引擎（Unity、Godot 等）的精灵动画
- 视频编辑软件的帧序列
- 自定义桌宠程序的素材

### 生成 GIF

使用 `process_sprites.py` 脚本将 PNG 关键帧合成为 GIF 动图：

```bash
python3 process_sprites.py
```

生成的 GIF 文件将保存至 `lucheng-gifs/` 目录。

### 统一关键帧质量

替换或新增 PNG 后，可在仓库根目录运行：

```bash
go run ./scripts/enhance_pet_keyframes -root lucheng-sprites -size 1248 -key-white
```

脚本会把所有动作帧统一为 `1248x1248` RGBA 透明 PNG，并使用 Catmull-Rom 高质量重采样。检查背景边缘透明度：

```bash
go run ./scripts/enhance_pet_keyframes -root lucheng-sprites -check-only
```

若某组动作贴到画布边缘，可给该目录加透明安全边：

```bash
go run ./scripts/enhance_pet_keyframes -root lucheng-sprites/headpat -size 1248 -padding 64 -key-white
```

若低清关键帧出现灰色脏遮罩，可从干净高分辨率帧重新衍生前 6 张：

```bash
go run ./scripts/derive_lucheng_keyframes
go run ./scripts/enhance_pet_keyframes -root lucheng-sprites -size 1248 -key-white -clean-alpha -alpha-cutoff 24 -alpha-radius 0
```

### 扩展新状态

1. 在 `lucheng-sprites/` 下新建文件夹（如 `sleep/`）
2. 放入 PNG 关键帧（frame1.png, frame2.png, frame3.png）
3. 运行 `process_sprites.py` 重新生成 GIF
4. 在 React 应用中添加对应状态配置

## 版权说明

本资源包基于 AI 生成，角色形象参考陆沉（Lucheng）纸片人设计。
