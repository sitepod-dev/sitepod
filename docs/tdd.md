# SitePod — Self-hosted static deployments

## 技术设计文档 (TDD)

**版本**: v1.1
**日期**: 2025-01-15
**项目主页**: https://sitepod.dev

> 产品需求详见 [prd.md](./prd.md)
> 运维手册详见 [ops.md](./ops.md)
> 品牌规范详见 [brand.md](./brand.md)

---

## 1. 系统架构

### 1.1 控制面 vs 数据面

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              控制面 (Control Plane)                      │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                         PocketBase                               │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐    │   │
│  │  │ Auth     │  │ API      │  │ Admin UI │  │ GC/Cleanup   │    │   │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────────┘    │   │
│  │                          │                                      │   │
│  │                          ▼                                      │   │
│  │  ┌──────────────────────────────────────────────────────────┐  │   │
│  │  │ SQLite (控制面 SSOT): 鉴权、审计日志、历史、GC roots     │  │   │
│  │  └──────────────────────────────────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                              │                                          │
│                              │ Release 时写入                           │
│                              ▼                                          │
└─────────────────────────────────────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                              数据面 (Data Plane)                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    Storage Backend (数据面 SSOT)                 │   │
│  │                                                                  │   │
│  │   refs/{project}/{env}.json  ◀── Caddy 直接读取，不经过 DB      │   │
│  │   blobs/{hash[0:2]}/{hash}   ◀── 文件内容                       │   │
│  │                                                                  │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                              ▲                                          │
│                              │                                          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                         Caddy Server                             │   │
│  │  ┌──────────┐  ┌──────────────────┐  ┌────────────────────┐    │   │
│  │  │ TLS      │  │ sitepod module   │  │ Reverse Proxy      │    │   │
│  │  │ (ACME)   │  │ (读 ref + blob)  │  │ → PocketBase API   │    │   │
│  │  └──────────┘  └──────────────────┘  └────────────────────┘    │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
```

### 1.2 SSOT 设计原则

| 层 | SSOT | 职责 | 容灾 |
|---|------|------|------|
| **数据面** | `refs/{project}/{env}.json` | 决定每个请求返回哪个版本 | DB 挂了不影响线上服务 |
| **控制面** | SQLite | 鉴权、审计、历史查询、GC roots | 可丢失可重建，异步对账 |

**关键不变量:**
- Caddy 服务请求**只读 Storage**，不读 DB
- Release API 写入顺序：先写 `refs/` → 成功后写 SQLite 审计
- 控制面与数据面解耦，DB 故障不影响线上访问

### 1.3 技术选型

| 组件 | 选择 | 原因 |
|------|------|------|
| 后端框架 | PocketBase | Go 单二进制，内置认证/REST API/Admin UI |
| Web 服务器 | Caddy (嵌入) | 自动 HTTPS，file_server 原生支持 range/etag/gzip |
| 数据库 | SQLite (WAL) | 零配置，高并发读性能 |
| CLI | Rust | 单二进制，跨平台，高性能 |
| CAS 哈希 | BLAKE3 | 快速，安全，适合内容寻址 |
| 上传校验 | SHA256 | S3 原生支持 `x-amz-checksum-sha256` |

---

## 2. 数据模型

### 2.1 Storage 层数据结构

#### Ref 文件 (数据面 SSOT)

```
refs/{project}/{env}.json

示例: refs/my-app/prod.json
```

```json
{
  "image_id": "img_abc123",
  "content_hash": "7d865e959b2466918...",
  "manifest": {
    "index.html": {
      "hash": "a1b2c3d4...",
      "size": 1234
    },
    "assets/app.js": {
      "hash": "e5f6g7h8...",
      "size": 56789
    }
  },
  "updated_at": "2025-01-15T10:30:00Z"
}
```

**设计要点:**
- Ref 文件包含完整 manifest，Caddy 一次读取即可服务
- 文件小（通常 < 100KB），易缓存
- 原子写入（先写 tmp，再 rename）

#### Blob 存储

```
blobs/{hash[0:2]}/{hash}

示例: blobs/a1/a1b2c3d4e5f6...
```

- 纯二进制内容，无元数据
- 按 BLAKE3 hash 分片存储
- 跨版本、跨项目去重

### 2.2 SQLite 数据模型 (控制面)

#### projects

```sql
CREATE TABLE projects (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  owner_id TEXT REFERENCES users(id) NOT NULL,
  routing_mode TEXT CHECK(routing_mode IN ('subdomain', 'path')) DEFAULT 'subdomain',
  -- subdomain: 子域名模式，系统分配子域名，可绑定自定义域名（slug='/'）
  -- path: 路径模式，用户配置 domain+slug，不分配系统子域名
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(name, owner_id)        -- 同一用户下项目名唯一
);

