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
          --repository.url "https://github.com/ansys/aali-flowkit" \
          --repository.default-branch main \
          --repository.path "/" \
          ./pkg/grpcserver

# Function definitions (AST parsing)
mkdir -p "$API_REF_DIR/functiondefinitions"
gomarkdoc --output "$API_REF_DIR/functiondefinitions/index.md" \
          --repository.url "https://github.com/ansys/aali-flowkit" \
          --repository.default-branch main \
          --repository.path "/" \
          ./pkg/functiondefinitions

# Internal states
mkdir -p "$API_REF_DIR/internalstates"
gomarkdoc --output "$API_REF_DIR/internalstates/index.md" \
          --repository.url "https://github.com/ansys/aali-flowkit" \
          --repository.default-branch main \
          --repository.path "/" \
          ./pkg/internalstates

# Generate consolidated docs for externalfunctions
echo "Generating documentation for external functions..."
mkdir -p "$API_REF_DIR/externalfunctions"
gomarkdoc --output "$API_REF_DIR/externalfunctions/index.md" \
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
                      --repository.url "https://github.com/ansys/aali-flowkit" \
                      --repository.default-branch main \
                      --repository.path "/" \
                      ./pkg/privatefunctions/$subdir
        fi
    done
fi

echo "API documentation generation completed successfully!"
echo "Generated files:"
find "$API_REF_DIR" -name "*.md" -type f | sort
