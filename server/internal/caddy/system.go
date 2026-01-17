package caddy

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	pbmigrations "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/migrate"
	"go.uber.org/zap"
)

// initDatabaseSchema runs all database migrations to create/update schema
// This is needed when running PocketBase in headless/embedded mode
func (h *SitePodHandler) initDatabaseSchema() error {
	// Run all registered migrations (PocketBase core + SitePod app migrations)
	// Our migrations are registered via the import of github.com/sitepod/sitepod/migrations
	runner, err := migrate.NewRunner(h.app.DB(), pbmigrations.AppMigrations)
	if err != nil {
		return fmt.Errorf("failed to create migration runner: %w", err)
	}

	applied, err := runner.Up()
	if err != nil {
		h.logger.Warn("migration error (may be expected if already applied)", zap.Error(err))
	}

	if len(applied) > 0 {
		h.logger.Info("Applied migrations", zap.Int("count", len(applied)))
	}

	return nil
}

// ensureDefaultAdmin creates a default admin account if none exists
func (h *SitePodHandler) ensureDefaultAdmin() error {
	// Check if any admin exists
	total, err := h.app.Dao().TotalAdmins()
	if err != nil {
		return err
	}

	if total > 0 {
		return nil // Admin already exists
	}

	// Get credentials from environment or use defaults
	email := os.Getenv("SITEPOD_ADMIN_EMAIL")
	if email == "" {
		email = "admin@sitepod.local"
	}

	password := os.Getenv("SITEPOD_ADMIN_PASSWORD")
	if password == "" {
		password = "sitepod123"
		h.logger.Warn("SITEPOD_ADMIN_PASSWORD not set; default admin password in use")
	}

	// Create admin
	admin := &models.Admin{}
	admin.Email = email
	if err := admin.SetPassword(password); err != nil {
		return err
	}

	if err := h.app.Dao().SaveAdmin(admin); err != nil {
		return err
	}

	h.logger.Info("Default admin created",
		zap.String("email", email),
		zap.String("hint", "Change password via environment variables SITEPOD_ADMIN_EMAIL and SITEPOD_ADMIN_PASSWORD"),
	)

	return nil
}

// ensureConsoleAdmin creates or updates a console admin user based on env vars.
func (h *SitePodHandler) ensureConsoleAdmin() error {
	email := os.Getenv("SITEPOD_CONSOLE_ADMIN_EMAIL")
	password := os.Getenv("SITEPOD_CONSOLE_ADMIN_PASSWORD")
	if email == "" || password == "" {
		return nil
	}

	usersCollection, err := h.app.Dao().FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}

	user, err := h.app.Dao().FindAuthRecordByEmail("users", email)
	if err == nil {
		if err := user.SetPassword(password); err != nil {
			return err
		}
		if err := user.SetVerified(true); err != nil {
			return err
		}
		user.Set("is_admin", true)
		if err := h.app.Dao().SaveRecord(user); err != nil {
			return err
		}
		h.logger.Info("Console admin updated", zap.String("email", email))
		return nil
	}

	user = models.NewRecord(usersCollection)
	if err := user.SetEmail(email); err != nil {
		return err
	}
	if err := user.SetUsername(normalizeUsername(email)); err != nil {
		return err
	}
	if err := user.SetVerified(true); err != nil {
		return err
	}
	if err := user.SetPassword(password); err != nil {
		return err
	}
	user.Set("is_admin", true)

	if err := h.app.Dao().SaveRecord(user); err != nil {
		// Retry with a random suffix if username conflict
		if err := user.SetUsername(fmt.Sprintf("admin-%s", uuid.New().String()[:6])); err != nil {
			return err
		}
		if err := h.app.Dao().SaveRecord(user); err != nil {
			return err
		}
	}

	h.logger.Info("Console admin created", zap.String("email", email))

	return nil
}

