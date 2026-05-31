# Chiikawa Live Pet - Sprite Assets

这是 Chiikawa 直播精灵的切图资源包，按照 shark 项目的目录结构组织。

## 目录结构

```
chiikawa-sprites/
├── idle/              # 待机状态
├── talking/           # 说话状态
├── happy/             # 开心状态
├── confused/          # 困惑状态
├── sleeping/          # 睡眠状态
├── angry/             # 生气状态
├── shy/               # 害羞状态
├── surprised/         # 惊讶状态
├── tired/             # 疲倦状态
└── wink/              # 眨眼状态
```

## 使用说明

每个状态文件夹内包含该状态的 PNG 图像，所有图像均为**透明背景**，可直接用于：

- OBS 浏览器源集成
- 直播精灵应用
- 其他动画应用

## 图像规格

- **格式**: PNG (透明背景)
- **分辨率**: 192x192 像素
- **用途**: 直播精灵动画帧

## 多帧命名

每帧文件名为 `frame01.png`、`frame02.png`…（两位数字），便于排序。基准立绘放在 `frame01.png`；可由仓库内脚本从 `frame01` 衍生中间帧：

```bash
go run ./scripts/gen_chiikawa_keyframes
```

（脚本仅做轻微缩放与位移合成「呼吸 / 说话挤压 / 抖动」等过渡感，并非重绘；改画风请替换 `frame01` 后重新跑脚本覆盖 `frame02`…`frame07`。）

## 集成方式

### 用于 OBS

1. 在 OBS 中添加浏览器源
2. 指向直播精灵应用的 URL
3. 设置分辨率和位置

### 用于其他应用

可以按照相同的目录结构组织，每个状态文件夹可包含多个帧以实现逐帧动画。

---

**Created**: 2026-05-07  
**Project**: Chiikawa Live Pet  
**Reference**: Based on shark project structure
