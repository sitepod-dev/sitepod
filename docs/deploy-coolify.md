# SitePod Coolify 部署指南

本文档介绍如何在 Coolify 上部署 SitePod。

## 前置要求

- 已安装并运行的 Coolify 实例
- 一个域名，可配置 DNS

## 部署步骤

### 1. 配置 DNS

在你的 DNS 提供商添加以下记录，指向 Coolify 服务器：

| 类型 | 主机记录 | 记录值 |
|------|---------|--------|
| A | sitepod | Coolify 服务器 IP |
| A | *.sitepod | Coolify 服务器 IP |

> 假设你的域名是 `example.com`，则配置后：
> - 主域名 (Console + API)：`sitepod.example.com`
> - 项目域名 (生产)：`myapp.sitepod.example.com`
> - 项目域名 (Beta)：`myapp-beta.sitepod.example.com`

### 2. 在 Coolify 创建服务

1. 登录 Coolify 控制台
2. 选择 **New Resource** → **Docker Compose**
3. 连接到 SitePod 的 Git 仓库：`https://github.com/cosformula/sitepod.dev`
4. 配置文件选择：
   - **Docker Compose**: `docker-compose.coolify.yml`
   - **Dockerfile**: `Dockerfile.coolify`

### 3. 配置环境变量

在 Coolify 的 **Environment Variables** 中添加：

| 变量 | 值 | 说明 |
|------|---|------|
| `SITEPOD_DOMAIN` | `sitepod.example.com` | **必填**，你的基础域名 |
| `SITEPOD_STORAGE_TYPE` | `local` | 存储类型，可选 |

**可选的 S3 存储配置**（如使用 Cloudflare R2）：

```
SITEPOD_STORAGE_TYPE=r2
SITEPOD_S3_BUCKET=your-bucket
SITEPOD_S3_ENDPOINT=https://ACCOUNT_ID.r2.cloudflarestorage.com
AWS_ACCESS_KEY_ID=xxx
AWS_SECRET_ACCESS_KEY=xxx
```

### 4. 配置域名（重要！）

在 Coolify 的 **Domains** 设置中，需要添加 **两个** 域名：

| 域名 | 用途 |
|------|------|
| `sitepod.example.com` | 主域名 (Console + API) |
| `*.sitepod.example.com` | 用户站点 (生产和 Beta) |

> **注意**：Beta 环境使用 `-beta` 后缀 (如 `myapp-beta.sitepod.example.com`)，因此只需要一个通配符记录。

### 5. 部署

点击 **Deploy** 按钮，等待构建完成。

首次构建可能需要几分钟（需要编译 Go 和 Rust 代码）。

## 验证部署

```bash
# 检查健康状态
curl https://sitepod.example.com/api/v1/health

# 应返回：
# {"status":"healthy","database":"ok","storage":"ok","uptime":"..."}
```

访问控制台：`https://sitepod.example.com`

## 使用 CLI 部署站点

在你的开发机器上：

```bash
# 安装 CLI
npm install -g sitepod

# 登录
sitepod login --endpoint https://sitepod.example.com

# 在项目目录初始化
cd my-website
sitepod init

# 部署
sitepod deploy        # 部署到 beta
sitepod deploy --prod # 部署到生产
```

## 文件说明

| 文件 | 说明 |
|------|------|
| `Dockerfile.coolify` | Coolify 专用 Dockerfile，禁用 SSL，使用 8080 端口 |
| `docker-compose.coolify.yml` | Coolify 专用 Compose，配置持久化存储 |
| `server/examples/Caddyfile.proxy` | 反向代理模式 Caddyfile |

## 常见问题

### Q: 通配符域名 SSL 证书失败

**原因**：Traefik 需要 DNS challenge 来获取通配符证书。

**解决方案**：
1. 在 Coolify 设置中配置 DNS provider（如 Cloudflare）
2. 添加 DNS API 凭证
3. 重新部署

### Q: 子域名访问返回 404

**排查**：
1. 检查 DNS 是否正确解析：`dig myapp.sitepod.example.com`
2. 检查 Coolify 域名配置是否包含通配符
3. 查看 Coolify 日志

### Q: 部署后无法访问

**排查**：
1. 检查容器是否运行：在 Coolify 中查看状态
2. 检查健康检查是否通过
3. 查看容器日志

## 数据持久化

Coolify 会自动管理 Docker volumes。数据存储在：

- `sitepod-data` volume → `/data` 目录
  - `blobs/` - 静态文件（内容寻址）
  - `refs/` - 环境指针
  - `sitepod.db` - SQLite 数据库

## 备份

```bash
# 在 Coolify 服务器上
docker volume inspect sitepod-data  # 查看 volume 位置

# 备份
tar -czvf sitepod-backup-$(date +%Y%m%d).tar.gz /var/lib/docker/volumes/sitepod-data/_data
```

## 更新

在 Coolify 中点击 **Redeploy** 即可拉取最新代码并重新构建。