func normalizeUsername(email string) string {
	username := strings.Split(email, "@")[0]
	var cleanUsername strings.Builder
	for _, c := range username {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			cleanUsername.WriteRune(c)
		}
	}
	result := cleanUsername.String()
	if len(result) < 3 {
		result = "user" + result
	}
	return result
}

// ensureSystemUser creates the system user for internal projects
func (h *SitePodHandler) ensureSystemUser() (*models.Record, error) {
	systemEmail := os.Getenv("SITEPOD_SYSTEM_EMAIL")
	if systemEmail == "" {
		systemEmail = "system@sitepod.local"
	}

	// Check if system user already exists
	user, err := h.app.Dao().FindAuthRecordByEmail("users", systemEmail)
	if err == nil {
		return user, nil // Already exists
	}

	// Create system user
	usersCollection, err := h.app.Dao().FindCollectionByNameOrId("users")
	if err != nil {
		return nil, err
	}

	user = models.NewRecord(usersCollection)
	if err := user.SetEmail(systemEmail); err != nil {
		return nil, err
	}
	if err := user.SetUsername("system"); err != nil {
		return nil, err
	}
	if err := user.SetVerified(true); err != nil {
		return nil, err
	}
	// System user doesn't need a real password since it can't be logged into
	if err := user.SetPassword("__system_user_no_login__" + uuid.New().String()); err != nil {
		return nil, err
	}

	if err := h.app.Dao().SaveRecord(user); err != nil {
		return nil, err
	}

	h.logger.Info("System user created",
		zap.String("email", systemEmail),
		zap.String("purpose", "Owner of system projects like welcome, console"),
	)

	return user, nil
}

// ensureDemoUser creates a demo user if IS_DEMO environment variable is set
func (h *SitePodHandler) ensureDemoUser() error {
	// Only create demo user if IS_DEMO is set
	if os.Getenv("IS_DEMO") != "1" && os.Getenv("IS_DEMO") != "true" {
		return nil
	}

	demoEmail := "demo@sitepod.dev"
	demoPassword := "demo123"

	// Check if demo user already exists
	user, err := h.app.Dao().FindAuthRecordByEmail("users", demoEmail)
	if err == nil {
		// User exists, update password if needed
		if err := user.SetPassword(demoPassword); err != nil {
			return err
		}
		if err := user.SetVerified(true); err != nil {
			return err
		}
		if err := h.app.Dao().SaveRecord(user); err != nil {
			return err
		}
		h.logger.Info("Demo user updated", zap.String("email", demoEmail))
		return nil
	}

	// Create demo user
	usersCollection, err := h.app.Dao().FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}

	user = models.NewRecord(usersCollection)
	if err := user.SetEmail(demoEmail); err != nil {
		return err
	}
	if err := user.SetUsername("demo"); err != nil {
		return err
	}
	if err := user.SetVerified(true); err != nil {
		return err
	}
	if err := user.SetPassword(demoPassword); err != nil {
		return err
	}

	if err := h.app.Dao().SaveRecord(user); err != nil {
		return err
	}

	h.logger.Info("Demo user created",
		zap.String("email", demoEmail),
		zap.String("password", demoPassword),
		zap.String("purpose", "Demo login for IS_DEMO mode"),
	)

	return nil
}

// getSystemUser returns the system user record
func (h *SitePodHandler) getSystemUser() (*models.Record, error) {
	systemEmail := os.Getenv("SITEPOD_SYSTEM_EMAIL")
	if systemEmail == "" {
		systemEmail = "system@sitepod.local"
	}
	return h.app.Dao().FindAuthRecordByEmail("users", systemEmail)
}

