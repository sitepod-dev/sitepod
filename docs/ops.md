# SitePod — Self-hosted static releases

## 运维手册 (OPS)

**版本**: v1.0
**日期**: 2025-01-15
**项目主页**: https://sitepod.dev

> 产品需求详见 [prd.md](./prd.md)
> 技术设计详见 [tdd.md](./tdd.md)
> 品牌规范详见 [brand.md](./brand.md)

---

## 1. 部署指南

### 1.1 系统要求

| 项目 | 最低配置 | 推荐配置 |
|------|----------|----------|
| CPU | 1 核 | 2 核+ |
| 内存 | 512MB | 2GB+ |
| 磁盘 | 10GB | 50GB+ (取决于项目数量) |
| 系统 | Linux/macOS/Windows | Linux |

### 1.2 单二进制部署

```bash
# 下载
curl -fsSL https://github.com/sitepod-dev/sitepod/releases/latest/download/sitepod-linux-amd64 -o sitepod
chmod +x sitepod

# 运行
./sitepod serve --http :8080 --data /var/sitepod-data
```

### 1.3 Docker 部署

```yaml
# docker-compose.yml
version: '3.8'
services:
  sitepod:
    image: ghcr.io/sitepod-dev/sitepod:latest
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - sitepod-data:/data
    environment:
      - SITEPOD_DOMAIN=example.com
      - SITEPOD_ADMIN_EMAIL=admin@example.com
      - SITEPOD_ADMIN_PASSWORD=YourSecurePassword123
    restart: unless-stopped

volumes:
  sitepod-data:
```

```bash
docker-compose up -d
```

### 1.4 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SITEPOD_DOMAIN` | 主域名 | 必填 |
| `SITEPOD_DATA_DIR` | 数据目录 | `/data` |
| `SITEPOD_ADMIN_EMAIL` | PocketBase 管理员邮箱（仅 PB 管理后台） | `admin@sitepod.local` |
| `SITEPOD_ADMIN_PASSWORD` | PocketBase 管理员密码（仅 PB 管理后台） | `sitepod123` |
| `SITEPOD_CONSOLE_ADMIN_EMAIL` | Console 管理员邮箱（users.is_admin） | - |
| `SITEPOD_CONSOLE_ADMIN_PASSWORD` | Console 管理员密码（users.is_admin） | - |
| `SITEPOD_STORAGE_TYPE` | 存储类型 (local/s3/oss/r2) | `local` |
| `SITEPOD_S3_BUCKET` | S3 桶名 | - |
| `SITEPOD_S3_REGION` | S3 区域 | - |
| `SITEPOD_S3_ACCESS_KEY` | S3 Access Key | - |
| `SITEPOD_S3_SECRET_KEY` | S3 Secret Key | - |

> **安全提示**: 生产环境请设置 `SITEPOD_ADMIN_EMAIL` / `SITEPOD_ADMIN_PASSWORD`（PB 管理后台）以及 `SITEPOD_CONSOLE_ADMIN_EMAIL` / `SITEPOD_CONSOLE_ADMIN_PASSWORD`（Console 管理员）。

---

## 2. 配置说明

### 2.1 服务端配置文件

```toml
# /etc/sitepod/config.toml

[server]
http_addr = ":80"
https_addr = ":443"
admin_addr = ":8090"  # 内部 API

[domain]
primary = "example.com"
acme_email = "admin@example.com"

[storage]
type = "local"  # local | s3 | oss | r2
path = "/data/blobs"

# S3 配置 (type = "s3" 时)
[storage.s3]
bucket = "my-sitepod-bucket"
region = "us-east-1"
access_key = "${SITEPOD_S3_ACCESS_KEY}"
secret_key = "${SITEPOD_S3_SECRET_KEY}"

[database]
path = "/data/sitepod.db"

[cache]
manifest_ttl = "5s"
max_entries = 1000

[gc]
enabled = true
interval = "24h"
grace_period = "1h"
min_versions = 5
keep_days = 30

[log]
level = "info"  # debug | info | warn | error
format = "json"  # json | text
```

### 2.2 CLI 配置

```toml
# ~/.sitepod/config.toml (全局)
# ./sitepod.toml (项目级)

[server]
endpoint = "https://sitepod.example.com"

[auth]
token = "xxx"  # 或使用环境变量 SITEPOD_TOKEN

[project]
name = "my-app"

[build]
directory = "./dist"

[deploy]
ignore = ["**/*.map", ".*", "node_modules/**"]
concurrent = 20
```

---

## 3. 常用运维命令

### 3.1 服务管理

```bash
# 启动服务
./sitepod serve

# 后台运行 (systemd)
sudo systemctl start sitepod
sudo systemctl enable sitepod

# 查看状态
sudo systemctl status sitepod

# 查看日志
journalctl -u sitepod -f
```

### 3.2 数据管理

