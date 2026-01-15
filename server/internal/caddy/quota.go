package caddy

import (
	"fmt"
	"os"
	"strconv"

	"github.com/pocketbase/pocketbase/models"
)

// QuotaConfig holds quota limits for deployments
// Can be overridden via environment variables
type QuotaConfig struct {
	MaxFilesPerDeploy  int   // SITEPOD_MAX_FILES_PER_DEPLOY (default: 10000)
	MaxFileSizeBytes   int64 // SITEPOD_MAX_FILE_SIZE (default: 100MB)
	MaxDeploySizeBytes int64 // SITEPOD_MAX_DEPLOY_SIZE (default: 500MB)
	MaxProjectsPerUser int   // SITEPOD_MAX_PROJECTS_PER_USER (default: 100)
	AnonMaxProjects    int   // SITEPOD_ANON_MAX_PROJECTS (default: 5)
	AnonMaxDeploySize  int64 // SITEPOD_ANON_MAX_DEPLOY_SIZE (default: 50MB)
}

// Global quota config (loaded once at startup)
var quotaConfig = loadQuotaConfig()

func loadQuotaConfig() QuotaConfig {
	return QuotaConfig{
		MaxFilesPerDeploy:  getEnvInt("SITEPOD_MAX_FILES_PER_DEPLOY", 10000),
		MaxFileSizeBytes:   getEnvInt64("SITEPOD_MAX_FILE_SIZE", 100*1024*1024),       // 100MB
		MaxDeploySizeBytes: getEnvInt64("SITEPOD_MAX_DEPLOY_SIZE", 500*1024*1024),     // 500MB
		MaxProjectsPerUser: getEnvInt("SITEPOD_MAX_PROJECTS_PER_USER", 100),
		AnonMaxProjects:    getEnvInt("SITEPOD_ANON_MAX_PROJECTS", 5),
		AnonMaxDeploySize:  getEnvInt64("SITEPOD_ANON_MAX_DEPLOY_SIZE", 50*1024*1024), // 50MB
	}
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvInt64(key string, defaultVal int64) int64 {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
	}
	return defaultVal
}

// formatBytes formats bytes to human-readable string
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

// FileEntry for quota checking
type FileEntry struct {
	Path        string `json:"path"`
	Blake3      string `json:"blake3"`
	SHA256      string `json:"sha256"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}

// checkDeployQuotas validates deployment against size and file count limits
func (h *SitePodHandler) checkDeployQuotas(files []FileEntry, isAnonymous bool) error {
	// Check file count
	if len(files) > quotaConfig.MaxFilesPerDeploy {
		return fmt.Errorf("too many files: %d (max: %d)", len(files), quotaConfig.MaxFilesPerDeploy)
	}

	// Calculate total size and check individual file sizes
	var totalSize int64
	for _, f := range files {
		// Check individual file size
		if f.Size > quotaConfig.MaxFileSizeBytes {
			return fmt.Errorf("file too large: %s (%s, max: %s)",
				f.Path, formatBytes(f.Size), formatBytes(quotaConfig.MaxFileSizeBytes))
		}
		totalSize += f.Size
	}

	// Check total deploy size
	maxDeploySize := quotaConfig.MaxDeploySizeBytes
	if isAnonymous {
		maxDeploySize = quotaConfig.AnonMaxDeploySize
	}
	if totalSize > maxDeploySize {
		return fmt.Errorf("deployment too large: %s (max: %s)",
			formatBytes(totalSize), formatBytes(maxDeploySize))
	}

	return nil
}

// checkProjectCountQuota validates user hasn't exceeded project limit
func (h *SitePodHandler) checkProjectCountQuota(userID string, isAnonymous bool) error {
	projects, err := h.app.Dao().FindRecordsByFilter(
		"projects", "owner_id = {:owner_id}", "", 1000, 0,
		map[string]any{"owner_id": userID},
	)
	if err != nil {
		projects = []*models.Record{}
	}

	maxProjects := quotaConfig.MaxProjectsPerUser
	if isAnonymous {
		maxProjects = quotaConfig.AnonMaxProjects
	}

	if len(projects) >= maxProjects {
		if isAnonymous {
			return fmt.Errorf("anonymous users can have at most %d projects. Verify your email to increase limits", maxProjects)
		}
		return fmt.Errorf("project limit reached: %d (max: %d)", len(projects), maxProjects)
	}

	return nil
}
