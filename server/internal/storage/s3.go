package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Backend implements storage using S3-compatible object storage
type S3Backend struct {
	client   *s3.Client
	bucket   string
	presign  *s3.PresignClient
	endpoint string
}

// NewS3Backend creates a new S3 storage backend
func NewS3Backend(bucket, region, endpoint string) (*S3Backend, error) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	var client *s3.Client
	if endpoint != "" {
		// Custom endpoint (for OSS/R2/MinIO)
		client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		})
	} else {
		client = s3.NewFromConfig(cfg)
	}

	return &S3Backend{
		client:   client,
		bucket:   bucket,
		presign:  s3.NewPresignClient(client),
		endpoint: endpoint,
	}, nil
}

func (b *S3Backend) blobKey(hash string) string {
	if len(hash) < 2 {
		return path.Join("blobs", hash)
	}
	return path.Join("blobs", hash[:2], hash)
}

func (b *S3Backend) refKey(project, env string) string {
	return path.Join("refs", project, env+".json")
}

func (b *S3Backend) previewKey(project, slug string) string {
	return path.Join("previews", project, slug+".json")
}

// BlobBasePath returns empty for S3 (not used directly)
func (b *S3Backend) BlobBasePath() string {
	return ""
}

// UploadMode returns the upload mode for this backend
func (b *S3Backend) UploadMode() string {
	return "presigned"
}

// PutBlob uploads a blob to S3
func (b *S3Backend) PutBlob(hash string, r io.Reader, size int64) error {
	ctx := context.Background()

	input := &s3.PutObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(b.blobKey(hash)),
		Body:   r,
	}
	if size >= 0 {
		input.ContentLength = aws.Int64(size)
	}

	_, err := b.client.PutObject(ctx, input)

	return err
}

// GetBlob retrieves a blob from S3
func (b *S3Backend) GetBlob(hash string) (io.ReadCloser, error) {
	ctx := context.Background()

	result, err := b.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(b.blobKey(hash)),
	})
	if err != nil {
		return nil, err
	}

	return result.Body, nil
}

// HasBlob checks if a blob exists
func (b *S3Backend) HasBlob(hash string) (bool, error) {
	ctx := context.Background()

	_, err := b.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(b.blobKey(hash)),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// DeleteBlob removes a blob from S3
func (b *S3Backend) DeleteBlob(hash string) error {
	ctx := context.Background()

	_, err := b.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(b.blobKey(hash)),
	})

	return err
}

// ListBlobs returns all blob hashes
func (b *S3Backend) ListBlobs() ([]string, error) {
	ctx := context.Background()
	var hashes []string

	paginator := s3.NewListObjectsV2Paginator(b.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(b.bucket),
		Prefix: aws.String("blobs/"),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, obj := range page.Contents {
			// Extract hash from key (blobs/xx/hash)
			parts := strings.Split(*obj.Key, "/")
			if len(parts) >= 3 {
				hashes = append(hashes, parts[len(parts)-1])
			}
		}
	}

	return hashes, nil
}

// StatBlob returns metadata about a blob
func (b *S3Backend) StatBlob(hash string) (*BlobInfo, error) {
	ctx := context.Background()

	result, err := b.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(b.blobKey(hash)),
	})
	if err != nil {
		return nil, err
	}

	return &BlobInfo{
		Hash:    hash,
		Size:    *result.ContentLength,
		ModTime: *result.LastModified,
	}, nil
}

// PutRef writes a ref file
func (b *S3Backend) PutRef(project, env string, data []byte) error {
	ctx := context.Background()

	_, err := b.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(b.bucket),
		Key:         aws.String(b.refKey(project, env)),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
	})

	return err
}

// GetRef reads a ref file
func (b *S3Backend) GetRef(project, env string) ([]byte, error) {
	ctx := context.Background()

	result, err := b.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(b.refKey(project, env)),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

// DeleteRef removes a ref file
func (b *S3Backend) DeleteRef(project, env string) error {
	ctx := context.Background()

	_, err := b.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(b.refKey(project, env)),
	})

	return err
}

// PutPreview writes a preview file
func (b *S3Backend) PutPreview(project, slug string, data []byte) error {
	ctx := context.Background()

	_, err := b.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(b.bucket),
		Key:         aws.String(b.previewKey(project, slug)),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
	})

	return err
}

// GetPreview reads a preview file
func (b *S3Backend) GetPreview(project, slug string) ([]byte, error) {
	ctx := context.Background()

	result, err := b.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(b.previewKey(project, slug)),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

// DeletePreview removes a preview file
func (b *S3Backend) DeletePreview(project, slug string) error {
	ctx := context.Background()

	_, err := b.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(b.previewKey(project, slug)),
	})

	return err
}

// PutRouting writes the routing index file
func (b *S3Backend) PutRouting(data []byte) error {
	ctx := context.Background()

	_, err := b.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(b.bucket),
		Key:         aws.String("routing/index.json"),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
	})

	return err
}

// GetRouting reads the routing index file
func (b *S3Backend) GetRouting() ([]byte, error) {
	ctx := context.Background()

	result, err := b.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String("routing/index.json"),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

// GenerateUploadURL generates a presigned URL for uploading
func (b *S3Backend) GenerateUploadURL(hash, sha256Base64 string, size int64) (string, error) {
	ctx := context.Background()

	input := &s3.PutObjectInput{
		Bucket:        aws.String(b.bucket),
		Key:           aws.String(b.blobKey(hash)),
		ContentLength: aws.Int64(size),
	}

	if sha256Base64 != "" {
		input.ChecksumSHA256 = aws.String(sha256Base64)
	}

	presignResult, err := b.presign.PresignPutObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = 15 * time.Minute
	})
	if err != nil {
		return "", err
	}

	return presignResult.URL, nil
}
