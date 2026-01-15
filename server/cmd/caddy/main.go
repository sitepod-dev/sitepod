// Caddy server with SitePod module
package main

import (
	caddycmd "github.com/caddyserver/caddy/v2/cmd"

	// Plug in Caddy modules
	_ "github.com/caddyserver/caddy/v2/modules/standard"

	// Plug in SitePod module
	_ "github.com/sitepod/sitepod/internal/caddy"

	// Load PocketBase migrations
	_ "github.com/sitepod/sitepod/migrations"
)

func main() {
	caddycmd.Main()
}
