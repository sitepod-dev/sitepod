# SitePod — Self-hosted static deployments

## 产品需求文档 (PRD)

**版本**: v1.0
**日期**: 2025-01-15
**项目主页**: https://sitepod.dev

> 技术设计详见 [tdd.md](./tdd.md)
> 用户故事详见 [docs/user-stories.md](./docs/user-stories.md)
> 运维手册详见 [ops.md](./ops.md)
> 品牌规范详见 [brand.md](./brand.md)

---

## 1. 产品概述

### 1.1 背景

现有静态站点部署方案的问题：

| 方案 | 问题 |
|------|------|
| Vercel/Netlify | 海外服务，国内访问慢；数据存储在境外；企业版价格高 |
| 手动上传 CDN | 流程繁琐，无版本管理，回滚困难 |
| CI/CD 脚本 | 每个项目重复造轮子，缺乏统一管理 |

### 1.2 产品定位

**SitePod** 是一个可自托管的静态站点部署平台，提供：

- 一行命令完成上传、发布
- 基于内容哈希的版本管理
- 秒级回滚能力
- 多环境支持 (prod/beta/preview)
- 可插拔存储后端 (本地/S3/OSS/R2)

### 1.3 目标用户

- 需要私有化部署的企业前端团队
- 对数据合规有要求的公司
- 希望降低 CDN 成本的团队
- 个人开发者搭建自己的部署平台

### 1.4 核心价值

```
传统流程: 构建 → 手动上传 → 配置 → 验证 → 切换 → (出问题) → 手动回滚
                    ↓ 多步骤，易出错，难追溯

SitePod:  sitepod deploy --prod
                    ↓ 一步完成，自动版本管理，秒级回滚
```

---

## 2. 核心概念

| 概念 | 说明 | 类比 |
|------|------|------|
| **Project** | 单个站点/应用 | GitHub Repository |
| **Image** | 部署快照，不可变 | Docker Image / Git Commit |
| **Manifest** | 版本清单，列出所有文件 | package-lock.json |
| **Blob** | 按内容哈希存储的文件 | Git Object |
| **Ref** | 环境指向的版本引用 | Git Branch |
| **Environment** | 部署环境 (prod/beta/preview) | Vercel Preview |

---

## 3. 用户故事

### 3.1 首次部署

```
作为 前端开发者
我想要 一行命令部署我的静态站点
以便于 快速上线，不用手动配置 CDN
```

**验收标准：**
- `sitepod deploy` 自动检测 dist 目录
- 首次运行时交互式创建项目
- 部署成功后输出访问 URL

### 3.2 增量更新

```
作为 前端开发者
我想要 只上传改动的文件
以便于 加快部署速度，节省带宽
```

**验收标准：**
- 改 1 个文件只上传 1 个文件
- 显示复用了多少文件

### 3.3 秒级回滚

```
作为 运维人员
我想要 一键回滚到之前的版本
以便于 快速恢复线上问题
```

**验收标准：**
- `sitepod rollback` 列出历史版本
- 选择后立即生效
- 回滚操作有审计日志

### 3.4 预览部署

```
作为 产品经理
我想要 在上线前预览新版本
以便于 验收功能，确认无误后再发布
```

**验收标准：**
- `sitepod preview` 生成临时预览链接
- 默认 24 小时过期
- 不影响线上环境

---

## 4. 功能需求

### 4.1 CLI 命令

| 命令 | 说明 |
|------|------|
| `sitepod deploy` | 一键部署（自动登录、初始化） |
| `sitepod deploy --prod` | 部署到生产环境 |
| `sitepod login` | 登录认证（邮箱验证） |
| `sitepod bind` | 匿名账户绑定邮箱升级 |
| `sitepod init` | 初始化项目配置 |
| `sitepod preview` | 创建预览部署 |
| `sitepod rollback` | 回滚到历史版本 |
| `sitepod history` | 查看部署历史 |
| `sitepod domain add` | 添加自定义域名 |
| `sitepod domain verify` | 验证域名所有权 |
| `sitepod domain rename` | 修改系统分配的子域名 |
| `sitepod domain list` | 列出已添加的域名 |
| `sitepod domain remove` | 移除域名 |

