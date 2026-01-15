// Package caddy provides a Caddy module plugin entry point for SitePod.
// This package is meant to be imported by xcaddy to build a custom Caddy
// with the SitePod module included.
//
// Usage with xcaddy:
//
//	xcaddy build --with github.com/sitepod/sitepod/caddy
package caddy

import (
	// Register the SitePod Caddy module
	_ "github.com/sitepod/sitepod/internal/caddy"
)
