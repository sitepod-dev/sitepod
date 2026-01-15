#!/bin/bash
# Sync version across Cargo.toml and npm packages

set -e

if [ -z "$1" ]; then
    # Read version from Cargo.toml
    VERSION=$(grep '^version = ' cli/Cargo.toml | head -1 | sed 's/version = "\(.*\)"/\1/')
else
    VERSION="$1"
fi

echo "Syncing version: $VERSION"

# Update Cargo.toml
sed -i.bak "s/^version = \".*\"/version = \"$VERSION\"/" cli/Cargo.toml && rm cli/Cargo.toml.bak

# Update main npm package
jq --arg v "$VERSION" '.version = $v | .optionalDependencies |= with_entries(.value = $v)' \
    npm-packages/sitepod/package.json > tmp.json && mv tmp.json npm-packages/sitepod/package.json

# Update platform packages
for pkg in darwin-arm64 darwin-x64 linux-x64 linux-arm64 win32-x64; do
    jq --arg v "$VERSION" '.version = $v' \
        npm-packages/@sitepod/$pkg/package.json > tmp.json && mv tmp.json npm-packages/@sitepod/$pkg/package.json
done

# Update website version constant
sed -i.bak "s/export const VERSION = \".*\"/export const VERSION = \"$VERSION\"/" www/src/consts.ts && rm www/src/consts.ts.bak

echo "✓ All versions updated to $VERSION"
echo ""
echo "Updated files:"
echo "  - cli/Cargo.toml"
echo "  - npm-packages/sitepod/package.json"
echo "  - npm-packages/@sitepod/*/package.json"
echo "  - www/src/consts.ts"
