package storage

import (
	"bytes"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/zeebo/blake3"
)

func TestLocalBackend(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "sitepod-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create backend
	backend, err := NewLocalBackend(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("PutBlob", func(t *testing.T) {
		content := []byte("hello world")
		hash := computeHash(content)

		err := backend.PutBlob(hash, bytes.NewReader(content), int64(len(content)))
		if err != nil {
			t.Fatal(err)
		}

		// Verify file exists
		blobPath := filepath.Join(tmpDir, "blobs", hash[:2], hash)
		if _, err := os.Stat(blobPath); os.IsNotExist(err) {
			t.Error("blob file not created")
		}
	})

	t.Run("GetBlob", func(t *testing.T) {
		content := []byte("test content")
		hash := computeHash(content)

		// Put first
		err := backend.PutBlob(hash, bytes.NewReader(content), int64(len(content)))
		if err != nil {
			t.Fatal(err)
		}

		// Get
		reader, err := backend.GetBlob(hash)
		if err != nil {
			t.Fatal(err)
		}
		defer reader.Close()

		got, err := io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(got, content) {
			t.Errorf("content mismatch: got %q, want %q", got, content)
		}
	})

	t.Run("HasBlob", func(t *testing.T) {
		content := []byte("exists check")
		hash := computeHash(content)

		// Should not exist initially
		exists, err := backend.HasBlob(hash)
		if err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Error("blob should not exist")
		}

		// Put
		err = backend.PutBlob(hash, bytes.NewReader(content), int64(len(content)))
		if err != nil {
			t.Fatal(err)
		}

		// Should exist now
		exists, err = backend.HasBlob(hash)
		if err != nil {
			t.Fatal(err)
		}
		if !exists {
			t.Error("blob should exist")
		}
	})

	t.Run("HashMismatch", func(t *testing.T) {
		content := []byte("original content")
		wrongHash := "0000000000000000000000000000000000000000000000000000000000000000"

		err := backend.PutBlob(wrongHash, bytes.NewReader(content), int64(len(content)))
		if err == nil {
			t.Error("expected hash mismatch error")
		}
		if _, ok := err.(*HashMismatchError); !ok {
			t.Errorf("expected HashMismatchError, got %T", err)
		}
	})

	t.Run("PutRef", func(t *testing.T) {
		data := []byte(`{"image_id":"img_123","content_hash":"abc"}`)

		err := backend.PutRef("myproject", "prod", data)
		if err != nil {
			t.Fatal(err)
		}

		// Verify file exists
		refPath := filepath.Join(tmpDir, "refs", "myproject", "prod.json")
		if _, err := os.Stat(refPath); os.IsNotExist(err) {
			t.Error("ref file not created")
		}
	})

	t.Run("GetRef", func(t *testing.T) {
		data := []byte(`{"image_id":"img_456","content_hash":"def"}`)

		err := backend.PutRef("gettest", "beta", data)
		if err != nil {
			t.Fatal(err)
		}

		got, err := backend.GetRef("gettest", "beta")
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(got, data) {
			t.Errorf("ref data mismatch: got %q, want %q", got, data)
		}
	})

	t.Run("GetRefNotFound", func(t *testing.T) {
		_, err := backend.GetRef("nonexistent", "prod")
		if err == nil {
			t.Error("expected error for nonexistent ref")
		}
		if _, ok := err.(*RefNotFoundError); !ok {
			t.Errorf("expected RefNotFoundError, got %T", err)
		}
	})

	t.Run("PutPreview", func(t *testing.T) {
		data := []byte(`{"image_id":"img_789","expires_at":"2025-01-01T00:00:00Z"}`)

		err := backend.PutPreview("myproject", "abc123", data)
		if err != nil {
			t.Fatal(err)
		}

		got, err := backend.GetPreview("myproject", "abc123")
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(got, data) {
			t.Errorf("preview data mismatch")
		}
	})

	t.Run("Idempotent", func(t *testing.T) {
		content := []byte("idempotent test")
		hash := computeHash(content)

		// Put twice - should not error
		err := backend.PutBlob(hash, bytes.NewReader(content), int64(len(content)))
		if err != nil {
			t.Fatal(err)
		}

		err = backend.PutBlob(hash, bytes.NewReader(content), int64(len(content)))
		if err != nil {
			t.Fatal(err)
		}
	})
}

func computeHash(content []byte) string {
	hasher := blake3.New()
	if _, err := hasher.Write(content); err != nil {
		panic(err)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}
