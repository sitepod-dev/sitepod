package caddy

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/sitepod/sitepod/internal/storage"
	"go.uber.org/zap"
)

// API: Cleanup - removes expired anonymous accounts and previews
func (h *SitePodHandler) apiCleanup(w http.ResponseWriter, r *http.Request) error {
	if err := h.requireAdminToken(r); err != nil {
		if errors.Is(err, errAdminTokenMissing) {
			return h.jsonError(w, http.StatusForbidden, "admin token not configured")
		}
		return h.jsonError(w, http.StatusForbidden, "forbidden")
	}

	h.logger.Info("Starting cleanup task")

	result := map[string]any{
		"expired_users_deleted":    0,
		"expired_previews_deleted": 0,
		"errors":                   []string{},
	}
	errs := []string{}

	// 1. Find and delete expired anonymous users
	expiredUsers, err := h.app.FindRecordsByFilter(
		"users",
		"is_anonymous = true AND anonymous_expires_at < {:now}",
		"", 1000, 0,
		map[string]any{"now": time.Now()},
	)
	if err != nil {
		h.logger.Debug("No expired users found or error", zap.Error(err))
		expiredUsers = []*core.Record{}
	}

	for _, user := range expiredUsers {
		h.logger.Info("Deleting expired anonymous user",
			zap.String("user_id", user.Id),
			zap.String("email", user.GetString("email")))

		if err := h.deleteUserCascade(user); err != nil {
			errMsg := fmt.Sprintf("failed to delete user %s: %v", user.Id, err)
			errs = append(errs, errMsg)
			h.logger.Warn(errMsg)
		} else {
			result["expired_users_deleted"] = result["expired_users_deleted"].(int) + 1
		}
	}

	// 2. Find and delete expired previews
	previewsDeleted := 0
	projects, _ := h.app.FindRecordsByFilter("projects", "1=1", "", 1000, 0, nil)
	for _, project := range projects {
		projectName := project.GetString("name")
		// Check previews in storage
		previewsDeleted += h.cleanupExpiredPreviews(projectName)
	}
	result["expired_previews_deleted"] = previewsDeleted

	result["errors"] = errs

	h.logger.Info("Cleanup completed",
		zap.Int("expired_users_deleted", result["expired_users_deleted"].(int)),
		zap.Int("expired_previews_deleted", previewsDeleted))

	return h.jsonResponse(w, http.StatusOK, result)
}

