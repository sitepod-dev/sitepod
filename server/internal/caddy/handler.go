package caddy

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/models"
	"github.com/sitepod/sitepod/internal/gc"
	"github.com/sitepod/sitepod/internal/storage"
	"go.uber.org/zap"

	// Import migrations to register them
	_ "github.com/sitepod/sitepod/migrations"
)

// SitePodHandler implements the Caddy module for SitePod
// It handles both API requests and static file serving
type SitePodHandler struct {
	// Configuration
	StoragePath string `json:"storage_path,omitempty"`
	DataDir     string `json:"data_dir,omitempty"`
	Domain      string `json:"domain,omitempty"`
	CacheTTL    string `json:"cache_ttl,omitempty"`

	// Runtime
	storage      storage.Backend
	app          *pocketbase.PocketBase
	pbRouter     *echo.Echo
	cache        *refCache
	routingCache *routingCache
	cacheTTL     time.Duration
	gc           *gc.GC
	logger       *zap.Logger
	startTime    time.Time
}

// CaddyModule returns the Caddy module information
func (SitePodHandler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.sitepod",
		New: func() caddy.Module { return new(SitePodHandler) },
	}
}

// Provision sets up the handler
func (h *SitePodHandler) Provision(ctx caddy.Context) error {
	h.logger = ctx.Logger(h)
	h.startTime = time.Now()

	// Defaults
	if h.StoragePath == "" {
		h.StoragePath = "./data"
	}
	if h.DataDir == "" {
		h.DataDir = h.StoragePath
	}
	if h.Domain == "" {
		h.Domain = "localhost"
	}
	if h.CacheTTL == "" {
		h.CacheTTL = "5s"
	}

	// Parse cache TTL
	ttl, err := time.ParseDuration(h.CacheTTL)
	if err != nil {
		ttl = 5 * time.Second
	}
	h.cacheTTL = ttl

	// Initialize storage
	storageType := os.Getenv("SITEPOD_STORAGE_TYPE")
	if storageType == "" {
		storageType = "local"
	}

	switch storageType {
	case "s3", "oss", "r2":
		backend, err := storage.NewS3Backend(
			os.Getenv("SITEPOD_S3_BUCKET"),
			os.Getenv("SITEPOD_S3_REGION"),
			os.Getenv("SITEPOD_S3_ENDPOINT"),
		)
		if err != nil {
			return err
		}
		h.storage = backend
	default:
		backend, err := storage.NewLocalBackend(h.StoragePath)
		if err != nil {
			return err
		}
		h.storage = backend
	}

	// Initialize PocketBase (headless mode - no HTTP server)
	h.app = pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: h.DataDir,
	})

	// Bootstrap the app (initialize config and logger)
	if err := h.app.Bootstrap(); err != nil {
		return err
	}

	// Initialize database schema
	if err := h.initDatabaseSchema(); err != nil {
		return err
	}

	// Create default admin if none exists
	if err := h.ensureDefaultAdmin(); err != nil {
		h.logger.Warn("failed to create default admin", zap.Error(err))
	}

	// Create system user for internal projects
	if _, err := h.ensureSystemUser(); err != nil {
		h.logger.Warn("failed to create system user", zap.Error(err))
	}

	// Initialize caches
	h.cache = newRefCache(h.cacheTTL)
	h.routingCache = newRoutingCache(h.cacheTTL)

	// Start GC background worker
	h.gc = gc.New(h.app, h.storage, gc.DefaultConfig())
	go h.gc.Start(ctx.Context)

	// Get PocketBase router for forwarding requests
	pbRouter, err := apis.InitApi(h.app)
	if err != nil {
		return err
	}
	h.pbRouter = pbRouter

	// Print startup banner
	h.printStartupBanner()

	// Ensure system sites exist (welcome page, console)
	go h.ensureSystemSites()

	return nil
}

// Validate validates the handler configuration
func (h *SitePodHandler) Validate() error {
	return nil
}

// ServeHTTP handles all HTTP requests
func (h *SitePodHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	start := time.Now()
	path := r.URL.Path

	// API routes
	if strings.HasPrefix(path, "/api/v1/") {
		return h.handleAPI(w, r)
	}

	// PocketBase admin UI and API routes - forward to PocketBase router
	if strings.HasPrefix(path, "/_/") || strings.HasPrefix(path, "/api/") {
		h.pbRouter.ServeHTTP(w, r)
		return nil
	}

	// Static file serving with logging
	err := h.handleStatic(w, r)
	duration := time.Since(start)

	// Log static requests
	// - Always log errors and slow requests (>10ms) at DEBUG level
	// - Log all requests at INFO level if SITEPOD_ACCESS_LOG=1
	accessLog := os.Getenv("SITEPOD_ACCESS_LOG") == "1"
	status := 200
	if err != nil {
		status = 404
	}

	if accessLog {
		h.logger.Info("static",
			zap.String("host", r.Host),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("duration", duration),
		)
	} else if duration > 10*time.Millisecond || err != nil {
		h.logger.Debug("static",
			zap.String("host", r.Host),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("duration", duration),
		)
	}

	return err
}

