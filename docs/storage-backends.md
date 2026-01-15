# SitePod 存储后端配置

SitePod 支持多种存储后端，可以根据需求选择本地存储或 S3 兼容的对象存储。

## 目录

- [存储后端概览](#存储后端概览)
- [本地存储](#本地存储)
- [AWS S3](#aws-s3)
- [Cloudflare R2](#cloudflare-r2)
- [阿里云 OSS](#阿里云-oss)
- [MinIO](#minio)
- [存储结构](#存储结构)
- [迁移指南](#迁移指南)

---

## 存储后端概览

| 后端 | 类型 | 上传模式 | 适用场景 |
|------|------|----------|----------|
| Local | 本地文件系统 | direct | 开发、小型部署 |
| S3 | AWS S3 | presigned | AWS 生态、大规模部署 |
| R2 | Cloudflare R2 | presigned | 低成本、全球边缘 |
| OSS | 阿里云 OSS | presigned | 中国区域部署 |
| MinIO | 自托管 S3 | presigned | 私有云、离线环境 |

**上传模式说明**：
- `direct`: CLI 直接上传到服务器，服务器存储到本地
- `presigned`: CLI 获取预签名 URL，直接上传到对象存储

---

## 本地存储

最简单的配置，适合开发和小型部署。

### 配置

```yaml
# docker-compose.yml
services:
  sitepod:
    environment:
      - SITEPOD_STORAGE_TYPE=local
      - SITEPOD_DATA_DIR=/data
    volumes:
      - sitepod-data:/data

volumes:
  sitepod-data:
```

### 注意事项

- 确保数据卷有足够空间
- 定期备份数据
- 单机部署，不支持水平扩展

---

## AWS S3

使用 Amazon S3 作为存储后端，适合 AWS 生态和大规模部署。

### 创建 S3 存储桶

1. 登录 AWS Console
2. 创建 S3 存储桶
3. 建议启用版本控制和服务器端加密

### IAM 策略

创建具有最小权限的 IAM 用户：

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:GetObject",
                "s3:PutObject",
                "s3:DeleteObject",
                "s3:ListBucket"
            ],
            "Resource": [
                "arn:aws:s3:::your-bucket-name",
                "arn:aws:s3:::your-bucket-name/*"
            ]
        }
    ]
}
```

### 配置

```yaml
# docker-compose.yml
services:
  sitepod:
    environment:
      - SITEPOD_STORAGE_TYPE=s3
      - SITEPOD_S3_BUCKET=your-bucket-name
      - SITEPOD_S3_REGION=us-east-1
      - AWS_ACCESS_KEY_ID=AKIAXXXXXXXXXXXXXXXX
      - AWS_SECRET_ACCESS_KEY=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

### 环境变量

| 变量 | 说明 | 示例 |
|------|------|------|
| `SITEPOD_S3_BUCKET` | 存储桶名称 | `my-sitepod-bucket` |
| `SITEPOD_S3_REGION` | AWS 区域 | `us-east-1` |
| `AWS_ACCESS_KEY_ID` | 访问密钥 ID | `AKIA...` |
| `AWS_SECRET_ACCESS_KEY` | 访问密钥 | `...` |

---

## Cloudflare R2

Cloudflare R2 提供零出口费用的 S3 兼容存储。

### 创建 R2 存储桶

1. 登录 Cloudflare Dashboard
2. 进入 R2 Object Storage
3. 创建存储桶

### 创建 API Token

1. 在 R2 页面点击 "Manage R2 API Tokens"
2. 创建新的 API Token
3. 选择权限：Object Read & Write
4. 保存 Access Key ID 和 Secret Access Key

### 配置

```yaml
# docker-compose.yml
services:
  sitepod:
    environment:
      - SITEPOD_STORAGE_TYPE=r2
      - SITEPOD_S3_BUCKET=your-bucket-name
      - SITEPOD_S3_ENDPOINT=https://ACCOUNT_ID.r2.cloudflarestorage.com
      - AWS_ACCESS_KEY_ID=your-r2-access-key-id
      - AWS_SECRET_ACCESS_KEY=your-r2-secret-access-key
```

**注意**：
- `ACCOUNT_ID` 是你的 Cloudflare 账户 ID
- R2 不需要指定 region

### 公开访问（可选）

如果需要通过 R2 公开域直接访问静态文件：

1. 在 R2 设置中启用 "Public access"
2. 配置自定义域名

---

## 阿里云 OSS

阿里云 OSS 适合中国区域部署。

### 创建 OSS 存储桶

1. 登录阿里云控制台
2. 进入对象存储 OSS
3. 创建 Bucket（建议选择私有读写）

### 创建 AccessKey

1. 进入 RAM 访问控制
2. 创建用户并授权 OSS 权限
3. 创建 AccessKey

### 配置

```yaml
# docker-compose.yml
services:
  sitepod:
    environment:
      - SITEPOD_STORAGE_TYPE=oss
      - SITEPOD_S3_BUCKET=your-bucket-name
      - SITEPOD_S3_REGION=oss-cn-hangzhou
      - SITEPOD_S3_ENDPOINT=https://oss-cn-hangzhou.aliyuncs.com
      - AWS_ACCESS_KEY_ID=your-access-key-id
      - AWS_SECRET_ACCESS_KEY=your-access-key-secret
```

### OSS 区域端点

| 区域 | 端点 |
|------|------|
| 华东 1（杭州） | `oss-cn-hangzhou.aliyuncs.com` |
| 华东 2（上海） | `oss-cn-shanghai.aliyuncs.com` |
| 华北 2（北京） | `oss-cn-beijing.aliyuncs.com` |
| 华南 1（深圳） | `oss-cn-shenzhen.aliyuncs.com` |

完整列表参考：[OSS 访问域名](https://help.aliyun.com/document_detail/31837.html)

---

## MinIO

MinIO 是自托管的 S3 兼容存储，适合私有云和离线环境。

### 部署 MinIO

```yaml
# docker-compose.yml
services:
  minio:
    image: minio/minio
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - minio-data:/data
    environment:
      - MINIO_ROOT_USER=minioadmin
      - MINIO_ROOT_PASSWORD=minioadmin
    command: server /data --console-address ":9001"

  sitepod:
    environment:
      - SITEPOD_STORAGE_TYPE=s3
      - SITEPOD_S3_BUCKET=sitepod
      - SITEPOD_S3_REGION=us-east-1
      - SITEPOD_S3_ENDPOINT=http://minio:9000
      - AWS_ACCESS_KEY_ID=minioadmin
      - AWS_SECRET_ACCESS_KEY=minioadmin
    depends_on:
      - minio

volumes:
  minio-data:
```

### 创建存储桶

```bash
# 使用 mc 客户端
mc alias set myminio http://localhost:9000 minioadmin minioadmin
mc mb myminio/sitepod
```

---

## 存储结构

无论使用哪种存储后端，数据结构都相同：

```
/
├── blobs/                    # 内容寻址的文件存储
│   ├── ab/                   # 前两位哈希分片
│   │   ├── abcd1234...       # 实际文件内容
│   │   └── abef5678...
│   └── cd/
│       └── cdef9012...
├── refs/                     # 环境引用
│   └── {project}/
│       ├── beta.json         # Beta 环境部署信息
│       └── prod.json         # 生产环境部署信息
├── previews/                 # 预览部署
│   └── {project}/
│       └── {slug}.json
└── sitepod.db               # SQLite 数据库（仅本地存储）
```

### refs 文件格式

```json
{
  "image_id": "img_abc123",
  "created_at": "2024-01-15T10:30:00Z",
  "manifest": {
    "index.html": {
      "hash": "abcd1234...",
      "size": 1024,
      "content_type": "text/html"
    }
  }
}
```

---

## 迁移指南

### 从本地存储迁移到 S3

1. **停止服务**
   ```bash
   docker compose stop
   ```

2. **同步数据到 S3**
   ```bash
   # 使用 AWS CLI
   aws s3 sync /path/to/data/blobs s3://your-bucket/blobs
   aws s3 sync /path/to/data/refs s3://your-bucket/refs
   ```

3. **更新配置**
   ```yaml
   environment:
     - SITEPOD_STORAGE_TYPE=s3
     - SITEPOD_S3_BUCKET=your-bucket
     # ... 其他 S3 配置
   ```

4. **启动服务**
   ```bash
   docker compose up -d
   ```

### 从 S3 迁移到本地存储

```bash
# 同步数据
aws s3 sync s3://your-bucket/blobs /path/to/data/blobs
aws s3 sync s3://your-bucket/refs /path/to/data/refs

# 更新配置为 local
SITEPOD_STORAGE_TYPE=local
```

---

## 最佳实践

1. **生产环境使用对象存储**：更可靠、可扩展
2. **启用版本控制**：防止误删除
3. **配置生命周期策略**：自动清理旧版本
4. **使用 IAM 角色**：在 AWS 环境中避免使用 AccessKey
5. **定期备份数据库**：SQLite 数据库需要单独备份
