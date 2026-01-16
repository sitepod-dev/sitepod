package caddy

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tokens"
	"go.uber.org/zap"
)

// API: Anonymous Auth
func (h *SitePodHandler) apiAnonymousAuth(w http.ResponseWriter, r *http.Request) error {
	usersCollection, err := h.app.Dao().FindCollectionByNameOrId("users")
	if err != nil {
		return h.jsonError(w, http.StatusInternalServerError, "users collection not found")
	}

	anonID := "anon_" + uuid.New().String()[:8]
	anonEmail := anonID + "@anonymous.sitepod.local"
	expiresAt := time.Now().Add(24 * time.Hour)

	user := models.NewRecord(usersCollection)
	user.Set("email", anonEmail)
	user.Set("username", anonID)
	user.Set("verified", false)
	user.Set("is_anonymous", true)
	user.Set("anonymous_expires_at", expiresAt)
	user.SetPassword(uuid.New().String())

	if err := h.app.Dao().SaveRecord(user); err != nil {
		return h.jsonError(w, http.StatusInternalServerError, "failed to create anonymous user")
	}

	token, err := tokens.NewRecordAuthToken(h.app, user)
	if err != nil {
		return h.jsonError(w, http.StatusInternalServerError, "failed to generate token")
	}

	return h.jsonResponse(w, http.StatusOK, map[string]any{
		"token":      token,
		"user_id":    user.Id,
		"expires_at": expiresAt,
		"message":    "Anonymous session created. Verify your email within 24 hours to keep your deployments.",
	})
}

// API: Bind Email
func (h *SitePodHandler) apiBindEmail(w http.ResponseWriter, r *http.Request, user *models.Record) error {
	return h.jsonError(w, http.StatusNotImplemented, "email binding is not supported; use anonymous or password login")
}

// API: Delete Account - cascade deletes all user data
func (h *SitePodHandler) apiDeleteAccount(w http.ResponseWriter, r *http.Request, user *models.Record) error {
	// Prevent deletion of system user
	if user.GetString("email") == "system@sitepod.local" {
		return h.jsonError(w, http.StatusForbidden, "cannot delete system user")
	}

	h.logger.Info("Deleting user account",
		zap.String("user_id", user.Id),
		zap.String("email", user.GetString("email")))

	// Find all projects owned by this user
	projects, err := h.app.Dao().FindRecordsByFilter(
		"projects", "owner_id = {:owner_id}", "", 1000, 0,
		map[string]any{"owner_id": user.Id},
	)
	if err != nil {
		h.logger.Debug("No projects found for user", zap.String("user_id", user.Id))
		projects = []*models.Record{}
	}

	// For each project, delete related data
	for _, project := range projects {
		projectID := project.Id
		projectName := project.GetString("name")

		h.logger.Info("Deleting project", zap.String("project", projectName))

		// Delete domains
		domains, _ := h.app.Dao().FindRecordsByFilter(
			"domains", "project_id = {:project_id}", "", 1000, 0,
			map[string]any{"project_id": projectID},
		)
		for _, domain := range domains {
			h.app.Dao().DeleteRecord(domain)
		}

		// Delete images
		images, _ := h.app.Dao().FindRecordsByFilter(
			"images", "project_id = {:project_id}", "", 1000, 0,
			map[string]any{"project_id": projectID},
		)
		for _, image := range images {
			h.app.Dao().DeleteRecord(image)
		}

		// Delete deploy_events
		events, _ := h.app.Dao().FindRecordsByFilter(
			"deploy_events", "project_id = {:project_id}", "", 1000, 0,
			map[string]any{"project_id": projectID},
		)
		for _, event := range events {
			h.app.Dao().DeleteRecord(event)
		}

		// Delete plans
		plans, _ := h.app.Dao().FindRecordsByFilter(
			"plans", "project_id = {:project_id}", "", 1000, 0,
			map[string]any{"project_id": projectID},
		)
		for _, plan := range plans {
			h.app.Dao().DeleteRecord(plan)
		}

		// Delete ref files from storage
		h.storage.DeleteRef(projectName, "beta")
		h.storage.DeleteRef(projectName, "prod")

		// Delete the project record
		if err := h.app.Dao().DeleteRecord(project); err != nil {
			h.logger.Warn("Failed to delete project", zap.String("project", projectName), zap.Error(err))
		}
	}

	// Delete the user record
	if err := h.app.Dao().DeleteRecord(user); err != nil {
		return h.jsonError(w, http.StatusInternalServerError, "failed to delete account")
	}

	h.logger.Info("Account deleted successfully", zap.String("user_id", user.Id))

	return h.jsonResponse(w, http.StatusOK, map[string]any{
		"message":          "Account deleted successfully",
		"deleted_projects": len(projects),
	})
}