#### 智能部署流程

`sitepod deploy` 命令会自动处理所有前置条件：

```
sitepod deploy
    │
    ├─ 未登录? ──→ 自动创建匿名账户 (24h 有效)
    │
    ├─ 无 sitepod.toml? ──→ 交互式初始化
    │
    └─ 执行部署
```

#### 匿名账户机制

| 特性 | 说明 |
|------|------|
| 自动创建 | 未登录时执行 deploy/preview 自动创建 |
| 有效期 | 24 小时后过期，部署一并删除 |
| 升级方式 | `sitepod bind` 绑定邮箱升级为正式账户 |
| 升级后 | 账户永久保留，域名不变 |

#### 系统域名

初始化时用户可自定义子域名：

```bash
$ sitepod init
? 项目名称: my-blog
? 子域名 (my-blog): my-blog
✗ my-blog.sitepod.dev 已被占用
? 子域名: alice-blog
✓ alice-blog.sitepod.dev 可用
```

- 默认值 = 项目名（slugify 后）
- 检查可用性，冲突则提示重新输入
- 用户也可输入 `-` 使用随机 ID（如 `my-blog-7x3k`）
- 创建后可通过 `sitepod domain rename` 修改

### 4.2 部署模式

SitePod 支持两种部署模式，适应不同的基础设施环境：

#### 子域名模式 (Subdomain Mode)

适用于有通配符 DNS 和证书的环境。每个项目自动分配子域名。

**默认域名格式：** `{project}-{id}.sitepod.dev`

| 环境 | URL 格式 | 示例 |
|------|----------|------|
| prod | `{project}.{domain}` | `my-app.sitepod.dev` |
| beta | `{project}-beta.{domain}` | `my-app-beta.sitepod.dev` |
| preview | `{project}.{domain}/__preview__/{slug}/` | `my-app.sitepod.dev/__preview__/feat1/` |

> **说明：** 使用 `{project}-{短随机ID}` 格式确保全局唯一。用户可通过 `sitepod domain rename` 修改。

#### 路径模式 (Path Mode)

适用于无法配置通配符 DNS 的环境（如 Coolify、自定义域名）。每个项目独立配置域名和路径前缀。

| 环境 | URL 格式 | 示例 |
|------|----------|------|
| prod | `{domain}/{slug}/` | `h5.example.com/my-app/` |
| beta | Cookie/Query 切换 | `h5.example.com/my-app/?env=beta` |
| preview | Cookie/Query 切换 | `h5.example.com/my-app/?preview=abc123` |

**路径模式工作原理：**
1. 入口 URL 带 `?env=beta` 或 `?preview=abc123` query 参数
2. 服务端设置 Cookie 并重定向到干净 URL
3. 后续请求通过 Cookie 识别环境/预览版本
4. 静态资源路径不受影响，避免 base URL 问题

**单域名模式：** 当 `slug = "/"` 时，项目独占整个域名，无路径前缀。

#### 模式选择

| 场景 | 推荐模式 |
|------|----------|
| 自建平台，有通配符 DNS | 子域名模式 |
| Coolify/PaaS 部署 | 路径模式（单域名） |
| 多项目共享域名 | 路径模式（多 slug） |
| 已有自定义域名 | 路径模式（单域名） |

### 4.3 自定义域名

使用路径模式时，需要先验证域名所有权。

#### 添加域名流程

```
1. 用户在 CLI 或 Admin UI 添加自定义域名
2. 系统生成验证令牌
3. 用户添加 DNS 记录证明所有权
4. 系统验证通过后，域名可用于部署
```

#### 验证方式

