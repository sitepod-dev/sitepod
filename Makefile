.PHONY: all build build-server build-cli run dev test lint lint-server lint-cli clean docker docker-push npm-prepare npm-link npm-publish bump-patch bump-minor bump-major release

# Default target
all: build

# Build everything (server + CLI)
build: build-server build-cli
	@echo "✓ Server and CLI built successfully"

# Build Caddy with embedded SitePod API
build-server:
	cd server && go build -ldflags="-s -w" -o ../bin/sitepod-server ./cmd/caddy

# Build Rust CLI
build-cli:
	cd cli && cargo build --release
	mkdir -p bin
	cp cli/target/release/sitepod bin/sitepod

# Alias for build-server
build-caddy: build-server

# Run server (single binary - Caddy with embedded API)
run:
	mkdir -p data
	SITEPOD_DATA_DIR=./data SITEPOD_DOMAIN=localhost:8080 \
	./bin/sitepod-server run --config server/Caddyfile.local

# Run with hot reload (requires air)
dev:
	cd server && air

# Run tests
test: test-server test-cli

test-server:
	cd server && go test ./...

test-cli:
	cd cli && cargo test

# Run linters
lint: lint-server lint-cli

lint-server:
	cd server && golangci-lint run --timeout=5m

lint-cli:
	cd cli && cargo fmt --check
	cd cli && cargo clippy -- -D warnings

# Clean build artifacts and data
clean:
	rm -rf bin/
	rm -rf server/data/
	rm -rf data/

# Docker commands
docker-build:
	docker build -t sitepod:latest .

docker-run:
	docker run -d --name sitepod \
		-p 80:80 -p 443:443 \
		-v sitepod-data:/data \
		-e SITEPOD_DOMAIN=localhost \
		sitepod:latest

docker-stop:
	docker rm -f sitepod 2>/dev/null || true

docker-logs:
	docker logs -f sitepod

# Build and push to ghcr.io (for linux/amd64)
docker-push:
	docker buildx build --platform linux/amd64 -t ghcr.io/sitepod-dev/sitepod:latest --push .

# Install dependencies
deps:
	cd server && go mod download
	cd cli && cargo fetch

# Create directories
init:
	mkdir -p bin data

# Install CLI globally
install-cli: build-cli
	@echo "Installing sitepod CLI to /usr/local/bin..."
	sudo cp bin/sitepod /usr/local/bin/sitepod
	@echo "Done! Run 'sitepod --help' to get started."

# Quick start
quick-start: init deps build
	@echo ""
	@echo "✓ SitePod is ready!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Start the server:     make run"
	@echo "  2. Login:                ./bin/sitepod login --endpoint http://localhost:8080"
	@echo "  3. Deploy example:       cd examples/simple-site && ../../bin/sitepod deploy"
	@echo "  4. Visit your site:      http://demo-site-beta.localhost:8080"
	@echo ""

# npm packages - prepare for local testing
npm-prepare: build-cli
	@echo "Preparing npm packages with local binary..."
	@# Detect current platform
	@PLATFORM=$$(uname -s)-$$(uname -m); \
	case $$PLATFORM in \
		Darwin-arm64) PKG=darwin-arm64 ;; \
		Darwin-x86_64) PKG=darwin-x64 ;; \
		Linux-x86_64) PKG=linux-x64 ;; \
		Linux-aarch64) PKG=linux-arm64 ;; \
		*) echo "Unsupported platform: $$PLATFORM"; exit 1 ;; \
	esac; \
	echo "Copying binary to @sitepod/$$PKG..."; \
	cp bin/sitepod npm-packages/@sitepod/$$PKG/bin/sitepod; \
	chmod +x npm-packages/@sitepod/$$PKG/bin/sitepod
	@echo "✓ npm packages ready for local testing"

# npm packages - link for local development
npm-link: npm-prepare
	@echo "Linking npm packages for local development..."
	@PLATFORM=$$(uname -s)-$$(uname -m); \
	case $$PLATFORM in \
		Darwin-arm64) PKG=darwin-arm64 ;; \
		Darwin-x86_64) PKG=darwin-x64 ;; \
		Linux-x86_64) PKG=linux-x64 ;; \
		Linux-aarch64) PKG=linux-arm64 ;; \
		*) echo "Unsupported platform: $$PLATFORM"; exit 1 ;; \
	esac; \
	cd npm-packages/@sitepod/$$PKG && npm link
	cd npm-packages/sitepod && npm link @sitepod/$$PKG && npm link
	@echo "✓ Run 'sitepod --help' to test the linked package"

# npm packages - publish to npm (local publish)
npm-publish: npm-prepare
	@echo "Publishing npm packages..."
	@# First publish platform-specific packages
	@for pkg in darwin-arm64 darwin-x64 linux-x64 linux-arm64 win32-x64; do \
		echo "Publishing @sitepod/$$pkg..."; \
		cd npm-packages/@sitepod/$$pkg && npm publish --access public || true; \
		cd ../../..; \
	done
	@# Then publish main package
	@echo "Publishing sitepod..."
	cd npm-packages/sitepod && npm publish --access public
	@echo "✓ All packages published to npm"

# Bump version (patch/minor/major)
bump-patch:
	@VERSION=$$(grep '^version = ' cli/Cargo.toml | head -1 | sed 's/version = "\(.*\)"/\1/'); \
	MAJOR=$$(echo $$VERSION | cut -d. -f1); \
	MINOR=$$(echo $$VERSION | cut -d. -f2); \
	PATCH=$$(echo $$VERSION | cut -d. -f3); \
	NEW_VERSION="$$MAJOR.$$MINOR.$$((PATCH + 1))"; \
	./scripts/sync-versions.sh $$NEW_VERSION

bump-minor:
	@VERSION=$$(grep '^version = ' cli/Cargo.toml | head -1 | sed 's/version = "\(.*\)"/\1/'); \
	MAJOR=$$(echo $$VERSION | cut -d. -f1); \
	MINOR=$$(echo $$VERSION | cut -d. -f2); \
	NEW_VERSION="$$MAJOR.$$((MINOR + 1)).0"; \
	./scripts/sync-versions.sh $$NEW_VERSION

bump-major:
	@VERSION=$$(grep '^version = ' cli/Cargo.toml | head -1 | sed 's/version = "\(.*\)"/\1/'); \
	MAJOR=$$(echo $$VERSION | cut -d. -f1); \
	NEW_VERSION="$$((MAJOR + 1)).0.0"; \
	./scripts/sync-versions.sh $$NEW_VERSION

# Create release tag and push (triggers GitHub Actions)
release:
	@VERSION=$$(grep '^version = ' cli/Cargo.toml | head -1 | sed 's/version = "\(.*\)"/\1/'); \
	echo "Creating release v$$VERSION..."; \
	git tag -a "v$$VERSION" -m "Release v$$VERSION"; \
	git push origin "v$$VERSION"; \
	echo "✓ Release v$$VERSION pushed. GitHub Actions will build and publish."
