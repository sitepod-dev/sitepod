package caddy

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tokens"
	"go.uber.org/zap"
)

// API: Register or Login - creates account if not exists, logs in if exists
func (h *SitePodHandler) apiRegisterOrLogin(w http.ResponseWriter, r *http.Request) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return h.jsonError(w, http.StatusBadRequest, "invalid request")
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	password := req.Password

	if email == "" || password == "" {
		return h.jsonError(w, http.StatusBadRequest, "email and password required")
	}

	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return h.jsonError(w, http.StatusBadRequest, "invalid email format")
	}

	if len(password) < 6 {
		return h.jsonError(w, http.StatusBadRequest, "password must be at least 6 characters")
	}

	// Try to find existing user
	user, err := h.app.Dao().FindAuthRecordByEmail("users", email)

	if err != nil {
		// User doesn't exist - create new account
		usersCollection, err := h.app.Dao().FindCollectionByNameOrId("users")
		if err != nil {
			return h.jsonError(w, http.StatusInternalServerError, "users collection not found")
		}

		// Generate username from email
		username := strings.Split(email, "@")[0]
		// Remove non-alphanumeric characters
		var cleanUsername strings.Builder
		for _, c := range username {
			if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
				cleanUsername.WriteRune(c)
			}
		}
		username = cleanUsername.String()
		if len(username) < 3 {
			username = "user" + username
		}

		user = models.NewRecord(usersCollection)
		if err := user.SetEmail(email); err != nil {
			return h.jsonError(w, http.StatusInternalServerError, "failed to create account")
		}
		if err := user.SetUsername(username); err != nil {
			return h.jsonError(w, http.StatusInternalServerError, "failed to create account")
		}
		if err := user.SetVerified(true); err != nil {
			return h.jsonError(w, http.StatusInternalServerError, "failed to create account")
		}
		if err := user.SetPassword(password); err != nil {
			return h.jsonError(w, http.StatusInternalServerError, "failed to create account")
		}

		if err := h.app.Dao().SaveRecord(user); err != nil {
			// Check if username conflict, try with suffix
			if err := user.SetUsername(username + user.Id[:4]); err != nil {
				h.logger.Error("failed to set username", zap.Error(err))
				return h.jsonError(w, http.StatusInternalServerError, "failed to create account")
			}
			if err := h.app.Dao().SaveRecord(user); err != nil {
				h.logger.Error("failed to create user", zap.Error(err))
				return h.jsonError(w, http.StatusInternalServerError, "failed to create account")
			}
		}

		h.logger.Info("New user registered", zap.String("email", email))

		token, err := tokens.NewRecordAuthToken(h.app, user)
		if err != nil {
			return h.jsonError(w, http.StatusInternalServerError, "failed to generate token")
		}

		return h.jsonResponse(w, http.StatusOK, map[string]any{
			"token":   token,
			"user_id": user.Id,
			"email":   email,
			"created": true,
			"message": "Account created successfully",
		})
	}

	// User exists - verify password
	if !user.ValidatePassword(password) {
		return h.jsonError(w, http.StatusUnauthorized, "invalid password")
	}

	token, err := tokens.NewRecordAuthToken(h.app, user)
	if err != nil {
		return h.jsonError(w, http.StatusInternalServerError, "failed to generate token")
	}

	return h.jsonResponse(w, http.StatusOK, map[string]any{
		"token":   token,
		"user_id": user.Id,
		"email":   email,
		"created": false,
		"message": "Logged in successfully",
	})
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
			if err := h.app.Dao().DeleteRecord(domain); err != nil {
				h.logger.Warn("Failed to delete domain", zap.String("domain_id", domain.Id), zap.Error(err))
			}
		}

		// Delete images
		images, _ := h.app.Dao().FindRecordsByFilter(
			"images", "project_id = {:project_id}", "", 1000, 0,
			map[string]any{"project_id": projectID},
		)
		for _, image := range images {
			if err := h.app.Dao().DeleteRecord(image); err != nil {
				h.logger.Warn("Failed to delete image", zap.String("image_id", image.Id), zap.Error(err))
			}
		}

		// Delete deploy_events
		events, _ := h.app.Dao().FindRecordsByFilter(
			"deploy_events", "project_id = {:project_id}", "", 1000, 0,
			map[string]any{"project_id": projectID},
		)
		for _, event := range events {
			if err := h.app.Dao().DeleteRecord(event); err != nil {
				h.logger.Warn("Failed to delete deploy event", zap.String("event_id", event.Id), zap.Error(err))
			}
		}

		// Delete plans
		plans, _ := h.app.Dao().FindRecordsByFilter(
			"plans", "project_id = {:project_id}", "", 1000, 0,
			map[string]any{"project_id": projectID},
		)
		for _, plan := range plans {
			if err := h.app.Dao().DeleteRecord(plan); err != nil {
				h.logger.Warn("Failed to delete plan", zap.String("plan_id", plan.Id), zap.Error(err))
			}
		}

		// Delete ref files from storage
		if err := h.storage.DeleteRef(projectName, "beta"); err != nil {
			h.logger.Warn("Failed to delete beta ref", zap.String("project", projectName), zap.Error(err))
		}
		if err := h.storage.DeleteRef(projectName, "prod"); err != nil {
			h.logger.Warn("Failed to delete prod ref", zap.String("project", projectName), zap.Error(err))
		}

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

// API: Auth Info - returns current user/admin info
func (h *SitePodHandler) apiAuthInfo(w http.ResponseWriter, r *http.Request) error {
	authCtx, err := h.authenticateAny(r)
	if err != nil {
		return h.jsonError(w, http.StatusUnauthorized, "authentication required")
	}

	if authCtx.IsAdmin() {
		return h.jsonResponse(w, http.StatusOK, map[string]any{
			"id":           authCtx.Admin.Id,
			"email":        authCtx.Admin.Email,
			"is_admin":     true,
			"is_anonymous": false,
		})
	}

	// Regular user
	isAnonymous := authCtx.User.GetBool("is_anonymous")
	return h.jsonResponse(w, http.StatusOK, map[string]any{
		"id":           authCtx.User.Id,
		"email":        authCtx.User.GetString("email"),
		"is_admin":     false,
		"is_anonymous": isAnonymous,
	})
}
