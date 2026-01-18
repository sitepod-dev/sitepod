# SitePod - 品牌规范

**版本**: v1.0
**日期**: 2025-01-15

---

## 1. 品牌标识

### 1.1 名称与副标题

**主名称**
```
SitePod
```

**英文副标题** (任选一)
```
SitePod — Self-hosted static releases
SitePod — Static release & rollback platform
```

**中文副标题**
```
SitePod — 自托管静态站点发布与回滚平台
```

**使用规则:**
- 名称与副标题必须同时出现在首次展示场景
- GitHub README 首行、官网 Hero 区、CLI `--help` 输出
- 后续引用可只用 `SitePod`

### 1.2 首页 Hero 文案

```
SitePod — Self-hosted static releases

Release with one command.
Rollback in seconds with immutable versions.
```

中文版:
```
SitePod — 自托管静态站点发布与回滚平台

一条命令发布，秒级回滚。
把你的站点变成不可变的 Pod。
```

---

## 2. 核心概念定义

### 2.1 什么是 Pod

> **Pod = 一个不可变的站点快照 (manifest + blobs)，环境只是一个可变的 ref 指针。**

这是 SitePod 区别于托管型平台的核心设计：

| | 传统平台 | SitePod |
|---|----------|---------|
| 模型 | 构建 → 发布 → 覆盖 | Pod 快照 + Ref 切换 |
| 回滚 | 重新构建 | 切换指针 (< 1s) |
| 存储 | 全量复制 | 内容寻址去重 |
| 可变性 | 部署是可变的 | Pod 是不可变的 |

### 2.2 README 首段模板

```markdown
# SitePod

**Self-hosted static releases with instant rollback.**

SitePod treats every deployment as an immutable **Pod** — a content-addressed
snapshot of your site. Environments (prod, beta, preview) are just refs pointing
to pods. Switch versions in seconds, not minutes.

- 🚀 One command release: `sitepod deploy --prod`
- ⚡ Instant rollback: switch refs, not rebuild
- 👀 Preview URLs: share work-in-progress safely
- 📦 Incremental uploads: only upload what changed
- 🔒 Self-hosted: your data, your infrastructure
```

---

## 3. 命名规范

### 3.1 统一命名表

| 场景 | 命名 | 说明 |
|------|------|------|
| **域名** | `sitepod.dev` | 主域名 |
| **文档** | `docs.sitepod.dev` | 文档站 |
| **下载** | `get.sitepod.dev` | 下载页/安装脚本 |
| **GitHub Repo** | `sitepod/sitepod` | 主仓库 |
| **CLI 命令** | `sitepod` | 唯一 CLI 命令名 |
| **Docker 镜像** | `ghcr.io/sitepod-dev/sitepod` | 官方镜像 |
| **配置文件** | `sitepod.toml` | 项目配置 |
| **环境变量前缀** | `SITEPOD_` | 如 `SITEPOD_TOKEN` |

### 3.2 CLI 命令

```bash
# 安装
curl -fsSL https://get.sitepod.dev | sh

# 命令名统一为 sitepod
sitepod login
sitepod deploy
sitepod deploy --prod
sitepod rollback
sitepod preview
sitepod history
```

**注意:** 不要使用 `pod` 作为命令名，避免与 Kubernetes `kubectl get pods` 混淆。

### 3.3 Docker

```bash
# 官方镜像
docker pull ghcr.io/sitepod-dev/sitepod:latest
docker pull ghcr.io/sitepod-dev/sitepod:v1.0.0

# 运行
docker run -d \
  -p 80:80 -p 443:443 \
  -v sitepod-data:/data \
  -e SITEPOD_DOMAIN=example.com \
  ghcr.io/sitepod-dev/sitepod:latest
```

---

## 4. 视觉规范

### 4.1 设计风格：Ops-grade Minimal

**核心理念：** 冷静、工程化、可信