| 方式 | DNS 记录 | 适用场景 |
|------|----------|----------|
| TXT 验证 | `_sitepod.example.com TXT "sitepod-verify=xxx"` | 通用 |
| CNAME 验证 | `_sitepod.example.com CNAME xxx.verify.sitepod.dev` | 需要自动续期 |

**示例：**
```
$ sitepod domain add h5.example.com

添加 DNS 记录以验证域名所有权：

  类型: TXT
  名称: _sitepod.h5.example.com
  值:   sitepod-verify=abc123def456

添加后运行: sitepod domain verify h5.example.com
```

#### 域名状态

| 状态 | 说明 |
|------|------|
| `pending` | 待验证，已添加但未通过验证 |
| `verified` | 已验证，可用于部署 |
| `failed` | 验证失败（DNS 记录错误或已删除） |

#### CLI 命令

| 命令 | 说明 |
|------|------|
| `sitepod domain add <domain>` | 添加自定义域名 |
| `sitepod domain verify <domain>` | 验证域名所有权 |
| `sitepod domain list` | 列出已添加的域名 |
| `sitepod domain remove <domain>` | 移除域名 |

### 4.4 配置文件

```toml
# sitepod.toml
[project]
name = "my-app"

[build]
directory = "./dist"

[deploy]
ignore = ["**/*.map", ".*", "node_modules/**"]
concurrent = 20

# 可选：自定义域名配置（路径模式）
# 不配置则使用系统默认的子域名模式
[deploy.routing]
domain = "h5.example.com"   # 自定义域名
slug = "/my-app"            # URL 路径前缀，"/" 表示单域名模式
```

#### 配置示例

**子域名模式（默认）：**
```toml
[project]
name = "my-app"
# routing_mode = "subdomain"  # 默认值，可省略

# 系统自动分配子域名:
# → https://my-app.sitepod.dev (prod)
# → https://my-app-beta.sitepod.dev (beta)

# 可修改子域名: sitepod domain rename my-preferred-name
# 可绑定自定义域名: sitepod domain add www.mysite.com
```

**路径模式 - 多项目共享域名：**
```toml
[project]
name = "blog-web"
routing_mode = "path"

[deploy.routing]
domain = "www.example.com"
slug = "/blog"
# → https://www.example.com/blog/

# 路径模式不分配系统子域名
```

**路径模式 - 单域名：**
```toml
[project]
name = "company-website"
routing_mode = "path"

[deploy.routing]
domain = "www.example.com"
slug = "/"
# → https://www.example.com/

# 也可以绑定多个域名指向同一项目
# sitepod domain add example.com --slug /
```

---

## 5. API 契约

### 5.1 认证

```
Authorization: Bearer <token>
```

| Token 类型 | 获取方式 | 用途 |
|------------|----------|------|
| User Token | `sitepod login` | 交互式操作 |
| API Token | Admin UI 创建 | CI/CD |

### 5.2 部署流程 (Plan/Commit)

**Step 1: Plan - 提交文件清单**

```yaml
POST /api/v1/plan
Body:
  project: string
  files: [{ path, hash, size, content_type }]
  git?: { commit, branch, message }
Response:
  plan_id: string
  content_hash: string
  missing: [{ path, hash, upload_url }]  # 需要上传的文件
  reusable: number                        # 可复用的文件数
```

**Step 2: Commit - 确认完成**

```yaml
POST /api/v1/commit
Body:
  plan_id: string
Response:
  image_id: string
  content_hash: string
```

### 5.3 发布

```yaml
POST /api/v1/release
Body:
  project: string
  environment: "prod" | "beta"
  image_id?: string  # 不传则使用最新
Response:
  url: string
```

### 5.4 回滚

```yaml
POST /api/v1/rollback
Body:
  project: string
  environment: "prod" | "beta"
  image_id: string
Response:
  url: string
  previous_image_id: string
```

### 5.5 预览

