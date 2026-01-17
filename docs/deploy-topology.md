# SitePod 部署拓扑

本文档介绍 SitePod 的三种部署拓扑，按复杂度和鲁棒性排序。

## 架构背景

SitePod 采用控制面/数据面分离架构：

- **控制面**：PocketBase (SQLite) 处理认证、审计日志、历史记录
- **数据面**：Storage (`refs/` + `blobs/`) 存储部署内容，Caddy 直接读取

关键设计：线上请求只读 Storage，不依赖数据库。即使 DB 故障，已部署站点仍可访问。

---

## 拓扑 A：单机直出（最简单）

```
┌─────────────────────────────────────────┐
│              Internet                    │
└─────────────────┬───────────────────────┘
                  │ :443 (TLS)
                  ▼
┌─────────────────────────────────────────┐
│              SitePod                     │
│  ┌─────────────────────────────────────┐│
│  │  Caddy (嵌入式)                      ││
│  │  - TLS (Let's Encrypt / ACME)       ││
│  │  - 通配符路由                        ││
│  │  - 静态文件服务                      ││
│  └─────────────────────────────────────┘│
│  ┌─────────────────────────────────────┐│
│  │  PocketBase API                      ││
│  │  - 认证 / 部署 / 审计                ││
│  └─────────────────────────────────────┘│
└─────────────────────────────────────────┘
```

### 配置

```bash
# 环境变量
SITEPOD_DOMAIN=sitepod.example.com

# 不设置 SITEPOD_PROXY_MODE（使用默认 Caddyfile）
```

### 适用场景

- 全新服务器，无现有入口
- 快速测试、个人项目
- 不需要额外的反向代理

### 优点

- 部署最简单，开箱即用
- 自动 HTTPS 证书管理

### 缺点

- 如果已有入口（Coolify/Traefik/Nginx），会产生"双入口/双 TLS"冲突
- 通配符证书需要 DNS challenge 配置

---

## 拓扑 B：反向代理后端（最常见）

```
┌─────────────────────────────────────────┐
│              Internet                    │
└─────────────────┬───────────────────────┘
                  │ :443 (TLS)
                  ▼
┌─────────────────────────────────────────┐
│     外层入口 (Coolify/Traefik/Nginx)     │
│     - TLS 终止                           │
│     - Host 路由                          │
└─────────────────┬───────────────────────┘
                  │ :8080 (HTTP)
                  ▼
┌─────────────────────────────────────────┐
│              SitePod                     │
│  ┌─────────────────────────────────────┐│
│  │  Caddy (代理模式)                    ││
│  │  - 无 TLS，监听 8080                 ││
│  │  - 静态文件服务                      ││
│  └─────────────────────────────────────┘│
│  ┌─────────────────────────────────────┐│
│  │  PocketBase API                      ││
│  └─────────────────────────────────────┘│
└─────────────────────────────────────────┘
```

### 配置

```bash
# 环境变量
SITEPOD_DOMAIN=sitepod.example.com
SITEPOD_PROXY_MODE=1  # 关键：禁用 TLS，监听 8080
```

### 三个关键不变量

#### 1. Host 头必须原样透传

数据面依赖 Host 头解析 `{project}` 和 `{env}`，决定读取哪个 `refs/{project}/{env}.json`。

```nginx
# Nginx 示例
proxy_set_header Host $host;

# Traefik 默认透传，无需配置
```

#### 2. TLS 只终止一次

- 外层终止 TLS → SitePod 跑 HTTP（`SITEPOD_PROXY_MODE=1`）
- 或者外层做 TCP passthrough（按 SNI 转发）→ SitePod 终止 TLS

**不要两层都终止 TLS**。

#### 3. API 和站点分开路由

| 域名 | 用途 |
|------|------|
| `sitepod.example.com` | 控制面 (Console + API) |
| `*.sitepod.example.com` | 数据面 (项目站点) |

外层入口需要配置通配符域名路由到 SitePod。

### 适用场景