**关键词：** 留白、克制、系统感、可读性极强

**定位：** 像一套"运维/工程工具"而不是"花哨 SaaS"

---

### 4.2 视觉隐喻：控制面 vs 数据面

SitePod 架构中最有辨识度的是**控制面/数据面分离**、ref 作为数据面 SSOT 的设计。将其转化为视觉语言：

| 层级 | 职责 | 视觉表达 | 色温 |
|------|------|----------|------|
| **控制面** | 审计、历史、权限、管理 | 更柔和的边界、更暖的中性色 | 暖灰 |
| **数据面** | refs/blobs、确定性、高速 | 更冷的主色、强对比、硬边界 | 冷青 |

**视觉元素：**
- 双层卡片结构
- 双轨道布局
- 双态配色

用户一眼理解"稳定"。

---

### 4.3 配色系统

#### 核心原则
- **1 主色 + 1 强调色 + 大量中性灰**
- 中性灰阶占 **80% 以上**，让界面像"工具"
- 从一开始就做 **Light/Dark 双主题**（开发者产品的默认期望）

#### 色板定义

| 用途 | Light Mode | Dark Mode | 说明 |
|------|------------|-----------|------|
| **主色** | `#0891B2` (cyan-600) | `#22D3EE` (cyan-400) | 偏冷的青蓝，传达可靠、技术、速度 |
| **强调色** | `#84CC16` (lime-500) | `#A3E635` (lime-400) | 酸性黄绿，仅用于高价值状态 |
| **背景-主** | `#FFFFFF` | `#0F172A` (slate-900) | |
| **背景-次** | `#F8FAFC` (slate-50) | `#1E293B` (slate-800) | |
| **边框** | `#E2E8F0` (slate-200) | `#334155` (slate-700) | |
| **文字-主** | `#0F172A` (slate-900) | `#F8FAFC` (slate-50) | |
| **文字-次** | `#64748B` (slate-500) | `#94A3B8` (slate-400) | |

#### 状态色

| 状态 | 颜色 | 使用场景 |
|------|------|----------|
| **成功** | `#22C55E` (green-500) | 部署完成、发布成功 |
| **警告** | `#F59E0B` (amber-500) | `--prod` 确认、注意事项 |
| **错误** | `#EF4444` (red-500) | 错误、失败 |
| **信息** | `#0891B2` (cyan-600) | 提示、进度 |

#### 控制面 vs 数据面配色

```
控制面（管理界面）:
- 背景：slate-50 / slate-800
- 边框：slate-200 / slate-700
- 暖灰基调

数据面（部署状态、refs）:
- 背景：带主色调的微妙色彩
- 边框：主色系边框
- 冷青基调
```

---

### 4.4 字体与排版

#### 核心原则
把"**读得快**"当核心品牌价值

#### 字体选择

| 用途 | 字体 | 备选 |
|------|------|------|
| **标题** | Inter | SF Pro Display, -apple-system |
| **正文** | Inter | SF Pro Text, -apple-system |
| **代码/ID** | JetBrains Mono | SF Mono, Menlo |

#### 字体规格

```css
/* 标题 - 几何无衬线，干净现代 */
.heading {
  font-family: 'Inter', -apple-system, sans-serif;
  font-weight: 600;
  letter-spacing: -0.02em;
}

/* 正文 - 高可读无衬线 */
.body {
  font-family: 'Inter', -apple-system, sans-serif;
  font-weight: 400;
  line-height: 1.6;
}

/* 代码/ID - 等宽，字重稍高 */
.code {
  font-family: 'JetBrains Mono', 'SF Mono', monospace;
  font-weight: 500;
}
```

#### 排版风格

- **大标题 + 短句**：减少长段 marketing 话术
- **代码块/终端块**：卖点本来就硬，让代码说话
- **Hash/ID 突出显示**：image_id、content_hash、refs 路径用等宽字体

