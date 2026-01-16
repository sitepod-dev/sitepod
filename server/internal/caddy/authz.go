package caddy

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase/models"
)

var errForbidden = errors.New("forbidden")
var errAdminTokenMissing = errors.New("admin token not configured")

func (h *SitePodHandler) isSystemUser(user *models.Record) bool {
	systemEmail := os.Getenv("SITEPOD_SYSTEM_EMAIL")
	if systemEmail == "" {
		systemEmail = "system@sitepod.local"
	}
	return user.GetString("email") == systemEmail
}

func (h *SitePodHandler) userOwnsProject(user *models.Record, project *models.Record) bool {
	ownerID := project.GetString("owner_id")
	if ownerID == "" {
		return h.isSystemUser(user)
	}
	return ownerID == user.Id
}

func (h *SitePodHandler) requireProjectOwnerByName(projectName string, user *models.Record) (*models.Record, error) {
	project, err := h.app.Dao().FindFirstRecordByData("projects", "name", projectName)
	if err != nil {
		return nil, err
	}
	if !h.userOwnsProject(user, project) {
		return nil, errForbidden
	}
	return project, nil
}

func (h *SitePodHandler) requireProjectOwnerByID(projectID string, user *models.Record) (*models.Record, error) {
	project, err := h.app.Dao().FindRecordById("projects", projectID)
	if err != nil {
		return nil, err
	}
	if !h.userOwnsProject(user, project) {
		return nil, errForbidden
	}
	return project, nil
}

func (h *SitePodHandler) requirePlanOwner(planID string, user *models.Record) (*models.Record, *models.Record, error) {
	plan, err := h.app.Dao().FindFirstRecordByData("plans", "plan_id", planID)
	if err != nil {
		return nil, nil, err
	}
	project, err := h.app.Dao().FindRecordById("projects", plan.GetString("project_id"))
	if err != nil {
		return nil, nil, err
	}
	if !h.userOwnsProject(user, project) {
		return nil, nil, errForbidden
	}
	return plan, project, nil
}

func (h *SitePodHandler) requireImageOwner(imageID string, user *models.Record) (*models.Record, *models.Record, error) {
	image, err := h.app.Dao().FindFirstRecordByData("images", "image_id", imageID)
	if err != nil {
		return nil, nil, err
	}
	project, err := h.app.Dao().FindRecordById("projects", image.GetString("project_id"))
	if err != nil {
		return nil, nil, err
	}
	if !h.userOwnsProject(user, project) {
		return nil, nil, errForbidden
	}
	return image, project, nil
}

func (h *SitePodHandler) requireDomainOwner(domain string, user *models.Record) (*models.Record, *models.Record, error) {
	domainRecord, err := h.app.Dao().FindFirstRecordByData("domains", "domain", domain)
	if err != nil {
		return nil, nil, err
	}
	project, err := h.app.Dao().FindRecordById("projects", domainRecord.GetString("project_id"))
	if err != nil {
		return nil, nil, err
	}
	if !h.userOwnsProject(user, project) {
		return nil, nil, errForbidden
	}
	return domainRecord, project, nil
}

func (h *SitePodHandler) requireAdminToken(r *http.Request) error {
	token := os.Getenv("SITEPOD_ADMIN_TOKEN")
	if token == "" {
		return errAdminTokenMissing
	}

	if header := r.Header.Get("X-Sitepod-Admin-Token"); header != "" && header == token {
		return nil
	}

	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") && strings.TrimPrefix(auth, "Bearer ") == token {
		return nil
	}

	return errForbidden
}
