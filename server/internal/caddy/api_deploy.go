package caddy

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pocketbase/pocketbase/models"
	"github.com/sitepod/sitepod/internal/storage"
	"github.com/zeebo/blake3"
)

// API: Plan
func (h *SitePodHandler) apiPlan(w http.ResponseWriter, r *http.Request, user *models.Record) error {
	var req struct {
		Project string      `json:"project"`
		Files   []FileEntry `json:"files"`
		Git     *struct {
			Commit  string `json:"commit"`
			Branch  string `json:"branch"`
			Message string `json:"message"`
		} `json:"git"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return h.jsonError(w, http.StatusBadRequest, "invalid request")
	}
	if req.Project == "" {
		return h.jsonError(w, http.StatusBadRequest, "project required")
	}

	// Check quotas
	isAnonymous := user.GetBool("is_anonymous")
	if err := h.checkDeployQuotas(req.Files, isAnonymous); err != nil {
		return h.jsonError(w, http.StatusBadRequest, err.Error())
	}

	// Check project count quota (only for new projects)
	existingProject, _ := h.app.Dao().FindFirstRecordByData("projects", "name", req.Project)
	if existingProject == nil {
		if err := h.checkProjectCountQuota(user.Id, isAnonymous); err != nil {
			return h.jsonError(w, http.StatusBadRequest, err.Error())
		}
	}

	project, err := h.getOrCreateProjectWithOwner(req.Project, user.Id)
	if err != nil {
		if errors.Is(err, errForbidden) {
			return h.jsonError(w, http.StatusForbidden, "forbidden")
		}
		return h.jsonError(w, http.StatusInternalServerError, err.Error())
	}

	// Calculate content hash (stable order by path)
	files := append([]FileEntry(nil), req.Files...)
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
	hasher := blake3.New()
	for _, f := range files {
		hasher.Write([]byte(f.Path))
		hasher.Write([]byte(f.Blake3))
	}
	contentHash := hex.EncodeToString(hasher.Sum(nil))

	// Check which blobs are missing
	type missingBlob struct {
		Path      string `json:"path"`
		Hash      string `json:"hash"`
		Size      int64  `json:"size"`
		UploadURL string `json:"upload_url"`
	}

	missing := make([]missingBlob, 0)
	reusable := 0

	for _, f := range req.Files {
		exists, err := h.storage.HasBlob(f.Blake3)
		if err != nil {
			return h.jsonError(w, http.StatusInternalServerError, err.Error())
		}

		if exists {
			reusable++
		} else {
			var uploadURL string
			if h.storage.UploadMode() == "presigned" {
				uploadURL, _ = h.storage.GenerateUploadURL(f.Blake3, f.SHA256, f.Size)
			}
			missing = append(missing, missingBlob{
				Path:      f.Path,
				Hash:      f.Blake3,
				Size:      f.Size,
				UploadURL: uploadURL,
			})
		}
	}

	// Create plan record
	planID := "plan_" + uuid.New().String()[:8]

	manifest := make(map[string]storage.FileEntry)
	for _, f := range req.Files {
		manifest[f.Path] = storage.FileEntry{
			Hash:        f.Blake3,
			Size:        f.Size,
			ContentType: f.ContentType,
		}
	}

	manifestJSON, _ := json.Marshal(manifest)
	missingJSON, _ := json.Marshal(missing)

	plansCollection, _ := h.app.Dao().FindCollectionByNameOrId("plans")
	if plansCollection == nil {
		return h.jsonError(w, http.StatusInternalServerError, "plans collection not found")
	}

	planRecord := models.NewRecord(plansCollection)
	planRecord.Set("plan_id", planID)
	planRecord.Set("project_id", project.Id)
	planRecord.Set("content_hash", contentHash)
	planRecord.Set("manifest", string(manifestJSON))
	planRecord.Set("missing_blobs", string(missingJSON))
	planRecord.Set("upload_mode", h.storage.UploadMode())
	planRecord.Set("status", "pending")
	planRecord.Set("expires_at", time.Now().Add(30*time.Minute))

	if req.Git != nil {
		planRecord.Set("git_commit", req.Git.Commit)
		planRecord.Set("git_branch", req.Git.Branch)
		planRecord.Set("git_message", req.Git.Message)
	}

	if err := h.app.Dao().SaveRecord(planRecord); err != nil {
		return h.jsonError(w, http.StatusInternalServerError, err.Error())
	}

	// Set upload URLs for direct mode
	if h.storage.UploadMode() == "direct" {
		for i := range missing {
			missing[i].UploadURL = fmt.Sprintf("/api/v1/upload/%s/%s", planID, missing[i].Hash)
		}
	}

	return h.jsonResponse(w, http.StatusOK, map[string]any{
		"plan_id":      planID,
		"content_hash": contentHash,
		"upload_mode":  h.storage.UploadMode(),
		"missing":      missing,
		"reusable":     reusable,
	})
}

// API: Upload
func (h *SitePodHandler) apiUpload(w http.ResponseWriter, r *http.Request, path string, user *models.Record) error {
	// Parse: /upload/{plan_id}/{hash}
	parts := strings.Split(strings.TrimPrefix(path, "/upload/"), "/")
	if len(parts) != 2 {
		return h.jsonError(w, http.StatusBadRequest, "invalid upload path")
	}

	planID, hash := parts[0], parts[1]

	plan, project, err := h.requirePlanOwner(planID, user)
	if err != nil {
		if errors.Is(err, errForbidden) {
			return h.jsonError(w, http.StatusForbidden, "forbidden")
		}
		return h.jsonError(w, http.StatusNotFound, "plan not found")
	}

	if plan.GetString("status") != "pending" {
		return h.jsonError(w, http.StatusBadRequest, "plan is not pending")
	}
	if planExpired(plan) {
		plan.Set("status", "expired")
		h.app.Dao().SaveRecord(plan)
		return h.jsonError(w, http.StatusBadRequest, "plan expired")
	}

	if project == nil {
		return h.jsonError(w, http.StatusNotFound, "project not found")
	}

	var missing []struct {
		Hash string `json:"hash"`
		Size int64  `json:"size"`
	}
	if err := json.Unmarshal([]byte(plan.GetString("missing_blobs")), &missing); err != nil {
		return h.jsonError(w, http.StatusInternalServerError, "invalid plan")
	}

	var expectedSize int64 = -1
	for _, entry := range missing {
		if entry.Hash == hash {
			expectedSize = entry.Size
			break
		}
	}
	if expectedSize < 0 {
		return h.jsonError(w, http.StatusBadRequest, "blob not in plan")
	}

	if r.ContentLength < 0 {
		return h.jsonError(w, http.StatusBadRequest, "content-length required")
	}
	if expectedSize != r.ContentLength {
		return h.jsonError(w, http.StatusBadRequest, "size mismatch")
	}

	// Store the blob
	if err := h.storage.PutBlob(hash, r.Body, r.ContentLength); err != nil {
		if _, ok := err.(*storage.HashMismatchError); ok {
			return h.jsonError(w, http.StatusBadRequest, "hash mismatch")
		}
		return h.jsonError(w, http.StatusInternalServerError, err.Error())
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

// API: Commit
func (h *SitePodHandler) apiCommit(w http.ResponseWriter, r *http.Request, user *models.Record) error {
	var req struct {
		PlanID string `json:"plan_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return h.jsonError(w, http.StatusBadRequest, "invalid request")
	}

	plan, project, err := h.requirePlanOwner(req.PlanID, user)
	if err != nil {
		if errors.Is(err, errForbidden) {
			return h.jsonError(w, http.StatusForbidden, "forbidden")
		}
		return h.jsonError(w, http.StatusNotFound, "plan not found")
	}

	if plan.GetString("status") != "pending" {
		return h.jsonError(w, http.StatusBadRequest, "plan is not pending")
	}
	if planExpired(plan) {
		plan.Set("status", "expired")
		h.app.Dao().SaveRecord(plan)
		return h.jsonError(w, http.StatusBadRequest, "plan expired")
	}

	if project == nil {
		return h.jsonError(w, http.StatusNotFound, "project not found")
	}

	// Verify all blobs exist
	var manifest map[string]storage.FileEntry
	if err := json.Unmarshal([]byte(plan.GetString("manifest")), &manifest); err != nil {
		return h.jsonError(w, http.StatusInternalServerError, "invalid manifest")
	}

	for _, entry := range manifest {
		exists, err := h.storage.HasBlob(entry.Hash)
		if err != nil {
			return h.jsonError(w, http.StatusInternalServerError, err.Error())
		}
		if !exists {
			return h.jsonError(w, http.StatusBadRequest, "missing blob: "+entry.Hash)
		}
	}

	// Create image record
	imageID := "img_" + uuid.New().String()[:8]
	contentHash := plan.GetString("content_hash")

	imagesCollection, _ := h.app.Dao().FindCollectionByNameOrId("images")
	if imagesCollection == nil {
		return h.jsonError(w, http.StatusInternalServerError, "images collection not found")
	}

	imageRecord := models.NewRecord(imagesCollection)
	imageRecord.Set("image_id", imageID)
	imageRecord.Set("project_id", plan.GetString("project_id"))
	imageRecord.Set("content_hash", contentHash)
	imageRecord.Set("manifest", plan.GetString("manifest"))
	imageRecord.Set("file_count", len(manifest))
	imageRecord.Set("git_commit", plan.GetString("git_commit"))
	imageRecord.Set("git_branch", plan.GetString("git_branch"))
	imageRecord.Set("git_message", plan.GetString("git_message"))

	var totalSize int64
	for _, entry := range manifest {
		totalSize += entry.Size
	}
	imageRecord.Set("total_size", totalSize)

	if err := h.app.Dao().SaveRecord(imageRecord); err != nil {
		return h.jsonError(w, http.StatusInternalServerError, err.Error())
	}

	plan.Set("status", "committed")
	h.app.Dao().SaveRecord(plan)

	return h.jsonResponse(w, http.StatusOK, map[string]string{
		"image_id":     imageID,
		"content_hash": contentHash,
	})
}

