package caddy

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/pocketbase/pocketbase/models"
)

// API: Add Domain
func (h *SitePodHandler) apiAddDomain(w http.ResponseWriter, r *http.Request, user *models.Record) error {
	var req struct {
		Project string `json:"project"`
		Domain  string `json:"domain"`
		Slug    string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return h.jsonError(w, http.StatusBadRequest, "invalid request")
	}

	if req.Domain == "" {
		return h.jsonError(w, http.StatusBadRequest, "domain required")
	}

	domain := strings.ToLower(strings.TrimSpace(req.Domain))
	slug := req.Slug
	if slug == "" {
		slug = "/"
	}
	if !strings.HasPrefix(slug, "/") {
		slug = "/" + slug
	}

	project, err := h.requireProjectOwnerByName(req.Project, user)
	if err != nil {
		if errors.Is(err, errForbidden) {
			return h.jsonError(w, http.StatusForbidden, "forbidden")
		}
		return h.jsonError(w, http.StatusNotFound, "project not found")
	}

	existing, _ := h.app.Dao().FindFirstRecordByFilter(
		"domains", "domain = {:domain} AND slug = {:slug}",
		map[string]any{"domain": domain, "slug": slug},
	)
	if existing != nil {
		return h.jsonError(w, http.StatusConflict, "domain+slug combination already exists")
	}

	baseDomain := h.Domain
	domainType := "custom"
	status := "pending"
	verificationToken := ""

	if strings.HasSuffix(domain, "."+baseDomain) || domain == baseDomain {
		domainType = "system"
		status = "active"
	} else {
		verificationToken = "sitepod-verify-" + uuid.New().String()[:16]
	}

	domainsCollection, err := h.app.Dao().FindCollectionByNameOrId("domains")
	if err != nil {
		return h.jsonError(w, http.StatusInternalServerError, "domains collection not found")
	}

	domainRecord := models.NewRecord(domainsCollection)
	domainRecord.Set("domain", domain)
	domainRecord.Set("slug", slug)
	domainRecord.Set("project_id", project.Id)
	domainRecord.Set("type", domainType)
	domainRecord.Set("status", status)
	domainRecord.Set("verification_token", verificationToken)
	domainRecord.Set("is_primary", false)

	if err := h.app.Dao().SaveRecord(domainRecord); err != nil {
		return h.jsonError(w, http.StatusInternalServerError, "failed to create domain")
	}

	if status == "active" {
		if err := h.rebuildRoutingIndex(); err != nil {
			return h.jsonError(w, http.StatusInternalServerError, "failed to rebuild routing index")
		}
	}

	resp := map[string]any{
		"domain": domain,
		"slug":   slug,
		"status": status,
	}

	if verificationToken != "" {
		resp["verification_token"] = verificationToken
		resp["verification_txt"] = fmt.Sprintf("_sitepod.%s TXT \"%s\"", domain, verificationToken)
	}

	return h.jsonResponse(w, http.StatusOK, resp)
}

// API: List Domains
func (h *SitePodHandler) apiListDomains(w http.ResponseWriter, r *http.Request, user *models.Record) error {
	projectName := r.URL.Query().Get("project")
	if projectName == "" {
		return h.jsonError(w, http.StatusBadRequest, "project required")
	}

	project, err := h.requireProjectOwnerByName(projectName, user)
	if err != nil {
		if errors.Is(err, errForbidden) {
			return h.jsonError(w, http.StatusForbidden, "forbidden")
		}
		return h.jsonError(w, http.StatusNotFound, "project not found")
	}

	domains, err := h.app.Dao().FindRecordsByFilter(
		"domains", "project_id = {:project_id}", "-is_primary,-created", 100, 0,
		map[string]any{"project_id": project.Id},
	)
	if err != nil {
		return h.jsonError(w, http.StatusInternalServerError, "failed to list domains")
	}

	result := make([]map[string]any, len(domains))
	for i, d := range domains {
		result[i] = map[string]any{
			"domain":     d.GetString("domain"),
			"slug":       d.GetString("slug"),
			"type":       d.GetString("type"),
			"status":     d.GetString("status"),
			"is_primary": d.GetBool("is_primary"),
			"created_at": d.Created.String(),
		}
	}

	return h.jsonResponse(w, http.StatusOK, map[string]any{"domains": result})
}

// API: Verify Domain
func (h *SitePodHandler) apiVerifyDomain(w http.ResponseWriter, r *http.Request, domain string, user *models.Record) error {
	domainRecord, _, err := h.requireDomainOwner(domain, user)
	if err != nil {
		if errors.Is(err, errForbidden) {
			return h.jsonError(w, http.StatusForbidden, "forbidden")
		}
		return h.jsonError(w, http.StatusNotFound, "domain not found")
	}

	if domainRecord.GetString("status") == "active" {
		return h.jsonResponse(w, http.StatusOK, map[string]any{
			"domain":   domain,
			"status":   "active",
			"verified": true,
			"message":  "Domain is already verified and active",
		})
	}

	verificationToken := domainRecord.GetString("verification_token")
	if verificationToken == "" {
		return h.jsonError(w, http.StatusBadRequest, "no verification token for this domain")
	}

	// DNS TXT lookup
	txtRecords, err := net.LookupTXT("_sitepod." + domain)
	verified := false
	if err == nil {
		for _, txt := range txtRecords {
			if txt == verificationToken {
				verified = true
				break
			}
		}
	}

	if verified {
		domainRecord.Set("status", "active")
		if err := h.app.Dao().SaveRecord(domainRecord); err != nil {
			return h.jsonError(w, http.StatusInternalServerError, "failed to update domain")
		}
		if err := h.rebuildRoutingIndex(); err != nil {
			return h.jsonError(w, http.StatusInternalServerError, "failed to rebuild routing index")
		}

		return h.jsonResponse(w, http.StatusOK, map[string]any{
			"domain":   domain,
			"status":   "active",
			"verified": true,
			"message":  "Domain verified successfully",
		})
	}

	return h.jsonResponse(w, http.StatusOK, map[string]any{
		"domain":   domain,
		"status":   "pending",
		"verified": false,
		"message":  fmt.Sprintf("DNS TXT record not found. Please add: _sitepod.%s TXT \"%s\"", domain, verificationToken),
	})
}