CREATE INDEX idx_projects_owner ON projects(owner_id);
```

**两种模式：**

| 模式 | 系统子域名 | 自定义域名 | slug |
|------|-----------|-----------|------|
| subdomain | ✅ 自动分配 | ✅ 可绑定多个 | 固定 '/' |
| path | ❌ 不分配 | ✅ 必须配置 | 用户指定 |

#### domains（统一管理所有域名映射）

```sql
CREATE TABLE domains (
  id TEXT PRIMARY KEY,
  domain TEXT NOT NULL,                -- 完整域名
  slug TEXT NOT NULL DEFAULT '/',      -- 路径前缀，默认 "/"
  project_id TEXT REFERENCES projects(id),
  type TEXT CHECK(type IN ('system', 'custom')) NOT NULL,
  -- system: 系统自动生成的子域名，如 "my-app-7x3k.sitepod.dev"
  -- custom: 用户自定义域名，如 "www.example.com"
  status TEXT CHECK(status IN ('pending', 'verified', 'active')) DEFAULT 'active',
  -- pending: 待验证（仅 custom 类型需要验证）
  -- verified: 已验证
  -- active: 可用（system 类型直接 active）
  verification_token TEXT,             -- 验证令牌（仅 custom 类型）
  is_primary BOOLEAN DEFAULT FALSE,    -- 是否为项目主域名
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(domain, slug)                 -- 域名+路径唯一
);

CREATE INDEX idx_domains_project ON domains(project_id);
CREATE INDEX idx_domains_lookup ON domains(domain, slug);
```

**域名类型：**

| 类型 | 示例 | 说明 |
|------|------|------|
| system | `my-app-7x3k.sitepod.dev` | 创建项目时自动生成，无需验证 |
| custom | `www.example.com` | 用户添加，需 DNS 验证 |

**创建项目时的域名处理：**

```sql
-- 子域名模式（默认）：自动生成系统域名
INSERT INTO projects (id, name, owner_id, routing_mode)
VALUES ('proj_1', 'my-app', 'user_alice', 'subdomain');

INSERT INTO domains (domain, slug, project_id, type, status, is_primary)
VALUES ('my-app-7x3k.sitepod.dev', '/', 'proj_1', 'system', 'active', TRUE);

-- 路径模式：用户必须配置域名，不生成系统域名
INSERT INTO projects (id, name, owner_id, routing_mode)
VALUES ('proj_2', 'blog', 'user_bob', 'path');

-- 用户后续添加自定义域名+slug
INSERT INTO domains (domain, slug, project_id, type, status, verification_token)
VALUES ('h5.example.com', '/blog', 'proj_2', 'custom', 'pending', 'xyz789');
```

**子域名模式下绑定自定义域名：**
```sql
-- 项目已有系统域名 my-app-7x3k.sitepod.dev
-- 用户额外绑定 www.mysite.com（slug 必须是 '/'）
INSERT INTO domains (domain, slug, project_id, type, status, verification_token)
VALUES ('www.mysite.com', '/', 'proj_1', 'custom', 'pending', 'abc123');

-- 验证通过后
UPDATE domains SET status = 'verified' WHERE domain = 'www.mysite.com';
```

**约束检查：**
```sql
-- 子域名模式的项目，只能绑定 slug='/' 的域名
-- 通过触发器或应用层检查：
-- IF project.routing_mode = 'subdomain' AND new_domain.slug != '/' THEN REJECT
```

**路由查找（统一逻辑）：**
```sql
SELECT d.*, p.* FROM domains d
JOIN projects p ON d.project_id = p.id
WHERE d.domain = ?                    -- 请求的 Host
  AND ? LIKE d.slug || '%'            -- 请求的 Path 以 slug 开头
  AND d.status IN ('verified', 'active')