```
✗ 避免: "我们提供业界领先的部署解决方案，通过先进的..."
✓ 推荐: "sitepod deploy → 上传 → 发布 → 完成"
```

---

### 4.5 图形与插画

#### 核心原则
**不要插画"讲故事"，要"讲结构"**

#### 避免使用
- ❌ 人物插画
- ❌ 复杂 3D 图形
- ❌ 抽象装饰性元素

#### 推荐使用

**系统示意图：**
```
┌─────────────┐     ┌─────────────┐
│   Plan      │ ──→ │   Commit    │
│  (manifest) │     │  (upload)   │
└─────────────┘     └─────────────┘
                           │
                           ▼
┌─────────────┐     ┌─────────────┐
│    Ref      │ ←── │    Pod      │
│  (pointer)  │     │  (snapshot) │
└─────────────┘     └─────────────┘
```

**推荐图形内容：**
- Plan/Commit 流程图
- Ref 指向示意
- Blob 去重可视化
- Rollback 路径图

**线条风格：**
- 圆角（4-8px）
- 细线（1-2px）
- 低饱和色彩
- 少阴影或无阴影

#### 动效规范

**原则：** 只做"状态变化"动效

| 场景 | 时长 | 缓动 |
|------|------|------|
| Deploy 进度 | 120-180ms | ease-out |
| Ref 切换 | 150ms | ease-in-out |
| Rollback 生效 | 120ms | ease-out |
| Hover 状态 | 100ms | ease |

**快节奏**：120-180ms，传达"快速、确定"的品牌感

---

### 4.6 组件风格

#### 设计定位
偏"控制台/运维面板"，但比传统运维工具**更精致**

#### 卡片

```css
.card {
  border: 1px solid var(--border);     /* 轻边框 */
  border-radius: 6px;                   /* 圆角偏小，更专业 */
  box-shadow: 0 1px 2px rgba(0,0,0,0.05); /* 微弱阴影 */
  /* 或仅边框，无阴影 */
}
```

#### 按钮

| 类型 | 样式 | 使用场景 |
|------|------|----------|
| **主按钮** | 主色填充 | Deploy, Release, 唯一主操作 |
| **次级按钮** | 中性灰边框 | Preview, History, 次要操作 |
| **危险按钮** | 红色边框/填充 | Delete, 不可逆操作 |
| **文字按钮** | 无边框 | Cancel, 辅助操作 |

```
原则：主按钮只留一个强主色，其余次级按钮都中性
```

#### 状态徽章 (Badge/Label)

统一的状态标签体系，保持信息密度高但不乱：

| 状态 | 样式 | 示例 |
|------|------|------|
| **环境** | 填充色 + 白字 | `prod` `beta` `preview` |
| **发布状态** | 边框 + 色字 | `released` `rollback` `pending` |
| **数据指标** | 浅底 + 深字 | `92% reused` `12 new` |

```html
<!-- 环境标签 -->
<span class="badge badge-prod">prod</span>
<span class="badge badge-beta">beta</span>
<span class="badge badge-preview">preview</span>

<!-- 状态标签 -->
<span class="badge badge-success">released</span>
<span class="badge badge-warning">pending</span>
<span class="badge badge-info">rollback</span>

<!-- 指标标签 -->
<span class="badge badge-metric">92% reused</span>
```

---

### 4.7 CLI 输出作为品牌基因

**核心观点：** CLI 是 SitePod 的核心触点，用户最信任的是"命令行给我的确定反馈"。

**让 CLI 输出样式反向定义 UI 的视觉语言：**

#### 成功状态
```bash
✓ Released to prod
```
- 单一强调色（绿色）+ 勾号
- UI 复用：绿色徽章、成功提示

#### 警告状态
```bash
⚠ You are deploying to PRODUCTION
  Press Enter to confirm, Ctrl+C to cancel
```
- 琥珀色 + 明确确认提示
- UI 复用：黄色警告框、确认对话框

