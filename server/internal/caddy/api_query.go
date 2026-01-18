package caddy

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/sitepod/sitepod/internal/storage"
	"go.uber.org/zap"
)

// API: Health
func (h *SitePodHandler) apiHealth(w http.ResponseWriter, r *http.Request) error {
	dbStatus := "ok"
	if _, err := h.app.FindCollectionByNameOrId("projects"); err != nil {
		dbStatus = "error"
	}

	storageStatus := "ok"
	if _, err := h.storage.ListBlobs(); err != nil {
		storageStatus = "error"
	}

	status := "healthy"
	if dbStatus != "ok" || storageStatus != "ok" {
		status = "degraded"
	}

	return h.jsonResponse(w, http.StatusOK, map[string]any{
		"status":   status,
		"database": dbStatus,
		"storage":  storageStatus,
		"uptime":   time.Since(h.startTime).String(),
	})
}

// API: Config (public endpoint for frontend configuration)
func (h *SitePodHandler) apiConfig(w http.ResponseWriter, r *http.Request) error {
	isDemo := os.Getenv("IS_DEMO") == "1" || os.Getenv("IS_DEMO") == "true"

	return h.jsonResponse(w, http.StatusOK, map[string]any{
		"domain":  h.Domain,
		"is_demo": isDemo,
	})
}

// API: Metrics (Prometheus format)
func (h *SitePodHandler) apiMetrics(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	// Basic metrics - can be extended
	fmt.Fprintf(w, "# HELP sitepod_up SitePod is running\n")
	fmt.Fprintf(w, "# TYPE sitepod_up gauge\n")
	fmt.Fprintf(w, "sitepod_up 1\n")
	fmt.Fprintf(w, "# HELP sitepod_uptime_seconds Uptime in seconds\n")
	fmt.Fprintf(w, "# TYPE sitepod_uptime_seconds counter\n")
	fmt.Fprintf(w, "sitepod_uptime_seconds %.0f\n", time.Since(h.startTime).Seconds())
	return nil
}

// API: Get Current
func (h *SitePodHandler) apiGetCurrent(w http.ResponseWriter, r *http.Request, user *core.Record) error {
	projectName := r.URL.Query().Get("project")
	env := r.URL.Query().Get("environment")

	if projectName == "" || env == "" {
		return h.jsonError(w, http.StatusBadRequest, "project and environment required")
	}

	project, err := h.requireProjectOwnerByName(projectName, user)
	if err != nil {
		if errors.Is(err, errForbidden) {
			return h.jsonError(w, http.StatusForbidden, "forbidden")
		}
		return h.jsonError(w, http.StatusNotFound, "project not found")
	}

	refData, err := h.storage.GetRef(projectName, env)
	if err != nil {
		return h.jsonError(w, http.StatusNotFound, "ref not found")
	}

	var ref storage.RefData
	if err := json.Unmarshal(refData, &ref); err != nil {
		return h.jsonError(w, http.StatusInternalServerError, "invalid ref data")
	}

	var fileCount *int
	if ref.ImageID != "" {
		if image, err := h.app.FindFirstRecordByData("images", "image_id", ref.ImageID); err == nil {
			if image.GetString("project_id") == project.Id {
				count := image.GetInt("file_count")
				fileCount = &count
			}
		}
	}

	resp := map[string]any{
		"image_id":     ref.ImageID,
		"content_hash": ref.ContentHash,
		"deployed_at":  ref.UpdatedAt,
	}
	if fileCount != nil {
		resp["file_count"] = *fileCount
	}

	return h.jsonResponse(w, http.StatusOK, resp)
}

// API: Get History
func (h *SitePodHandler) apiGetHistory(w http.ResponseWriter, r *http.Request, user *core.Record) error {
	projectName := r.URL.Query().Get("project")

	project, err := h.requireProjectOwnerByName(projectName, user)
	if err != nil {
		if errors.Is(err, errForbidden) {
			return h.jsonError(w, http.StatusForbidden, "forbidden")
		}
		return h.jsonError(w, http.StatusNotFound, "project not found")
	}

	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			if parsed > 200 {
				parsed = 200
			}
			limit = parsed
		}
	}

	images, err := h.app.FindRecordsByFilter(
		"images", "project_id = {:project_id}", "-created", limit, 0,
		map[string]any{"project_id": project.Id},
	)
	if err != nil {
		return h.jsonError(w, http.StatusInternalServerError, err.Error())
	}

	items := make([]map[string]any, len(images))
	for i, img := range images {
		items[i] = map[string]any{
			"image_id":     img.GetString("image_id"),
			"content_hash": img.GetString("content_hash"),
			"created_at":   img.GetDateTime("created").Time(),
			"git_commit":   img.GetString("git_commit"),
		}
	}

	return h.jsonResponse(w, http.StatusOK, map[string]any{"items": items})
}

