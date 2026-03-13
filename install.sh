#!/usr/bin/env bash
# Install scli from GitHub releases and ensure 'scli' is on PATH.
# Usage: curl -sSL https://raw.githubusercontent.com/sessiondb/scli/main/install.sh | bash
#    or: curl -sSL ... | bash -s -- v1.0.0
set -e

REPO="sessiondb/scli"
VERSION="${1:-latest}"
INSTALL_DIR=""
BINARY_NAME="scli"

# Resolve latest to the newest release tag
if [ "$VERSION" = "latest" ]; then
  VERSION=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p')
  [ -n "$VERSION" ] || { echo "Could not resolve latest release"; exit 1; }
fi

# Normalize version (strip v if present for URL)
TAG="$VERSION"
[ "${TAG#v}" != "$TAG" ] || TAG="v${TAG}"

# Detect OS and arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$OS" in
  darwin) OS=darwin ;;
  linux)  OS=linux ;;
  mingw*|msys*|cygwin*) OS=windows ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac
case "$ARCH" in
  x86_64|amd64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *) echo "Unsupported arch: $ARCH"; exit 1 ;;
esac

# Windows: .exe suffix
SUFFIX=""
[ "$OS" = "windows" ] && SUFFIX=".exe"

FILE="scli-${TAG}-${OS}-${ARCH}${SUFFIX}"
URL="https://github.com/${REPO}/releases/download/${TAG}/${FILE}"

echo "Installing scli ${TAG} (${OS}/${ARCH})..."
echo "  ${URL}"

# Prefer /usr/local/bin; fallback to $HOME/.local/bin
if [ -w /usr/local/bin ] 2>/dev/null; then
  INSTALL_DIR="/usr/local/bin"
elif [ -n "$HOME" ] && [ -w "$HOME" ]; then
  INSTALL_DIR="${HOME}/.local/bin"
  mkdir -p "$INSTALL_DIR"
else
  echo "No writable install directory (tried /usr/local/bin and \$HOME/.local/bin)"; exit 1
fi

# Download (follow redirects)
if command -v curl >/dev/null 2>&1; then
  curl -sSL -o "${INSTALL_DIR}/scli${SUFFIX}.tmp" "$URL"
elif command -v wget >/dev/null 2>&1; then
  wget -q -O "${INSTALL_DIR}/scli${SUFFIX}.tmp" "$URL"
else
  echo "Need curl or wget"; exit 1
fi

chmod +x "${INSTALL_DIR}/scli${SUFFIX}.tmp"
mv "${INSTALL_DIR}/scli${SUFFIX}.tmp" "${INSTALL_DIR}/${BINARY_NAME}${SUFFIX}"

echo ""
echo "Installed to: ${INSTALL_DIR}/${BINARY_NAME}${SUFFIX}"
echo ""

# Ensure install dir is on PATH
add_path_line() {
  local file="$1"
  local line="$2"
  [ -f "$file" ] || touch "$file"
  if grep -qF "$line" "$file" 2>/dev/null; then
    return
  fi
  echo "" >> "$file"
  echo "# scli" >> "$file"
  echo "$line" >> "$file"
  echo "Added to $file — run 'source $file' or open a new terminal"
}

if ! command -v scli >/dev/null 2>&1; then
  if [ "$INSTALL_DIR" = "${HOME}/.local/bin" ]; then
    SHELL_RC=""
    [ -f "${HOME}/.zshrc" ] && SHELL_RC="${HOME}/.zshrc"
    [ -f "${HOME}/.bashrc" ] && SHELL_RC="${HOME}/.bashrc"
    [ -z "$SHELL_RC" ] && [ -f "${HOME}/.profile" ] && SHELL_RC="${HOME}/.profile"
    if [ -n "$SHELL_RC" ]; then
      add_path_line "$SHELL_RC" 'export PATH="$PATH:$HOME/.local/bin"'
    else
      echo "Add to your PATH: export PATH=\"\$PATH:${INSTALL_DIR}\""
    fi
  else
    echo "Ensure ${INSTALL_DIR} is on your PATH:"
    echo "  export PATH=\"\$PATH:${INSTALL_DIR}\""
  fi
else
  echo "scli is on your PATH. Run: scli --help  # or scli init"
fi