// API: Release
func (h *SitePodHandler) apiRelease(w http.ResponseWriter, r *http.Request, user *models.Record) error {
	var req struct {
		Project     string `json:"project"`
		ProjectID   string `json:"project_id"`
		Environment string `json:"environment"`
		ImageID     string `json:"image_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return h.jsonError(w, http.StatusBadRequest, "invalid request")
	}

	if req.Environment != "prod" && req.Environment != "beta" {
		return h.jsonError(w, http.StatusBadRequest, "environment must be 'prod' or 'beta'")
	}

	var project *models.Record
	var err error
	if req.ProjectID != "" {
		project, err = h.app.Dao().FindRecordById("projects", req.ProjectID)
	} else if req.Project != "" {
		project, err = h.app.Dao().FindFirstRecordByData("projects", "name", req.Project)
	} else {
		return h.jsonError(w, http.StatusBadRequest, "project or project_id required")
	}
	if err != nil {
		return h.jsonError(w, http.StatusNotFound, "project not found")
	}
	if !h.userOwnsProject(user, project) {
		return h.jsonError(w, http.StatusForbidden, "forbidden")
	}
	projectName := project.GetString("name")

	// Find image
	var image *models.Record
	if req.ImageID != "" {
		image, err = h.app.Dao().FindFirstRecordByData("images", "image_id", req.ImageID)
	} else {
		images, err := h.app.Dao().FindRecordsByFilter(
			"images", "project_id = {:project_id}", "-created", 1, 0,
			map[string]any{"project_id": project.Id},
		)
		if err == nil && len(images) > 0 {
			image = images[0]
		}
	}

	if err != nil || image == nil {
		return h.jsonError(w, http.StatusNotFound, "image not found")
	}
	if image.GetString("project_id") != project.Id {
		return h.jsonError(w, http.StatusForbidden, "forbidden")
	}

	// Get current ref for audit
	var previousImageID string
	if refData, err := h.storage.GetRef(projectName, req.Environment); err == nil {
		var currentRef storage.RefData
		if json.Unmarshal(refData, &currentRef) == nil {
			previousImageID = currentRef.ImageID
		}
	}

	// Build and write ref
	var manifest map[string]storage.FileEntry
	json.Unmarshal([]byte(image.GetString("manifest")), &manifest)

	refData := storage.RefData{
		ImageID:     image.GetString("image_id"),
		ContentHash: image.GetString("content_hash"),
		Manifest:    manifest,
		UpdatedAt:   time.Now(),
	}

	refJSON, _ := json.Marshal(refData)

	if err := h.storage.PutRef(projectName, req.Environment, refJSON); err != nil {
		return h.jsonError(w, http.StatusInternalServerError, "failed to write ref: "+err.Error())
	}

	h.cache.Delete(projectName + ":" + req.Environment)

	// Record deploy event
	eventsCollection, _ := h.app.Dao().FindCollectionByNameOrId("deploy_events")
	if eventsCollection != nil {
		eventRecord := models.NewRecord(eventsCollection)
		eventRecord.Set("project_id", project.Id)
		eventRecord.Set("image_id", image.Id)
		eventRecord.Set("environment", req.Environment)
		eventRecord.Set("action", "deploy")
		eventRecord.Set("previous_image_id", previousImageID)
		h.app.Dao().SaveRecord(eventRecord)
	}

	url := h.buildURL(projectName, req.Environment)

	return h.jsonResponse(w, http.StatusOK, map[string]string{"url": url})
}

// API: Rollback
func (h *SitePodHandler) apiRollback(w http.ResponseWriter, r *http.Request, user *models.Record) error {
	var req struct {
		Project     string `json:"project"`
		Environment string `json:"environment"`
		ImageID     string `json:"image_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return h.jsonError(w, http.StatusBadRequest, "invalid request")
	}
	if req.Environment != "prod" && req.Environment != "beta" {
		return h.jsonError(w, http.StatusBadRequest, "environment must be 'prod' or 'beta'")
	}

	project, err := h.app.Dao().FindFirstRecordByData("projects", "name", req.Project)
	if err != nil {
		return h.jsonError(w, http.StatusNotFound, "project not found")
	}
	if !h.userOwnsProject(user, project) {
		return h.jsonError(w, http.StatusForbidden, "forbidden")
	}

	image, err := h.app.Dao().FindFirstRecordByData("images", "image_id", req.ImageID)
	if err != nil {
		return h.jsonError(w, http.StatusNotFound, "image not found")
	}
	if image.GetString("project_id") != project.Id {
		return h.jsonError(w, http.StatusForbidden, "forbidden")
	}

	// Get current ref
	var previousImageID string
	if refData, err := h.storage.GetRef(req.Project, req.Environment); err == nil {
		var currentRef storage.RefData
		if json.Unmarshal(refData, &currentRef) == nil {
			previousImageID = currentRef.ImageID
		}
	}

	// Build and write ref
	var manifest map[string]storage.FileEntry
	json.Unmarshal([]byte(image.GetString("manifest")), &manifest)

	refData := storage.RefData{
		ImageID:     image.GetString("image_id"),
		ContentHash: image.GetString("content_hash"),
		Manifest:    manifest,
		UpdatedAt:   time.Now(),
	}

	refJSON, _ := json.Marshal(refData)

	if err := h.storage.PutRef(req.Project, req.Environment, refJSON); err != nil {
		return h.jsonError(w, http.StatusInternalServerError, "failed to write ref")
	}

	h.cache.Delete(req.Project + ":" + req.Environment)

	// Record rollback event
	eventsCollection, _ := h.app.Dao().FindCollectionByNameOrId("deploy_events")
	if eventsCollection != nil {
		eventRecord := models.NewRecord(eventsCollection)
		eventRecord.Set("project_id", project.Id)
		eventRecord.Set("image_id", image.Id)
		eventRecord.Set("environment", req.Environment)
		eventRecord.Set("action", "rollback")
		eventRecord.Set("previous_image_id", previousImageID)
		h.app.Dao().SaveRecord(eventRecord)
	}

	url := h.buildURL(req.Project, req.Environment)

	return h.jsonResponse(w, http.StatusOK, map[string]any{
		"url":               url,
		"previous_image_id": previousImageID,
	})
}

