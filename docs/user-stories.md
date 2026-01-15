# SitePod 用户故事

从开发者体验 (DX) 角度描述的典型使用场景。

---

## 0. 极速体验：一条命令部署

**新用户第一次使用 SitePod，无需任何准备。**

```bash
$ cd my-project
$ npx sitepod deploy

# 检测到未登录，自动创建匿名账户
✓ 已创建匿名账户 (24小时有效)
  提示: 运行 sitepod bind 绑定邮箱，保留账户和部署

# 检测到没有 sitepod.toml，自动初始化
? 项目名称: my-project
? 子域名 (my-project): my-project
✓ my-project.sitepod.dev 可用
? 构建目录 (检测到 dist/): dist
✓ 已创建 sitepod.toml

# 自动部署
扫描 ./dist... 42 个文件
上传中... 42/42 (1.2 MB)
✓ 已部署到 beta

  https://my-project.beta.sitepod.dev

  ⚠ 匿名账户将在 24 小时后过期
  运行 sitepod bind 绑定邮箱以保留部署
```

**子域名冲突时：**

```bash
? 子域名 (my-blog): my-blog
✗ my-blog.sitepod.dev 已被占用
? 子域名: alice-blog
✓ alice-blog.sitepod.dev 可用
```

**使用随机 ID：**

```bash
? 子域名 (my-blog): -
✓ 使用随机子域名: my-blog-7x3k.sitepod.dev
```

### 绑定邮箱升级账户

```bash
$ sitepod bind
? 邮箱: alice@example.com
✓ 验证邮件已发送，请查收

# 点击邮件中的链接后
✓ 账户已升级为正式账户

  邮箱: alice@example.com
  域名: https://my-project.sitepod.dev (保持不变)

  账户和部署已永久保留
```

---

## 1. 子域名模式：个人开发者快速部署

**Alice 是一个前端开发者，想快速部署她的个人博客。**

### 首次部署（已有账户）

```bash
# 1. 登录（可选，已登录则跳过）
$ sitepod login
? 邮箱: alice@example.com
✓ 验证邮件已发送

# 点击邮件链接后自动登录
✓ 已登录为 alice

# 2. 直接部署（自动初始化）
$ cd my-blog
$ sitepod deploy
? 项目名称: my-blog
? 构建目录: ./dist
✓ 已创建 sitepod.toml

扫描 ./dist... 42 个文件
上传中... 42/42 (1.2 MB)
✓ 已部署到 beta

  https://my-blog-7x3k.beta.sitepod.dev

# 3. 确认没问题，发布到生产
$ sitepod deploy --prod
✓ 已部署到 prod

  https://my-blog-7x3k.sitepod.dev
```

### 绑定自定义域名

```bash
# 5. 绑定自定义域名
$ sitepod domain add blog.alice.dev

添加 DNS 记录以验证域名所有权：
  类型: CNAME
  名称: blog.alice.dev
  值:   my-blog-7x3k.sitepod.dev

添加后运行: sitepod domain verify blog.alice.dev

# 6. 验证域名
$ sitepod domain verify blog.alice.dev
✓ 域名已验证

现在可以通过以下地址访问：
  https://my-blog-7x3k.sitepod.dev  (系统域名)
  https://blog.alice.dev              (自定义域名)
```

### 生成的配置文件

```toml
# sitepod.toml
[project]
name = "my-blog"
# routing_mode = "subdomain"  # 默认值，可省略

[build]
directory = "./dist"
```

---

## 2. 路径模式：多项目共享域名

**Bob 在一家公司工作，公司有多个 H5 项目，都部署在 `h5.company.com` 下。**

### 初始化项目

```bash
# 1. 登录公司内部 SitePod
$ sitepod login --endpoint https://sitepod.company.com
✓ 已登录为 bob

# 2. 初始化播客管理后台项目
$ cd blog-admin
$ sitepod init
? 项目名称: blog-admin
? 构建目录: ./dist
? 路由模式: 路径模式 (多项目共享域名)
? 域名: h5.company.com
? 路径前缀: /blog-admin
✓ 已创建 sitepod.toml
```

### 生成的配置文件

```toml
# sitepod.toml
[project]
name = "blog-admin"
routing_mode = "path"

[build]
directory = "./dist"

[deploy.routing]
domain = "h5.company.com"
slug = "/blog-admin"
```

### 部署

```bash
# 3. 部署到 beta
$ sitepod deploy
扫描 ./dist... 156 个文件
上传中... 156/156 (3.8 MB)
✓ 已部署到 beta

  https://h5.company.com/blog-admin/?env=beta

# 4. 部署到生产
$ sitepod deploy --prod
✓ 已部署到 prod

  https://h5.company.com/blog-admin/
```

### 同事部署另一个项目

```bash
# Carol 部署 user-center 到同一域名
$ cd user-center
$ sitepod init
? 项目名称: user-center
? 路由模式: 路径模式
? 域名: h5.company.com        # 同一个域名
? 路径前缀: /user-center      # 不同的路径

$ sitepod deploy --prod
✓ 已部署到 prod

  https://h5.company.com/user-center/
```

### 最终效果

`h5.company.com` 下有多个项目：

| 项目 | URL |
|------|-----|
| blog-admin | `https://h5.company.com/blog-admin/` |
| user-center | `https://h5.company.com/user-center/` |
| ... | `https://h5.company.com/...` |

---

## 3. 路径模式：单域名

**Dave 为客户开发了一个官网，客户要求部署到 `www.client.com`。**

