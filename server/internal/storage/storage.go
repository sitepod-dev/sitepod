package storage

import (
	"io"
	"time"
)

// BlobInfo contains metadata about a blob
type BlobInfo struct {
	Hash    string
	Size    int64
	ModTime time.Time
}

// Backend defines the interface for storage backends
type Backend interface {
	// Blob operations
	PutBlob(hash string, r io.Reader, size int64) error
	GetBlob(hash string) (io.ReadCloser, error)
	HasBlob(hash string) (bool, error)
	DeleteBlob(hash string) error
	ListBlobs() ([]string, error)
	StatBlob(hash string) (*BlobInfo, error)

	// Ref operations (environment pointers)
	PutRef(project, env string, data []byte) error
	GetRef(project, env string) ([]byte, error)
	DeleteRef(project, env string) error

	// Preview operations
	PutPreview(project, slug string, data []byte) error
	GetPreview(project, slug string) ([]byte, error)
	DeletePreview(project, slug string) error

	// Routing index operations (for path mode)
	PutRouting(data []byte) error
	GetRouting() ([]byte, error)

	// Upload URL generation (for remote backends)
	GenerateUploadURL(hash, sha256Base64 string, size int64) (string, error)

	// Upload mode
	UploadMode() string

	// Get the base path for blobs (used by Caddy)
	BlobBasePath() string
}

// FileEntry represents a file in a manifest
type FileEntry struct {
	Hash        string `json:"hash"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type,omitempty"`
}

// RefData represents the content of a ref file
type RefData struct {
	ImageID     string               `json:"image_id"`
	ContentHash string               `json:"content_hash"`
	Manifest    map[string]FileEntry `json:"manifest"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

// PreviewRef represents a preview deployment
type PreviewRef struct {
	ImageID   string               `json:"image_id"`
	Manifest  map[string]FileEntry `json:"manifest"`
	ExpiresAt time.Time            `json:"expires_at"`
	CreatedAt time.Time            `json:"created_at"`
}
