# SitePod Roadmap

## Current Status

SitePod 核心功能已完成，包括：
- 部署流程（plan/commit/release）
- 匿名登录和账户绑定
- Console 管理界面
- CLI 工具（deploy、rollback、history、preview 等）
- E2E 测试
- 账户删除功能

---

## High Priority

### 1. 过期数据清理任务 ✅ 已完成

**目标**：自动清理过期的匿名账户和 Preview 部署

**任务**：
- [x] 实现匿名账户过期检查（24h 未验证邮箱）
- [x] 级联删除过期账户的所有数据（复用 deleteUserCascade 逻辑）
- [x] 实现 Preview 部署过期检查（基础框架）
- [x] 添加 API 触发机制：`POST /api/v1/cleanup`
- [x] 添加清理日志记录

**API 用法**：
```bash
curl -X POST http://localhost:8080/api/v1/cleanup
# 返回：{"expired_users_deleted": 0, "expired_previews_deleted": 0, "errors": []}
```

### 2. Blob GC（垃圾回收）✅ 已完成

**目标**：删除不再被任何 ref 引用的 blobs，防止存储无限增长

**任务**：
- [x] 实现标记清除算法
- [x] 扫描所有 refs 和 images，收集被引用的 blob hashes
- [x] 对比 blobs 目录，找出未被引用的 blobs
- [x] 安全删除未引用的 blobs
- [x] 添加 GC API 端点：`POST /api/v1/gc`
- [x] 添加 GC 统计日志（删除了多少 blobs，释放了多少空间）

**API 用法**：
```bash
curl -X POST http://localhost:8080/api/v1/gc
# 返回：{"referenced_blobs": 10, "total_blobs": 15, "deleted_blobs": 5, "freed_bytes": 1024, "errors": []}
```

**注意**：cleanup 和 gc API 不需要认证，应在生产环境中通过防火墙保护或添加 secret token

---

## Medium Priority

### 3. 生产部署文档 ✅ 已完成

**目标**：提供完整的生产环境部署指南

**任务**：
- [x] Docker Compose 生产配置（带持久化存储）
- [x] S3 存储后端配置示例
- [x] Cloudflare R2 配置示例
- [x] 阿里云 OSS 配置示例
- [x] 反向代理配置（Nginx/Caddy/Cloudflare）
- [x] HTTPS/TLS 证书配置
- [x] 环境变量说明文档
- [x] 备份和恢复指南

**相关文件**：
- `docs/deployment.md` - 部署指南
- `docs/storage-backends.md` - 存储后端配置
- `docker-compose.prod.yml` - 生产 Docker Compose
- `.env.example` - 环境变量示例

### 4. CLI 版本检查 ✅ 已完成

**目标**：帮助用户及时更新到最新版本

**任务**：
- [x] 启动时检查 GitHub releases
- [x] 比较当前版本和最新版本（semver 比较）
- [x] 如有新版本，显示升级提示
- [x] 添加 `--skip-update-check` 参数跳过检查
- [x] 缓存检查结果（24h 检查一次）

**实现细节**：
- 检查在后台异步执行，不阻塞主命令
- 缓存保存在 `~/.cache/sitepod/version-check.json`
- 可通过 `SITEPOD_SKIP_UPDATE_CHECK=1` 环境变量跳过

**相关文件**：
- `cli/src/update.rs` - 版本检查逻辑
- `cli/src/main.rs` - 集成版本检查

### 5. 配额和限制 ✅ 已完成

**目标**：防止资源滥用，保护服务稳定性

**任务**：
- [x] 单项目最大文件数量限制（10,000 文件）
- [x] 单文件最大大小限制（100MB）
- [x] 单次部署最大总大小限制（500MB）
- [x] 单用户最大项目数量限制（100 个）
- [x] 匿名用户更严格的限制（5 项目，50MB/部署）
- [x] 超限时返回友好的错误信息
- [x] 配置化限制参数（通过环境变量）

**环境变量配置**：
```bash
SITEPOD_MAX_FILES_PER_DEPLOY=10000    # 单次部署最大文件数
SITEPOD_MAX_FILE_SIZE=104857600       # 单文件最大大小 (100MB)
SITEPOD_MAX_DEPLOY_SIZE=524288000     # 单次部署最大总大小 (500MB)
SITEPOD_MAX_PROJECTS_PER_USER=100     # 用户最大项目数
SITEPOD_ANON_MAX_PROJECTS=5           # 匿名用户最大项目数
SITEPOD_ANON_MAX_DEPLOY_SIZE=52428800 # 匿名用户部署大小限制 (50MB)
```

**相关文件**：
- `server/internal/caddy/handler.go` - 配额配置和检查逻辑

---

## Low Priority

### 6. 监控和指标增强

**目标**：更全面的运行状态监控

**任务**：
- [ ] 部署成功/失败计数
- [ ] 部署延迟分布
- [ ] 存储使用量（blobs 总大小、数量）
- [ ] 活跃用户数（日/周/月）
- [ ] 项目数量统计
- [ ] API 请求量和延迟
- [ ] 错误率监控

**相关文件**：
- `server/internal/caddy/metrics.go` - 新建或扩展

### 7. Webhook 通知

**目标**：部署事件通知到外部系统

**任务**：
- [ ] 项目级 Webhook 配置
- [ ] 支持的事件：deploy.success, deploy.failed, rollback
- [ ] Webhook payload 格式设计
- [ ] 重试机制
- [ ] Webhook 调用日志
- [ ] 预设模板：Slack、Discord、飞书

**相关文件**：
- `server/internal/webhook/` - 新建包
- 数据库 migrations - 添加 webhooks 表

### 8. Console UI 增强

**目标**：更完善的管理界面

**任务**：
- [ ] 项目设置页面（修改 subdomain、删除项目）
- [ ] 部署详情页面（查看文件列表、manifest）
- [ ] 用户设置页面（修改邮箱、删除账户）
- [ ] 深色模式支持
- [ ] 移动端适配优化

**相关文件**：
- `server/console/` - Svelte 组件

---

## Future Ideas

这些是更长远的想法，暂不列入计划：

- **团队协作**：多用户共享项目
- **CI/CD 集成**：GitHub Actions、GitLab CI 模板
- **边缘部署**：CDN 集成、多区域部署
- **A/B 测试**：流量分割、金丝雀发布
- **分析统计**：访问量、带宽统计
- **自定义构建**：集成构建流程（如 npm build）

---

## Changelog

| 日期 | 更新内容 |
|------|----------|
| 2026-01-16 | 完成生产部署文档 |
| 2026-01-16 | 完成配额和限制功能 |
| 2026-01-16 | 完成 CLI 版本检查功能 |
| 2026-01-16 | 完成高优先级任务：过期数据清理、Blob GC |
| 2026-01-16 | 初始版本，整理待办事项 |
