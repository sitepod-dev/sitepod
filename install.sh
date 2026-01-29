#!/bin/bash
# SitePod CLI Installation Script
# Usage: curl -fsSL https://get.sitepod.dev | sh

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
GITHUB_REPO="sitepod-dev/sitepod"
BINARY_NAME="sitepod"
INSTALL_DIR="${SITEPOD_INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        mingw*|msys*|cygwin*)
            OS="windows"
            ;;
        *)
            echo -e "${RED}Unsupported operating system: $OS${NC}"
            exit 1
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            echo -e "${RED}Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
}

# Get latest version from GitHub
get_latest_version() {
    VERSION=$(curl -fsSL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$VERSION" ]; then
        echo -e "${YELLOW}Could not determine latest version, using 'latest'${NC}"
        VERSION="latest"
    fi
}

# Download and install
install() {
    echo -e "${GREEN}SitePod CLI Installer${NC}"
    echo ""

    detect_platform
    echo "Detected platform: $PLATFORM"

    get_latest_version
    echo "Installing version: $VERSION"
    echo ""

    # Construct download URL
    if [ "$VERSION" = "latest" ]; then
        DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/latest/download/${BINARY_NAME}-${PLATFORM}"
    else
        DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${BINARY_NAME}-${PLATFORM}"
    fi

    # Add .exe for Windows
    if [ "$OS" = "windows" ]; then
        DOWNLOAD_URL="${DOWNLOAD_URL}.exe"
        BINARY_NAME="${BINARY_NAME}.exe"
    fi

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    TMP_FILE="${TMP_DIR}/${BINARY_NAME}"

    echo "Downloading from: $DOWNLOAD_URL"

    # Download
    if command -v curl &> /dev/null; then
        curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE"
    elif command -v wget &> /dev/null; then
        wget -q "$DOWNLOAD_URL" -O "$TMP_FILE"
    else
        echo -e "${RED}Neither curl nor wget found. Please install one of them.${NC}"
        exit 1
    fi

    # Make executable
    chmod +x "$TMP_FILE"

    # Install
    echo ""
    echo "Installing to: ${INSTALL_DIR}/${BINARY_NAME}"

    if [ -w "$INSTALL_DIR" ]; then
        mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        echo "Requesting sudo access to install to $INSTALL_DIR"
        sudo mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    # Cleanup
    rm -rf "$TMP_DIR"

    # Verify installation
    if command -v sitepod &> /dev/null; then
        echo ""
        echo -e "${GREEN}✓ SitePod CLI installed successfully!${NC}"
        echo ""
        sitepod --version
        echo ""
        echo "Get started:"
        echo "  sitepod login       # Login to your server"
        echo "  sitepod init        # Initialize a project"
        echo "  sitepod deploy      # Deploy your site"
    else
        echo ""
        echo -e "${YELLOW}Installation complete, but 'sitepod' command not found in PATH.${NC}"
        echo "You may need to add ${INSTALL_DIR} to your PATH:"
        echo ""
        echo "  export PATH=\"\$PATH:${INSTALL_DIR}\""
    fi
}

# Run installation
install
