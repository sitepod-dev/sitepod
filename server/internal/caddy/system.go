package caddy

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pocketbase/pocketbase/core"
	"go.uber.org/zap"
)

// initDatabaseSchema runs all database migrations to create/update schema
// This is needed when running PocketBase in headless/embedded mode
func (h *SitePodHandler) initDatabaseSchema() error {
	// In PocketBase 0.36+, migrations are handled automatically by Bootstrap
	// Our custom migrations are registered via the import of github.com/sitepod/sitepod/migrations
	// and will be run during Bootstrap()

	// Run pending migrations
	if err := h.app.RunAllMigrations(); err != nil {
		h.logger.Warn("migration error (may be expected if already applied)", zap.Error(err))
	}

	return nil
}

// ensureDefaultAdmin creates a default superuser account if none exists
// and ensures an API admin user exists in the users collection
func (h *SitePodHandler) ensureDefaultAdmin() error {
	// Get credentials from environment or use defaults
	email := os.Getenv("SITEPOD_ADMIN_EMAIL")
	if email == "" {
		email = "admin@sitepod.local"
	}

	password := os.Getenv("SITEPOD_ADMIN_PASSWORD")
	if password == "" {
		password = "sitepod123"
	}

	// Always ensure API admin user exists (even if superuser already exists)
	if err := h.ensureAPIAdminUser(email, password); err != nil {
		h.logger.Warn("failed to ensure API admin user", zap.Error(err))
	}

	// Check if any superuser exists
	superusers, err := h.app.FindAllRecords(core.CollectionNameSuperusers)
	if err != nil {
		return err
	}

	if len(superusers) > 0 {
		return nil // Superuser already exists
	}

	// Log warning if using default password
	if os.Getenv("SITEPOD_ADMIN_PASSWORD") == "" {
		h.logger.Warn("SITEPOD_ADMIN_PASSWORD not set; default admin password in use")
	}

	// Create superuser
	superusersCollection, err := h.app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		return err
	}

	superuser := core.NewRecord(superusersCollection)
	superuser.SetEmail(email)
	superuser.SetPassword(password)

	if err := h.app.Save(superuser); err != nil {
		return err
	}

	h.logger.Info("Default admin created",
		zap.String("email", email),
		zap.String("hint", "Change password via environment variables SITEPOD_ADMIN_EMAIL and SITEPOD_ADMIN_PASSWORD"),
	)

	return nil
}

// ensureAPIAdminUser creates an admin user in the users collection for API access
func (h *SitePodHandler) ensureAPIAdminUser(email, password string) error {
	// Check if user already exists
	existingUser, _ := h.app.FindAuthRecordByEmail("users", email)
	if existingUser != nil {
		// User exists, ensure is_admin is true
		if !existingUser.GetBool("is_admin") {
			existingUser.Set("is_admin", true)
			if err := h.app.Save(existingUser); err != nil {
				return err
			}
			h.logger.Info("Updated existing user to admin", zap.String("email", email))
		}
		return nil
	}

	// Create new admin user
	usersCollection, err := h.app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}

	user := core.NewRecord(usersCollection)
	user.SetEmail(email)
	user.Set("username", "admin")
	user.SetVerified(true)
	user.SetPassword(password)
	user.Set("is_admin", true)

	if err := h.app.Save(user); err != nil {
		return err
	}

	h.logger.Info("API admin user created", zap.String("email", email))
	return nil
}