- Coolify、Traefik、Nginx 等现有入口
- 企业内网部署
- 需要统一入口管理多个服务

### 优点

- 与现有基础设施集成
- 入口层统一管理 TLS 和路由

### 缺点

- 通配符证书仍需在入口层配置 DNS challenge
- 配置略复杂

### 平台特定指南

- [Coolify 部署指南](deploy-coolify.md)

---

## 拓扑 C：Cloudflare 入口（推荐）

```
┌─────────────────────────────────────────┐
│              Internet                    │
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│           Cloudflare Edge                │
│  - 边缘 TLS 终止                         │
│  - 通配符证书（自动）                     │
│  - CDN 缓存                              │
│  - WAF / DDoS 防护                       │
└─────────────────┬───────────────────────┘
                  │ HTTP / Origin Cert
                  ▼
┌─────────────────────────────────────────┐
│              SitePod                     │
│  ┌─────────────────────────────────────┐│
│  │  Caddy (代理模式)                    ││
│  │  - 监听 8080                         ││
│  └─────────────────────────────────────┘│
│  ┌─────────────────────────────────────┐│
│  │  PocketBase API                      ││
│  └─────────────────────────────────────┘│
└─────────────────────────────────────────┘
```

### 为什么推荐 Cloudflare

1. **通配符 TLS 自动处理**：无需配置 DNS challenge，Cloudflare 边缘证书自动覆盖 `*.example.com`
2. **全球 CDN**：静态资源边缘缓存，降低源站压力
3. **WAF / DDoS 防护**：企业级安全，开箱即用
4. **简化源站**：源站只需可达，无需公网 IP（可用 Cloudflare Tunnel）

### 配置

#### DNS 设置

在 Cloudflare 添加 DNS 记录，开启代理（橙云）：

| 类型 | 名称 | 内容 | 代理状态 |
|------|------|------|----------|
| A | sitepod | 源站 IP | 已代理 |
| A | *.sitepod | 源站 IP | 已代理 |

#### SitePod 环境变量

```bash
SITEPOD_DOMAIN=sitepod.example.com
SITEPOD_PROXY_MODE=1
```

#### SSL/TLS 设置

在 Cloudflare Dashboard → SSL/TLS：

- **加密模式**：Full 或 Full (Strict)
- 如果源站无证书，可使用 Cloudflare Origin Certificate

### 缓存策略

SitePod 已设置合理的 Cache-Control 头：

| 资源类型 | Cache-Control | 说明 |
|----------|---------------|------|
| `assets/*`, `_next/*` | `max-age=31536000, immutable` | 带 hash 的静态资源，长期缓存 |
| `index.html` 等入口 | `max-age=0, must-revalidate` | 每次验证，确保版本切换即时生效 |

---

## Cloudflare 集成等级

根据需求选择集成深度：

### 等级 1：DNS + CDN + TLS（最少改动）

立即收益：
- 通配符 HTTPS 自动处理
- 静态资源全球缓存
- 基础 DDoS 防护

无需代码改动，只需 DNS 配置。

### 等级 2：边缘短缓存（可选优化）

SitePod 默认对入口文件设置 `max-age=0, must-revalidate`，Cloudflare 不会缓存。这保证了版本切换立即生效。

对于高流量站点，如果想减少回源压力，可以启用"边缘短缓存"：

**方案 A：接受秒级延迟（推荐）**

使用 `Cloudflare-CDN-Cache-Control` 头分离浏览器缓存和边缘缓存：

```
Cache-Control: no-store                    # 浏览器不缓存
Cloudflare-CDN-Cache-Control: max-age=2    # 边缘缓存 2 秒
```

- 优点：不需要 Purge API，边缘缓存减少回源
- 缺点：版本切换最坏延迟 2 秒（可调整 TTL）

> ⚠️ 此功能需要修改代码，尚未实现。

**方案 B：调用 Purge API**

如果需要严格即时生效，可在 Release/Rollback 后调用 Cloudflare API 清除缓存：