// deleteUserCascade deletes a user and all their data
func (h *SitePodHandler) deleteUserCascade(user *core.Record) error {
	var firstErr error
	recordErr := func(err error) {
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	// Find all projects owned by this user
	projects, err := h.app.FindRecordsByFilter(
		"projects", "owner_id = {:owner_id}", "", 1000, 0,
		map[string]any{"owner_id": user.Id},
	)
	if err != nil {
		projects = []*core.Record{}
	}

	// For each project, delete related data
	for _, project := range projects {
		projectID := project.Id
		projectName := project.GetString("name")

		// Delete domains
		domains, _ := h.app.FindRecordsByFilter(
			"domains", "project_id = {:project_id}", "", 1000, 0,
			map[string]any{"project_id": projectID},
		)
		for _, domain := range domains {
			recordErr(h.app.Delete(domain))
		}

		// Delete images
		images, _ := h.app.FindRecordsByFilter(
			"images", "project_id = {:project_id}", "", 1000, 0,
			map[string]any{"project_id": projectID},
		)
		for _, image := range images {
			recordErr(h.app.Delete(image))
		}

		// Delete deploy_events
		events, _ := h.app.FindRecordsByFilter(
			"deploy_events", "project_id = {:project_id}", "", 1000, 0,
			map[string]any{"project_id": projectID},
		)
		for _, event := range events {
			recordErr(h.app.Delete(event))
		}

		// Delete plans
		plans, _ := h.app.FindRecordsByFilter(
			"plans", "project_id = {:project_id}", "", 1000, 0,
			map[string]any{"project_id": projectID},
		)
		for _, plan := range plans {
			recordErr(h.app.Delete(plan))
		}

		// Delete ref files from storage
		recordErr(h.storage.DeleteRef(projectName, "beta"))
		recordErr(h.storage.DeleteRef(projectName, "prod"))

		// Delete the project record
		recordErr(h.app.Delete(project))
	}

	// Delete the user record
	recordErr(h.app.Delete(user))

	return firstErr
}

// cleanupExpiredPreviews removes expired preview deployments for a project
func (h *SitePodHandler) cleanupExpiredPreviews(projectName string) int {
	deleted := 0
	collection, err := h.app.FindCollectionByNameOrId("previews")
	if err != nil || collection == nil {
		return 0
	}

	now := time.Now().UTC().Format("2006-01-02 15:04:05.000Z")
	previews, err := h.app.FindRecordsByFilter(
		"previews",
		"project = {:project} && expires_at < {:now}",
		"",
		1000,
		0,
		map[string]any{"project": projectName, "now": now},
	)
	if err != nil {
		return 0
	}

	for _, preview := range previews {
		slug := preview.GetString("slug")
		if err := h.storage.DeletePreview(projectName, slug); err == nil {
			deleted++
		}
		_ = h.app.Delete(preview)
	}

	return deleted
}

// API: Garbage Collection - removes unreferenced blobs
func (h *SitePodHandler) apiGarbageCollect(w http.ResponseWriter, r *http.Request) error {
	if err := h.requireAdminToken(r); err != nil {
		if errors.Is(err, errAdminTokenMissing) {
			return h.jsonError(w, http.StatusForbidden, "admin token not configured")
		}
		return h.jsonError(w, http.StatusForbidden, "forbidden")
	}

	h.logger.Info("Starting garbage collection")

	// 1. Collect all referenced blob hashes from refs and previews
	referencedHashes := make(map[string]bool)

	// Scan all projects' refs
	projects, _ := h.app.FindRecordsByFilter("projects", "1=1", "", 10000, 0, nil)
	for _, project := range projects {
		projectName := project.GetString("name")

		// Check beta ref
		if refData, err := h.storage.GetRef(projectName, "beta"); err == nil {
			var ref storage.RefData
			if json.Unmarshal(refData, &ref) == nil {
				for _, entry := range ref.Manifest {
					referencedHashes[entry.Hash] = true
				}
			}
		}

		// Check prod ref
		if refData, err := h.storage.GetRef(projectName, "prod"); err == nil {
			var ref storage.RefData
			if json.Unmarshal(refData, &ref) == nil {
				for _, entry := range ref.Manifest {
					referencedHashes[entry.Hash] = true
				}
			}
		}
	}

	// Also check images in database for recently deployed content
	images, _ := h.app.FindRecordsByFilter("images", "1=1", "", 10000, 0, nil)
	for _, image := range images {
		manifestStr := image.GetString("manifest")
		var manifest map[string]storage.FileEntry
		if json.Unmarshal([]byte(manifestStr), &manifest) == nil {
			for _, entry := range manifest {
				referencedHashes[entry.Hash] = true
			}
		}
	}

	h.logger.Info("Found referenced blobs", zap.Int("count", len(referencedHashes)))

	// 2. List all blobs in storage
	allBlobs, err := h.storage.ListBlobs()
	if err != nil {
		return h.jsonError(w, http.StatusInternalServerError, "failed to list blobs: "+err.Error())
	}

	h.logger.Info("Total blobs in storage", zap.Int("count", len(allBlobs)))

	// 3. Find and delete unreferenced blobs
	var deletedCount int
	var deletedSize int64
	var errs []string

	for _, hash := range allBlobs {
		if !referencedHashes[hash] {
			// Get blob size before deletion
			if info, err := h.storage.StatBlob(hash); err == nil {
				deletedSize += info.Size
			}

			if err := h.storage.DeleteBlob(hash); err != nil {
				errs = append(errs, fmt.Sprintf("failed to delete blob %s: %v", hash, err))
			} else {
				deletedCount++
			}
		}
	}

	h.logger.Info("Garbage collection completed",
		zap.Int("deleted_blobs", deletedCount),
		zap.Int64("freed_bytes", deletedSize))

	return h.jsonResponse(w, http.StatusOK, map[string]any{
		"referenced_blobs": len(referencedHashes),
		"total_blobs":      len(allBlobs),
		"deleted_blobs":    deletedCount,
		"freed_bytes":      deletedSize,
		"errors":           errs,
	})
}

// API: Invalidate Cache
func (h *SitePodHandler) apiInvalidateCache(w http.ResponseWriter, r *http.Request) error {
	project := r.URL.Query().Get("project")
	env := r.URL.Query().Get("env")

	if project != "" && env != "" {
		h.cache.Delete(project + ":" + env)
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

// API: Rebuild Routing
func (h *SitePodHandler) apiRebuildRouting(w http.ResponseWriter, r *http.Request) error {
	if err := h.rebuildRoutingIndex(); err != nil {
		return h.jsonError(w, http.StatusInternalServerError, "failed to rebuild routing index: "+err.Error())
	}
	return h.jsonResponse(w, http.StatusOK, map[string]string{"message": "Routing index rebuilt successfully"})
}
