#!/bin/bash
set -e

# GoMarkDoc API Documentation Generator for FlowKit
# Inspired by successful implementations in aali-agent and aali-llm

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOCS_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$DOCS_DIR")"
API_REF_DIR="$DOCS_DIR/source/api_reference"

echo "Starting API documentation generation..."
echo "Project root: $PROJECT_ROOT"
echo "API reference directory: $API_REF_DIR"

# Check if API docs already exist
if [ -f "$API_REF_DIR/externalfunctions/index.md" ] && \
   [ -f "$API_REF_DIR/functiondefinitions/index.md" ] && \
   [ -f "$API_REF_DIR/grpcserver/index.md" ]; then
    echo "API documentation files already exist. Skipping generation."
    echo "To regenerate, delete the existing .md files in $API_REF_DIR"
    exit 0
fi

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Cannot generate API documentation."
    echo "Using existing documentation files if available."
    exit 0
fi

# Install gomarkdoc if not present
if ! command -v gomarkdoc &> /dev/null; then
    echo "Installing gomarkdoc..."
    go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest
fi

# Clean existing docs (preserve index.rst files)
echo "Cleaning existing API docs..."
find "$API_REF_DIR" -name "*.md" -type f -delete

# Change to project root for proper package paths
cd "$PROJECT_ROOT"

# Generate markdown for core packages
echo "Generating documentation for core packages..."

# gRPC Server
mkdir -p "$API_REF_DIR/grpcserver"
gomarkdoc --output "$API_REF_DIR/grpcserver/index.md" \
          --format plain \
          --repository.url "https://github.com/ansys/aali-flowkit" \
          --repository.default-branch main \
          --repository.path "/" \
          ./pkg/grpcserver

# Function definitions (AST parsing)
mkdir -p "$API_REF_DIR/functiondefinitions"
gomarkdoc --output "$API_REF_DIR/functiondefinitions/index.md" \
          --format plain \
          --repository.url "https://github.com/ansys/aali-flowkit" \
          --repository.default-branch main \
          --repository.path "/" \
          ./pkg/functiondefinitions

# Internal states
mkdir -p "$API_REF_DIR/internalstates"
gomarkdoc --output "$API_REF_DIR/internalstates/index.md" \
          --format plain \
          --repository.url "https://github.com/ansys/aali-flowkit" \
          --repository.default-branch main \
          --repository.path "/" \
          ./pkg/internalstates

# Function testing
mkdir -p "$API_REF_DIR/functiontesting"
gomarkdoc --output "$API_REF_DIR/functiontesting/index.md" \
          --format plain \
          --repository.url "https://github.com/ansys/aali-flowkit" \
          --repository.default-branch main \
          --repository.path "/" \
          ./pkg/functiontesting

# Generate consolidated docs for externalfunctions
echo "Generating documentation for external functions..."
mkdir -p "$API_REF_DIR/externalfunctions"
gomarkdoc --output "$API_REF_DIR/externalfunctions/index.md" \
          --format plain \
          --repository.url "https://github.com/ansys/aali-flowkit" \
          --repository.default-branch main \
          --repository.path "/" \
          ./pkg/externalfunctions

# Also generate docs for private functions subdirectories
if [ -d "./pkg/privatefunctions" ]; then
    echo "Generating documentation for private functions..."

    # Generate docs for each subdirectory
    for subdir in codegeneration generic graphdb qdrant; do
        if [ -d "./pkg/privatefunctions/$subdir" ]; then
            echo "  Processing privatefunctions/$subdir..."
            mkdir -p "$API_REF_DIR/privatefunctions/$subdir"
            gomarkdoc --output "$API_REF_DIR/privatefunctions/$subdir/index.md" \
                      --format plain \
                      --repository.url "https://github.com/ansys/aali-flowkit" \
                      --repository.default-branch main \
                      --repository.path "/" \
                      ./pkg/privatefunctions/$subdir
        fi
    done
fi

# Generate docs for meshpilot subdirectories
if [ -d "./pkg/meshpilot" ]; then
    echo "Generating documentation for meshpilot packages..."

    # Generate docs for each subdirectory
    for subdir in ampgraphdb azure; do
        if [ -d "./pkg/meshpilot/$subdir" ]; then
            echo "  Processing meshpilot/$subdir..."
            mkdir -p "$API_REF_DIR/meshpilot/$subdir"
            gomarkdoc --output "$API_REF_DIR/meshpilot/$subdir/index.md" \
                      --format plain \
                      --repository.url "https://github.com/ansys/aali-flowkit" \
                      --repository.default-branch main \
                      --repository.path "/" \
                      ./pkg/meshpilot/$subdir
        fi
    done
fi

echo "API documentation generation completed successfully!"
echo "Generated files:"
find "$API_REF_DIR" -name "*.md" -type f | sort
