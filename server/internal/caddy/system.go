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
	admin.SetPassword(password)

	if err := h.app.Dao().SaveAdmin(admin); err != nil {
		return err
	}

	h.logger.Info("Default admin created",
		zap.String("email", email),
		zap.String("hint", "Change password via environment variables SITEPOD_ADMIN_EMAIL and SITEPOD_ADMIN_PASSWORD"),
	)

	return nil
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
	user.SetEmail(systemEmail)
	user.SetUsername("system")
	user.SetVerified(true)
	// System user doesn't need a real password since it can't be logged into
	user.SetPassword("__system_user_no_login__" + uuid.New().String())

	if err := h.app.Dao().SaveRecord(user); err != nil {
		return nil, err
	}

	h.logger.Info("System user created",
		zap.String("email", systemEmail),
		zap.String("purpose", "Owner of system projects like welcome, console"),
	)

	return user, nil
}

// getSystemUser returns the system user record
func (h *SitePodHandler) getSystemUser() (*models.Record, error) {
	systemEmail := os.Getenv("SITEPOD_SYSTEM_EMAIL")
	if systemEmail == "" {
		systemEmail = "system@sitepod.local"
	}
	return h.app.Dao().FindAuthRecordByEmail("users", systemEmail)
}

func (h *SitePodHandler) getOrCreateProject(name string) (*models.Record, error) {
	return h.getOrCreateProjectWithOwner(name, "")
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

	h.app.Dao().SaveRecord(domain)
	h.rebuildRoutingIndex()
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