// handleAPI routes API requests to the appropriate handler
func (h *SitePodHandler) handleAPI(w http.ResponseWriter, r *http.Request) error {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1")

	// Route matching - no auth required
	switch {
	// Health & Metrics
	case path == "/health" && r.Method == "GET":
		return h.apiHealth(w, r)
	case path == "/metrics" && r.Method == "GET":
		return h.apiMetrics(w, r)

	// Cleanup & GC (should be protected by firewall in production)
	case path == "/cleanup" && r.Method == "POST":
		return h.apiCleanup(w, r)
	case path == "/gc" && r.Method == "POST":
		return h.apiGarbageCollect(w, r)

	// Auth
	case path == "/auth/anonymous" && r.Method == "POST":
		return h.apiAnonymousAuth(w, r)

	// Domain check (called by Caddy for on-demand TLS)
	case path == "/domains/check" && r.Method == "GET":
		return h.apiCheckDomain(w, r)
	}

	// Auth required routes
	user, err := h.authenticate(r)
	if err != nil {
		return h.jsonError(w, http.StatusUnauthorized, "authentication required")
	}

	switch {
	// Auth
	case path == "/auth/bind" && r.Method == "POST":
		return h.apiBindEmail(w, r, user)

	// Account management
	case path == "/account" && r.Method == "DELETE":
		return h.apiDeleteAccount(w, r, user)

	// Projects
	case path == "/projects" && r.Method == "GET":
		return h.apiListProjects(w, r, user)
	case strings.HasPrefix(path, "/projects/") && r.Method == "GET":
		projectName := strings.TrimPrefix(path, "/projects/")
		return h.apiGetProject(w, r, projectName, user)

	// Subdomain check
	case path == "/subdomain/check" && r.Method == "GET":
		return h.apiCheckSubdomain(w, r)

	// Plan/Commit flow
	case path == "/plan" && r.Method == "POST":
		return h.apiPlan(w, r, user)
	case strings.HasPrefix(path, "/upload/") && r.Method == "POST":
		return h.apiUpload(w, r, path, user)
	case path == "/commit" && r.Method == "POST":
		return h.apiCommit(w, r, user)

	// Release/Rollback
	case path == "/release" && r.Method == "POST":
		return h.apiRelease(w, r, user)
	case path == "/rollback" && r.Method == "POST":
		return h.apiRollback(w, r, user)

	// Preview
	case path == "/preview" && r.Method == "POST":
		return h.apiCreatePreview(w, r, user)

	// Query
	case path == "/current" && r.Method == "GET":
		return h.apiGetCurrent(w, r, user)
	case path == "/history" && r.Method == "GET":
		return h.apiGetHistory(w, r, user)
	case path == "/images" && r.Method == "GET":
		return h.apiListImages(w, r, user)

	// Domain management
	case path == "/domains" && r.Method == "POST":
		return h.apiAddDomain(w, r, user)
	case path == "/domains" && r.Method == "GET":
		return h.apiListDomains(w, r, user)
	case strings.HasPrefix(path, "/domains/") && strings.HasSuffix(path, "/verify") && r.Method == "POST":
		domain := strings.TrimSuffix(strings.TrimPrefix(path, "/domains/"), "/verify")
		return h.apiVerifyDomain(w, r, domain, user)
	case strings.HasPrefix(path, "/domains/") && r.Method == "DELETE":
		domain := strings.TrimPrefix(path, "/domains/")
		return h.apiRemoveDomain(w, r, domain, user)
	case path == "/domains/rename" && r.Method == "PUT":
		return h.apiRenameDomain(w, r, user)

	// Admin routes
	case path == "/admin/cache/invalidate" && r.Method == "POST":
		return h.apiInvalidateCache(w, r)
	case path == "/admin/routing/rebuild" && r.Method == "POST":
		return h.apiRebuildRouting(w, r)

	default:
		return h.jsonError(w, http.StatusNotFound, "endpoint not found")
	}
}

// authenticate extracts and validates the auth token
func (h *SitePodHandler) authenticate(r *http.Request) (*models.Record, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return nil, errors.New("no authorization header")
	}

	token := strings.TrimPrefix(auth, "Bearer ")
	if token == auth {
		return nil, errors.New("invalid authorization format")
	}

	record, err := h.app.Dao().FindAuthRecordByToken(token, h.app.Settings().RecordAuthToken.Secret)
	if err != nil {
		return nil, err
	}

	if record.GetBool("is_anonymous") {
		expiresAt := record.GetDateTime("anonymous_expires_at")
		if !expiresAt.IsZero() && time.Now().After(expiresAt.Time()) {
			return nil, errors.New("anonymous session expired")
		}
	}

	return record, nil
}

// jsonResponse writes a JSON response
func (h *SitePodHandler) jsonResponse(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// jsonError writes a JSON error response
func (h *SitePodHandler) jsonError(w http.ResponseWriter, status int, message string) error {
	return h.jsonResponse(w, status, map[string]string{"error": message})
}
