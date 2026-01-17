# SitePod 独立服务器部署指南

本文档详细说明如何在一台独立的 VPS/云服务器上部署 SitePod。

> 📖 关于部署拓扑的详细说明，请参阅 [部署拓扑文档](deploy-topology.md)。本文档涵盖 **拓扑 A（单机直出）** 和 **拓扑 B（反向代理后端）**。

## 目录

- [前置要求](#前置要求)
- [服务器准备](#服务器准备)
- [DNS 配置](#dns-配置)
- [安装部署](#安装部署)
  - [方式一：Docker 一键部署](#方式一docker-一键部署推荐)
  - [方式二：Docker Compose](#方式二docker-compose-部署)
  - [方式三：R2 存储](#方式三使用-cloudflare-r2-存储)
  - [方式四：已有反向代理](#方式四已有-nginxcaddy-反向代理)
- [验证部署](#验证部署)
- [CLI 安装](#cli-安装)
- [第一次部署](#第一次部署)
- [运维管理](#运维管理)
- [常见问题](#常见问题)

---

## 前置要求

### 服务器要求

| 项目 | 最低配置 | 推荐配置 |
|------|---------|---------|
| CPU | 1 核 | 2 核 |
| 内存 | 512MB | 1GB |
| 磁盘 | 10GB | 20GB+ |
| 系统 | Linux (amd64/arm64) | Ubuntu 22.04 LTS |

### 网络要求

- 公网 IP 地址
- 开放端口：80 (HTTP)、443 (HTTPS)
- 一个域名（用于访问 SitePod 和部署的站点）

### 软件要求

- Docker 20.10+
- Docker Compose (可选，推荐)

---

## 服务器准备

### 1. 安装 Docker

**Ubuntu/Debian:**

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装 Docker
curl -fsSL https://get.docker.com | sh

# 将当前用户加入 docker 组（避免每次 sudo）
sudo usermod -aG docker $USER

# 重新登录使权限生效
exit
# 重新 SSH 登录
```

**CentOS/RHEL:**

```bash
sudo yum install -y yum-utils
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo yum install -y docker-ce docker-ce-cli containerd.io
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker $USER
```

### 2. 开放防火墙端口

**Ubuntu (UFW):**

```bash
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw reload
```

**CentOS (firewalld):**

```bash
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
```

**云服务商安全组:**

在阿里云、腾讯云、AWS 等控制台的安全组中，添加入站规则：
- TCP 80
- TCP 443

---

## DNS 配置

假设你的域名是 `sitepod.example.com`，服务器 IP 是 `1.2.3.4`。

只需要 **2 条 DNS 记录**：

| 类型 | 主机记录 | 记录值 | 说明 |
|------|---------|--------|------|
| A | sitepod | 1.2.3.4 | 主域名 (Console + API) |
| A | *.sitepod | 1.2.3.4 | 用户站点 |

部署后的访问地址：

| URL | 用途 |
|-----|------|
| `https://sitepod.example.com` | Console + API |
| `https://myapp.sitepod.example.com` | 用户站点 (生产环境) |
| `https://myapp-beta.sitepod.example.com` | 用户站点 (Beta 环境) |
| `https://welcome.sitepod.example.com` | 欢迎页 |

> **注意**：Beta 环境使用 `-beta` 后缀而非子域名，因此只需要一个通配符记录。

### 示例：使用其他子域名

如果想把 SitePod 部署在 `pods.example.com` 下：

| 类型 | 主机记录 | 记录值 |
|------|---------|--------|
| A | pods | 1.2.3.4 |
| A | *.pods | 1.2.3.4 |

### 验证 DNS

等待 DNS 生效后验证：

```bash
# 在本地执行
dig sitepod.example.com
dig test.sitepod.example.com

# 应该都返回你的服务器 IP
```

---

## 安装部署

### 方式一：Docker 一键部署（推荐）

```bash
# 创建数据目录
sudo mkdir -p /opt/sitepod

# 启动 SitePod
docker run -d \
  --name sitepod \
  --restart unless-stopped \
  -p 80:80 \
  -p 443:443 \
  -v /opt/sitepod:/data \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  -e SITEPOD_ADMIN_EMAIL=admin@example.com \
  -e SITEPOD_ADMIN_PASSWORD=YourSecurePassword123 \
  ghcr.io/sitepod-dev/sitepod:latest
```

**环境变量说明：**

| 变量 | 必填 | 默认值 | 说明 |
|------|-----|--------|------|
| `SITEPOD_DOMAIN` | 是 | - | 你的域名 |
| `SITEPOD_ADMIN_EMAIL` | 否 | `admin@sitepod.local` | 管理员邮箱 |
| `SITEPOD_ADMIN_PASSWORD` | 否 | `sitepod123` | 管理员密码（首次启动若未设置） |
| `SITEPOD_STORAGE_TYPE` | 否 | `local` | 存储类型：local, s3, oss, r2 |

**配额限制（可选）：**

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `SITEPOD_MAX_FILES_PER_DEPLOY` | `10000` | 单次部署最大文件数 |
| `SITEPOD_MAX_FILE_SIZE` | `104857600` | 单文件最大大小（100MB） |
| `SITEPOD_MAX_DEPLOY_SIZE` | `524288000` | 单次部署最大总大小（500MB） |
| `SITEPOD_MAX_PROJECTS_PER_USER` | `100` | 用户最大项目数 |
| `SITEPOD_ANON_MAX_PROJECTS` | `5` | 匿名用户最大项目数 |
| `SITEPOD_ANON_MAX_DEPLOY_SIZE` | `52428800` | 匿名用户部署大小限制（50MB） |

> **安全提示**: 生产环境建议设置 `SITEPOD_ADMIN_EMAIL` 和 `SITEPOD_ADMIN_PASSWORD`！

### 方式二：Docker Compose 部署

创建 `/opt/sitepod/docker-compose.yml`：

```yaml
version: "3.8"

services:
  sitepod:
    image: ghcr.io/sitepod-dev/sitepod:latest
    container_name: sitepod
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    environment:
      - SITEPOD_DOMAIN=sitepod.example.com
      - SITEPOD_ADMIN_EMAIL=admin@example.com
      - SITEPOD_ADMIN_PASSWORD=YourSecurePassword123
      - SITEPOD_STORAGE_TYPE=local
    volumes:
      - ./data:/data
      - caddy-data:/caddy-data
      - caddy-config:/caddy-config

volumes:
  caddy-data:
  caddy-config:
```

启动：

```bash
cd /opt/sitepod
docker compose up -d
```

### 方式三：使用 Cloudflare R2 存储

如果你希望将静态文件存储在 Cloudflare R2：

```bash
docker run -d \
  --name sitepod \
  --restart unless-stopped \
  -p 80:80 \
  -p 443:443 \
  -v /opt/sitepod:/data \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  -e SITEPOD_ADMIN_EMAIL=admin@example.com \
  -e SITEPOD_ADMIN_PASSWORD=YourSecurePassword123 \
  -e SITEPOD_STORAGE_TYPE=r2 \
  -e SITEPOD_S3_BUCKET=your-bucket-name \
  -e SITEPOD_S3_REGION=auto \
  -e SITEPOD_S3_ENDPOINT=https://YOUR_ACCOUNT_ID.r2.cloudflarestorage.com \
  -e AWS_ACCESS_KEY_ID=your-access-key \
  -e AWS_SECRET_ACCESS_KEY=your-secret-key \
  ghcr.io/sitepod-dev/sitepod:latest
```

### 方式四：已有 Nginx/Caddy 反向代理

如果服务器上已经运行了 Nginx 或 Caddy 处理其他服务，SitePod 需要运行在内部端口，由现有的反向代理转发流量。

#### 架构说明

```
                                    ┌─────────────────────┐
Internet ──► Nginx/Caddy (80/443) ──┤ SSL 终止            │
                    │               │ 路由到不同服务      │
                    │               └─────────────────────┘
                    │
        ┌───────────┼───────────┐
        ▼           ▼           ▼
   其他服务    SitePod:8080   其他服务
```

#### 步骤 1：创建 SitePod 配置文件

```bash
# 创建配置目录
sudo mkdir -p /opt/sitepod

# 创建 Caddyfile（SitePod 内部用，禁用 HTTPS）
cat > /opt/sitepod/Caddyfile << 'EOF'
{
    admin off
    auto_https off
    order sitepod first
}

:8080 {
    sitepod {
        storage_path /data
        data_dir /data
        domain {$SITEPOD_DOMAIN}
    }
}
EOF
```

#### 步骤 2：启动 SitePod 容器

```bash
docker run -d \
  --name sitepod \
  --restart unless-stopped \
  -p 127.0.0.1:8080:8080 \
  -v /opt/sitepod/data:/data \
  -v /opt/sitepod/Caddyfile:/etc/caddy/Caddyfile \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  -e SITEPOD_ADMIN_EMAIL=admin@example.com \
  -e SITEPOD_ADMIN_PASSWORD=YourSecurePassword123 \
  ghcr.io/sitepod-dev/sitepod:latest
```

> **注意**: `-p 127.0.0.1:8080:8080` 只绑定本地，不对外暴露。

#### 步骤 3：配置 Nginx 反向代理

在 `/etc/nginx/sites-available/sitepod` 创建配置：

```nginx
# SitePod 主域名和泛域名
server {
    listen 80;
    listen [::]:80;
    server_name sitepod.example.com *.sitepod.example.com;

    # 重定向到 HTTPS
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name sitepod.example.com *.sitepod.example.com;

    # SSL 证书（使用 certbot 或其他方式获取）
    ssl_certificate /etc/letsencrypt/live/sitepod.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/sitepod.example.com/privkey.pem;

    # SSL 配置
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers off;

    # 代理到 SitePod
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket 支持（如果需要）
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        # 上传文件大小限制
        client_max_body_size 100M;
    }
}
```

启用配置：

```bash
sudo ln -s /etc/nginx/sites-available/sitepod /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

#### 步骤 3（替代）：配置 Caddy 反向代理

如果使用 Caddy 作为主反向代理，在 `/etc/caddy/Caddyfile` 添加：

```caddyfile
# SitePod 泛域名
sitepod.example.com, *.sitepod.example.com {
    reverse_proxy localhost:8080
}
```

重载 Caddy：

```bash
sudo systemctl reload caddy
```

#### 步骤 4：申请泛域名 SSL 证书

**使用 Certbot + Cloudflare DNS 验证：**

```bash
# 安装 certbot 和 cloudflare 插件
sudo apt install certbot python3-certbot-dns-cloudflare

# 创建 Cloudflare API 凭证文件
cat > ~/.cloudflare.ini << EOF
dns_cloudflare_api_token = YOUR_CLOUDFLARE_API_TOKEN
EOF
chmod 600 ~/.cloudflare.ini

# 申请泛域名证书
sudo certbot certonly \
  --dns-cloudflare \
  --dns-cloudflare-credentials ~/.cloudflare.ini \
  -d sitepod.example.com \
  -d "*.sitepod.example.com"
```

**使用 acme.sh + 阿里云 DNS：**

```bash
# 安装 acme.sh
curl https://get.acme.sh | sh

# 设置阿里云 API
export Ali_Key="YOUR_ACCESS_KEY"
export Ali_Secret="YOUR_SECRET_KEY"

# 申请证书
acme.sh --issue --dns dns_ali \
  -d sitepod.example.com \
  -d "*.sitepod.example.com"

# 安装证书到 Nginx
acme.sh --install-cert -d sitepod.example.com \
  --key-file /etc/nginx/ssl/sitepod.key \
  --fullchain-file /etc/nginx/ssl/sitepod.crt \
  --reloadcmd "systemctl reload nginx"
```

#### Docker Compose 版本

```yaml
version: "3.8"

services:
  sitepod:
    image: ghcr.io/sitepod-dev/sitepod:latest
    container_name: sitepod
    restart: unless-stopped
    ports:
      - "127.0.0.1:8080:8080"  # 只绑定本地
    environment:
      - SITEPOD_DOMAIN=sitepod.example.com
      - SITEPOD_ADMIN_EMAIL=admin@example.com
      - SITEPOD_ADMIN_PASSWORD=YourSecurePassword123
    volumes:
      - ./data:/data
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
```

---

## 验证部署

### 1. 检查容器状态

```bash
docker ps
# 应该看到 sitepod 容器状态为 Up

docker logs sitepod
# 查看启动日志，确认没有错误
```

### 2. 健康检查

```bash
# 本地检查
curl http://localhost/api/v1/health

# 应该返回：
# {"status":"healthy","database":"ok","storage":"ok","uptime":"..."}
```

### 3. 通过域名访问

在浏览器中访问：
- `https://sitepod.example.com/api/v1/health`

首次访问时，Caddy 会自动申请 Let's Encrypt SSL 证书。

### 4. 访问管理后台

访问 `https://sitepod.example.com/_/` 进入 PocketBase 管理后台。

使用你设置的管理员邮箱和密码登录。

---

## CLI 安装

在你的**开发机器**（不是服务器）上安装 SitePod CLI：

### macOS

```bash
# 使用 Homebrew
brew install sitepod/tap/sitepod

# 或直接下载
curl -fsSL https://github.com/sitepod-dev/sitepod/releases/latest/download/sitepod-darwin-arm64 -o /usr/local/bin/sitepod
chmod +x /usr/local/bin/sitepod
```

### Linux

```bash
curl -fsSL https://github.com/sitepod-dev/sitepod/releases/latest/download/sitepod-linux-amd64 -o /usr/local/bin/sitepod
chmod +x /usr/local/bin/sitepod
```

### 验证安装

```bash
sitepod --version
```

---

## 第一次部署

### 1. 登录

```bash
sitepod login --endpoint https://sitepod.example.com

# 选择登录方式：
# - Email: 输入邮箱，收到验证码后输入
# - Anonymous: 快速测试，24小时后过期
```

### 2. 初始化项目

在你的静态站点项目目录中：

```bash
cd my-website
sitepod init

# 按提示输入项目名称，如: my-website
# 会创建 sitepod.toml 配置文件
```

### 3. 部署

```bash
# 部署到 Beta 环境
sitepod deploy

# 部署到生产环境
sitepod deploy --prod
```

部署成功后会显示访问地址：
- Beta: `https://my-website-beta.sitepod.example.com`
- Prod: `https://my-website.sitepod.example.com`

### 4. 查看部署历史

```bash
sitepod history
```

### 5. 回滚

```bash
sitepod rollback
# 交互式选择要回滚到的版本
```

---

## 运维管理

### 查看日志

```bash
# 实时日志
docker logs -f sitepod

# 最近 100 行
docker logs --tail 100 sitepod
```

### 更新版本

```bash
# 拉取最新镜像
docker pull ghcr.io/sitepod-dev/sitepod:latest

# 停止并删除旧容器
docker stop sitepod
docker rm sitepod

# 重新启动（使用相同的参数）
docker run -d \
  --name sitepod \
  --restart unless-stopped \
  -p 80:80 \
  -p 443:443 \
  -v /opt/sitepod:/data \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  -e SITEPOD_ADMIN_EMAIL=admin@example.com \
  -e SITEPOD_ADMIN_PASSWORD=YourSecurePassword123 \
  ghcr.io/sitepod-dev/sitepod:latest
```

或使用 Docker Compose：

```bash
cd /opt/sitepod
docker compose pull
docker compose up -d
```

### 数据备份

```bash
# 停止容器（可选，确保数据一致性）
docker stop sitepod

# 备份数据目录
tar -czvf sitepod-backup-$(date +%Y%m%d).tar.gz /opt/sitepod

# 重新启动
docker start sitepod
```

### 数据恢复

```bash
# 停止容器
docker stop sitepod

# 恢复数据
tar -xzvf sitepod-backup-20240101.tar.gz -C /

# 重新启动
docker start sitepod
```

### 监控

**健康检查脚本** (`/opt/sitepod/healthcheck.sh`)：

```bash
#!/bin/bash
response=$(curl -s http://localhost/api/v1/health)
status=$(echo $response | jq -r '.status')

if [ "$status" != "healthy" ]; then
    echo "SitePod is unhealthy: $response"
    # 发送告警通知
    # curl -X POST https://your-webhook-url -d "SitePod is down"
    exit 1
fi

echo "SitePod is healthy"
```

添加到 crontab：

```bash
# 每 5 分钟检查一次
*/5 * * * * /opt/sitepod/healthcheck.sh >> /var/log/sitepod-health.log 2>&1
```

---

## 常见问题

### Q: SSL 证书申请失败

**症状**: 访问 HTTPS 时提示证书错误

**解决方案**:
1. 确认端口 80 和 443 对外开放
2. 确认 DNS 已正确指向服务器 IP
3. 查看日志：`docker logs sitepod | grep -i cert`
4. 等待几分钟让 Caddy 重试

### Q: 泛域名证书问题

**症状**: `*.sitepod.example.com` 证书申请失败

**原因**: Let's Encrypt 的 HTTP 验证不支持泛域名

**解决方案**: 使用 DNS 验证（需要 Cloudflare 等 DNS 服务商）

```bash
# 下载 DNS 验证配置
curl -O https://raw.githubusercontent.com/sitepod/sitepod/main/server/examples/Caddyfile.wildcard

# 修改配置中的域名和 API Token

# 重新启动并挂载配置
docker run -d \
  --name sitepod \
  -p 80:80 -p 443:443 \
  -v /opt/sitepod:/data \
  -v ./Caddyfile.wildcard:/etc/caddy/Caddyfile \
  -e SITEPOD_DOMAIN=sitepod.example.com \
  -e CF_API_TOKEN=your-cloudflare-api-token \
  ghcr.io/sitepod-dev/sitepod:latest
```

### Q: 子域名无法访问

**症状**: 部署成功但访问 `myapp.sitepod.example.com` 返回错误

**排查步骤**:
1. 检查 DNS：`dig myapp.sitepod.example.com`
2. 确认泛域名记录已配置
3. 检查 `SITEPOD_DOMAIN` 环境变量是否正确
4. 查看容器日志

### Q: 磁盘空间不足

**症状**: 部署失败，日志显示磁盘空间错误

**解决方案**:
1. 检查磁盘使用：`df -h`
2. 清理 Docker 无用镜像：`docker system prune -a`
3. 清理旧版本（GC 会自动清理，但可以手动触发）

### Q: 数据库锁定

**症状**: 日志显示 `database is locked`

**原因**: SQLite 不支持多实例并发写入

**解决方案**:
1. 确保只运行一个 SitePod 实例
2. 如果需要高可用，考虑使用 S3/R2 存储

### Q: 如何修改管理员密码

**方法 1**: 通过环境变量（推荐）
- 修改 `SITEPOD_ADMIN_PASSWORD` 环境变量
- 删除数据库重新初始化（会丢失数据）

**方法 2**: 通过管理后台
1. 访问 `https://sitepod.example.com/_/`
2. 使用当前密码登录
3. 在设置中修改密码

---

## 架构说明

```
┌─────────────────────────────────────────────────────────────┐
│                        Internet                              │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Your VPS Server                          │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                   Docker Container                     │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │              SitePod (Single Binary)             │  │  │
│  │  │  ┌─────────────┐    ┌──────────────────────┐   │  │  │
│  │  │  │    Caddy    │    │   Embedded PocketBase │   │  │  │
│  │  │  │  Port 80/443│───▶│   - API Handlers      │   │  │  │
│  │  │  │  Auto HTTPS │    │   - SQLite Database   │   │  │  │
│  │  │  │  Static Files│   │   - Auth              │   │  │  │
│  │  │  └─────────────┘    └──────────────────────┘   │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  │                          │                             │  │
│  │                          ▼                             │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │                    /data                         │  │  │
│  │  │  ├── blobs/      (静态文件，内容寻址)            │  │  │
│  │  │  ├── refs/       (环境指针)                      │  │  │
│  │  │  ├── data.db     (SQLite 数据库)                 │  │  │
│  │  │  └── previews/   (预览部署)                      │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────────┘  │
│                              │                               │
│                              ▼                               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │              Host Volume: /opt/sitepod                 │  │
│  │                    (数据持久化)                         │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

---

## 下一步

- 阅读 [CLI 完整文档](./cli.md) 了解更多命令
- 阅读 [自定义域名](./custom-domain.md) 配置独立域名
- 阅读 [API 文档](./api.md) 进行集成开发
