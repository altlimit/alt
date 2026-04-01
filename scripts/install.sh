#!/bin/sh
# alt — One-line installer for Unix (Linux & macOS)
# Usage: curl -fsSL https://raw.githubusercontent.com/altlimit/alt/main/scripts/install.sh | sh

set -e

REPO="altlimit/alt"
INSTALL_DIR="$HOME/.local/share/alt/internal"
BIN_DIR="$HOME/.local/share/alt/bin"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
    linux)  OS="linux" ;;
    darwin) OS="darwin" ;;
    *)
        echo "Error: Unsupported operating system: $OS"
        echo "alt supports Linux and macOS. For Windows, use the PowerShell installer."
        exit 1
        ;;
esac

# Detect Architecture
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64)   ARCH="amd64" ;;
    aarch64|arm64)   ARCH="arm64" ;;
    *)
        echo "Error: Unsupported architecture: $ARCH"
        echo "alt supports amd64 and arm64."
        exit 1
        ;;
esac

echo "Detected: ${OS}/${ARCH}"

# Get latest release tag
RELEASE_URL="https://api.github.com/repos/${REPO}/releases/latest"
if [ -n "$GITHUB_TOKEN" ]; then
    API_RESPONSE=$(curl -sSL -H "Accept: application/vnd.github+json" -H "User-Agent: alt-cli/1.0" -H "Authorization: Bearer ${GITHUB_TOKEN}" "$RELEASE_URL" 2>&1) || true
else
    API_RESPONSE=$(curl -sSL -H "Accept: application/vnd.github+json" -H "User-Agent: alt-cli/1.0" "$RELEASE_URL" 2>&1) || true
fi
LATEST=$(echo "$API_RESPONSE" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
if [ -z "$LATEST" ]; then
    echo "Error: Could not determine latest release."
    case "$API_RESPONSE" in
        *rate*limit*|*rate*exceeded*|*403*)
            echo ""
            echo "GitHub API rate limit exceeded. Set GITHUB_TOKEN to increase your limit:"
            echo "  export GITHUB_TOKEN=your_token"
            ;;
        *Not*Found*|*404*)
            echo ""
            echo "No releases found for ${REPO}."
            echo "Check: https://github.com/${REPO}/releases"
            ;;
        *)
            echo ""
            echo "API response: $API_RESPONSE"
            ;;
    esac
    exit 1
fi

echo "Latest version: ${LATEST}"

# Build download URL
BINARY_NAME="alt_${OS}_${ARCH}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST}/${BINARY_NAME}"

# Create directories
mkdir -p "$INSTALL_DIR"
mkdir -p "$BIN_DIR"

# Download
echo "Downloading ${DOWNLOAD_URL}..."
if ! curl -fsSL "$DOWNLOAD_URL" -o "${INSTALL_DIR}/alt"; then
    echo ""
    echo "Error: Download failed."
    echo "Check that a release exists at: https://github.com/${REPO}/releases"
    exit 1
fi
chmod +x "${INSTALL_DIR}/alt"

# Create symlink in bin dir
ln -sf "${INSTALL_DIR}/alt" "${BIN_DIR}/alt"
echo "Installed alt to ${BIN_DIR}/alt"

# Update PATH in shell profile
update_profile() {
    profile="$1"
    if [ -f "$profile" ]; then
        if ! grep -q '\.local/share/alt/bin' "$profile" 2>/dev/null; then
            echo '' >> "$profile"
            echo '# Added by alt installer' >> "$profile"
            echo 'export PATH="$HOME/.local/share/alt/bin:$PATH"' >> "$profile"
            echo "Updated PATH in ${profile}"
            return 0
        fi
        return 1
    fi
    return 1
}

PATH_UPDATED=false
SHELL_NAME="$(basename "$SHELL" 2>/dev/null || echo "sh")"

case "$SHELL_NAME" in
    zsh)
        update_profile "$HOME/.zshrc" && PATH_UPDATED=true
        ;;
    bash)
        update_profile "$HOME/.bashrc" && PATH_UPDATED=true
        if [ "$PATH_UPDATED" = "false" ]; then
            update_profile "$HOME/.bash_profile" && PATH_UPDATED=true
        fi
        ;;
    *)
        update_profile "$HOME/.profile" && PATH_UPDATED=true
        ;;
esac

echo ""
echo "✓ alt ${LATEST} installed successfully!"
echo ""

if [ "$PATH_UPDATED" = "true" ]; then
    echo "Restart your shell or run:"
    echo "  export PATH=\"\$HOME/.local/share/alt/bin:\$PATH\""
else
    # Check if alt/bin is already in PATH
    case ":$PATH:" in
        *:$HOME/.local/share/alt/bin:*) ;;
        *)
            echo "Add this to your shell profile:"
            echo "  export PATH=\"\$HOME/.local/share/alt/bin:\$PATH\""
            ;;
    esac
fi
echo ""
echo "Get started: alt install user/repo"
