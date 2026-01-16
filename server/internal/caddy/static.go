package caddy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/sitepod/sitepod/internal/storage"
)

// handleStatic serves static files for deployed sites
func (h *SitePodHandler) handleStatic(w http.ResponseWriter, r *http.Request) error {
	host := r.Host
	path := r.URL.Path

	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	project, env, stripPath := h.resolveRouting(host, path)
	if project == "" {
		return caddyhttp.Error(http.StatusNotFound, errors.New("site not found"))
	}

	// Handle preview paths
	if stripPath == "" && strings.HasPrefix(path, "/__preview__/") {
		return h.servePreview(w, r, project)
	}

	// Handle path mode preview via cookie
	if stripPath != "" {
		if cookie, err := r.Cookie("sitepod_preview"); err == nil && cookie.Value != "" {
			return h.servePreviewBySlug(w, r, project, cookie.Value, path, stripPath)
		}
		if previewSlug := r.URL.Query().Get("preview"); previewSlug != "" {
			http.SetCookie(w, &http.Cookie{
				Name: "sitepod_preview", Value: previewSlug,
				Path: "/", MaxAge: 86400, HttpOnly: true,
			})
			http.Redirect(w, r, r.URL.Path, http.StatusFound)
			return nil
		}
		if envParam := r.URL.Query().Get("env"); envParam == "beta" {
			http.SetCookie(w, &http.Cookie{
				Name: "sitepod_env", Value: "beta",
				Path: "/", MaxAge: 86400, HttpOnly: true,
			})
			http.Redirect(w, r, r.URL.Path, http.StatusFound)
			return nil
		}
		if cookie, err := r.Cookie("sitepod_env"); err == nil && cookie.Value == "beta" {
			env = "beta"
		}
	}

	servePath := path
	if stripPath != "" && strings.HasPrefix(path, stripPath) {
		servePath = strings.TrimPrefix(path, stripPath)
		if servePath == "" {
			servePath = "/"
		}
	}

	return h.serveStatic(w, r, project, env, servePath)
}

// resolveRouting determines the project and environment from host and path
func (h *SitePodHandler) resolveRouting(host, path string) (string, string, string) {
	index := h.getRoutingIndex()
	if index != nil {
		var matches []RoutingEntry
		for _, route := range index.Entries {
			if route.Domain == host && strings.HasPrefix(path, route.Slug) {
				matches = append(matches, route)
			}
		}

		sort.Slice(matches, func(i, j int) bool {
			return len(matches[i].Slug) > len(matches[j].Slug)
		})

		if len(matches) > 0 {
			match := matches[0]
			stripPath := match.Slug
			if stripPath == "/" {
				stripPath = ""
			}
			env := match.Env
			if env == "" {
				env = "prod"
			}
			return match.Project, env, stripPath
		}
	}

	project, env := h.extractProjectAndEnv(host)
	return project, env, ""
}

// extractProjectAndEnv parses the host to determine project and environment
// Domain structure:
// - {domain} (root) → console project
// - {project}.{domain} → prod env
// - {project}-beta.{domain} → beta env
// - {project}--{slug}.{domain}/__preview__/{slug}/ → preview
func (h *SitePodHandler) extractProjectAndEnv(host string) (string, string) {
	// Strip the base domain to get subdomain
	baseDomain := h.Domain
	if !strings.HasSuffix(host, baseDomain) {
		// Host doesn't match our domain
		return "", ""
	}

	// Remove base domain and trailing dot
	subdomain := strings.TrimSuffix(host, baseDomain)
	subdomain = strings.TrimSuffix(subdomain, ".")

	// Root domain → console
	if subdomain == "" {
		return "console", "prod"
	}

	// Check for beta suffix: {project}-beta
	if strings.HasSuffix(subdomain, "-beta") {
		project := strings.TrimSuffix(subdomain, "-beta")
		return project, "beta"
	}

	// Check for preview: {project}--{slug}
	if idx := strings.Index(subdomain, "--"); idx > 0 {
		return subdomain[:idx], "preview"
	}

	// Default: {project} → prod
	return subdomain, "prod"
}

// getRoutingIndex retrieves the cached routing index
func (h *SitePodHandler) getRoutingIndex() *RoutingIndex {
	if cached, ok := h.routingCache.Get(); ok {
		return cached
	}

	data, err := h.storage.GetRouting()
	if err != nil {
		return nil
	}

	var index RoutingIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil
	}

	h.routingCache.Set(&index)
	return &index
}