// ensureConsoleAdmin creates or updates a console admin user based on env vars.
// In Demo mode or local development, if SITEPOD_CONSOLE_ADMIN_* is not set, uses PocketBase admin credentials.
func (h *SitePodHandler) ensureConsoleAdmin() error {
	email := os.Getenv("SITEPOD_CONSOLE_ADMIN_EMAIL")
	password := os.Getenv("SITEPOD_CONSOLE_ADMIN_PASSWORD")

	// In Demo mode or local dev, fallback to PocketBase admin credentials if Console admin not set
	isDemo := os.Getenv("IS_DEMO") == "1" || os.Getenv("IS_DEMO") == "true"
	isLocalDev := strings.HasPrefix(h.Domain, "localhost")
	if email == "" || password == "" {
		if isDemo || isLocalDev {
			email = os.Getenv("SITEPOD_ADMIN_EMAIL")
			if email == "" {
				email = "admin@sitepod.local"
			}
			password = os.Getenv("SITEPOD_ADMIN_PASSWORD")
			if password == "" {
				password = "sitepod123"
			}
		} else {
			return nil
		}
	}

	usersCollection, err := h.app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}

	user, err := h.app.FindAuthRecordByEmail("users", email)
	if err == nil {
		user.SetPassword(password)
		user.SetVerified(true)
		user.Set("is_admin", true)
		if err := h.app.Save(user); err != nil {
			return err
		}
		h.logger.Info("Console admin updated", zap.String("email", email))
		return nil
	}

	user = core.NewRecord(usersCollection)
	user.SetEmail(email)
	user.Set("username", normalizeUsername(email))
	user.SetVerified(true)
	user.SetPassword(password)
	user.Set("is_admin", true)

	if err := h.app.Save(user); err != nil {
		// Retry with a random suffix if username conflict
		user.Set("username", fmt.Sprintf("admin-%s", uuid.New().String()[:6]))
		if err := h.app.Save(user); err != nil {
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
func (h *SitePodHandler) ensureSystemUser() (*core.Record, error) {
	systemEmail := os.Getenv("SITEPOD_SYSTEM_EMAIL")
	if systemEmail == "" {
		systemEmail = "system@sitepod.local"
	}

	// Check if system user already exists
	user, err := h.app.FindAuthRecordByEmail("users", systemEmail)
	if err == nil {
		return user, nil // Already exists
	}

	// Create system user
	usersCollection, err := h.app.FindCollectionByNameOrId("users")
	if err != nil {
		return nil, err
	}

	user = core.NewRecord(usersCollection)
	user.SetEmail(systemEmail)
	user.Set("username", "system")
	user.SetVerified(true)
	// System user doesn't need a real password since it can't be logged into
	user.SetPassword("__system_user_no_login__" + uuid.New().String())

	if err := h.app.Save(user); err != nil {
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
	user, err := h.app.FindAuthRecordByEmail("users", demoEmail)
	if err == nil {
		// User exists, update password if needed
		user.SetPassword(demoPassword)
		user.SetVerified(true)
		if err := h.app.Save(user); err != nil {
			return err
		}
		h.logger.Info("Demo user updated", zap.String("email", demoEmail))
		return nil
	}

	// Create demo user
	usersCollection, err := h.app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}

	user = core.NewRecord(usersCollection)
	user.SetEmail(demoEmail)
	user.Set("username", "demo")
	user.SetVerified(true)
	user.SetPassword(demoPassword)

	if err := h.app.Save(user); err != nil {
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
func (h *SitePodHandler) getSystemUser() (*core.Record, error) {
	systemEmail := os.Getenv("SITEPOD_SYSTEM_EMAIL")
	if systemEmail == "" {
		systemEmail = "system@sitepod.local"
	}
	return h.app.FindAuthRecordByEmail("users", systemEmail)
}

func (h *SitePodHandler) getOrCreateProjectWithOwner(name string, ownerID string) (*core.Record, error) {
	project, err := h.app.FindFirstRecordByData("projects", "name", name)
	if err == nil {
		if ownerID != "" {
			currentOwner := project.GetString("owner_id")
			if currentOwner == "" {
				systemUser, sysErr := h.getSystemUser()
				if sysErr == nil && systemUser != nil && systemUser.Id == ownerID {
					project.Set("owner_id", ownerID)
					if err := h.app.Save(project); err != nil {
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

	projectsCollection, err := h.app.FindCollectionByNameOrId("projects")
	if err != nil {
		return nil, err
	}

	subdomain := h.normalizeSubdomain(name)

	project = core.NewRecord(projectsCollection)
	project.Set("name", name)
	project.Set("subdomain", subdomain)
	project.Set("routing_mode", "subdomain")
	if ownerID != "" {
		project.Set("owner_id", ownerID)
	}

	if err := h.app.Save(project); err != nil {
		return nil, err
	}

	// Create system domain
	h.createSystemDomain(project, subdomain)

	return project, nil
}

func (h *SitePodHandler) createSystemDomain(project *core.Record, subdomain string) {
	domainsCollection, err := h.app.FindCollectionByNameOrId("domains")
	if err != nil {
		return
	}

	fullDomain := subdomain + "." + h.Domain

	domain := core.NewRecord(domainsCollection)
	domain.Set("domain", fullDomain)
	domain.Set("slug", "/")
	domain.Set("project_id", project.Id)
	domain.Set("type", "system")
	domain.Set("status", "active")
	domain.Set("is_primary", true)

	if err := h.app.Save(domain); err != nil {
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
	domains, err := h.app.FindRecordsByFilter(
		"domains", "status = 'active'", "-is_primary", 1000, 0, nil,
	)
	if err != nil {
		return err
	}

	entries := make([]RoutingEntry, 0, len(domains))
	for _, d := range domains {
		projectID := d.GetString("project_id")
		project, err := h.app.FindRecordById("projects", projectID)
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
