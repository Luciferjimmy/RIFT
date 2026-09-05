#!/usr/bin/env bash
set -e

# ==============================================================================
# RIFT — Consolidated Translation Engine Installer for macOS
# Automatically downloads, installs to Applications, and strips Gatekeeper quarantine.
# ==============================================================================

REPO="Luciferjimmy/RIFT"
APP_NAME="RIFT.app"

# Determine installation directory: /Applications if writable, else $HOME/Applications
if [ -w "/Applications" ]; then
    INSTALL_DIR="/Applications"
else
    INSTALL_DIR="$HOME/Applications"
    mkdir -p "$INSTALL_DIR"
fi
TARGET_APP="$INSTALL_DIR/$APP_NAME"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

echo -e "${CYAN}${BOLD}"
echo "    ____  ________  ______"
echo "   / __ \/  _/ __/_/_  __/"
echo "  / /_/ // // /_    / /   "
echo " / _, _// // __/   / /    "
echo "/_/ |_/___/_/     /_/     "
echo -e "Consolidated Translation Engine for Apple Silicon${NC}\n"

# 1. Architecture & OS Verification
OS="$(uname -s)"
ARCH="$(uname -m)"

if [ "$OS" != "Darwin" ]; then
    echo -e "${RED}[!] RIFT is exclusively built for macOS.${NC}"
    exit 1
fi

if [ "$ARCH" != "arm64" ]; then
    echo -e "${RED}[!] Warning: RIFT translation pipelines are optimized for Apple Silicon (M1/M2/M3/M4). Intel Macs are not supported.${NC}"
    exit 1
fi

# 2. Automated Rosetta 2 Verification & Installation
if [ ! -f "/Library/Apple/usr/libexec/oah/libRosettaRuntime" ]; then
    echo -e "${BLUE}[*] Rosetta 2 runtime missing. Automatically installing for x86 Windows translation...${NC}"
    softwareupdate --install-rosetta --agree-to-license 2>/dev/null || true
    if [ -f "/Library/Apple/usr/libexec/oah/libRosettaRuntime" ]; then
        echo -e "${GREEN}[✓] Rosetta 2 successfully installed!${NC}"
    else
        echo -e "${CYAN}[*] Attempting privileged Rosetta 2 install...${NC}"
        osascript -e 'do shell script "softwareupdate --install-rosetta --agree-to-license" with administrator privileges' 2>/dev/null || true
    fi
else
    echo -e "${GREEN}[✓] Rosetta 2 runtime detected.${NC}"
fi

echo -e "${BLUE}[*] Fetching latest RIFT release from GitHub ($REPO)...${NC}"

# 3. Get latest release download URL
LATEST_URL=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep -o 'https://[^"]*RIFT[^"]*\.dmg' | head -n 1 || true)

if [ -z "$LATEST_URL" ]; then
    # Fallback to general zip if DMG not found
    LATEST_URL=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep -o 'https://[^"]*RIFT[^"]*\.zip' | head -n 1 || true)
fi

TMP_DIR="$(mktemp -d)"
MOUNT_DIR=""

cleanup() {
    local exit_code=$?
    if [ -n "$MOUNT_DIR" ] && [ -d "$MOUNT_DIR" ]; then
        hdiutil detach "$MOUNT_DIR" -force -quiet 2>/dev/null || true
    fi
    if [ -n "$TMP_DIR" ] && [ -d "$TMP_DIR" ]; then
        rm -rf "$TMP_DIR" 2>/dev/null || true
    fi
    exit $exit_code
}
trap cleanup EXIT INT TERM

if [ -n "$LATEST_URL" ]; then
    echo -e "${BLUE}[*] Downloading release binary...${NC}"
    ARCHIVE_FILE="$TMP_DIR/RIFT_installer"
    curl -# -fSL "$LATEST_URL" -o "$ARCHIVE_FILE"

    echo -e "${BLUE}[*] Installing to $INSTALL_DIR...${NC}"
    # Terminate running instance if updating
    pkill -x RIFT 2>/dev/null || true
    sleep 0.5

    if [[ "$LATEST_URL" == *.dmg ]]; then
        MOUNT_DIR="$TMP_DIR/mount"
        mkdir -p "$MOUNT_DIR"
        hdiutil attach "$ARCHIVE_FILE" -mountpoint "$MOUNT_DIR" -nobrowse -quiet
        rm -rf "$TARGET_APP" 2>/dev/null || true
        cp -R "$MOUNT_DIR/$APP_NAME" "$INSTALL_DIR/"
        hdiutil detach "$MOUNT_DIR" -quiet 2>/dev/null || true
        MOUNT_DIR=""
    else
        unzip -q "$ARCHIVE_FILE" -d "$TMP_DIR/unpacked"
        rm -rf "$TARGET_APP" 2>/dev/null || true
        cp -R "$TMP_DIR/unpacked/$APP_NAME" "$INSTALL_DIR/"
    fi
else
    echo -e "${RED}[!] No published GitHub release found for $REPO yet.${NC}"
    exit 1
fi

# 4. Suppress Gatekeeper quarantine (Unnotarized Bypass)
echo -e "${BLUE}[*] Suppressing macOS Gatekeeper quarantine flags...${NC}"
xattr -cr "$TARGET_APP" 2>/dev/null || true

echo -e "\n${GREEN}${BOLD}✓ RIFT successfully installed to $TARGET_APP!${NC}"
echo -e "${CYAN}Launching RIFT...${NC}"
open "$TARGET_APP"