// serveStatic serves a static file from a deployed environment
func (h *SitePodHandler) serveStatic(w http.ResponseWriter, r *http.Request, project, env, path string) error {
	ref, err := h.getRef(project, env)
	if err != nil {
		return caddyhttp.Error(http.StatusNotFound, err)
	}

	lookupPath := strings.TrimPrefix(path, "/")
	if lookupPath == "" || strings.HasSuffix(lookupPath, "/") {
		lookupPath += "index.html"
	}

	file, ok := ref.Manifest[lookupPath]
	if !ok {
		if !strings.Contains(lookupPath, ".") || strings.HasPrefix(lookupPath, "index") {
			if fallback, ok := ref.Manifest["index.html"]; ok {
				file = fallback
				lookupPath = "index.html"
			} else {
				return caddyhttp.Error(http.StatusNotFound, errors.New("file not found"))
			}
		} else {
			return caddyhttp.Error(http.StatusNotFound, errors.New("file not found"))
		}
	}

	return h.serveBlob(w, r, lookupPath, file)
}

// servePreview serves files from a preview deployment
func (h *SitePodHandler) servePreview(w http.ResponseWriter, r *http.Request, project string) error {
	pathParts := strings.TrimPrefix(r.URL.Path, "/__preview__/")
	parts := strings.SplitN(pathParts, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		return caddyhttp.Error(http.StatusNotFound, errors.New("invalid preview path"))
	}

	slug := parts[0]
	filePath := ""
	if len(parts) > 1 {
		filePath = parts[1]
	}

	return h.servePreviewBySlug(w, r, project, slug, filePath, "")
}

// servePreviewBySlug serves files from a preview deployment by slug
func (h *SitePodHandler) servePreviewBySlug(w http.ResponseWriter, r *http.Request, project, slug, filePath, stripPath string) error {
	if filePath == "" || strings.HasSuffix(filePath, "/") {
		filePath += "index.html"
	}

	if stripPath != "" && strings.HasPrefix(filePath, stripPath) {
		filePath = strings.TrimPrefix(filePath, stripPath)
		filePath = strings.TrimPrefix(filePath, "/")
	}

	previewData, err := h.storage.GetPreview(project, slug)
	if err != nil {
		return caddyhttp.Error(http.StatusNotFound, errors.New("preview not found"))
	}

	var preview storage.PreviewRef
	if err := json.Unmarshal(previewData, &preview); err != nil {
		return caddyhttp.Error(http.StatusInternalServerError, err)
	}

	if time.Now().After(preview.ExpiresAt) {
		_ = h.storage.DeletePreview(project, slug)
		return caddyhttp.Error(http.StatusGone, errors.New("preview expired"))
	}

	file, ok := preview.Manifest[filePath]
	if !ok {
		if fallback, ok := preview.Manifest["index.html"]; ok {
			file = fallback
			filePath = "index.html"
		} else {
			return caddyhttp.Error(http.StatusNotFound, errors.New("file not found in preview"))
		}
	}

	return h.serveBlob(w, r, filePath, file)
}

// getRef retrieves ref data for a project and environment
func (h *SitePodHandler) getRef(project, env string) (*storage.RefData, error) {
	cacheKey := project + ":" + env

	if cached, ok := h.cache.Get(cacheKey); ok {
		return cached, nil
	}

	data, err := h.storage.GetRef(project, env)
	if err != nil {
		return nil, err
	}

	var ref storage.RefData
	if err := json.Unmarshal(data, &ref); err != nil {
		return nil, err
	}

	h.cache.Set(cacheKey, &ref)
	return &ref, nil
}

// serveBlob serves a blob file with proper headers
func (h *SitePodHandler) serveBlob(w http.ResponseWriter, r *http.Request, path string, file storage.FileEntry) error {
	reader, err := h.storage.GetBlob(file.Hash)
	if err != nil {
		return caddyhttp.Error(http.StatusInternalServerError, err)
	}
	defer reader.Close()

	contentType := file.ContentType
	if contentType == "" {
		ext := filepath.Ext(path)
		contentType = mime.TypeByExtension(ext)
	}
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	etag := `"` + file.Hash[:16] + `"`
	w.Header().Set("ETag", etag)

	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return nil
	}

	if strings.HasPrefix(path, "assets/") || strings.HasPrefix(path, "_next/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=0, must-revalidate")
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")

	if file.Size > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Size))
	}

	_, err = io.Copy(w, reader)
	return err
}