// API: List Images
func (h *SitePodHandler) apiListImages(w http.ResponseWriter, r *http.Request, user *core.Record) error {
	projectID := r.URL.Query().Get("project_id")
	projectName := r.URL.Query().Get("project")

	var project *core.Record
	var err error

	if projectID != "" {
		project, err = h.requireProjectOwnerByID(projectID, user)
	} else if projectName != "" {
		project, err = h.requireProjectOwnerByName(projectName, user)
	} else {
		return h.jsonError(w, http.StatusBadRequest, "project_id or project required")
	}

	if err != nil {
		if errors.Is(err, errForbidden) {
			return h.jsonError(w, http.StatusForbidden, "forbidden")
		}
		return h.jsonError(w, http.StatusNotFound, "project not found")
	}

	// Get current deployed images for each environment
	prodRef, _ := h.storage.GetRef(project.GetString("name"), "prod")
	betaRef, _ := h.storage.GetRef(project.GetString("name"), "beta")

	var prodImageID, betaImageID string
	if prodRef != nil {
		var ref storage.RefData
		if json.Unmarshal(prodRef, &ref) == nil {
			prodImageID = ref.ImageID
		}
	}
	if betaRef != nil {
		var ref storage.RefData
		if json.Unmarshal(betaRef, &ref) == nil {
			betaImageID = ref.ImageID
		}
	}

	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			if parsed > 200 {
				parsed = 200
			}
			limit = parsed
		}
	}

	page := 1
	if v := r.URL.Query().Get("page"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			page = parsed
		}
	}

	offset := (page - 1) * limit

	total := 0
	allImages, err := h.app.FindRecordsByFilter(
		"images", "project_id = {:project_id}", "", 0, 0,
		map[string]any{"project_id": project.Id},
	)
	if err == nil {
		total = len(allImages)
	}

	images, err := h.app.FindRecordsByFilter(
		"images", "project_id = {:project_id}", "-created", limit, offset,
		map[string]any{"project_id": project.Id},
	)
	if err != nil {
		return h.jsonResponse(w, http.StatusOK, map[string]any{"images": []any{}, "total": total})
	}

	result := make([]map[string]any, len(images))
	for i, img := range images {
		imageID := img.GetString("image_id")
		deployedTo := []string{}
		if imageID == prodImageID {
			deployedTo = append(deployedTo, "prod")
		}
		if imageID == betaImageID {
			deployedTo = append(deployedTo, "beta")
		}

		result[i] = map[string]any{
			"id":           img.Id,
			"image_id":     imageID,
			"project_id":   img.GetString("project_id"),
			"content_hash": img.GetString("content_hash"),
			"file_count":   img.GetInt("file_count"),
			"total_size":   img.GetInt("total_size"),
			"git_commit":   img.GetString("git_commit"),
			"git_branch":   img.GetString("git_branch"),
			"git_message":  img.GetString("git_message"),
			"created_at":   img.GetDateTime("created").String(),
			"deployed_to":  deployedTo,
		}
	}

	return h.jsonResponse(w, http.StatusOK, map[string]any{"images": result, "total": total})
}

// API: List Projects (supports both admin and user tokens)
func (h *SitePodHandler) apiListProjectsAny(w http.ResponseWriter, r *http.Request) error {
	user, err := h.authenticate(r)
	if err != nil {
		return h.jsonError(w, http.StatusUnauthorized, "authentication required")
	}

	isAdmin := user.GetBool("is_admin")

	var projects []*core.Record

	if isAdmin {
		// Admin can see all projects
		projects, err = h.app.FindRecordsByFilter(
			"projects", "1=1", "-created", 100, 0, nil,
		)
		if err != nil {
			return h.jsonResponse(w, http.StatusOK, []any{})
		}
	} else {
		// Regular user can only see their own projects
		projects, err = h.app.FindRecordsByFilter(
			"projects", "owner_id = {:owner_id}", "-created", 100, 0,
			map[string]any{"owner_id": user.Id},
		)
		if err != nil {
			return h.jsonResponse(w, http.StatusOK, []any{})
		}
	}

	result := make([]map[string]any, len(projects))
	for i, p := range projects {
		item := map[string]any{
			"id":         p.Id,
			"name":       p.GetString("name"),
			"subdomain":  p.GetString("subdomain"),
			"owner_id":   p.GetString("owner_id"),
			"created_at": p.GetDateTime("created").String(),
			"updated_at": p.GetDateTime("updated").String(),
		}

		// Admin can see owner_email
		if isAdmin {
			ownerID := p.GetString("owner_id")
			if ownerID != "" {
				owner, err := h.app.FindRecordById("users", ownerID)
				if err == nil {
					item["owner_email"] = owner.GetString("email")
				}
			}
		}

		result[i] = item
	}

	return h.jsonResponse(w, http.StatusOK, result)
}

