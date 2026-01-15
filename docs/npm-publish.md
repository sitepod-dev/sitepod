# npm 发布指南

npm 发布结构已搭建完成。

## 创建的文件结构

```
npm-packages/
├── sitepod/                        # 主包
│   ├── package.json
│   ├── bin/sitepod.js             # 入口脚本
│   └── README.md
├── @sitepod/
│   ├── darwin-arm64/              # macOS Apple Silicon
│   ├── darwin-x64/                # macOS Intel
│   ├── linux-x64/                 # Linux x64
│   ├── linux-arm64/               # Linux ARM64
│   └── win32-x64/                 # Windows x64
scripts/
└── sync-versions.sh               # 版本同步脚本
```

## 发布流程

### 1. 自动发布

推送 `v*` tag 时，GitHub Actions 会自动：
- 构建所有平台的二进制
- 更新 npm 包版本
- 发布到 npm

### 2. 本地测试

```bash
# 构建并准备本地 npm 包
make npm-prepare

# 链接到全局进行测试
make npm-link
sitepod --help
```

### 3. 本地发布

```bash
# 先登录 npm
npm login

# 发布所有包（需要先准备好各平台的二进制）
make npm-publish
```

注意：本地发布只会包含当前平台的二进制文件。如需发布所有平台，建议使用 GitHub Actions 自动发布。

### 4. 版本管理

```bash
make bump-patch   # 0.1.0 → 0.1.1
make bump-minor   # 0.1.0 → 0.2.0
make bump-major   # 0.1.0 → 1.0.0
```

同步的文件：
- `cli/Cargo.toml` — Rust CLI
- `npm-packages/sitepod/package.json` — npm 主包
- `npm-packages/@sitepod/*/package.json` — 平台包
- `www/src/consts.ts` — 官网版本号

## 发布前准备

1. 在 npm 创建 organization `@sitepod`
2. 在 GitHub 仓库设置中添加 `NPM_TOKEN` secret
3. 更新 `package.json` 中的 repository URL

## 首次发布

```bash
# 1. 确认版本号（当前在 cli/Cargo.toml）
# 2. 提交并打 tag
git add .
git commit -m "Add npm publish support"
git tag v0.1.0
git push origin main --tags
```

GitHub Actions 会自动完成剩余的发布工作。
