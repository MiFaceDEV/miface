#!/usr/bin/env bash
# Quick start script for MiFace development

set -e

echo "🚀 MiFace Quick Start"
echo "===================="
echo "This script is intended to help you set up the MiFace development environment quickly, for end-users use the binary releases."

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later."
    echo "   Visit: https://golang.org/dl/"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✅ Go $GO_VERSION installed"

# Check OpenCV
if ! pkg-config --exists opencv4 2>/dev/null; then
    echo "⚠️  OpenCV 4 not found."
    echo "   Please install OpenCV 4 with development headers."
    echo "   See BUILDING.md for platform-specific instructions."
    read -p "   Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    OPENCV_VERSION=$(pkg-config --modversion opencv4)
    echo "✅ OpenCV $OPENCV_VERSION installed"
fi

# Download dependencies
echo ""
echo "📦 Downloading dependencies..."
go mod download
echo "✅ Dependencies downloaded"

# Build the project
echo ""
echo "🔨 Building MiFace..."
make build
echo "✅ Build successful"

# Run tests
echo ""
echo "🧪 Running tests..."
if make test; then
    echo "✅ All tests passed"
else
    echo "⚠️  Some tests failed (this is normal if no camera is available)"
fi

# Create example config if it doesn't exist
if [ ! -f config.toml ] && [ -f config.example.toml ]; then
    echo ""
    read -p "📝 Create config.toml from example? (Y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        cp config.example.toml config.toml
        echo "✅ Created config.toml"
    fi
fi

echo ""
echo "🎉 Setup complete!"
echo ""
echo "Next steps:"
echo "  1. Edit config.toml to customize settings"
echo "  2. Run: ./bin/miface -preview         # Test with camera preview"
echo "  3. Run: make help                     # See all available commands"
echo "  4. Read: CONTRIBUTING.md              # Learn how to contribute"
echo ""
echo "For more information, see README.md and BUILDING.md"