```bash
# 手动触发 GC
./sitepod gc --dry-run  # 预览
./sitepod gc            # 执行

# 导出数据
./sitepod export --output /backup/sitepod-backup.tar.gz

# 导入数据
./sitepod import --input /backup/sitepod-backup.tar.gz

# 数据库备份 (SQLite)
sqlite3 /data/sitepod.db ".backup /backup/sitepod.db"
```

### 3.3 用户管理

```bash
# 创建管理员
./sitepod admin create --email admin@example.com --password xxx

# 创建 API Token
./sitepod token create --name "ci-cd" --scope "deploy,preview"

# 列出 Token
./sitepod token list

# 撤销 Token
./sitepod token revoke <token-id>
```

### 3.4 项目管理

```bash
# 列出项目
./sitepod project list

# 查看项目详情
./sitepod project info my-app

# 删除项目 (危险操作)
./sitepod project delete my-app --confirm
```

---

## 4. 监控

### 4.1 健康检查

```bash
# HTTP 健康检查
curl http://localhost:8090/health

# 响应
{
  "status": "healthy",
  "database": "ok",
  "storage": "ok",
  "uptime": "72h30m"
}
```

### 4.2 Prometheus 指标

```bash
# 指标端点
curl http://localhost:8090/metrics
```

关键指标:

| 指标 | 说明 | 告警阈值建议 |
|------|------|--------------|
| `sitepod_deploys_total` | 部署总数 | - |
| `sitepod_deploy_errors_total` | 部署失败数 | > 5/min |
| `sitepod_storage_bytes` | 存储使用量 | > 80% 磁盘 |
| `sitepod_http_request_duration_seconds` | 请求延迟 | p99 > 1s |
| `sitepod_gc_duration_seconds` | GC 耗时 | > 1h |

### 4.3 告警规则

```yaml
# prometheus-alerts.yml
groups:
  - name: sitepod
    rules:
      - alert: SitePodHighErrorRate
        expr: rate(sitepod_deploy_errors_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "SitePod 部署错误率过高"

      - alert: SitePodStorageHigh
        expr: sitepod_storage_bytes / node_filesystem_size_bytes > 0.8
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "SitePod 存储使用率超过 80%"

      - alert: SitePodDown
        expr: up{job="sitepod"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "SitePod 服务不可用"
```

---

## 5. 故障排查

### 5.1 常见问题

#### 问题: 部署超时

```bash
# 检查网络连接
curl -v https://sitepod.example.com/health

# 检查服务日志
journalctl -u sitepod --since "10 minutes ago" | grep -i error

# 检查磁盘空间
df -h /data
```

#### 问题: 版本切换不生效

```bash
# 1. 检查 ref 文件是否写入成功
cat /data/refs/my-app/prod.json

# 2. 检查缓存是否失效
curl http://localhost:8090/debug/cache

# 3. 强制刷新缓存
curl -X POST http://localhost:8090/admin/cache/invalidate?project=my-app&env=prod
```

#### 问题: Blob 上传失败

```bash
# 检查存储后端连接
./sitepod storage check

# 检查文件权限
ls -la /data/blobs

# S3 存储时检查凭证
aws s3 ls s3://my-sitepod-bucket/
```

#### 问题: HTTPS 证书问题

```bash
# 检查证书状态
./sitepod cert status

# 强制更新证书
./sitepod cert renew --force

# 检查 ACME 日志
journalctl -u sitepod | grep -i acme
```

### 5.2 日志分析

```bash
# 查找错误日志
journalctl -u sitepod | jq 'select(.level == "error")'

# 按请求追踪
journalctl -u sitepod | jq 'select(.request_id == "xxx")'

# 统计错误类型
journalctl -u sitepod --since today | jq -r 'select(.level == "error") | .error_code' | sort | uniq -c
```

### 5.3 性能调优

```bash
# 查看慢请求
journalctl -u sitepod | jq 'select(.duration_ms > 1000)'

# 检查 SQLite 性能
sqlite3 /data/sitepod.db "PRAGMA integrity_check"
sqlite3 /data/sitepod.db "ANALYZE"

# 检查缓存命中率
curl http://localhost:8090/debug/cache/stats
```

---

## 6. 备份与恢复

### 6.1 备份策略

| 组件 | 备份频率 | 保留时间 |
|------|----------|----------|
| SQLite 数据库 | 每日 | 30 天 |
| Blobs (增量) | 每周 | 永久 |
| 配置文件 | 变更时 | Git 管理 |

### 6.2 备份脚本

