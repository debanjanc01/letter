#!/bin/bash

# Define the Go version you want to install
GO_VERSION="1.21.1"

# Determine the system architecture and OS
ARCHITECTURE=$(uname -m)
OS=$(uname -s)
case "$ARCHITECTURE" in
    "x86_64" | "amd64")
        GO_ARCH="amd64"
        ;;
    "aarch64" | "arm64")
        GO_ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCHITECTURE"
        exit 1
        ;;
esac

# Define the binary name and package name
BINARY_NAME="letter"
PACKAGE_NAME="github.com/debanjanc01/letter"

# Check if Go is already installed
GO_INSTALLED=false
if command -v go &> /dev/null; then
    echo "Necessary tools are already installed."
    GO_INSTALLED=true
else
    # Download and install Go
    echo "Getting the required tools..."
    case "$OS" in
        "Linux")
            DOWNLOAD_URL="https://dl.google.com/go/go$GO_VERSION.linux-$GO_ARCH.tar.gz"
            ;;
        "Darwin")
            DOWNLOAD_URL="https://dl.google.com/go/go$GO_VERSION.darwin-$GO_ARCH.tar.gz"
            ;;
        *)
            echo "Unsupported operating system: $OS"
            exit 1
            ;;
    esac
    curl -O "$DOWNLOAD_URL"
    mkdir -p golocal
    if [ -x "$(command -v tar)" ]; then
        tar -C golocal -xzf "go$GO_VERSION.$OS-$GO_ARCH.tar.gz"
    else
        echo "tar is not found. Please install a compatible archive utility (e.g., 7-Zip) and try again."
        exit 1
    fi
    export GOROOT=$PWD/golocal/go
    export PATH=$PWD/golocal/go/bin:$PATH
    echo "Tools acquired. Time to build..."
fi

# Build the Go binary
echo "Getting the necessary files..."
curl -s -LOJ "https://github.com/debanjanc01/letter/archive/refs/heads/main.zip"
echo "Reading the files..."
unzip -qq *.zip
cd letter-main
echo "Building the binary..."
go build -o $BINARY_NAME $PACKAGE_NAME
cd ..
cp letter-main/$BINARY_NAME .

echo "Time to clean up..."
rm -rf letter-main
rm letter-main.zip

# Uninstall Go only if it was downloaded during script execution
if [ "$GO_INSTALLED" = false ]; then
    echo "Tidying things up..."
    rm -rf "$PWD/golocal/go"
    unset GOROOT
    unset GOPATH
    unset PATH
    rm "go$GO_VERSION.$OS-$GO_ARCH.tar.gz"
fi

# Confirm the binary is built
if [ -f "$BINARY_NAME" ]; then
    echo "Binary $BINARY_NAME is built!"
else
    echo "Failed to build the binary."
    exit 1
fi

echo "Goodbye :)"
