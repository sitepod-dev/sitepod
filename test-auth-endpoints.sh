#!/bin/bash
# SitePod Auth Endpoints Test Script
# Usage: ./test-auth-endpoints.sh [server_url]
# Example: ./test-auth-endpoints.sh http://localhost:8080

set -e

SERVER_URL=${1:-http://localhost:8080}
echo "Testing SitePod auth endpoints on $SERVER_URL"
echo "================================================"

# Test 1: Health check
echo "1. Testing health endpoint..."
curl -f "$SERVER_URL/api/v1/health" | jq . || {
    echo "❌ Health check failed"
    exit 1
}
echo "✅ Health check passed"
echo

# Test 2: Test incorrect path (should 404)
echo "2. Testing incorrect health path (should 404)..."
if curl -f "$SERVER_URL/v1/health" 2>/dev/null; then
    echo "❌ Expected 404 for /v1/health but got success"
    exit 1
else
    echo "✅ Correctly returned 404 for /v1/health"
fi
echo

# Test 3: Registration
echo "3. Testing user registration..."
TEST_EMAIL="test-$(date +%s)@example.com"
REGISTER_RESPONSE=$(curl -f -X POST -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"testpass123\",\"action\":\"register\"}" \
    "$SERVER_URL/api/v1/auth/login")

echo "$REGISTER_RESPONSE" | jq .
TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r .token)

if [[ "$TOKEN" == "null" || -z "$TOKEN" ]]; then
    echo "❌ Registration failed - no token received"
    exit 1
fi
echo "✅ Registration successful, token: ${TOKEN:0:20}..."
echo

# Test 4: Test protected endpoint without auth (should 401)
echo "4. Testing protected endpoint without auth (should 401)..."
if curl -f "$SERVER_URL/api/v1/projects" 2>/dev/null; then
    echo "❌ Expected 401 for /api/v1/projects without auth but got success"
    exit 1
else
    echo "✅ Correctly returned 401 for protected endpoint"
fi
echo

# Test 5: Test protected endpoint with auth
echo "5. Testing protected endpoint with auth..."
curl -f -H "Authorization: Bearer $TOKEN" "$SERVER_URL/api/v1/projects" | jq . || {
    echo "❌ Authenticated request failed"
    exit 1
}
echo "✅ Authenticated request successful"
echo

# Test 6: Test auth info endpoint
echo "6. Testing auth info endpoint..."
curl -f -H "Authorization: Bearer $TOKEN" "$SERVER_URL/api/v1/auth/info" | jq . || {
    echo "❌ Auth info request failed"
    exit 1
}
echo "✅ Auth info request successful"
echo

echo "================================================"
echo "🎉 All authentication tests passed!"
echo "The SitePod auth system is working correctly."