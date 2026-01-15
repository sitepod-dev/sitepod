package storage

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/zeebo/blake3"
)

// LocalBackend implements storage on local filesystem
type LocalBackend struct {
	basePath string
	blobPath string
	refPath  string
	prevPath string
	tmpPath  string
}

// NewLocalBackend creates a new local storage backend
func NewLocalBackend(basePath string) (*LocalBackend, error) {
	b := &LocalBackend{
		basePath: basePath,
		blobPath: filepath.Join(basePath, "blobs"),
		refPath:  filepath.Join(basePath, "refs"),
		prevPath: filepath.Join(basePath, "previews"),
		tmpPath:  filepath.Join(basePath, "tmp"),
	}

	// Create directories
	dirs := []string{b.blobPath, b.refPath, b.prevPath, b.tmpPath}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return b, nil
}

// BlobBasePath returns the base path for blobs
func (b *LocalBackend) BlobBasePath() string {
	return b.blobPath
}

// UploadMode returns the upload mode for this backend
func (b *LocalBackend) UploadMode() string {
	return "direct"
}

// blobFilePath returns the file path for a blob
func (b *LocalBackend) blobFilePath(hash string) string {
	if len(hash) < 2 {
		return filepath.Join(b.blobPath, hash)
	}
	return filepath.Join(b.blobPath, hash[:2], hash)
}

// PutBlob stores a blob, verifying its hash
func (b *LocalBackend) PutBlob(hash string, r io.Reader, size int64) error {
	targetPath := b.blobFilePath(hash)

	// Check if already exists
	if _, err := os.Stat(targetPath); err == nil {
		return nil // Already exists, idempotent
	}

	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("failed to create blob directory: %w", err)
	}

	// Write to temp file while computing hash
	tmpFile := filepath.Join(b.tmpPath, uuid.New().String())
	f, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile)

	hasher := blake3.New()
	tee := io.TeeReader(r, hasher)

	written, err := io.Copy(f, tee)
	if err != nil {
		f.Close()
		return fmt.Errorf("failed to write blob: %w", err)
	}
	f.Close()

	// Verify hash
	actualHash := hex.EncodeToString(hasher.Sum(nil))
	if actualHash != hash {
		return &HashMismatchError{Expected: hash, Actual: actualHash}
	}

	// Verify size if provided
	if size > 0 && written != size {
		return fmt.Errorf("size mismatch: expected %d, got %d", size, written)
	}

	// Atomic move
	if err := os.Rename(tmpFile, targetPath); err != nil {
		return fmt.Errorf("failed to move blob to final location: %w", err)
	}

	return nil
}

// GetBlob retrieves a blob
func (b *LocalBackend) GetBlob(hash string) (io.ReadCloser, error) {
	path := b.blobFilePath(hash)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &BlobNotFoundError{Hash: hash}
		}
		return nil, err
	}
	return f, nil
}

// HasBlob checks if a blob exists
func (b *LocalBackend) HasBlob(hash string) (bool, error) {
	path := b.blobFilePath(hash)
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// DeleteBlob removes a blob
func (b *LocalBackend) DeleteBlob(hash string) error {
	path := b.blobFilePath(hash)
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// ListBlobs returns all blob hashes
func (b *LocalBackend) ListBlobs() ([]string, error) {
	var hashes []string

	err := filepath.Walk(b.blobPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		// Extract hash from path
		hash := filepath.Base(path)
		hashes = append(hashes, hash)
		return nil
	})

	return hashes, err
}

// StatBlob returns metadata about a blob
func (b *LocalBackend) StatBlob(hash string) (*BlobInfo, error) {
	path := b.blobFilePath(hash)
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &BlobNotFoundError{Hash: hash}
		}
		return nil, err
	}

	return &BlobInfo{
		Hash:    hash,
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}, nil
}

// PutRef writes a ref file atomically
func (b *LocalBackend) PutRef(project, env string, data []byte) error {
	dir := filepath.Join(b.refPath, project)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	targetPath := filepath.Join(dir, env+".json")
	tmpPath := targetPath + ".tmp." + uuid.New().String()

	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, targetPath); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
}

// GetRef reads a ref file
func (b *LocalBackend) GetRef(project, env string) ([]byte, error) {
	path := filepath.Join(b.refPath, project, env+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &RefNotFoundError{Project: project, Env: env}
		}
		return nil, err
	}
	return data, nil
}

// DeleteRef removes a ref file
func (b *LocalBackend) DeleteRef(project, env string) error {
	path := filepath.Join(b.refPath, project, env+".json")
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// PutPreview writes a preview file
func (b *LocalBackend) PutPreview(project, slug string, data []byte) error {
	dir := filepath.Join(b.prevPath, project)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := filepath.Join(dir, slug+".json")
	return os.WriteFile(path, data, 0644)
}

// GetPreview reads a preview file
func (b *LocalBackend) GetPreview(project, slug string) ([]byte, error) {
	path := filepath.Join(b.prevPath, project, slug+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &PreviewNotFoundError{Project: project, Slug: slug}
		}
		return nil, err
	}
	return data, nil
}

// DeletePreview removes a preview file
func (b *LocalBackend) DeletePreview(project, slug string) error {
	path := filepath.Join(b.prevPath, project, slug+".json")
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// PutRouting writes the routing index file
func (b *LocalBackend) PutRouting(data []byte) error {
	path := filepath.Join(b.basePath, "routing", "index.json")
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmpPath := path + ".tmp." + uuid.New().String()
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
}

// GetRouting reads the routing index file
func (b *LocalBackend) GetRouting() ([]byte, error) {
	path := filepath.Join(b.basePath, "routing", "index.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("routing index not found")
		}
		return nil, err
	}
	return data, nil
}

// GenerateUploadURL is not supported for local storage
func (b *LocalBackend) GenerateUploadURL(hash, sha256Base64 string, size int64) (string, error) {
	return "", errors.New("presigned URLs not supported for local storage")
}

// ListExpiredPreviews returns previews that should be cleaned up
func (b *LocalBackend) ListExpiredPreviews() ([]string, error) {
	var expired []string

	err := filepath.Walk(b.prevPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}
		expired = append(expired, path)
		return nil
	})

	return expired, err
}

// Error types

type HashMismatchError struct {
	Expected string
	Actual   string
}

func (e *HashMismatchError) Error() string {
	return fmt.Sprintf("hash mismatch: expected %s, got %s", e.Expected, e.Actual)
}

type BlobNotFoundError struct {
	Hash string
}

func (e *BlobNotFoundError) Error() string {
	return fmt.Sprintf("blob not found: %s", e.Hash)
}

type RefNotFoundError struct {
	Project string
	Env     string
}

func (e *RefNotFoundError) Error() string {
	return fmt.Sprintf("ref not found: %s/%s", e.Project, e.Env)
}

type PreviewNotFoundError struct {
	Project string
	Slug    string
}

func (e *PreviewNotFoundError) Error() string {
	return fmt.Sprintf("preview not found: %s/%s", e.Project, e.Slug)
}
