package caddy

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/pocketbase/pocketbase/models"
	"github.com/sitepod/sitepod/internal/storage"
)

// API: Health
func (h *SitePodHandler) apiHealth(w http.ResponseWriter, r *http.Request) error {
	dbStatus := "ok"
	if _, err := h.app.Dao().FindCollectionByNameOrId("projects"); err != nil {
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

	allowAnonymous := os.Getenv("SITEPOD_ALLOW_ANONYMOUS") == "1" || os.Getenv("SITEPOD_ALLOW_ANONYMOUS") == "true"

	return h.jsonResponse(w, http.StatusOK, map[string]any{
		"status":          status,
		"database":        dbStatus,
		"storage":         storageStatus,
		"uptime":          time.Since(h.startTime).String(),
		"allow_anonymous": allowAnonymous,
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
func (h *SitePodHandler) apiGetCurrent(w http.ResponseWriter, r *http.Request, user *models.Record) error {
	project := r.URL.Query().Get("project")
	env := r.URL.Query().Get("environment")

	if project == "" || env == "" {
		return h.jsonError(w, http.StatusBadRequest, "project and environment required")
	}

	if _, err := h.requireProjectOwnerByName(project, user); err != nil {
		if errors.Is(err, errForbidden) {
			return h.jsonError(w, http.StatusForbidden, "forbidden")
		}
		return h.jsonError(w, http.StatusNotFound, "project not found")
	}

	refData, err := h.storage.GetRef(project, env)
	if err != nil {
		return h.jsonError(w, http.StatusNotFound, "ref not found")
	}

	var ref storage.RefData
	if err := json.Unmarshal(refData, &ref); err != nil {
		return h.jsonError(w, http.StatusInternalServerError, "invalid ref data")
	}

	return h.jsonResponse(w, http.StatusOK, map[string]any{
		"image_id":     ref.ImageID,
		"content_hash": ref.ContentHash,
		"deployed_at":  ref.UpdatedAt,
	})
}

// API: Get History
func (h *SitePodHandler) apiGetHistory(w http.ResponseWriter, r *http.Request, user *models.Record) error {
	projectName := r.URL.Query().Get("project")

	project, err := h.requireProjectOwnerByName(projectName, user)
	if err != nil {
		if errors.Is(err, errForbidden) {
			return h.jsonError(w, http.StatusForbidden, "forbidden")
		}
		return h.jsonError(w, http.StatusNotFound, "project not found")
	}

	images, err := h.app.Dao().FindRecordsByFilter(
		"images", "project_id = {:project_id}", "-created", 20, 0,
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
			"created_at":   img.Created.Time(),
			"git_commit":   img.GetString("git_commit"),
		}
	}

	return h.jsonResponse(w, http.StatusOK, map[string]any{"items": items})
}

// API: List Images
func (h *SitePodHandler) apiListImages(w http.ResponseWriter, r *http.Request, user *models.Record) error {
	projectID := r.URL.Query().Get("project_id")
	projectName := r.URL.Query().Get("project")

	var project *models.Record
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

	images, err := h.app.Dao().FindRecordsByFilter(
		"images", "project_id = {:project_id}", "-created", 50, 0,
		map[string]any{"project_id": project.Id},
	)
	if err != nil {
		return h.jsonResponse(w, http.StatusOK, map[string]any{"images": []any{}, "total": 0})
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
			"created_at":   img.Created.String(),
			"deployed_to":  deployedTo,
		}
	}

	return h.jsonResponse(w, http.StatusOK, map[string]any{"images": result, "total": len(result)})
}

// API: List Projects
func (h *SitePodHandler) apiListProjects(w http.ResponseWriter, r *http.Request, user *models.Record) error {
	// Get all projects owned by the user
	projects, err := h.app.Dao().FindRecordsByFilter(
		"projects", "owner_id = {:owner_id}", "-created", 100, 0,
		map[string]any{"owner_id": user.Id},
	)
	if err != nil {
		// If no projects found, return empty array
		return h.jsonResponse(w, http.StatusOK, []any{})
	}

	result := make([]map[string]any, len(projects))
	for i, p := range projects {
		result[i] = map[string]any{
			"id":         p.Id,
			"name":       p.GetString("name"),
			"subdomain":  p.GetString("subdomain"),
			"owner_id":   p.GetString("owner_id"),
			"created_at": p.Created.String(),
			"updated_at": p.Updated.String(),
		}
	}

	return h.jsonResponse(w, http.StatusOK, result)
}

// API: Get Project
func (h *SitePodHandler) apiGetProject(w http.ResponseWriter, r *http.Request, projectName string, user *models.Record) error {
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
		"created_at": project.Created.String(),
		"updated_at": project.Updated.String(),
	})
}