#### 错误状态
```bash
✗ Error: E1001 - Authentication failed
  → Run `sitepod login` to authenticate
```
- 错误码 + 下一步行动（ops 手册的排障思路）
- UI 复用：红色错误框、带操作建议

#### 进度状态
```bash
◐ Uploading 12 files... (3/12)
```
- 主色 + spinner
- UI 复用：进度条、加载状态

#### 完整示例

```bash
$ sitepod deploy --prod

◐ Scanning ./dist...
✓ Found 156 files

◐ Computing hashes...
✓ Done

◐ Planning deployment...
✓ Plan ready
  → 12 new, 144 reused (92%)

⚠ You are deploying to PRODUCTION
  Press Enter to confirm, Ctrl+C to cancel

◐ Uploading 12 files...
✓ Upload complete

✓ Released to prod

  image: img_a1b2c3d4
  url:   https://my-app.example.com
```

**UI 设计原则：**
- Admin UI 和官网全部复用同一套状态色、徽章、语气
- 形成"统一且可信"的品牌体验

---

### 4.8 文案语气

#### 核心原则
**短、确定、可执行**（像 CLI 输出一样）

#### 示例对比

| ❌ 避免 | ✓ 推荐 |
|--------|--------|
| "部署正在进行中，请稍候..." | "Deploying..." |
| "操作已成功完成！" | "Done" |
| "是否确定要执行此操作？" | "Deploy to prod?" |
| "发生了一个错误，请重试" | "Error: E1001. Run `sitepod login`" |

#### 语气特点
- 使用祈使句
- 省略不必要的词
- 提供下一步行动
- 像 Unix 工具一样简洁

---

### 4.9 Logo 规范

#### 图标构成

```
┌─────────────────────────────┐
│                             │
│    ╭─────────────────╮      │
│   ╱                   ╲     │
│  ╱    ┌─────────┐      ╲    │   外层：六边形边框（控制面）
│ │     │  ◢██◣   │       │   │   颜色：Slate-500 (#64748B)
│ │     │ ██████  │       │   │
│ │     │ ◥██◤   │       │   │   内层：3D 立方体（数据面/Pod）
│  ╲    └─────────┘      ╱    │   颜色：Cyan-600 (#0891B2)
│   ╲                   ╱     │
│    ╰─────────────────╯      │
│                             │
└─────────────────────────────┘
```

**设计理念：**
- **外层六边形**：代表控制面（管理、审计、权限）— 稳定的容器边界
- **内层立方体**：代表数据面（Pod/Blob）— 不可变的内容快照
- **双层结构**：直观传达"控制面 vs 数据面"的架构特色

#### 色彩规格

| 元素 | Light Mode | Dark Mode |
|------|------------|-----------|
| 外层边框 | `#64748B` (slate-500) | `#94A3B8` (slate-400) |
| 立方体主面 | `#0891B2` (cyan-600) | `#22D3EE` (cyan-400) |
| 立方体暗面 | `#0E7490` (cyan-700) | `#06B6D4` (cyan-500) |
| 文字 | `#0F172A` (slate-900) | `#F8FAFC` (slate-50) |

#### Logo 版本

| 版本 | 文件名 | 使用场景 |
|------|--------|----------|
| **横版完整** | `logo.svg` | 官网 Header、README |
| **纯图标** | `logo-icon.svg` | Favicon、App Icon、小尺寸 |
| **深色背景** | `logo-dark.svg` | 深色页面、Dark Mode |
| **单色版** | `logo-mono.svg` | 打印、水印、低对比场景 |

#### 最小尺寸

| 版本 | 最小宽度 |
|------|----------|
| 横版完整 | 120px |
| 纯图标 | 24px |

#### 安全区域

图标周围保留 **图标高度的 25%** 作为安全区域，确保视觉呼吸空间。

#### 禁止用法

