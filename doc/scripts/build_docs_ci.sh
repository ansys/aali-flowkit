#!/bin/bash
# Wrapper script for CI documentation build
# This script handles the case where Go might not be available or the wrong version

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOCS_DIR="$(dirname "$SCRIPT_DIR")"

echo "CI Documentation Build Wrapper"
echo "=============================="

# Check if we're in CI environment
if [ -n "$CI" ]; then
    echo "Running in CI environment"

    # Check if Go is available and what version
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        echo "Go is installed: version $GO_VERSION"
    else
        echo "Go is not installed or not available"
    fi
fi

# Change to docs directory
cd "$DOCS_DIR"

# Try to run the standard build first
if make html; then
    echo "Documentation build completed successfully"
    exit 0
else
    echo "Standard build failed, trying alternative build without Go dependency"

    # If standard build fails, try html-only which doesn't regenerate API docs
    if make html-only; then
        echo "Documentation build completed successfully (using existing API docs)"
        exit 0
    else
        echo "Documentation build failed"
        exit 1
    fi
fi
