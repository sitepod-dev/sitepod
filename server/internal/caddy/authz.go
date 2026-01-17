package caddy

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

var errForbidden = errors.New("forbidden")
var errAdminTokenMissing = errors.New("admin token not configured")

func (h *SitePodHandler) isSystemUser(user *core.Record) bool {
	systemEmail := os.Getenv("SITEPOD_SYSTEM_EMAIL")
	if systemEmail == "" {
		systemEmail = "system@sitepod.local"
	}
	return user.GetString("email") == systemEmail
}

func (h *SitePodHandler) userOwnsProject(user *core.Record, project *core.Record) bool {
	ownerID := project.GetString("owner_id")
	if user.GetBool("is_admin") {
		return true
	}
	if ownerID == "" {
		return h.isSystemUser(user)
	}
	return ownerID == user.Id
}

func (h *SitePodHandler) requireProjectOwnerByName(projectName string, user *core.Record) (*core.Record, error) {
	project, err := h.app.FindFirstRecordByData("projects", "name", projectName)
	if err != nil {
		return nil, err
	}
	if !h.userOwnsProject(user, project) {
		return nil, errForbidden
	}
	return project, nil
}

func (h *SitePodHandler) requireProjectOwnerByID(projectID string, user *core.Record) (*core.Record, error) {
	project, err := h.app.FindRecordById("projects", projectID)
	if err != nil {
		return nil, err
	}
	if !h.userOwnsProject(user, project) {
		return nil, errForbidden
	}
	return project, nil
}

func (h *SitePodHandler) requirePlanOwner(planID string, user *core.Record) (*core.Record, *core.Record, error) {
	plan, err := h.app.FindFirstRecordByData("plans", "plan_id", planID)
	if err != nil {
		return nil, nil, err
	}
	project, err := h.app.FindRecordById("projects", plan.GetString("project_id"))
	if err != nil {
		return nil, nil, err
	}
	if !h.userOwnsProject(user, project) {
		return nil, nil, errForbidden
	}
	return plan, project, nil
}

func (h *SitePodHandler) requireDomainOwner(domain string, user *core.Record) (*core.Record, *core.Record, error) {
	domainRecord, err := h.app.FindFirstRecordByData("domains", "domain", domain)
	if err != nil {
		return nil, nil, err
	}
	project, err := h.app.FindRecordById("projects", domainRecord.GetString("project_id"))
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
	if token != "" {
		if header := r.Header.Get("X-Sitepod-Admin-Token"); header != "" && header == token {
			return nil
		}

		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") && strings.TrimPrefix(auth, "Bearer ") == token {
			return nil
		}
	}

	if authCtx, err := h.authenticateAny(r); err == nil {
		if authCtx.IsAdmin() {
			return nil
		}
		if authCtx.User != nil && authCtx.User.GetBool("is_admin") {
			return nil
		}
	}

	if token == "" {
		return errAdminTokenMissing
	}
	return errForbidden
}