- ❌ 拉伸或压缩比例
- ❌ 旋转图标
- ❌ 添加阴影或特效
- ❌ 更改品牌色
- ❌ 在复杂背景上使用（需使用带底色版本）

#### 资源文件位置

```
www/public/
├── logo.svg              # 横版完整（Light）
├── logo-dark.svg         # 横版完整（Dark）
├── logo-icon.svg         # 纯图标（Light）
├── logo-icon-dark.svg    # 纯图标（Dark）
├── logo-mono.svg         # 单色版
├── favicon.svg           # Favicon
├── favicon.ico           # Favicon (ICO)
├── apple-touch-icon.png  # iOS 图标 (180x180)
└── og-image.png          # Open Graph 图片 (1200x630)
```

---

## 5. 文档结构

### 5.1 官网页面

```
sitepod.dev/
├── /                    # Hero + 核心价值
├── /docs               # → docs.sitepod.dev
├── /pricing            # 开源免费 + 企业支持
├── /blog               # 更新日志、技术文章
└── /community          # Discord/GitHub Discussions
```

### 5.2 文档站结构

```
docs.sitepod.dev/
├── /getting-started    # 快速开始
│   ├── /install        # 安装 CLI
│   ├── /first-deploy   # 第一次部署
│   └── /configuration  # 配置文件
├── /guides             # 指南
│   ├── /ci-cd          # CI/CD 集成
│   ├── /rollback       # 回滚操作
│   └── /preview        # 预览部署
├── /self-hosting       # 自托管
│   ├── /docker         # Docker 部署
│   ├── /kubernetes     # K8s 部署
│   └── /storage        # 存储后端配置
├── /api                # API 参考
└── /cli                # CLI 参考
```

---

## 6. 宣传文案

### 6.1 一句话介绍

```
SitePod: Self-hosted static release & rollback platform.
```

```
SitePod: 自托管静态站点发布与回滚平台。
```

### 6.2 三点价值

```
1. 🚀 One command release — deploy to prod/beta/preview
2. ⚡ Instant rollback — switch refs, no rebuild
3. 🔒 Self-hosted — your data stays on your infrastructure
```

### 6.3 技术亮点

```
- Immutable pods: content-addressed storage, never overwrite
- Incremental uploads: only upload what changed (Plan/Commit)
- Ref-based releases: roll back by switching pointers
- Zero vendor lock-in: pluggable storage (Local/S3/OSS/R2)
- Single binary: PocketBase + Caddy, zero external dependencies
```

---

## 7. SEO & 社交媒体

### 7.1 Meta Tags

```html
<title>SitePod — Self-hosted static releases</title>
<meta name="description" content="Self-hosted static releases with instant rollback. One-command deploys, content-hash versions, preview environments, and pluggable storage.">
<meta name="keywords" content="static site, release, deployment, rollback, self-hosted, preview, content-addressed storage, multi-environment">
```

### 7.2 Open Graph

```html
<meta property="og:title" content="SitePod — Self-hosted static releases">
<meta property="og:description" content="Release once, rollback in seconds. Your sites as immutable pods.">
<meta property="og:image" content="https://sitepod.dev/og-image.png">
<meta property="og:url" content="https://sitepod.dev">
```

### 7.3 GitHub Topics

```
static-site, release, deployment, rollback, self-hosted,
preview, content-addressed, cdn, devops, cli, rust, golang
```

---

## 8. Checklist: 品牌一致性

发布前检查:

- [ ] GitHub README 首行有副标题
- [ ] 官网 Hero 区有副标题 + Pod 定义
- [ ] CLI `--help` 输出有副标题
- [ ] Docker Hub/GHCR 描述有副标题
- [ ] 所有命名使用 `sitepod` (不是 `pod`)
- [ ] 环境变量使用 `SITEPOD_` 前缀
- [ ] 配置文件名为 `sitepod.toml`
