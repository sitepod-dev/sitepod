#!/bin/bash
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Config
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ENDPOINT="http://localhost:8080"
DATA_DIR="$SCRIPT_DIR/data"
SERVER_PID=""
ADMIN_TOKEN="test-admin-token"

cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"
    if [ -n "$SERVER_PID" ]; then
        kill $SERVER_PID 2>/dev/null || true
    fi
    rm -rf "$DATA_DIR"
    rm -rf /tmp/sitepod-test-project
    rm -f ~/.sitepod/config.toml 2>/dev/null || true
}

trap cleanup EXIT

fail() {
    echo -e "${RED}✗ $1${NC}"
    exit 1
}

pass() {
    echo -e "${GREEN}✓ $1${NC}"
}

info() {
    echo -e "${YELLOW}→ $1${NC}"
}

# Build if needed
if [ ! -f "$SCRIPT_DIR/bin/sitepod-server" ] || [ ! -f "$SCRIPT_DIR/bin/sitepod" ]; then
    info "Building..."
    cd "$SCRIPT_DIR"
    make build-server
    cd cli && cargo build --release && cp target/release/sitepod ../bin/
    cd "$SCRIPT_DIR"
fi

# Start server with IS_DEMO=1 for demo mode tests
# Also set SITEPOD_CONSOLE_ADMIN_* for console admin user tests
info "Starting server..."
rm -rf "$DATA_DIR"
mkdir -p "$DATA_DIR"
cd "$SCRIPT_DIR"
IS_DEMO=1 SITEPOD_ADMIN_TOKEN="$ADMIN_TOKEN" \
  SITEPOD_CONSOLE_ADMIN_EMAIL="console-admin@test.local" \
  SITEPOD_CONSOLE_ADMIN_PASSWORD="consoleadmin123" \
  "$SCRIPT_DIR/bin/sitepod-server" run --config server/Caddyfile.local > /tmp/sitepod-test.log 2>&1 &
SERVER_PID=$!
sleep 8

# Check server is running
if ! kill -0 $SERVER_PID 2>/dev/null; then
    cat /tmp/sitepod-test.log
    fail "Server failed to start"
fi
pass "Server started"

# Test health endpoint
info "Testing health endpoint..."
HEALTH=$(curl -s "$ENDPOINT/api/v1/health")
if echo "$HEALTH" | grep -q "ok"; then
    pass "Health check passed"
else
    fail "Health check failed: $HEALTH"
fi

# Test config endpoint (Demo mode)
info "Testing config endpoint..."
CONFIG=$(curl -s "$ENDPOINT/api/v1/config")
if echo "$CONFIG" | grep -q '"is_demo":true'; then
    pass "Config endpoint returns is_demo=true"
else
    fail "Config endpoint failed: $CONFIG"
fi
if echo "$CONFIG" | grep -q '"domain"'; then
    pass "Config endpoint returns domain"
else
    fail "Config missing domain: $CONFIG"
fi