ORDER BY length(d.slug) DESC          -- 最长匹配
LIMIT 1;
```

#### images

```sql
CREATE TABLE images (
  id TEXT PRIMARY KEY,
  project_id TEXT REFERENCES projects(id),
  content_hash TEXT NOT NULL,           -- BLAKE3, 用于去重
  manifest JSON NOT NULL,               -- Map<path, {hash, size}>
  file_count INTEGER,
  total_size INTEGER,
  git_commit TEXT,
  git_branch TEXT,
  git_message TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_images_project ON images(project_id, created_at DESC);
CREATE UNIQUE INDEX idx_images_hash ON images(project_id, content_hash);
```

#### deploy_events (审计日志)

```sql
CREATE TABLE deploy_events (
  id TEXT PRIMARY KEY,
  project_id TEXT REFERENCES projects(id),
  image_id TEXT REFERENCES images(id),
  environment TEXT CHECK(environment IN ('prod', 'beta')),
  action TEXT CHECK(action IN ('deploy', 'rollback')),
  previous_image_id TEXT REFERENCES images(id),
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  created_by TEXT REFERENCES users(id)
);

CREATE INDEX idx_deploy_events_lookup
ON deploy_events(project_id, environment, created_at DESC);
```

#### plans (部署计划，临时)

```sql
CREATE TABLE plans (
  id TEXT PRIMARY KEY,
  project_id TEXT REFERENCES projects(id),
  content_hash TEXT NOT NULL,
  manifest JSON NOT NULL,
  missing_blobs JSON,                   -- [{hash, size, sha256}]
  upload_mode TEXT CHECK(upload_mode IN ('presigned', 'direct')),
  status TEXT CHECK(status IN ('pending', 'committed', 'expired')),
  expires_at DATETIME,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_plans_expires ON plans(status, expires_at);
```

#### previews

```sql
CREATE TABLE previews (
  id TEXT PRIMARY KEY,
  project_id TEXT REFERENCES projects(id),
  image_id TEXT REFERENCES images(id),
  slug TEXT NOT NULL,
  expires_at DATETIME,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  created_by TEXT REFERENCES users(id),
  UNIQUE(project_id, slug)
);

CREATE INDEX idx_previews_expires ON previews(expires_at);
```

### 2.3 SQLite WAL 配置

```go
db.Exec("PRAGMA journal_mode=WAL")
db.Exec("PRAGMA busy_timeout=5000")
db.Exec("PRAGMA synchronous=NORMAL")
db.Exec("PRAGMA cache_size=-64000")  // 64MB
```

---

## 3. 哈希与校验设计

### 3.1 双哈希策略

| 哈希 | 用途 | 计算时机 |
|------|------|----------|
| **BLAKE3** | CAS key、去重、版本 content_hash | CLI 扫描时 |
| **SHA256** | S3 上传校验 (`x-amz-checksum-sha256`) | CLI 上传前 |

**为什么需要双哈希:**
- BLAKE3 快（~3x SHA256），适合 CLI 批量扫描
- S3/OSS/R2 原生只支持 SHA256 校验，无法用 BLAKE3
- 两者职责分离：BLAKE3 做寻址，SHA256 做传输校验

### 3.2 CLI 端计算

```rust
use blake3::Hasher as Blake3Hasher;
use sha2::{Sha256, Digest};

pub struct FileHashes {
    pub blake3: String,   // CAS key
    pub sha256: String,   // S3 校验用，base64 编码
    pub size: u64,
}

pub fn compute_hashes(path: &Path) -> Result<FileHashes> {
    let file = File::open(path)?;
    let mut reader = BufReader::new(file);

    let mut blake3_hasher = Blake3Hasher::new();
    let mut sha256_hasher = Sha256::new();
    let mut size = 0u64;

    let mut buffer = [0u8; 64 * 1024];
    loop {
        let n = reader.read(&mut buffer)?;
        if n == 0 { break; }
        blake3_hasher.update(&buffer[..n]);
        sha256_hasher.update(&buffer[..n]);
        size += n as u64;
    }

    Ok(FileHashes {
        blake3: blake3_hasher.finalize().to_hex().to_string(),
        sha256: base64::encode(sha256_hasher.finalize()),
        size,
    })
}
```

### 3.3 上传校验

#### Remote 存储 (S3/OSS/R2)

```go
// Server 生成 presigned URL 时，嵌入 SHA256 校验
func (s *S3Storage) GenerateUploadURL(hash, sha256Base64 string, size int64) (string, error) {
    req, _ := s.client.PutObjectRequest(&s3.PutObjectInput{
        Bucket:            &s.bucket,
        Key:               aws.String(s.blobKey(hash)),
        ContentLength:     aws.Int64(size),
        ChecksumSHA256:    aws.String(sha256Base64),
    })
    return req.Presign(15 * time.Minute)
}
```

CLI 上传时，S3 自动校验 SHA256，不匹配则拒绝。

#### Local 存储

```go
// Server 端验证（流式读取时边算边验）
func (s *LocalStorage) ReceiveBlob(expectedHash string, r io.Reader) error {
    hasher := blake3.New()
    tee := io.TeeReader(r, hasher)

    // 写临时文件
    tmpPath := filepath.Join(s.tmpDir, uuid.New().String())
    f, _ := os.Create(tmpPath)
    defer os.Remove(tmpPath)

    if _, err := io.Copy(f, tee); err != nil {
        return err
    }
    f.Close()

    // 验证 BLAKE3
    actualHash := hex.EncodeToString(hasher.Sum(nil))
    if actualHash != expectedHash {
        return ErrHashMismatch{Expected: expectedHash, Actual: actualHash}
    }

    // 原子移动
    targetPath := s.blobPath(actualHash)
    os.MkdirAll(filepath.Dir(targetPath), 0755)
    return os.Rename(tmpPath, targetPath)
}
```

---

## 4. 上传通道设计

### 4.1 两种上传模式

| 模式 | 适用后端 | 机制 |
|------|----------|------|
| **presigned** | S3/OSS/R2 | Plan 返回 presigned URL，CLI 直传对象存储 |
| **direct** | Local | Plan 返回 `/api/v1/upload/{plan_id}/{hash}`，CLI POST 到 Server |

### 4.2 Plan API 响应

```yaml
POST /api/v1/plan
Request:
  project: string
  files:
    - path: string
      blake3: string      # CAS key
      sha256: string      # Base64, S3 校验用
      size: number

Response:
  plan_id: string
  content_hash: string
  upload_mode: "presigned" | "direct"
  missing:
    - path: string
      hash: string        # BLAKE3
      size: number
      upload_url: string  # presigned URL 或 /api/v1/upload/...
  reusable: number
```

### 4.3 Local 模式上传 API

```yaml
POST /api/v1/upload/{plan_id}/{hash}
Headers:
  Content-Type: application/octet-stream
  Content-Length: {size}
Body: <binary>

Response:
  200 OK

Errors:
  400 HASH_MISMATCH: 内容校验失败
  404 PLAN_NOT_FOUND: plan_id 无效或已过期
  409 ALREADY_EXISTS: blob 已存在（幂等，返回成功）
```

### 4.4 CLI 上传逻辑

```rust
pub async fn upload_missing(plan: &PlanResponse, files: &HashMap<String, PathBuf>) -> Result<()> {
    let semaphore = Arc::new(Semaphore::new(20)); // 并发控制

    let uploads: Vec<_> = plan.missing.iter().map(|m| {
        let sem = semaphore.clone();
        let path = files.get(&m.path).unwrap().clone();
        let url = m.upload_url.clone();

        async move {
            let _permit = sem.acquire().await?;
            upload_file_with_retry(&url, &path, 3).await
        }
    }).collect();

    futures::future::try_join_all(uploads).await?;
    Ok(())
}

async fn upload_file_with_retry(url: &str, path: &Path, max_retries: u32) -> Result<()> {
    let file = tokio::fs::File::open(path).await?;
    let stream = FramedRead::new(file, BytesCodec::new());
    let body = Body::wrap_stream(stream);

    for attempt in 0..max_retries {
        let resp = reqwest::Client::new()
            .put(url)
            .header("Content-Type", "application/octet-stream")
            .body(body.try_clone().unwrap())
            .send()
            .await;

        match resp {
            Ok(r) if r.status().is_success() => return Ok(()),
            Ok(r) if r.status() == 409 => return Ok(()), // 已存在，幂等
            Err(e) if attempt < max_retries - 1 => {
                tokio::time::sleep(Duration::from_secs(1 << attempt)).await;
                continue;
            }
            Err(e) => return Err(e.into()),
            Ok(r) => return Err(anyhow!("Upload failed: {}", r.status())),
        }
    }
    unreachable!()
}
```

---

## 5. 版本发布流程

### 5.1 Release 写入顺序

```
CLI                              Server                           Storage
 │                                  │                                 │
 │  POST /api/v1/release            │                                 │
 │  {project, env, image_id}        │                                 │
 │ ─────────────────────────────────▶                                 │
 │                                  │                                 │
 │                          1. 从 images 表获取 manifest               │
 │                                  │                                 │
 │                          2. 构造 ref JSON                          │
 │                                  │                                 │
 │                          3. 原子写入 ref ─────────────────────────▶│
 │                                  │     refs/{project}/{env}.json   │
 │                                  │                                 │
 │                                  │◀─────────────────── 写入成功 ───│
 │                                  │                                 │
 │                          4. 写入 deploy_events (审计)              │
 │                                  │                                 │
 │                          5. 失效内存缓存                            │
 │                                  │                                 │
 │  {url}                           │                                 │
 ◀─────────────────────────────────│                                 │
```

**关键点:**
- Ref 写入成功才算发布成功
- deploy_events 写入失败不影响发布，后续可对账补录
- 缓存失效确保下次请求读取新 ref

### 5.2 Ref 原子写入

```go
func (s *Storage) WriteRef(project, env string, ref RefData) error {
    key := fmt.Sprintf("refs/%s/%s.json", project, env)
    data, _ := json.Marshal(ref)

    // 1. 写临时文件
    tmpKey := key + ".tmp." + uuid.New().String()
    if err := s.Put(tmpKey, bytes.NewReader(data)); err != nil {
        return err
    }

    // 2. 原子重命名 (Local) 或 Copy+Delete (S3)
    if err := s.Rename(tmpKey, key); err != nil {
        s.Delete(tmpKey)
        return err
    }

    return nil
}
```

---

## 6. Caddy 数据面实现

### 6.1 统一路由查找

所有请求都通过 domains 表查找项目，逻辑统一：

```
请求: GET https://my-app-7x3k.sitepod.dev/assets/app.js
  或: GET https://h5.example.com/blog/assets/app.js
  或: GET https://www.mysite.com/index.html

┌─────────────────────────────────────────────────────────────┐
│ 1. 从 domains 表查找                                        │
│                                                              │
│    SELECT d.*, p.* FROM domains d                            │
│    JOIN projects p ON d.project_id = p.id                    │
│    WHERE d.domain = 'my-app-7x3k.sitepod.dev'              │
│      AND '/assets/app.js' LIKE d.slug || '%'                │
│      AND d.status IN ('verified', 'active')                 │
│    ORDER BY length(d.slug) DESC                              │
│    LIMIT 1;                                                  │
│                                                              │
│ 2. 环境判断                                                  │
│    - 域名包含 .beta. → env="beta"                           │
│    - 域名包含 .preview. → 解析 preview slug                 │
│    - Cookie: __sitepod_env → 覆盖环境                       │
│    - Cookie: __sitepod_preview → 覆盖为预览                 │
│    - 默认 → env="prod"                                       │
│                                                              │
│ 3. 计算文件路径                                              │
│    请求路径 - slug 前缀 = 实际文件路径                       │
│    /blog/assets/app.js - /blog = /assets/app.js       │
└─────────────────────────────────────────────────────────────┘
```

**示例：**

| 请求 | 查找条件 | 结果 |
|------|----------|------|
| `my-app-7x3k.sitepod.dev/index.html` | `domain='my-app-7x3k.sitepod.dev'` | project=my-app, env=prod |
| `my-app-7x3k.beta.sitepod.dev/index.html` | `domain='my-app-7x3k.beta.sitepod.dev'` | project=my-app, env=beta |
| `h5.example.com/blog/app.js` | `domain='h5.example.com', slug='/blog'` | project=blog, file=/app.js |
| `www.mysite.com/about.html` | `domain='www.mysite.com', slug='/'` | project=mysite, file=/about.html |

### 6.2 请求处理流程

```
┌─────────────────────────────────────────────────────────────┐
│ 1. 查 domains 表 → 获取 project + slug                       │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. 读 Ref (带缓存)                                          │
│    cache key: "{project}:{env}" 或 "{project}:preview:{slug}"│
│    cache miss → 从 Storage 读 refs/{project}/{env}.json     │
│    cache TTL: 5s                                             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. 计算文件路径并查 manifest                                 │
│    file_path = request_path - domain.slug                   │
│    查 manifest[file_path]                                    │
│    → {hash: "e5f6g7h8...", size: 56789}                     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. 提供文件                                                  │
│    路径: blobs/{hash[0:2]}/{hash}                           │
│    自动处理: Range, ETag, Gzip/Brotli                       │
└─────────────────────────────────────────────────────────────┘
```

### 6.3 Caddy 模块实现

```go
// sitepod_handler.go
func (h *SitePodHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
    // 1. 从 domains 表查找项目
    ctx, err := h.resolveContext(r)
    if err != nil {
        return caddyhttp.Error(http.StatusNotFound, err)
    }

    // 2. 处理 Query 参数（设置 Cookie 并重定向）
    if redirect := h.handleEnvSwitch(w, r, ctx); redirect {
        return nil
    }

    // 3. 提供静态内容
    return h.serveStatic(w, r, ctx)
}

// resolveContext 通过 domains 表统一查找
func (h *SitePodHandler) resolveContext(r *http.Request) (*RequestContext, error) {
    host := r.Host
    path := r.URL.Path

    // 统一查询 domains 表
    // SELECT d.*, p.* FROM domains d
    // JOIN projects p ON d.project_id = p.id
    // WHERE d.domain = ? AND ? LIKE d.slug || '%'
    // ORDER BY length(d.slug) DESC LIMIT 1
    domain := h.findDomain(host, path)
    if domain == nil {
        return nil, errors.New("no matching domain")
    }

    // 判断环境
    env := h.determineEnv(host, r.Cookies())
    previewSlug := h.getPreviewSlug(host, r.Cookies())

    return &RequestContext{
        Project:     domain.Project,
        Domain:      domain,
        Env:         env,
        PreviewSlug: previewSlug,
        PathPrefix:  domain.Slug,
    }, nil
}

// determineEnv 判断环境
func (h *SitePodHandler) determineEnv(host string, cookies []*http.Cookie) string {
    // 1. Cookie 优先
    for _, c := range cookies {
        if c.Name == "__sitepod_env" {
            return c.Value
        }
    }

    // 2. 域名判断
    if strings.Contains(host, ".beta.") {
        return "beta"
    }

    return "prod"
}

func (h *SitePodHandler) serveStatic(w http.ResponseWriter, r *http.Request, ctx *RequestContext) error {
    // 1. 获取 ref (带缓存)
    var ref *RefData
    var err error
    if ctx.PreviewSlug != "" {
        ref, err = h.getPreviewRef(ctx.Project.Name, ctx.PreviewSlug)
    } else {
        ref, err = h.getRef(ctx.Project.Name, ctx.Env)
    }
    if err != nil {
        return caddyhttp.Error(http.StatusNotFound, err)
    }

    // 2. 计算文件路径
    path := r.URL.Path

    // 路径模式：剥离 slug 前缀
    if ctx.PathPrefix != "" && ctx.PathPrefix != "/" {
        path = strings.TrimPrefix(path, ctx.PathPrefix)
    }
    path = strings.TrimPrefix(path, "/")

    if path == "" || strings.HasSuffix(path, "/") {
        path += "index.html"
    }

    // 3. 查找文件
    file, ok := ref.Manifest[path]
    if !ok {
        // SPA fallback: 尝试 index.html
        if file, ok = ref.Manifest["index.html"]; !ok {
            return caddyhttp.Error(http.StatusNotFound, errors.New("file not found"))
        }
    }

    // 4. 重写请求路径，委托给 file_server
    blobPath := fmt.Sprintf("/blobs/%s/%s", file.Hash[:2], file.Hash)
    r.URL.Path = blobPath

    // 5. 设置 Content-Type (由扩展名决定)
    ext := filepath.Ext(path)
    if ct := mime.TypeByExtension(ext); ct != "" {
        w.Header().Set("Content-Type", ct)
    }

    // 6. 安全头
    w.Header().Set("X-Content-Type-Options", "nosniff")

    return next.ServeHTTP(w, r)
}

// handleEnvSwitch 处理环境切换 Query 参数
func (h *SitePodHandler) handleEnvSwitch(w http.ResponseWriter, r *http.Request, ctx *RequestContext) bool {
    q := r.URL.Query()

    // ?env=beta → 设置 Cookie 并重定向
    if env := q.Get("env"); env != "" {
        http.SetCookie(w, &http.Cookie{
            Name:     "__sitepod_env",
            Value:    env,
            Path:     ctx.PathPrefix,
            MaxAge:   3600, // 1 hour
            HttpOnly: true,
            SameSite: http.SameSiteLaxMode,
        })
        q.Del("env")
        r.URL.RawQuery = q.Encode()
        http.Redirect(w, r, r.URL.String(), http.StatusFound)
        return true
    }

    // ?preview=abc → 设置 Cookie 并重定向
    if preview := q.Get("preview"); preview != "" {
        http.SetCookie(w, &http.Cookie{
            Name:     "__sitepod_preview",
            Value:    preview,
            Path:     ctx.PathPrefix,
            MaxAge:   86400, // 24 hours
            HttpOnly: true,
            SameSite: http.SameSiteLaxMode,
        })
        q.Del("preview")
        r.URL.RawQuery = q.Encode()
        http.Redirect(w, r, r.URL.String(), http.StatusFound)
        return true
    }

    return false
}

func (h *SitePodHandler) getRef(project, env string) (*RefData, error) {
    cacheKey := project + ":" + env

    // 检查缓存
    if cached, ok := h.cache.Get(cacheKey); ok {
        return cached.(*RefData), nil
    }

    // 从 Storage 读取
    key := fmt.Sprintf("refs/%s/%s.json", project, env)
    data, err := h.storage.Get(key)
    if err != nil {
        return nil, err
    }

    var ref RefData
    if err := json.Unmarshal(data, &ref); err != nil {
        return nil, err
    }

    // 写入缓存 (TTL 5s)
    h.cache.Set(cacheKey, &ref, 5*time.Second)
    return &ref, nil
}
```

### 6.4 Caddyfile 配置

#### 子域名模式

```caddyfile
{
    auto_https on
    on_demand_tls {
        ask http://localhost:8090/api/v1/domains/check
    }
}

# 主域名 - API 和管理界面
sitepod.dev {
    route /api/* {
        reverse_proxy localhost:8090
    }
    # 管理界面...
}

# 项目子域名
*.sitepod.dev, *.beta.sitepod.dev, *.preview.sitepod.dev {
    route /api/* {
        reverse_proxy localhost:8090
    }

    route {
        sitepod {
            storage_path /data
            cache_ttl 5s
            mode subdomain
            domain sitepod.dev
        }
        file_server {
            root /data/blobs
            precompressed gzip br
        }
    }
}
```

#### 路径模式（自定义域名）

路径模式下，Caddy 需要监听用户配置的自定义域名：

```caddyfile
# 用户自定义域名（通过 on_demand_tls 动态处理）
:443 {
    tls {
        on_demand
    }

    route /api/* {
        reverse_proxy localhost:8090
    }

    route {
        sitepod {
            storage_path /data
            cache_ttl 5s
            mode path
        }
        file_server {
            root /data/blobs
            precompressed gzip br
        }
    }
}
```

#### 混合模式

同时支持子域名和路径模式：

```caddyfile
{
    auto_https on
    on_demand_tls {
        ask http://localhost:8090/api/v1/domains/check
    }
}

# 子域名模式
*.sitepod.dev, *.beta.sitepod.dev, *.preview.sitepod.dev {
    sitepod {
        mode subdomain
        domain sitepod.dev
    }
}

# 路径模式（自定义域名）
:443 {
    @custom_domain not host *.sitepod.dev
    handle @custom_domain {
        sitepod {
            mode path
        }
    }
}
```

---

## 7. GC (垃圾回收)

### 7.1 可回收对象

| 对象 | 位置 | 回收条件 |
|------|------|----------|
| Blob | Storage | 无任何 image manifest 引用 |
| Image | SQLite | 无任何 ref/preview 引用，且超过保留期 |
| Plan | SQLite | status=expired 或超过 TTL |
| Preview | SQLite + Storage | 已过期 |

### 7.2 Mark-Sweep with Grace Period

```go
func (gc *GC) Run(ctx context.Context) error {
    // 1. Mark: 从所有 images 收集被引用的 blob hash
    referencedBlobs := make(map[string]bool)

    images, _ := gc.db.ListAllImages()
    for _, img := range images {
        for path, file := range img.Manifest {
            referencedBlobs[file.Hash] = true
        }
    }

    // 2. Sweep: 删除未被引用且超过 grace period 的 blobs
    allBlobs, _ := gc.storage.ListBlobs()
    var deleted int

    for _, hash := range allBlobs {
        if referencedBlobs[hash] {
            continue
        }

        // Grace period: 1 小时内创建的不删除
        info, err := gc.storage.Stat(hash)
        if err != nil || time.Since(info.ModTime) < time.Hour {
            continue
        }

        if err := gc.storage.DeleteBlob(hash); err == nil {
            deleted++
        }
    }

    log.Printf("GC completed: deleted %d blobs", deleted)
    return nil
}
```

### 7.3 保留策略

```yaml
gc:
  interval: 24h
  grace_period: 1h

retention:
  min_images_per_env: 5      # 每环境至少保留 5 个版本
  keep_days: 30              # 30 天内的 image 不删除
```

---

## 8. 错误处理

### 8.1 错误码定义

```go
var (
    ErrProjectNotFound  = APIError{Code: "PROJECT_NOT_FOUND", Status: 404}
    ErrImageNotFound    = APIError{Code: "IMAGE_NOT_FOUND", Status: 404}
    ErrPlanNotFound     = APIError{Code: "PLAN_NOT_FOUND", Status: 404}
    ErrPlanExpired      = APIError{Code: "PLAN_EXPIRED", Status: 410}
    ErrPreviewExpired   = APIError{Code: "PREVIEW_EXPIRED", Status: 410}
    ErrHashMismatch     = APIError{Code: "HASH_MISMATCH", Status: 400}
    ErrBlobMissing      = APIError{Code: "BLOB_MISSING", Status: 400}
    ErrUnauthorized     = APIError{Code: "UNAUTHORIZED", Status: 401}
    ErrForbidden        = APIError{Code: "FORBIDDEN", Status: 403}
)
```

### 8.2 CLI 重试策略

```rust
const MAX_RETRIES: u32 = 3;
const RETRY_DELAYS: [u64; 3] = [1, 2, 4]; // 指数退避

async fn with_retry<F, T, E>(op: F) -> Result<T, E>
where
    F: Fn() -> impl Future<Output = Result<T, E>>,
    E: std::error::Error + IsRetryable,
{
    for (attempt, delay) in RETRY_DELAYS.iter().enumerate() {
        match op().await {
            Ok(v) => return Ok(v),
            Err(e) if e.is_retryable() && attempt < MAX_RETRIES as usize - 1 => {
                tokio::time::sleep(Duration::from_secs(*delay)).await;
            }
            Err(e) => return Err(e),
        }
    }
    unreachable!()
}
```

---

## 9. 监控指标

| 指标 | 类型 | 说明 |
|------|------|------|
| `sitepod_deploys_total` | Counter | 部署次数 (labels: project, env, status) |
| `sitepod_deploy_duration_seconds` | Histogram | 部署耗时 |
| `sitepod_uploads_total` | Counter | 上传次数 |
| `sitepod_uploads_bytes_total` | Counter | 上传字节数 |
| `sitepod_blob_reuse_ratio` | Gauge | Blob 复用率 |
| `sitepod_ref_cache_hits_total` | Counter | Ref 缓存命中 |
| `sitepod_ref_cache_misses_total` | Counter | Ref 缓存未命中 |
| `sitepod_http_requests_total` | Counter | HTTP 请求数 |
| `sitepod_http_request_duration_seconds` | Histogram | 请求耗时 |
| `sitepod_storage_bytes` | Gauge | 存储使用量 |

---

## 附录 A: Storage Backend 接口

```go
type StorageBackend interface {
    // Blob 操作
    PutBlob(hash string, r io.Reader) error
    GetBlob(hash string) (io.ReadCloser, error)
    HasBlob(hash string) (bool, error)
    DeleteBlob(hash string) error
    ListBlobs() ([]string, error)
    StatBlob(hash string) (BlobInfo, error)

    // Ref 操作
    PutRef(project, env string, data []byte) error
    GetRef(project, env string) ([]byte, error)
    DeleteRef(project, env string) error

    // Preview 操作
    PutPreview(project, slug string, data []byte) error
    GetPreview(project, slug string) ([]byte, error)
    DeletePreview(project, slug string) error

    // 上传 URL (仅 Remote 后端实现)
    GenerateUploadURL(hash, sha256 string, size int64) (string, error)

    // 通用
    Put(key string, r io.Reader) error
    Get(key string) ([]byte, error)
    Rename(oldKey, newKey string) error
}
```

## 附录 B: CLI 命令 UX

```
$ sitepod deploy

Scanning ./dist... 156 files
Computing hashes... done

Planning deployment...
  Project: my-app
  Environment: beta
  Files: 156 total, 12 new, 144 reused (92%)

Uploading 12 files...
  [████████████████████████████████] 12/12 (1.2 MB)

Committing...
  Image: img_7x8y9z
  Content hash: 7d865e959b...

Released to beta
  URL: https://my-app.beta.example.com

$ sitepod deploy --prod

⚠ Deploying to production
  Continue? [y/N] y

Released to prod
  URL: https://my-app.example.com
```