### 配置和部署

```bash
$ cd client-website
$ sitepod init
? 项目名称: client-website
? 路由模式: 路径模式 (自定义域名)
? 域名: www.client.com
? 路径前缀: /                 # 独占整个域名
✓ 已创建 sitepod.toml

# 添加并验证域名
$ sitepod domain add www.client.com

添加 DNS 记录：
  类型: CNAME
  名称: www.client.com
  值:   cname.sitepod.dev

$ sitepod domain verify www.client.com
✓ 域名已验证

# 部署
$ sitepod deploy --prod
✓ 已部署到 prod

  https://www.client.com/
```

### 生成的配置文件

```toml
# sitepod.toml
[project]
name = "client-website"
routing_mode = "path"

[build]
directory = "./dist"

[deploy.routing]
domain = "www.client.com"
slug = "/"
```

---

## 4. 预览部署

**Eve 开发了一个新功能，需要让产品经理 Frank 验收。**

### 子域名模式下的预览

```bash
$ sitepod preview
扫描 ./dist... 89 个文件
上传中... 12/12 (新增文件)
✓ 已创建预览

  https://my-app-3k7m--feat-login.preview.sitepod.dev
  有效期: 24 小时

# 发给 Frank 验收
# Frank 打开链接，确认没问题

# Eve 部署到生产
$ sitepod deploy --prod
```

### 路径模式下的预览

```bash
$ sitepod preview
✓ 已创建预览

  https://h5.company.com/blog-admin/?preview=feat-login
  有效期: 24 小时

# Frank 打开链接
# 服务端设置 Cookie，后续访问都是预览版本
# 关闭浏览器或 Cookie 过期后恢复正常
```

### 自定义预览标识

```bash
$ sitepod preview --slug pr-123
✓ 已创建预览

  https://my-app-3k7m--pr-123.preview.sitepod.dev
  有效期: 24 小时
```

---

## 5. 回滚

**线上出现问题，需要紧急回滚。**

### 交互式回滚

```bash
$ sitepod rollback
? 选择要回滚到的版本:
  > v3 (当前) - 2 分钟前 - feat: add dark mode
    v2 - 1 小时前 - fix: login bug
    v1 - 昨天 - initial release

? 确认回滚到 v2? Yes
✓ 已回滚到 v2

  https://my-app-9f2x.sitepod.dev
  回滚耗时: 0.3s
```

### 直接指定版本回滚

```bash
$ sitepod rollback --to v1
✓ 已回滚到 v1

  https://my-app-9f2x.sitepod.dev
```

### 查看部署历史

```bash
$ sitepod history
版本  状态    时间          Git Commit  说明
v3    当前    2 分钟前      a1b2c3d     feat: add dark mode
v2    -       1 小时前      e4f5g6h     fix: login bug
v1    -       昨天          i7j8k9l     initial release
```

---

## 6. CI/CD 集成

**在 GitHub Actions 中自动部署。**

### 生成 API Token

```bash
$ sitepod token create --name "github-actions"
✓ 已创建 API Token

  Token: sitepod_xxxxxxxxxxxxxxxxxxxx

  请妥善保存，此 Token 只显示一次。
  建议添加到 GitHub Secrets: SITEPOD_TOKEN
```

### GitHub Actions 配置

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Build
        run: npm ci && npm run build

      - name: Deploy to SitePod
        run: |
          curl -fsSL https://sitepod.dev/install.sh | sh
          sitepod deploy --prod
        env:
          SITEPOD_TOKEN: ${{ secrets.SITEPOD_TOKEN }}
          SITEPOD_ENDPOINT: https://sitepod.company.com
```

---

## 7. 本地开发快速体验

**新用户想快速体验 SitePod。**

```bash
# 1. 启动本地服务
$ git clone https://github.com/sitepod-dev/sitepod.git
$ cd sitepod
$ make quick-start

# 2. 另一个终端，匿名登录
$ ./bin/sitepod login --endpoint http://localhost:8080
? 选择登录方式:
  > Anonymous (快速体验，24小时有效)
    Email

✓ 已登录为 anonymous-abc123

# 3. 部署示例站点
$ cd examples/simple-site
$ ../../bin/sitepod deploy

✓ 已部署到 beta

  http://demo-site.localhost:8080
```

---

## 模式选择指南

| 场景 | 推荐模式 | 原因 |
|------|----------|------|
| 个人项目、博客 | 子域名模式 | 零配置，自动分配域名 |
| 公司多项目共享域名 | 路径模式 | 统一入口，便于管理 |
| 客户指定域名 | 路径模式 (slug='/') | 完全自定义 |
| Coolify/PaaS 部署 | 路径模式 (slug='/') | 无需通配符 DNS |

---

## 环境切换

### 子域名模式

```
生产环境:  https://my-app-9f2x.sitepod.dev
Beta环境:  https://my-app-9f2x.beta.sitepod.dev
预览环境:  https://my-app-9f2x--{slug}.preview.sitepod.dev
```

### 路径模式

```
生产环境:  https://h5.company.com/my-app/
Beta环境:  https://h5.company.com/my-app/?env=beta  (设置 Cookie 后重定向)
预览环境:  https://h5.company.com/my-app/?preview={slug}
```

路径模式下，首次访问带 `?env=beta` 或 `?preview=xxx` 参数，服务端会：
1. 设置 Cookie
2. 重定向到干净的 URL
3. 后续请求通过 Cookie 识别环境

清除 Cookie 或换浏览器即可恢复到生产环境。
