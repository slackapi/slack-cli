#!/bin/bash
# Install the latest Slack CLI pre-release from GitHub

set -e

RELEASE_TAG="${1:-latest}"
CLI_NAME="${2:-slack-prerelease}"

# Get the latest pre-release if "latest" is specified
if [ "$RELEASE_TAG" = "latest" ]; then
	RELEASE_TAG=$(curl -s https://api.github.com/repos/slackapi/slack-cli/releases | grep -m 1 '"tag_name"' | grep -v 'v[0-9]*\.[0-9]*\.[0-9]*"' | cut -d'"' -f4)
fi

# Determine platform and architecture
if [ "$(uname)" = "Darwin" ]; then
	ARCH=$(uname -m)
	[ "$ARCH" = "arm64" ] && ARCH="arm64" || ARCH="amd64"
	PLATFORM="macOS"
else
	PLATFORM="linux"
	ARCH="64-bit"
fi

# Download and extract
INSTALL_DIR="$HOME/.slack/prerelease"
mkdir -p "$INSTALL_DIR"

echo "📥 Downloading $RELEASE_TAG for $PLATFORM $ARCH..."
curl -fsSL "https://github.com/slackapi/slack-cli/releases/download/$RELEASE_TAG/slack_cli_*_${PLATFORM}_${ARCH}.tar.gz" -o "$INSTALL_DIR/slack-cli.tar.gz" 2>/dev/null || \
	curl -fsSL "https://github.com/slackapi/slack-cli/releases/download/$RELEASE_TAG/slack_cli_*_${PLATFORM}_64-bit.tar.gz" -o "$INSTALL_DIR/slack-cli.tar.gz"

tar -xzf "$INSTALL_DIR/slack-cli.tar.gz" -C "$INSTALL_DIR"
chmod +x "$INSTALL_DIR/bin/slack"
rm "$INSTALL_DIR/slack-cli.tar.gz"

# Create symlink
mkdir -p "$HOME/.local/bin"
ln -sf "$INSTALL_DIR/bin/slack" "$HOME/.local/bin/$CLI_NAME"

echo "✅ Installed $CLI_NAME"
echo "🔍 Version: $($HOME/.local/bin/$CLI_NAME --version)"