func planExpired(plan *models.Record) bool {
	expiresAt := plan.GetDateTime("expires_at")
	if expiresAt.IsZero() {
		return false
	}
	return time.Now().After(expiresAt.Time())
}

// API: Create Preview
func (h *SitePodHandler) apiCreatePreview(w http.ResponseWriter, r *http.Request, user *models.Record) error {
	var req struct {
		Project   string `json:"project"`
		ImageID   string `json:"image_id"`
		Slug      string `json:"slug"`
		ExpiresIn int64  `json:"expires_in"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return h.jsonError(w, http.StatusBadRequest, "invalid request")
	}

	project, err := h.requireProjectOwnerByName(req.Project, user)
	if err != nil {
		if errors.Is(err, errForbidden) {
			return h.jsonError(w, http.StatusForbidden, "forbidden")
		}
		return h.jsonError(w, http.StatusNotFound, "project not found")
	}

	image, err := h.app.Dao().FindFirstRecordByData("images", "image_id", req.ImageID)
	if err != nil {
		return h.jsonError(w, http.StatusNotFound, "image not found")
	}
	if image.GetString("project_id") != project.Id {
		return h.jsonError(w, http.StatusForbidden, "forbidden")
	}

	slug := req.Slug
	if slug == "" {
		slug = uuid.New().String()[:8]
	}

	expiresIn := req.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 86400
	}
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)

	var manifest map[string]storage.FileEntry
	json.Unmarshal([]byte(image.GetString("manifest")), &manifest)

	previewRef := storage.PreviewRef{
		ImageID:   image.GetString("image_id"),
		Manifest:  manifest,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	previewJSON, _ := json.Marshal(previewRef)

	if err := h.storage.PutPreview(req.Project, slug, previewJSON); err != nil {
		return h.jsonError(w, http.StatusInternalServerError, "failed to create preview")
	}

	previewsCollection, _ := h.app.Dao().FindCollectionByNameOrId("previews")
	if previewsCollection != nil {
		previewRecord := models.NewRecord(previewsCollection)
		previewRecord.Set("project", req.Project)
		previewRecord.Set("image_id", image.Id)
		previewRecord.Set("slug", slug)
		previewRecord.Set("expires_at", expiresAt)
		h.app.Dao().SaveRecord(previewRecord)
	}

	url := h.buildPreviewURL(req.Project, slug)

	return h.jsonResponse(w, http.StatusOK, map[string]any{
		"url":        url,
		"expires_at": expiresAt,
	})
}