```bash
#!/bin/bash
# /usr/local/bin/sitepod-backup.sh

set -e

BACKUP_DIR="/backup/sitepod"
DATE=$(date +%Y%m%d)

# 1. SQLite 在线备份
mkdir -p $BACKUP_DIR/db
sqlite3 /data/sitepod.db ".backup $BACKUP_DIR/db/sitepod-$DATE.db"

# 2. 压缩旧备份
find $BACKUP_DIR/db -name "*.db" -mtime +7 -exec gzip {} \;

# 3. 清理过期备份
find $BACKUP_DIR/db -name "*.gz" -mtime +30 -delete

# 4. 同步 Blobs 到远程 (可选)
# rclone sync /data/blobs remote:sitepod-blobs-backup
```

### 6.3 恢复流程

```bash
# 1. 停止服务
sudo systemctl stop sitepod

# 2. 恢复数据库
cp /backup/sitepod/db/sitepod-20250115.db /data/sitepod.db

# 3. 验证数据完整性
sqlite3 /data/sitepod.db "PRAGMA integrity_check"

# 4. 启动服务
sudo systemctl start sitepod

# 5. 验证服务
curl http://localhost:8090/health
```

---

## 7. 升级指南

### 7.1 升级前检查

```bash
# 1. 检查当前版本
./sitepod version

# 2. 查看 changelog
curl https://github.com/sitepod-dev/sitepod/releases

# 3. 备份数据
./sitepod-backup.sh

# 4. 在测试环境验证
```

### 7.2 升级步骤

```bash
# 1. 下载新版本
curl -fsSL https://github.com/sitepod-dev/sitepod/releases/download/vX.Y.Z/sitepod-linux-amd64 -o sitepod-new

# 2. 停止服务
sudo systemctl stop sitepod

# 3. 替换二进制
mv sitepod sitepod.bak
mv sitepod-new sitepod
chmod +x sitepod

# 4. 运行数据库迁移 (如需要)
./sitepod migrate

# 5. 启动服务
sudo systemctl start sitepod

# 6. 验证
curl http://localhost:8090/health
./sitepod version
```

### 7.3 回滚

```bash
# 1. 停止服务
sudo systemctl stop sitepod

# 2. 恢复旧版本
mv sitepod sitepod-failed
mv sitepod.bak sitepod

# 3. 恢复数据库 (如有迁移)
cp /backup/sitepod/db/sitepod-pre-upgrade.db /data/sitepod.db

# 4. 启动服务
sudo systemctl start sitepod
```

---

## 8. 安全加固

### 8.1 网络安全

```bash
# 限制 Admin API 访问
iptables -A INPUT -p tcp --dport 8090 -s 127.0.0.1 -j ACCEPT
iptables -A INPUT -p tcp --dport 8090 -j DROP

# 或使用防火墙规则
ufw allow from 192.168.1.0/24 to any port 8090
```

### 8.2 文件权限

```bash
# 数据目录权限
chown -R sitepod:sitepod /data
chmod 750 /data
chmod 640 /data/sitepod.db

# 配置文件权限
chmod 600 /etc/sitepod/config.toml
```

### 8.3 Token 管理

```bash
# 定期轮换 API Token
./sitepod token rotate --name "ci-cd"

# 审计 Token 使用
./sitepod token audit --name "ci-cd" --since "7 days ago"
```

---

## 9. Systemd 服务配置

```ini
# /etc/systemd/system/sitepod.service
[Unit]
Description=SitePod Static Site Deployment Platform
After=network.target

[Service]
Type=simple
User=sitepod
Group=sitepod
ExecStart=/usr/local/bin/sitepod serve --config /etc/sitepod/config.toml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# 安全限制
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/data

# 资源限制
LimitNOFILE=65535
MemoryMax=2G

[Install]
WantedBy=multi-user.target
```

```bash
# 重新加载配置
sudo systemctl daemon-reload
sudo systemctl enable sitepod
sudo systemctl start sitepod
```

---

## 附录: 快速参考

### A. 常用 CLI 命令

| 命令 | 说明 |
|------|------|
| `sitepod login` | 登录 |
| `sitepod deploy` | 部署到 beta |
| `sitepod deploy --prod` | 部署到 prod |
| `sitepod preview` | 创建预览 |
| `sitepod rollback` | 回滚 |
| `sitepod history` | 查看历史 |

### B. 常用 API 端点

| 端点 | 说明 |
|------|------|
| `GET /health` | 健康检查 |
| `GET /metrics` | Prometheus 指标 |
| `POST /api/v1/plan` | 提交部署计划 |
| `POST /api/v1/commit` | 确认部署 |
| `POST /api/v1/release` | 发布版本 |
| `POST /api/v1/rollback` | 回滚版本 |

### C. 重要文件路径

| 路径 | 说明 |
|------|------|
| `/data/sitepod.db` | SQLite 数据库 |
| `/data/blobs/` | Blob 存储 |
| `/data/refs/` | Ref 文件 (数据面 SSOT) |
| `/etc/sitepod/config.toml` | 服务配置 |
| `~/.sitepod/config.toml` | CLI 全局配置 |
| `./sitepod.toml` | 项目配置 |
