#!/bin/bash

# Verilis Installation Script
# This script automatically detects your platform and installs the appropriate binary

APP_NAME="verilis"
GITHUB_REPO="https://github.com/user/verilis"

# Function to get the latest version from GitHub
get_latest_version() {
  print_color "$BLUE" "Checking for the latest version..."
  
  if command_exists curl; then
    LATEST_VERSION=$(curl -s $GITHUB_REPO/releases/latest | grep -o 'tag/v[0-9]\+\.[0-9]\+\.[0-9]\+' | cut -d '/' -f 2 | cut -c 2-)
  elif command_exists wget; then
    LATEST_VERSION=$(wget -qO- $GITHUB_REPO/releases/latest | grep -o 'tag/v[0-9]\+\.[0-9]\+\.[0-9]\+' | cut -d '/' -f 2 | cut -c 2-)
  else
    # Default to the version passed to the script if we can't check
    LATEST_VERSION="0.1.0"
    print_color "$YELLOW" "Could not check for latest version. Using version $LATEST_VERSION."
  fi
  
  # If we couldn't get the latest version, use the provided version
  if [ -z "$LATEST_VERSION" ]; then
    LATEST_VERSION="0.1.0"
    print_color "$YELLOW" "Could not determine latest version. Using version $LATEST_VERSION."
  else
    print_color "$GREEN" "Latest version: $LATEST_VERSION"
  fi
  
  VERSION="$LATEST_VERSION"
  DOWNLOAD_BASE_URL="$GITHUB_REPO/releases/download/v$VERSION"
}

# Default installation directory (will be set based on platform)

# Check if we're installing from local files
LOCAL_INSTALL=false
if [ -d "./bin/release" ]; then
  LOCAL_INSTALL=true
fi

# ANSI color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print with color
print_color() {
  printf "%b%s%b\n" "$1" "$2" "$NC"
}

# Check if a command exists
command_exists() {
  command -v "$1" >/dev/null 2>&1
}

# Detect OS and architecture
detect_platform() {
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  ARCH="$(uname -m)"
  
  # Convert architecture to Go naming convention
  case "$ARCH" in
    x86_64)
      ARCH="amd64"
      ;;
    aarch64|arm64)
      ARCH="arm64"
      ;;
    *)
      print_color "$RED" "Unsupported architecture: $ARCH"
      exit 1
      ;;
  esac
  
  # Convert OS to Go naming convention
  case "$OS" in
    darwin)
      OS="darwin"
      ;;
    linux)
      OS="linux"
      ;;
    msys*|mingw*|cygwin*)
      OS="windows"
      ;;
    *)
      print_color "$RED" "Unsupported operating system: $OS"
      exit 1
      ;;
  esac
  
  PLATFORM="$OS-$ARCH"
  
  if [ "$OS" = "windows" ]; then
    BINARY_NAME="${APP_NAME}-${PLATFORM}.exe"
    ARCHIVE_NAME="${APP_NAME}-${VERSION}-${PLATFORM}.zip"
  else
    BINARY_NAME="${APP_NAME}-${PLATFORM}"
    ARCHIVE_NAME="${APP_NAME}-${VERSION}-${PLATFORM}.tar.gz"
  fi
}

# Determine the best installation directory based on platform
set_install_dir() {
  # Default installation directory
  INSTALL_DIR="$HOME/.local/bin"
  NEED_SUDO=false
  
  case "$OS" in
    linux)
      # Check if we can write to /usr/local/bin
      if [ -w "/usr/local/bin" ]; then
        INSTALL_DIR="/usr/local/bin"
      elif [ -d "/usr/local/bin" ]; then
        # We need sudo to write to /usr/local/bin
        INSTALL_DIR="/usr/local/bin"
        NEED_SUDO=true
      fi
      ;;
    darwin)
      # Check if we can write to /usr/local/bin
      if [ -w "/usr/local/bin" ]; then
        INSTALL_DIR="/usr/local/bin"
      elif [ -d "/usr/local/bin" ]; then
        # We need sudo to write to /usr/local/bin
        INSTALL_DIR="/usr/local/bin"
        NEED_SUDO=true
      fi
      ;;
    windows)
      # Windows doesn't have a standard bin directory, use user's home
      INSTALL_DIR="$HOME/bin"
      ;;
  esac
  
  print_color "$BLUE" "Installation directory: $INSTALL_DIR"
  if [ "$NEED_SUDO" = true ]; then
    print_color "$YELLOW" "Note: Installation to $INSTALL_DIR requires administrator privileges."
  fi
}

