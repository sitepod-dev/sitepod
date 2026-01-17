package caddy

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pocketbase/pocketbase/models"
	"github.com/sitepod/sitepod/internal/storage"
	"github.com/zeebo/blake3"
	"go.uber.org/zap"
)

// ensureSystemSites deploys system sites (welcome, console) on first startup
func (h *SitePodHandler) ensureSystemSites() {
	// Get system user
	systemUser, err := h.getSystemUser()
	if err != nil {
		h.logger.Warn("system user not found, skipping system sites setup", zap.Error(err))
		return
	}

	h.ensureWelcomeSite(systemUser.Id)
	h.ensureConsoleSite(systemUser.Id)
}

// ensureWelcomeSite creates a default welcome site on first startup
func (h *SitePodHandler) ensureWelcomeSite(ownerID string) {
	// Check if welcome project already exists
	if _, err := h.app.Dao().FindFirstRecordByData("projects", "name", "welcome"); err == nil {
		return // Already exists
	}

	h.logger.Info("Creating default welcome site...")

	// Create welcome project with system user as owner
	project, err := h.getOrCreateProjectWithOwner("welcome", ownerID)
	if err != nil {
		h.logger.Warn("failed to create welcome project", zap.Error(err))
		return
	}

	// Create welcome page content
	welcomeHTML := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to SitePod</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
        }
        .container {
            text-align: center;
            max-width: 600px;
        }
        h1 { font-size: 3rem; margin-bottom: 1rem; }
        .subtitle { font-size: 1.25rem; opacity: 0.9; margin-bottom: 2rem; }
        .card {
            background: rgba(255,255,255,0.1);
            backdrop-filter: blur(10px);
            border-radius: 16px;
            padding: 2rem;
            margin: 1rem 0;
            text-align: left;
        }
        .card h2 { font-size: 1.25rem; margin-bottom: 1rem; }
        code {
            background: rgba(0,0,0,0.3);
            padding: 0.5rem 1rem;
            border-radius: 8px;
            display: block;
            margin: 0.5rem 0;
            font-family: 'SF Mono', Monaco, monospace;
            font-size: 0.9rem;
            overflow-x: auto;
        }
        .features {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 1rem;
            margin-top: 2rem;
        }
        .feature {
            background: rgba(255,255,255,0.1);
            padding: 1rem;
            border-radius: 12px;
            text-align: center;
        }
        .feature-icon { font-size: 2rem; margin-bottom: 0.5rem; }
        a { color: #fff; }
    </style>
</head>
<body>
    <div class="container">
        <h1>SitePod</h1>
        <p class="subtitle">Your self-hosted static site deployment platform is ready!</p>

        <div class="card">
            <h2>Quick Start</h2>
            <code>sitepod login --endpoint {ENDPOINT}</code>
            <code>cd your-site && sitepod deploy</code>
        </div>

        <div class="features">
            <div class="feature">
                <div class="feature-icon">*</div>
                <div>Instant Deploy</div>
            </div>
            <div class="feature">
                <div class="feature-icon">~</div>
                <div>Fast Rollback</div>
            </div>
            <div class="feature">
                <div class="feature-icon">@</div>
                <div>Custom Domains</div>
            </div>
            <div class="feature">
                <div class="feature-icon">#</div>
                <div>Auto HTTPS</div>
            </div>
        </div>

        <p style="margin-top: 2rem; opacity: 0.8;">
            <a href="https://github.com/sitepod/sitepod">Documentation</a>
        </p>
    </div>
</body>
</html>`

	// Replace placeholder with actual endpoint
	scheme := "https"
	if h.Domain == "localhost" || strings.HasPrefix(h.Domain, "localhost:") {
		scheme = "http"
	}
	welcomeHTML = strings.Replace(welcomeHTML, "{ENDPOINT}", fmt.Sprintf("%s://%s", scheme, h.Domain), 1)

	// Calculate hash and store blob
	hasher := blake3.New()
	if _, err := hasher.Write([]byte(welcomeHTML)); err != nil {
		h.logger.Warn("failed to hash welcome page", zap.Error(err))
		return
	}
	hash := hex.EncodeToString(hasher.Sum(nil))

	htmlBytes := []byte(welcomeHTML)
	if err := h.storage.PutBlob(hash, strings.NewReader(welcomeHTML), int64(len(htmlBytes))); err != nil {
		h.logger.Warn("failed to write welcome page blob", zap.Error(err))
		return
	}

	// Create manifest
	manifest := map[string]storage.FileEntry{
		"index.html": {
			Hash: hash,
			Size: int64(len(welcomeHTML)),
		},
	}

	// Create image ID
	imageID := fmt.Sprintf("img_%s", uuid.New().String()[:8])

	// Calculate content hash
	manifestBytes, _ := json.Marshal(manifest)
	contentHasher := blake3.New()
	if _, err := contentHasher.Write(manifestBytes); err != nil {
		h.logger.Warn("failed to hash welcome manifest", zap.Error(err))
		return
	}
	contentHash := hex.EncodeToString(contentHasher.Sum(nil))

	// Save image record
	imagesCollection, err := h.app.Dao().FindCollectionByNameOrId("images")
	if err != nil {
		h.logger.Warn("failed to find images collection", zap.Error(err))
		return
	}

	image := models.NewRecord(imagesCollection)
	image.Set("image_id", imageID)
	image.Set("project_id", project.Id)
	image.Set("content_hash", contentHash)
	image.Set("manifest", manifest)
	image.Set("file_count", 1)
	image.Set("total_size", len(welcomeHTML))

	if err := h.app.Dao().SaveRecord(image); err != nil {
		h.logger.Warn("failed to save welcome image", zap.Error(err))
		return
	}

	// Write ref for prod environment
	refData := &storage.RefData{
		ImageID:     imageID,
		ContentHash: contentHash,
		Manifest:    manifest,
		UpdatedAt:   time.Now().UTC(),
	}

	refBytes, _ := json.Marshal(refData)
	if err := h.storage.PutRef("welcome", "prod", refBytes); err != nil {
		h.logger.Warn("failed to write welcome ref", zap.Error(err))
		return
	}

	h.logger.Info("Welcome site deployed", zap.String("url", fmt.Sprintf("%s://welcome.%s", scheme, h.Domain)))
}

// ensureConsoleSite deploys the pre-built console UI
func (h *SitePodHandler) ensureConsoleSite(ownerID string) {
	// Check if console project already exists
	if _, err := h.app.Dao().FindFirstRecordByData("projects", "name", "console"); err == nil {
		return // Already exists
	}

	// Look for console dist in common locations
	consolePaths := []string{
		"../console/dist",                   // Development (run from server/)
		"./console/dist",                    // Development (run from root)
		"/app/console/dist",                 // Docker
		filepath.Join(h.DataDir, "console"), // Bundled with data
	}

	var distPath string
	for _, p := range consolePaths {
		if _, err := os.Stat(filepath.Join(p, "index.html")); err == nil {
			distPath = p
			break
		}
	}

	if distPath == "" {
		h.logger.Debug("Console dist not found, skipping console site deployment")
		return
	}

	h.logger.Info("Deploying console site...", zap.String("source", distPath))

	// Create console project with system user as owner
	project, err := h.getOrCreateProjectWithOwner("console", ownerID)
	if err != nil {
		h.logger.Warn("failed to create console project", zap.Error(err))
		return
	}

	// Walk the dist directory and collect files
	manifest := make(map[string]storage.FileEntry)
	var totalSize int64

	err = filepath.Walk(distPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		relPath, err := filepath.Rel(distPath, path)
		if err != nil {
			return err
		}
		// Normalize path separators for consistency
		relPath = strings.ReplaceAll(relPath, "\\", "/")

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Calculate hash
		hasher := blake3.New()
		if _, err := hasher.Write(content); err != nil {
			return err
		}
		hash := hex.EncodeToString(hasher.Sum(nil))

		// Store blob
		if err := h.storage.PutBlob(hash, strings.NewReader(string(content)), int64(len(content))); err != nil {
			h.logger.Warn("failed to store console blob", zap.String("file", relPath), zap.Error(err))
			return nil // Continue with other files
		}

		// Detect content type
		contentType := mime.TypeByExtension(filepath.Ext(path))

		manifest[relPath] = storage.FileEntry{
			Hash:        hash,
			Size:        int64(len(content)),
			ContentType: contentType,
		}
		totalSize += int64(len(content))

		return nil
	})

	if err != nil {
		h.logger.Warn("failed to walk console dist", zap.Error(err))
		return
	}

	if len(manifest) == 0 {
		h.logger.Warn("no files found in console dist")
		return
	}

	// Create image record
	imageID := fmt.Sprintf("img_%s", uuid.New().String()[:8])

	// Calculate content hash from manifest
	manifestBytes, _ := json.Marshal(manifest)
	contentHasher := blake3.New()
	if _, err := contentHasher.Write(manifestBytes); err != nil {
		h.logger.Warn("failed to hash console manifest", zap.Error(err))
		return
	}
	contentHash := hex.EncodeToString(contentHasher.Sum(nil))

	imagesCollection, err := h.app.Dao().FindCollectionByNameOrId("images")
	if err != nil {
		h.logger.Warn("failed to find images collection", zap.Error(err))
		return
	}

	image := models.NewRecord(imagesCollection)
	image.Set("image_id", imageID)
	image.Set("project_id", project.Id)
	image.Set("content_hash", contentHash)
	image.Set("manifest", manifest)
	image.Set("file_count", len(manifest))
	image.Set("total_size", totalSize)

	if err := h.app.Dao().SaveRecord(image); err != nil {
		h.logger.Warn("failed to save console image", zap.Error(err))
		return
	}

	// Write ref for prod environment
	refData := &storage.RefData{
		ImageID:     imageID,
		ContentHash: contentHash,
		Manifest:    manifest,
		UpdatedAt:   time.Now().UTC(),
	}

	refBytes, _ := json.Marshal(refData)
	if err := h.storage.PutRef("console", "prod", refBytes); err != nil {
		h.logger.Warn("failed to write console ref", zap.Error(err))
		return
	}

	scheme := "https"
	if h.Domain == "localhost" || strings.HasPrefix(h.Domain, "localhost:") {
		scheme = "http"
	}

	h.logger.Info("Console site deployed",
		zap.String("url", fmt.Sprintf("%s://%s", scheme, h.Domain)),
		zap.Int("files", len(manifest)),
		zap.Int64("size", totalSize),
	)
}
