package caddy

import (
	"fmt"
	"os"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	caddy.RegisterModule(SitePodHandler{})
	httpcaddyfile.RegisterHandlerDirective("sitepod", parseCaddyfile)
}

// UnmarshalCaddyfile parses the Caddyfile configuration for the SitePod handler
func (h *SitePodHandler) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for d.NextBlock(0) {
			switch d.Val() {
			case "storage_path":
				if !d.NextArg() {
					return d.ArgErr()
				}
				h.StoragePath = d.Val()
			case "data_dir":
				if !d.NextArg() {
					return d.ArgErr()
				}
				h.DataDir = d.Val()
			case "cache_ttl":
				if !d.NextArg() {
					return d.ArgErr()
				}
				h.CacheTTL = d.Val()
			case "domain":
				if !d.NextArg() {
					return d.ArgErr()
				}
				h.Domain = d.Val()
			default:
				return d.Errf("unrecognized subdirective: %s", d.Val())
			}
		}
	}
	return nil
}

// printStartupBanner prints helpful information after server starts
func (h *SitePodHandler) printStartupBanner() {
	scheme := "https"
	isLocal := h.Domain == "localhost" || strings.HasPrefix(h.Domain, "localhost:") || strings.HasPrefix(h.Domain, "127.0.0.1")
	if isLocal {
		scheme = "http"
	}

	// Determine storage type from environment
	storageType := os.Getenv("SITEPOD_STORAGE_TYPE")
	if storageType == "" {
		storageType = "local"
	}

	baseURL := fmt.Sprintf("%s://%s", scheme, h.Domain)

	fmt.Println()
	fmt.Println("===============================================================")
	fmt.Println("                    SitePod is running!                        ")
	fmt.Println("===============================================================")
	fmt.Println("  Configuration:                                               ")
	fmt.Printf("    Domain:    %s\n", h.Domain)
	fmt.Printf("    Storage:   %s\n", storageType)
	fmt.Printf("    Data Dir:  %s\n", h.DataDir)
	fmt.Println("---------------------------------------------------------------")
	fmt.Println("  Endpoints:                                                   ")
	fmt.Printf("    Console:   %s\n", fmt.Sprintf("%s://%s", scheme, h.Domain))
	fmt.Printf("    Welcome:   %s\n", fmt.Sprintf("%s://welcome.%s", scheme, h.Domain))
	fmt.Printf("    Health:    %s\n", baseURL+"/api/v1/health")
	fmt.Println("---------------------------------------------------------------")
	// Show PocketBase Admin credentials
	adminEmail := os.Getenv("SITEPOD_ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@sitepod.local"
	}
	adminPassword := os.Getenv("SITEPOD_ADMIN_PASSWORD")
	logAdminPassword := os.Getenv("SITEPOD_LOG_ADMIN_PASSWORD") == "1"
	passwordDisplay := "(set via SITEPOD_ADMIN_PASSWORD)"
	if adminPassword == "" {
		passwordDisplay = "(default password set; change SITEPOD_ADMIN_PASSWORD)"
		if logAdminPassword {
			adminPassword = "sitepod123"
			passwordDisplay = adminPassword
		}
	} else if logAdminPassword {
		passwordDisplay = adminPassword
	}
	fmt.Println("  Admin Credentials:                                           ")
	fmt.Printf("    Email:     %s\n", adminEmail)
	fmt.Printf("    Password:  %s\n", passwordDisplay)
	if adminPassword == "" {
		fmt.Println("    WARNING:   DEFAULT ADMIN PASSWORD IN USE")
		fmt.Println("               Set SITEPOD_ADMIN_PASSWORD to change it")
		fmt.Println("               (set SITEPOD_LOG_ADMIN_PASSWORD=1 to log it here)")
	}
	fmt.Println("---------------------------------------------------------------")
	fmt.Println("  CLI Quick Start:                                             ")
	fmt.Printf("    sitepod login --endpoint %s\n", baseURL)
	fmt.Println("    sitepod deploy                                             ")
	fmt.Println("===============================================================")

	// Show DNS hint for local development
	if isLocal && strings.Contains(h.Domain, "localhost") {
		fmt.Println()
		fmt.Println("Note: Wildcard subdomains (*.localhost) may require DNS config:")
		fmt.Println("   macOS:  Works out of the box")
		fmt.Println("   Linux:  Add to /etc/hosts: 127.0.0.1 welcome.localhost")
		fmt.Println("           Or use dnsmasq / systemd-resolved for wildcard support")
	}
	fmt.Println()
}

// parseCaddyfile is the Caddyfile directive parser for the SitePod handler
func parseCaddyfile(helper httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var handler SitePodHandler
	err := handler.UnmarshalCaddyfile(helper.Dispenser)
	return &handler, err
}

// Interface guards
var (
	_ caddyhttp.MiddlewareHandler = (*SitePodHandler)(nil)
	_ caddy.Provisioner           = (*SitePodHandler)(nil)
	_ caddy.Validator             = (*SitePodHandler)(nil)
	_ caddyfile.Unmarshaler       = (*SitePodHandler)(nil)
)