// API: Get Project
func (h *SitePodHandler) apiGetProject(w http.ResponseWriter, r *http.Request, projectName string, user *core.Record) error {
	project, err := h.requireProjectOwnerByName(projectName, user)
	if err != nil {
		if errors.Is(err, errForbidden) {
			return h.jsonError(w, http.StatusForbidden, "forbidden")
		}
		return h.jsonError(w, http.StatusNotFound, "project not found")
	}

	return h.jsonResponse(w, http.StatusOK, map[string]any{
		"id":         project.Id,
		"name":       project.GetString("name"),
		"subdomain":  project.GetString("subdomain"),
		"owner_id":   project.GetString("owner_id"),
		"created_at": project.GetDateTime("created").String(),
		"updated_at": project.GetDateTime("updated").String(),
	})
}

// API: Delete Project
func (h *SitePodHandler) apiDeleteProject(w http.ResponseWriter, r *http.Request, projectName string, user *core.Record) error {
	if projectName == "" {
		return h.jsonError(w, http.StatusBadRequest, "project required")
	}

	project, err := h.requireProjectOwnerByName(projectName, user)
	if err != nil {
		if errors.Is(err, errForbidden) {
			return h.jsonError(w, http.StatusForbidden, "forbidden")
		}
		return h.jsonError(w, http.StatusNotFound, "project not found")
	}

	projectID := project.Id

	// Delete domains
	domains, _ := h.app.FindRecordsByFilter(
		"domains", "project_id = {:project_id}", "", 1000, 0,
		map[string]any{"project_id": projectID},
	)
	for _, domain := range domains {
		if err := h.app.Delete(domain); err != nil {
			h.logger.Warn("Failed to delete domain", zap.String("domain_id", domain.Id), zap.Error(err))
		}
	}

	// Delete images
	images, _ := h.app.FindRecordsByFilter(
		"images", "project_id = {:project_id}", "", 1000, 0,
		map[string]any{"project_id": projectID},
	)
	for _, image := range images {
		if err := h.app.Delete(image); err != nil {
			h.logger.Warn("Failed to delete image", zap.String("image_id", image.Id), zap.Error(err))
		}
	}

	// Delete deploy events
	events, _ := h.app.FindRecordsByFilter(
		"deploy_events", "project_id = {:project_id}", "", 1000, 0,
		map[string]any{"project_id": projectID},
	)
	for _, event := range events {
		if err := h.app.Delete(event); err != nil {
			h.logger.Warn("Failed to delete deploy event", zap.String("event_id", event.Id), zap.Error(err))
		}
	}

	// Delete plans
	plans, _ := h.app.FindRecordsByFilter(
		"plans", "project_id = {:project_id}", "", 1000, 0,
		map[string]any{"project_id": projectID},
	)
	for _, plan := range plans {
		if err := h.app.Delete(plan); err != nil {
			h.logger.Warn("Failed to delete plan", zap.String("plan_id", plan.Id), zap.Error(err))
		}
	}

	// Delete previews
	previews, _ := h.app.FindRecordsByFilter(
		"previews", "project = {:project}", "", 1000, 0,
		map[string]any{"project": projectName},
	)
	for _, preview := range previews {
		slug := preview.GetString("slug")
		if slug != "" {
			if err := h.storage.DeletePreview(projectName, slug); err != nil {
				h.logger.Warn("Failed to delete preview", zap.String("slug", slug), zap.Error(err))
			}
		}
		if err := h.app.Delete(preview); err != nil {
			h.logger.Warn("Failed to delete preview record", zap.String("preview_id", preview.Id), zap.Error(err))
		}
	}

	// Delete ref files from storage
	if err := h.storage.DeleteRef(projectName, "beta"); err != nil {
		h.logger.Warn("Failed to delete beta ref", zap.String("project", projectName), zap.Error(err))
	}
	if err := h.storage.DeleteRef(projectName, "prod"); err != nil {
		h.logger.Warn("Failed to delete prod ref", zap.String("project", projectName), zap.Error(err))
	}

	h.cache.Delete(projectName + ":prod")
	h.cache.Delete(projectName + ":beta")

	if err := h.app.Delete(project); err != nil {
		return h.jsonError(w, http.StatusInternalServerError, "failed to delete project")
	}

	if err := h.rebuildRoutingIndex(); err != nil {
		h.logger.Warn("failed to rebuild routing index", zap.Error(err))
	}

	return h.jsonResponse(w, http.StatusOK, map[string]any{
		"message": "Project deleted successfully",
	})
}