# Create installation directory if it doesn't exist
create_install_dir() {
  if [ ! -d "$INSTALL_DIR" ]; then
    print_color "$BLUE" "Creating installation directory: $INSTALL_DIR"
    
    if [ "$NEED_SUDO" = true ]; then
      # Ask for confirmation before using sudo
      read -p "Do you want to install to $INSTALL_DIR? This requires sudo privileges. (y/N): " confirm
      if [[ "$confirm" =~ ^[Yy]$ ]]; then
        sudo mkdir -p "$INSTALL_DIR"
      else
        # Fall back to user's home directory
        INSTALL_DIR="$HOME/.local/bin"
        NEED_SUDO=false
        print_color "$BLUE" "Installing to $INSTALL_DIR instead."
        mkdir -p "$INSTALL_DIR"
      fi
    else
      mkdir -p "$INSTALL_DIR"
    fi
  fi
  
  # Check if INSTALL_DIR is in PATH
  if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    print_color "$YELLOW" "Warning: $INSTALL_DIR is not in your PATH."
    print_color "$YELLOW" "Add the following line to your shell profile (.bashrc, .zshrc, etc.):"
    print_color "$YELLOW" "  export PATH=\"$INSTALL_DIR:$PATH\""
  fi
}

# Download or copy the binary
download_binary() {
  TMP_DIR="$(mktemp -d)"
  
  if [ "$LOCAL_INSTALL" = true ]; then
    # Local installation
    print_color "$BLUE" "Installing $APP_NAME v$VERSION for $PLATFORM from local files..."
    
    # Check if the binary exists locally
    if [ -f "./bin/release/$BINARY_NAME" ]; then
      cp "./bin/release/$BINARY_NAME" "$TMP_DIR/$BINARY_NAME"
    else
      print_color "$RED" "Error: Local binary not found: ./bin/release/$BINARY_NAME"
      exit 1
    fi
  else
    # Remote installation
    DOWNLOAD_URL="$DOWNLOAD_BASE_URL/$ARCHIVE_NAME"
    print_color "$BLUE" "Downloading $APP_NAME v$VERSION for $PLATFORM..."
    print_color "$BLUE" "URL: $DOWNLOAD_URL"
    
    if command_exists curl; then
      curl -L -o "$TMP_DIR/$ARCHIVE_NAME" "$DOWNLOAD_URL"
    elif command_exists wget; then
      wget -O "$TMP_DIR/$ARCHIVE_NAME" "$DOWNLOAD_URL"
    else
      print_color "$RED" "Error: Neither curl nor wget found. Please install one of them and try again."
      exit 1
    fi
    
    if [ $? -ne 0 ]; then
      print_color "$RED" "Error: Failed to download $APP_NAME."
      exit 1
    fi
    
    # Extract the archive
    print_color "$BLUE" "Extracting..."
    if [ "$OS" = "windows" ]; then
      if command_exists unzip; then
        unzip -o "$TMP_DIR/$ARCHIVE_NAME" -d "$TMP_DIR"
      else
        print_color "$RED" "Error: unzip command not found. Please install unzip and try again."
        exit 1
      fi
    else
      tar -xzf "$TMP_DIR/$ARCHIVE_NAME" -C "$TMP_DIR"
    fi
  fi
  
  # No extraction needed for local install
  
  # Install the binary
  if [ "$NEED_SUDO" = true ]; then
    sudo cp "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$APP_NAME"
    sudo chmod +x "$INSTALL_DIR/$APP_NAME"
  else
    cp "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$APP_NAME"
    chmod +x "$INSTALL_DIR/$APP_NAME"
  fi
  
  # Clean up
  rm -rf "$TMP_DIR"
}

# Main installation process
main() {
  print_color "$GREEN" "=== $APP_NAME Installer ==="
  
  # ANSI color codes need to be defined before get_latest_version
  get_latest_version
  
  detect_platform
  print_color "$BLUE" "Detected platform: $PLATFORM"
  
  set_install_dir
  create_install_dir
  download_binary
  
  print_color "$GREEN" "✓ $APP_NAME v$VERSION has been installed successfully!"
  print_color "$GREEN" "You can now run it using: $APP_NAME"
  
  # Verify installation
  if command_exists "$INSTALL_DIR/$APP_NAME"; then
    print_color "$BLUE" "Installation verified."
  else
    print_color "$YELLOW" "Warning: Installation could not be verified. Please check your PATH."
  fi
}

# Run the installation
main
