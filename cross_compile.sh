#!/bin/bash

# cross_compile.sh

set -e

# Function to display usage instructions
usage() {
    echo "Usage: $0 --os <GOOS> --arch <GOARCH> [--output <OUTPUT_DIR>] [--name <PROJECT_NAME>]"
    echo ""
    echo "Options:"
    echo "  --os        Target operating system (e.g., windows, linux, darwin)"
    echo "  --arch      Target architecture (e.g., amd64, arm64, arm)"
    echo "  --output    Output directory (default: build)"
    echo "  --name      Project name (default: biotechApi)"
    echo "  --help      Display this help message"
    exit 1
}

# Default values
PROJECT_NAME="biotechApi"
OUTPUT_DIR="build"

# Parse command-line arguments
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
        --os)
            GOOS="$2"
            shift # past argument
            shift # past value
            ;;
        --arch)
            GOARCH="$2"
            shift
            shift
            ;;
        --output)
            OUTPUT_DIR="$2"
            shift
            shift
            ;;
        --name)
            PROJECT_NAME="$2"
            shift
            shift
            ;;
        --help)
            usage
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

# Validate required arguments
if [[ -z "$GOOS" || -z "$GOARCH" ]]; then
    echo "Error: Both --os and --arch must be specified."
    usage
fi

# Save original environment variables
ORIGINAL_GOOS="${GOOS_ENV_ORIGINAL:-$GOOS}"
ORIGINAL_GOARCH="${GOARCH_ENV_ORIGINAL:-$GOARCH}"

# Optionally save existing GOOS and GOARCH if they are set
if [[ -n "$GOOS" ]]; then
    ORIGINAL_GOOS="$GOOS"
fi
if [[ -n "$GOARCH" ]]; then
    ORIGINAL_GOARCH="$GOARCH"
fi

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Determine output binary name
OUTPUT_NAME="${PROJECT_NAME}-${GOOS}-${GOARCH}"
if [ "$GOOS" = "windows" ]; then
    OUTPUT_NAME+=".exe"
fi

echo "Building for $GOOS/$GOARCH..."
# Set environment variables and build
env GOOS="$GOOS" GOARCH="$GOARCH" go build -o "${OUTPUT_DIR}/${OUTPUT_NAME}"

echo "Build successful: ${OUTPUT_DIR}/${OUTPUT_NAME}"

# Restore original environment variables
export GOOS="$ORIGINAL_GOOS"
export GOARCH="$ORIGINAL_GOARCH"

echo "Environment variables restored."
