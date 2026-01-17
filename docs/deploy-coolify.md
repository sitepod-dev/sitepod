# SitePod Coolify 部署指南

本文档介绍如何在 Coolify 上部署 SitePod。

> 📖 关于部署拓扑的详细说明，请参阅 [部署拓扑文档](deploy-topology.md)。Coolify 部署属于 **拓扑 B（反向代理后端）**。

## 前置要求

- 已安装并运行的 Coolify 实例
- 一个域名，可配置 DNS

## 部署方式

Coolify 支持三种部署方式，任选其一：

| 方式 | 优点 | 适用场景 |
|------|------|---------|
| **Docker Image** | 最简单，无需构建 | 推荐，快速部署 |
| **Dockerfile** | 可自定义构建 | 需要修改构建过程 |
| **Docker Compose** | 完整配置 | 需要多服务编排 |

## 方式一：Docker Image（推荐）

最简单的部署方式，直接使用预构建镜像。

### 1. 创建服务

1. 登录 Coolify → **New Resource** → **Docker Image**
2. 镜像地址：`ghcr.io/sitepod-dev/sitepod:latest`
3. 端口映射：`8080`

### 2. 配置

**环境变量：**
```
SITEPOD_DOMAIN=sitepod.example.com
SITEPOD_PROXY_MODE=1
```

> **重要**：`SITEPOD_PROXY_MODE=1` 让 SitePod 监听 8080 端口并禁用 SSL，由 Coolify/Traefik 处理 SSL。

**持久化存储：**
- 挂载路径：`/data`
- 选择或创建 volume

**域名配置（添加两个）：**
- `sitepod.example.com`
- `*.sitepod.example.com`

### 3. 部署

点击 **Deploy**，几秒内即可完成。

---

## 方式二：Dockerfile

从源码构建。

### 1. 创建服务

1. 登录 Coolify → **New Resource** → **Dockerfile**
2. Git 仓库：`https://github.com/sitepod-dev/sitepod`
3. Dockerfile 路径：`Dockerfile`（默认即可）

### 2. 配置

**环境变量：**
```
SITEPOD_DOMAIN=sitepod.example.com
SITEPOD_PROXY_MODE=1
```

其他配置同方式一（存储、域名）。

### 3. 部署

点击 **Deploy**，首次构建需要几分钟（编译 Go + Rust）。

---

## 方式三：Docker Compose

使用 Compose 文件，配置更完整。

### 1. 创建服务

1. 登录 Coolify → **New Resource** → **Docker Compose**
2. Git 仓库：`https://github.com/sitepod-dev/sitepod`
3. Compose 文件：`docker-compose.coolify.yml`

### 2. 配置

环境变量在 Coolify 界面设置：
```
SITEPOD_DOMAIN=sitepod.example.com
SITEPOD_PROXY_MODE=1
```

域名配置同上。

### 3. 部署

点击 **Deploy**。

---

## DNS 配置

在你的 DNS 提供商添加以下记录，指向 Coolify 服务器：

| 类型 | 主机记录 | 记录值 |
|------|---------|--------|
| A | sitepod | Coolify 服务器 IP |
| A | *.sitepod | Coolify 服务器 IP |

> 假设你的域名是 `example.com`，则配置后：
> - 主域名 (Console + API)：`sitepod.example.com`
> - 项目域名 (生产)：`myapp.sitepod.example.com`
> - 项目域名 (Beta)：`myapp-beta.sitepod.example.com`

## 环境变量

| 变量 | 值 | 说明 |
|------|---|------|
| `SITEPOD_DOMAIN` | `sitepod.example.com` | **必填**，你的基础域名 |
| `SITEPOD_PROXY_MODE` | `1` | **Coolify 必填**，禁用 SSL，监听 8080 |
| `SITEPOD_STORAGE_TYPE` | `local` | 存储类型，默认 local |

**可选的 S3 存储配置**（如使用 Cloudflare R2）：

```
SITEPOD_STORAGE_TYPE=r2
SITEPOD_S3_BUCKET=your-bucket
SITEPOD_S3_ENDPOINT=https://ACCOUNT_ID.r2.cloudflarestorage.com
AWS_ACCESS_KEY_ID=xxx
AWS_SECRET_ACCESS_KEY=xxx
```

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
| `Dockerfile` | 主 Dockerfile，通过 `SITEPOD_PROXY_MODE=1` 切换代理模式 |
| `docker-compose.coolify.yml` | Coolify 专用 Compose，预设了代理模式配置 |
| `server/docker-entrypoint.sh` | 启动脚本，根据环境变量选择 Caddyfile |

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
