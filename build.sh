#!/usr/bin/env bash

# Build script for service-center
# Supports building individual binaries or all at once

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

VERSION="${VERSION:-dev}"
BUILD_DIR="_output/bin"
BINARIES=("server" "client")

function usage() {
    cat <<EOF
Build script for service-center

Usage: $0 [OPTIONS] [BINARIES...]

OPTIONS:
    -h, --help              Show this help message
    -v, --version VERSION   Set version (default: dev)
    -o, --output DIR        Set output directory (default: _output/bin)
    -p, --platform OS/ARCH  Set target platform (default: current)
    --all-platforms         Build for all platforms (linux, darwin, windows)

BINARIES:
    server                  Build server binary
    client                  Build client binary
    (none)                  Build all binaries

EXAMPLES:
    $0                      # Build all binaries
    $0 server               # Build only server
    $0 --platform linux/amd64 server client
    $0 --all-platforms      # Build for all platforms
EOF
}

PLATFORM=""
ALL_PLATFORMS=false
BUILD_TARGETS=()

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -o|--output)
            BUILD_DIR="$2"
            shift 2
            ;;
        -p|--platform)
            PLATFORM="$2"
            shift 2
            ;;
        --all-platforms)
            ALL_PLATFORMS=true
            shift
            ;;
        server|client)
            BUILD_TARGETS+=("$1")
            shift
            ;;
        *)
            echo "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# If no targets specified, build all
if [ ${#BUILD_TARGETS[@]} -eq 0 ]; then
    BUILD_TARGETS=("${BINARIES[@]}")
fi

function build_binary() {
    local binary=$1
    local goos=$2
    local goarch=$3
    local output_dir="${BUILD_DIR}/${goos}/${goarch}"
    local output_file="${output_dir}/${binary}"
    
    if [ "$goos" = "windows" ]; then
        output_file="${output_file}.exe"
    fi
    
    echo "Building ${binary} for ${goos}/${goarch}..."
    mkdir -p "$output_dir"
    
    CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch \
        go build -ldflags "-w -s -X main.Version=${VERSION}" \
        -o "$output_file" \
        "./cmd/${binary}"
    
    echo "  ✓ Built: $output_file"
}

if [ "$ALL_PLATFORMS" = true ]; then
    # Build for all platforms
    platforms=(
        "linux/amd64"
        "linux/arm64"
        "darwin/amd64"
        "darwin/arm64"
        "windows/amd64"
    )
    
    for platform in "${platforms[@]}"; do
        IFS='/' read -r goos goarch <<< "$platform"
        for binary in "${BUILD_TARGETS[@]}"; do
            build_binary "$binary" "$goos" "$goarch"
        done
    done
elif [ -n "$PLATFORM" ]; then
    # Build for specific platform
    IFS='/' read -r goos goarch <<< "$PLATFORM"
    for binary in "${BUILD_TARGETS[@]}"; do
        build_binary "$binary" "$goos" "$goarch"
    done
else
    # Build for current platform
    goos=$(go env GOOS)
    goarch=$(go env GOARCH)
    for binary in "${BUILD_TARGETS[@]}"; do
        build_binary "$binary" "$goos" "$goarch"
    done
fi

echo ""
echo "✓ Build completed successfully!"
echo "Version: $VERSION"
echo "Output directory: $BUILD_DIR"