// API: Remove Domain
func (h *SitePodHandler) apiRemoveDomain(w http.ResponseWriter, r *http.Request, domain string, user *models.Record) error {
	domainRecord, _, err := h.requireDomainOwner(domain, user)
	if err != nil {
		if errors.Is(err, errForbidden) {
			return h.jsonError(w, http.StatusForbidden, "forbidden")
		}
		return h.jsonError(w, http.StatusNotFound, "domain not found")
	}

	if domainRecord.GetBool("is_primary") && domainRecord.GetString("type") == "system" {
		return h.jsonError(w, http.StatusBadRequest, "cannot remove primary system domain")
	}

	if err := h.app.Dao().DeleteRecord(domainRecord); err != nil {
		return h.jsonError(w, http.StatusInternalServerError, "failed to remove domain")
	}

	if err := h.rebuildRoutingIndex(); err != nil {
		return h.jsonError(w, http.StatusInternalServerError, "failed to rebuild routing index")
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

// API: Rename Domain
func (h *SitePodHandler) apiRenameDomain(w http.ResponseWriter, r *http.Request, user *models.Record) error {
	projectName := r.URL.Query().Get("project")
	if projectName == "" {
		return h.jsonError(w, http.StatusBadRequest, "project required")
	}

	var req struct {
		NewSubdomain string `json:"new_subdomain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return h.jsonError(w, http.StatusBadRequest, "invalid request")
	}

	newSubdomain := h.normalizeSubdomain(req.NewSubdomain)
	if newSubdomain == "" {
		return h.jsonError(w, http.StatusBadRequest, "valid subdomain required")
	}

	existing, _ := h.app.Dao().FindFirstRecordByData("projects", "subdomain", newSubdomain)
	if existing != nil {
		return h.jsonError(w, http.StatusConflict, "subdomain already in use")
	}

	project, err := h.requireProjectOwnerByName(projectName, user)
	if err != nil {
		if errors.Is(err, errForbidden) {
			return h.jsonError(w, http.StatusForbidden, "forbidden")
		}
		return h.jsonError(w, http.StatusNotFound, "project not found")
	}

	oldSubdomain := project.GetString("subdomain")
	project.Set("subdomain", newSubdomain)
	if err := h.app.Dao().SaveRecord(project); err != nil {
		return h.jsonError(w, http.StatusInternalServerError, "failed to update project")
	}

	// Update system domain
	oldFullDomain := oldSubdomain + "." + h.Domain
	newFullDomain := newSubdomain + "." + h.Domain

	domainRecord, _ := h.app.Dao().FindFirstRecordByFilter(
		"domains", "project_id = {:project_id} AND is_primary = true AND type = 'system'",
		map[string]any{"project_id": project.Id},
	)
	if domainRecord != nil {
		domainRecord.Set("domain", newFullDomain)
		if err := h.app.Dao().SaveRecord(domainRecord); err != nil {
			return h.jsonError(w, http.StatusInternalServerError, "failed to update domain")
		}
	}

	if err := h.rebuildRoutingIndex(); err != nil {
		return h.jsonError(w, http.StatusInternalServerError, "failed to rebuild routing index")
	}

	return h.jsonResponse(w, http.StatusOK, map[string]any{
		"old_domain": oldFullDomain,
		"new_domain": newFullDomain,
		"message":    "Subdomain renamed successfully",
	})
}

// API: Check Subdomain
func (h *SitePodHandler) apiCheckSubdomain(w http.ResponseWriter, r *http.Request) error {
	subdomain := r.URL.Query().Get("subdomain")
	if subdomain == "" {
		return h.jsonError(w, http.StatusBadRequest, "subdomain required")
	}

	subdomain = strings.ToLower(subdomain)

	existing, _ := h.app.Dao().FindFirstRecordByData("projects", "subdomain", subdomain)
	if existing != nil {
		suggestion := subdomain + "-" + uuid.New().String()[:4]
		return h.jsonResponse(w, http.StatusOK, map[string]any{
			"available":  false,
			"suggestion": suggestion,
		})
	}

	return h.jsonResponse(w, http.StatusOK, map[string]any{
		"available": true,
	})
}

// API: Check Domain (for on-demand TLS)
func (h *SitePodHandler) apiCheckDomain(w http.ResponseWriter, r *http.Request) error {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		return h.jsonError(w, http.StatusBadRequest, "domain required")
	}

	// Check if domain exists in domains table
	existing, _ := h.app.Dao().FindFirstRecordByData("domains", "domain", domain)
	if existing != nil && existing.GetString("status") == "active" {
		w.WriteHeader(http.StatusOK)
		return nil
	}

	return h.jsonError(w, http.StatusNotFound, "domain not allowed")
}
