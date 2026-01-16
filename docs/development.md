# SitePod 开发指南

## 本地开发环境

### 前置依赖

- Go 1.21+
- Rust 1.70+
- Make

### 构建

```bash
# 构建 server + CLI
make build

# 或分开构建
make build-server   # 构建 server
make build-cli      # 构建 CLI
```

### 启动服务器

```bash
# 启动本地服务器 (http://localhost:8080)
make run

# 或手动启动（指定端口）
SITEPOD_DOMAIN=localhost:9000 ./bin/sitepod-server run --config server/Caddyfile.local
```

### Clean Start（重置数据）

```bash
# 停止服务器
pkill sitepod-server

# 方式一：完全清理（数据 + 编译产物）
make clean
make build
make run

# 方式二：只清理数据
rm -rf data/
make run
```

### 数据目录

| 路径 | 内容 |
|------|------|
| `data/data.db` | SQLite 数据库（用户、项目、镜像记录） |
| `data/blobs/` | 部署的静态文件（内容寻址） |
| `data/refs/` | 环境指针（prod/beta） |
| `bin/` | 编译后的二进制文件 |

---

## CLI 本地测试

### 登录到本地服务器

```bash
# 使用 HTTP（本地开发）
./bin/sitepod login --endpoint http://localhost:8080

# 使用 IP 地址（局域网测试）
./bin/sitepod login --endpoint http://192.168.1.100:8080
```

登录时选择 **Anonymous (quick start, 24h limit)** 可跳过邮箱验证。

### 部署示例站点

```bash
cd examples/simple-site

# 部署到 beta
../../bin/sitepod deploy
# 访问: http://demo-site-beta.localhost:8080

# 部署到 prod
../../bin/sitepod deploy --prod
# 访问: http://demo-site.localhost:8080
```

### 测试自定义站点

```bash
cd /path/to/your/site

# 初始化项目
/path/to/sitepod/bin/sitepod init
# 输入项目名，如: my-test

# 部署
/path/to/sitepod/bin/sitepod deploy

# 访问
# http://my-test-beta.localhost:8080
```

### CLI 命令参考

```bash
./bin/sitepod --help              # 查看所有命令

# 认证
./bin/sitepod login               # 登录（交互式）
./bin/sitepod login --endpoint URL  # 指定服务器

# 部署
./bin/sitepod init                # 初始化项目（创建 sitepod.toml）
./bin/sitepod deploy              # 部署到 beta
./bin/sitepod deploy --prod       # 部署到 prod
./bin/sitepod preview             # 创建预览链接

# 管理
./bin/sitepod history             # 查看部署历史
./bin/sitepod rollback            # 回滚版本（交互式）

# 全局选项
./bin/sitepod --endpoint URL <command>  # 覆盖配置的服务器地址
```

---

## HTTP/HTTPS 模式

### 本地开发（HTTP）

CLI 默认支持 HTTP，无需特殊配置：

```bash
./bin/sitepod login --endpoint http://localhost:8080
./bin/sitepod login --endpoint http://192.168.1.100:8080
```

### 生产环境（HTTPS）

```bash
./bin/sitepod login --endpoint https://sitepod.example.com
```

### 服务器端 HTTP 模式

使用 `server/Caddyfile.local`（已禁用 HTTPS）：

```caddyfile
{
    admin off
    auto_https off      # 禁用自动 HTTPS
    order sitepod first
}

http://:8080, http://*.localhost:8080 {
    sitepod {
        storage_path ./data
        data_dir ./data
        domain localhost:8080
    }
}
```

---

## 运行测试

```bash
# 全部测试
make test

# 只测试 server
make test-server
# 或: cd server && go test ./...

# 只测试 CLI
make test-cli
# 或: cd cli && cargo test
```

---

## 项目结构

```
sitepod.dev/
├── server/                    # Go 服务端
│   ├── cmd/caddy/            # 入口（Caddy + 嵌入式 API）
│   ├── internal/
│   │   ├── caddy/            # Caddy 模块（API + 静态文件服务）
│   │   ├── storage/          # 存储后端（local, S3）
│   │   └── gc/               # 垃圾回收
│   └── migrations/           # 数据库迁移
├── cli/                       # Rust CLI
│   └── src/
│       ├── main.rs           # 入口
│       ├── api.rs            # API 客户端
│       ├── config.rs         # 配置管理
│       └── commands/         # 命令实现
├── examples/                  # 示例站点
├── docs/                      # 文档
└── Makefile
```

---

## 调试技巧

### 查看服务器日志

```bash
# make run 会输出日志到终端
# 日志级别是 DEBUG，包含详细请求信息
```

### 查看数据库内容

```bash
sqlite3 data/data.db

# 常用查询
.tables                          # 查看所有表
SELECT * FROM projects;          # 查看项目
SELECT * FROM images;            # 查看镜像
SELECT * FROM _admins;           # 查看管理员
```

### 查看 CLI 配置

```bash
cat ~/.config/sitepod/config.toml

# 包含：
# - endpoint: 服务器地址
# - token: 认证令牌
```

### 清理 CLI 配置

```bash
rm -rf ~/.config/sitepod/
```