```bash
SITEPOD_CF_ZONE_ID=your_zone_id
SITEPOD_CF_API_TOKEN=your_api_token
```

> ⚠️ 此功能尚未实现。大多数场景下，默认的 `max-age=0` 已足够。

### 等级 3：R2 + Worker 边缘数据面（最强形态）

将数据面完全搬到 Cloudflare 边缘：

```
┌─────────────────────────────────────────┐
│           Cloudflare Edge                │
│  ┌─────────────────────────────────────┐│
│  │  Worker (数据面)                     ││
│  │  - 读取 R2: refs/{project}/{env}    ││
│  │  - 解析 manifest                    ││
│  │  - 返回 blob 内容                   ││
│  └─────────────────────────────────────┘│
│  ┌─────────────────────────────────────┐│
│  │  R2 (Storage)                        ││
│  │  - refs/                             ││
│  │  - blobs/                            ││
│  └─────────────────────────────────────┘│
└─────────────────────────────────────────┘
           │
           │ API only
           ▼
┌─────────────────────────────────────────┐
│        SitePod (控制面)                  │
│  - Plan / Commit / Release              │
│  - 写入 R2                              │
│  - 不再 serve 静态文件                   │
└─────────────────────────────────────────┘
```

优势：
- 源站不承担静态流量，只处理 API
- 全球边缘分发，延迟最低
- 源站可以藏在内网/Tunnel 后面

> ⚠️ **此功能尚未实现**，需要开发 Worker 代码。

---

## 拓扑选择指南

| 场景 | 推荐拓扑 |
|------|----------|
| 快速测试 / 个人项目 | A (单机直出) |
| 已有 Coolify / Traefik | B (反向代理后端) |
| 生产环境 / 需要 CDN | C (Cloudflare) |
| 高流量 / 全球用户 | C + 等级 3 (R2 + Worker) |

---

## 环境变量参考

| 变量 | 说明 | 拓扑 A | 拓扑 B/C |
|------|------|--------|----------|
| `SITEPOD_DOMAIN` | 基础域名 | 必填 | 必填 |
| `SITEPOD_PROXY_MODE` | 代理模式 (禁用 TLS) | 不设置 | `1` |
| `SITEPOD_STORAGE_TYPE` | 存储类型 | `local` | `local` / `r2` |
| `SITEPOD_CF_ZONE_ID` | Cloudflare Zone ID | - | 可选 (等级 2) |
| `SITEPOD_CF_API_TOKEN` | Cloudflare API Token | - | 可选 (等级 2) |
| `SITEPOD_ALLOW_ANONYMOUS` | 允许匿名账户 | `1` (开发) | 不设置 (生产) |

> ⚠️ **安全提示**：`SITEPOD_ALLOW_ANONYMOUS=1` 仅建议在开发/测试环境使用。生产环境应禁用匿名账户。

---

## 域名结构

### 自托管实例 (如 `x.com`)

| 域名 | 用途 |
|------|------|
| `x.com` | Console (管理界面 + API) |
| `{project}.x.com` | 用户项目 (prod) |
| `{project}-beta.x.com` | 用户项目 (beta) |

默认行为，无需额外配置。根域名自动指向 console 项目。

### sitepod.dev 官方实例

| 域名 | 用途 | 配置方式 |
|------|------|----------|
| `sitepod.dev` | 官网 (Landing page) | 自定义域名 → landing 项目 |
| `www.sitepod.dev` | 官网 (Landing page) | 自定义域名 → landing 项目 |
| `console.sitepod.dev` | Console (管理界面 + API) | 默认规则 |
| `{project}.sitepod.dev` | 用户项目 | 默认规则 |

官方实例通过自定义域名功能将 apex 域名指向独立的 landing 项目，而 console 作为普通子域名访问。

### 保留子域名

以下子域名被系统保留，用户无法创建同名项目：

`console`, `welcome`, `www`, `api`, `admin`, `app`, `static`, `assets`, `cdn`, `mail`, `email`, `ftp`, `ssh`