# Test demo user login
info "Testing demo user login..."
DEMO_RESP=$(curl -s -X POST -H "Content-Type: application/json" -d '{"email":"demo@sitepod.dev","password":"demo123"}' "$ENDPOINT/api/v1/auth/login")
DEMO_TOKEN=$(echo "$DEMO_RESP" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
if [ -z "$DEMO_TOKEN" ]; then
    fail "Demo login failed: $DEMO_RESP"
fi
pass "Demo user login successful"

# Test demo user auth info
info "Testing demo user auth info..."
DEMO_INFO=$(curl -s -H "Authorization: Bearer $DEMO_TOKEN" "$ENDPOINT/api/v1/auth/info")
if echo "$DEMO_INFO" | grep -q '"is_admin":false'; then
    pass "Demo user is not admin"
else
    fail "Demo auth info failed: $DEMO_INFO"
fi

# Test console admin login (users.is_admin=true)
info "Testing console admin login..."
ADMIN_RESP=$(curl -s -X POST -H "Content-Type: application/json" -d '{"email":"console-admin@test.local","password":"consoleadmin123"}' "$ENDPOINT/api/v1/auth/login")
CONSOLE_ADMIN_TOKEN=$(echo "$ADMIN_RESP" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
if [ -z "$CONSOLE_ADMIN_TOKEN" ]; then
    fail "Console admin login failed: $ADMIN_RESP"
fi
pass "Console admin login successful"

# Test admin auth info
info "Testing admin auth info..."
ADMIN_INFO=$(curl -s -H "Authorization: Bearer $CONSOLE_ADMIN_TOKEN" "$ENDPOINT/api/v1/auth/info")
if echo "$ADMIN_INFO" | grep -q '"is_admin":true'; then
    pass "Admin auth info shows is_admin=true"
else
    fail "Admin auth info failed: $ADMIN_INFO"
fi

# Test admin can see all projects (including system ones)
info "Testing admin projects view..."
ADMIN_PROJECTS=$(curl -s -H "Authorization: Bearer $CONSOLE_ADMIN_TOKEN" "$ENDPOINT/api/v1/projects")
if echo "$ADMIN_PROJECTS" | grep -q '"owner_email"'; then
    pass "Admin can see owner_email in projects"
else
    fail "Admin projects missing owner_email: $ADMIN_PROJECTS"
fi
if echo "$ADMIN_PROJECTS" | grep -q '"console"'; then
    pass "Admin can see system project (console)"
else
    fail "Admin cannot see system projects: $ADMIN_PROJECTS"
fi

# Test email/password login (register or login)
info "Testing email/password login..."
AUTH_RESP=$(curl -s -X POST -H "Content-Type: application/json" -d '{"email":"test@example.com","password":"test123456"}' "$ENDPOINT/api/v1/auth/login")
TOKEN=$(echo "$AUTH_RESP" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
if [ -z "$TOKEN" ]; then
    fail "Email login failed: $AUTH_RESP"
fi
CREATED=$(echo "$AUTH_RESP" | grep -o '"created":true' || true)
if [ -n "$CREATED" ]; then
    pass "Account created and logged in"
else
    pass "Login successful"
fi

# Test welcome site
info "Testing welcome site..."
WELCOME=$(curl -s -o /dev/null -w "%{http_code}" "http://welcome.localhost:8080/")
if [ "$WELCOME" = "200" ]; then
    pass "Welcome site accessible"
else
    fail "Welcome site returned $WELCOME"
fi

# Test console site (now at root domain)
info "Testing console site..."
CONSOLE=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080/")
if [ "$CONSOLE" = "200" ]; then
    pass "Console site accessible"
else
    fail "Console site returned $CONSOLE"
fi

# Test projects API (should be empty for new user)
info "Testing projects API..."
PROJECTS=$(curl -s -H "Authorization: Bearer $TOKEN" "$ENDPOINT/api/v1/projects")
if [ "$PROJECTS" = "[]" ]; then
    pass "Projects API returns empty list for new user"
else
    fail "Projects API unexpected response: $PROJECTS"
fi

# Create test project
info "Creating test project..."
mkdir -p /tmp/sitepod-test-project/dist
echo "<html><body><h1>Test Site</h1></body></html>" > /tmp/sitepod-test-project/dist/index.html
echo "test content" > /tmp/sitepod-test-project/dist/test.txt
cd /tmp/sitepod-test-project

# Save token for CLI (CLI uses ~/.sitepod/config.toml)
mkdir -p ~/.sitepod
cat > ~/.sitepod/config.toml << EOF
[server]
endpoint = "$ENDPOINT"

[auth]
token = "$TOKEN"
EOF

# Create project config directly (init is interactive)
info "Creating project config..."
cat > sitepod.toml << EOF
[project]
name = "test-project"

[build]
directory = "./dist"

[deploy]
subdomain = "test-project"
EOF
pass "Project config created"

# Deploy
info "Testing sitepod deploy..."
$SCRIPT_DIR/bin/sitepod deploy > /tmp/sitepod-deploy.log 2>&1
if grep -q "Released to" /tmp/sitepod-deploy.log; then
    pass "Deploy successful"
else
    cat /tmp/sitepod-deploy.log
    fail "Deploy failed"
fi

# Verify deployed site (beta uses -beta suffix now)
info "Verifying deployed site..."
sleep 2
SITE=$(curl -s "http://test-project-beta.localhost:8080/")
if echo "$SITE" | grep -q "Test Site"; then
    pass "Deployed site accessible and correct"
else
    fail "Deployed site content wrong: $SITE"
fi

# Test projects API again (should have 1 project)
info "Verifying project in API..."
PROJECTS=$(curl -s -H "Authorization: Bearer $TOKEN" "$ENDPOINT/api/v1/projects")
if echo "$PROJECTS" | grep -q "test-project"; then
    pass "Project visible in API"
else
    fail "Project not in API: $PROJECTS"
fi

# Test images API
info "Testing images API..."
IMAGES=$(curl -s -H "Authorization: Bearer $TOKEN" "$ENDPOINT/api/v1/images?project=test-project")
if echo "$IMAGES" | grep -q "deployed_to"; then
    pass "Images API working"
else
    fail "Images API failed: $IMAGES"
fi

# Test history
info "Testing sitepod history..."
$SCRIPT_DIR/bin/sitepod history > /tmp/sitepod-history.log 2>&1
if grep -q "img_" /tmp/sitepod-history.log; then
    pass "History command working"
else
    cat /tmp/sitepod-history.log
    fail "History command failed"
fi

# Deploy again to test incremental
info "Testing incremental deploy..."
$SCRIPT_DIR/bin/sitepod deploy > /tmp/sitepod-deploy2.log 2>&1
if grep -q "reused" /tmp/sitepod-deploy2.log; then
    pass "Incremental deploy working (files reused)"
else
    cat /tmp/sitepod-deploy2.log
    fail "Incremental deploy failed"
fi

# Deploy to prod (use --yes to skip confirmation)
info "Testing deploy to prod..."
$SCRIPT_DIR/bin/sitepod deploy --prod --yes > /tmp/sitepod-deploy-prod.log 2>&1 || true
if grep -q "Released to" /tmp/sitepod-deploy-prod.log; then
    pass "Prod deploy successful"
else
    cat /tmp/sitepod-deploy-prod.log
    fail "Prod deploy failed"
fi

# Verify prod site
PROD_SITE=$(curl -s "http://test-project.localhost:8080/")
if echo "$PROD_SITE" | grep -q "Test Site"; then
    pass "Prod site accessible"
else
    fail "Prod site not accessible"
fi

# Test account deletion API
info "Testing account deletion..."
DELETE_RESP=$(curl -s -X DELETE -H "Authorization: Bearer $TOKEN" "$ENDPOINT/api/v1/account")
if echo "$DELETE_RESP" | grep -q "deleted_projects"; then
    pass "Account deletion API working"
else
    fail "Account deletion failed: $DELETE_RESP"
fi

# Verify projects are gone
PROJECTS_AFTER=$(curl -s -H "Authorization: Bearer $TOKEN" "$ENDPOINT/api/v1/projects")
if [ "$PROJECTS_AFTER" = '{"error":"authentication required"}' ]; then
    pass "User account fully deleted (auth fails)"
else
    fail "User still exists after deletion: $PROJECTS_AFTER"
fi

# Test cleanup API
info "Testing cleanup API..."
CLEANUP_RESP=$(curl -s -X POST -H "X-Sitepod-Admin-Token: $ADMIN_TOKEN" "$ENDPOINT/api/v1/cleanup")
if echo "$CLEANUP_RESP" | grep -q "expired_users_deleted"; then
    pass "Cleanup API working"
else
    fail "Cleanup API failed: $CLEANUP_RESP"
fi

# Test garbage collection API
info "Testing GC API..."
GC_RESP=$(curl -s -X POST -H "X-Sitepod-Admin-Token: $ADMIN_TOKEN" "$ENDPOINT/api/v1/gc")
if echo "$GC_RESP" | grep -q "deleted_blobs"; then
    pass "GC API working"
else
    fail "GC API failed: $GC_RESP"
fi

cd "$SCRIPT_DIR"

echo ""
echo -e "${GREEN}════════════════════════════════════════${NC}"
echo -e "${GREEN}  All E2E tests passed!${NC}"
echo -e "${GREEN}════════════════════════════════════════${NC}"