```yaml
POST /api/v1/preview
Body:
  project: string
  image_id: string
  slug?: string         # 预览标识，默认随机生成
  expires_in?: number   # 秒，默认 86400
Response:
  url: string           # 访问 URL（格式取决于部署模式）
  expires_at: datetime
```

**返回的 URL 格式：**

| 部署模式 | URL 格式 |
|----------|----------|
| 子域名模式 | `https://{project}--{slug}.preview.{domain}/` |
| 路径模式 | `https://{domain}/{slug}/?preview={preview_slug}` |

### 5.6 查询

```yaml
GET /api/v1/current?project=xxx&environment=prod
Response:
  image_id: string
  content_hash: string
  deployed_at: datetime

GET /api/v1/history?project=xxx&limit=20
Response:
  items: [{ image_id, content_hash, created_at, git_commit }]
```

### 5.7 域名管理

```yaml
# 添加自定义域名
POST /api/v1/domains
Body:
  domain: string              # 域名，如 "h5.example.com"
Response:
  domain: string
  status: "pending"
  verification:
    type: "txt"               # 验证类型
    name: string              # DNS 记录名称
    value: string             # DNS 记录值

# 验证域名
POST /api/v1/domains/{domain}/verify
Response:
  domain: string
  status: "verified" | "failed"
  error?: string              # 失败原因

# 列出域名
GET /api/v1/domains
Response:
  items: [{ domain, status, verified_at, projects }]

# 删除域名
DELETE /api/v1/domains/{domain}
Response:
  success: boolean
```

### 5.8 错误码

| 状态码 | 错误码 | 说明 |
|--------|--------|------|
| 400 | INVALID_REQUEST | 请求参数错误 |
| 401 | UNAUTHORIZED | 未认证或 token 无效 |
| 403 | FORBIDDEN | 无权限操作该项目 |
| 404 | NOT_FOUND | 项目/版本不存在 |
| 410 | GONE | Preview 已过期 |

---

## 6. 非功能需求

### 6.1 性能

| 指标 | 目标 |
|------|------|
| 增量部署 (改 1 文件) | < 3s |
| 版本切换 | < 1s |
| 回滚生效 | < 1s |

### 6.2 可用性

- 单二进制部署，零外部依赖
- 支持本地存储，无需云服务
- 自动 HTTPS (Let's Encrypt)

### 6.3 安全

- Token 只支持 Header 认证
- API Token 支持 scope 限制
- 所有部署操作有审计日志

### 6.4 存储

- 支持本地磁盘、S3、OSS、R2
- 按内容哈希存储，跨版本去重
- 无厂商锁定，可迁移

---

## 7. 约束与决策

### 7.1 架构决策

| 决策 | 选择 | 原因 |
|------|------|------|
| 后端框架 | PocketBase + 嵌入 Caddy | 单二进制，内置认证/存储/HTTPS |
| CLI 语言 | Rust | 单二进制，快速 |
| 存储模型 | CAS (Content-Addressed) | 增量上传，去重 |
| 版本切换 | Ref 文件 + Caddy 路由 | 无厂商锁定 |

### 7.2 MVP 不支持

| 功能 | 原因 | 后续计划 |
|------|------|----------|
| 自定义域名 | 需要 host→project 映射机制 | v1.1 |
| 团队协作 | 需要权限系统 | v1.2 |
| 自动构建 | 专注部署，构建交给 CI | 可能不做 |
| 图片压缩 | 应在 build pipeline 完成 | 不做 |

---

## 8. 里程碑

### Phase 1: MVP

- [x] CLI 基础命令 (login/deploy/rollback)
- [x] 本地存储后端
- [x] Plan/Commit 增量上传
- [x] 基本版本切换

### Phase 2: 完整功能

- [x] S3/OSS 存储后端
- [x] Preview 功能
- [x] 历史查询
- [x] Admin UI (PocketBase 内置)

### Phase 3: 生产就绪

- [x] 监控指标 (Prometheus 格式)
- [x] GC 清理
- [x] 文档完善
- [x] Docker 镜像