func (h *SitePodHandler) getOrCreateProjectWithOwner(name string, ownerID string) (*models.Record, error) {
	project, err := h.app.Dao().FindFirstRecordByData("projects", "name", name)
	if err == nil {
		if ownerID != "" {
			currentOwner := project.GetString("owner_id")
			if currentOwner == "" {
				systemUser, sysErr := h.getSystemUser()
				if sysErr == nil && systemUser != nil && systemUser.Id == ownerID {
					project.Set("owner_id", ownerID)
					if err := h.app.Dao().SaveRecord(project); err != nil {
						return nil, err
					}
					return project, nil
				}
				return nil, errForbidden
			}
			if currentOwner != ownerID {
				return nil, errForbidden
			}
		}
		return project, nil
	}

	projectsCollection, err := h.app.Dao().FindCollectionByNameOrId("projects")
	if err != nil {
		return nil, err
	}

	subdomain := h.normalizeSubdomain(name)

	project = models.NewRecord(projectsCollection)
	project.Set("name", name)
	project.Set("subdomain", subdomain)
	project.Set("routing_mode", "subdomain")
	if ownerID != "" {
		project.Set("owner_id", ownerID)
	}

	if err := h.app.Dao().SaveRecord(project); err != nil {
		return nil, err
	}

	// Create system domain
	h.createSystemDomain(project, subdomain)

	return project, nil
}

func (h *SitePodHandler) createSystemDomain(project *models.Record, subdomain string) {
	domainsCollection, err := h.app.Dao().FindCollectionByNameOrId("domains")
	if err != nil {
		return
	}

	fullDomain := subdomain + "." + h.Domain

	domain := models.NewRecord(domainsCollection)
	domain.Set("domain", fullDomain)
	domain.Set("slug", "/")
	domain.Set("project_id", project.Id)
	domain.Set("type", "system")
	domain.Set("status", "active")
	domain.Set("is_primary", true)

	if err := h.app.Dao().SaveRecord(domain); err != nil {
		h.logger.Warn("failed to save system domain", zap.Error(err))
		return
	}
	if err := h.rebuildRoutingIndex(); err != nil {
		h.logger.Warn("failed to rebuild routing index", zap.Error(err))
	}
}

func (h *SitePodHandler) normalizeSubdomain(name string) string {
	result := strings.ToLower(name)
	var normalized strings.Builder
	for _, c := range result {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			normalized.WriteRune(c)
		} else {
			normalized.WriteRune('-')
		}
	}
	return strings.Trim(normalized.String(), "-")
}

func (h *SitePodHandler) buildURL(project, env string) string {
	scheme := "https"
	if h.Domain == "localhost" || strings.HasPrefix(h.Domain, "localhost:") {
		scheme = "http"
	}
	if env == "beta" {
		return fmt.Sprintf("%s://%s-beta.%s", scheme, project, h.Domain)
	}
	return fmt.Sprintf("%s://%s.%s", scheme, project, h.Domain)
}

func (h *SitePodHandler) buildPreviewURL(project, slug string) string {
	scheme := "https"
	if h.Domain == "localhost" || strings.HasPrefix(h.Domain, "localhost:") {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s.%s/__preview__/%s/", scheme, project, h.Domain, slug)
}

func (h *SitePodHandler) rebuildRoutingIndex() error {
	domains, err := h.app.Dao().FindRecordsByFilter(
		"domains", "status = 'active'", "-is_primary", 1000, 0, nil,
	)
	if err != nil {
		return err
	}

	entries := make([]RoutingEntry, 0, len(domains))
	for _, d := range domains {
		projectID := d.GetString("project_id")
		project, err := h.app.Dao().FindRecordById("projects", projectID)
		if err != nil {
			continue
		}

		entries = append(entries, RoutingEntry{
			Domain:    d.GetString("domain"),
			Slug:      d.GetString("slug"),
			ProjectID: projectID,
			Project:   project.GetString("name"),
		})
	}

	index := RoutingIndex{
		Entries:   entries,
		UpdatedAt: time.Now(),
	}

	indexJSON, _ := json.Marshal(index)
	return h.storage.PutRouting(indexJSON)
}
