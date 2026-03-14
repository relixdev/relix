#!/bin/sh
# Relix installer — https://relix.sh/install
# Usage: curl -fsSL https://relix.sh/install | sh
set -e

REPO="relixdev/relix"
BINARY="relixctl"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
  linux|darwin) ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

echo "🐉 Installing Relix ($OS/$ARCH)..."

# Get latest release
LATEST=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "No releases found. Install from source:"
  echo "  git clone https://github.com/$REPO && cd relix/relixctl && go install ."
  exit 1
fi

# Download binary
URL="https://github.com/$REPO/releases/download/$LATEST/${BINARY}_${OS}_${ARCH}.tar.gz"
echo "Downloading $LATEST..."

TMPDIR=$(mktemp -d)
curl -fsSL "$URL" -o "$TMPDIR/relixctl.tar.gz"
tar -xzf "$TMPDIR/relixctl.tar.gz" -C "$TMPDIR"

# Install
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMPDIR/$BINARY" "$INSTALL_DIR/$BINARY"
else
  echo "Installing to $INSTALL_DIR (requires sudo)..."
  sudo mv "$TMPDIR/$BINARY" "$INSTALL_DIR/$BINARY"
fi

chmod +x "$INSTALL_DIR/$BINARY"
rm -rf "$TMPDIR"

echo ""
echo "✅ Relix installed successfully!"
echo ""
echo "Next steps:"
echo "  relixctl login    # Authenticate with GitHub"
echo "  relixctl pair     # Pair with the mobile app"
echo "  relixctl start    # Start the daemon"
echo ""
echo "Docs: https://relix.sh"
