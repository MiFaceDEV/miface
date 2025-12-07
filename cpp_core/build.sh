#!/bin/bash
# build.sh - Build script for MediaPipe bridge

set -e  # Exit on error

MODE="${1:-cpu}"  # cpu or gpu

echo "======================================"
echo "Building MediaPipe Bridge ($MODE mode)"
echo "======================================"

# Check for Bazel
if ! command -v bazel &> /dev/null; then
    echo "Error: Bazel is not installed"
    echo "Install: https://bazel.build/install"
    exit 1
fi

# Build configuration
BUILD_FLAGS=(
    "-c opt"                    # Optimized build
    "--cxxopt=-std=c++17"       # C++17 standard
    "--copt=-fPIC"              # Position independent code
    "--copt=-DMESA_EGL_NO_X11_HEADERS"  # Fix EGL headers
)

# Add GPU flags if requested
if [ "$MODE" = "gpu" ]; then
    echo "Building with GPU support..."
    TARGET="//cpp_core:libmediapipe_bridge_gpu.so"
    BUILD_FLAGS+=(
        "--define=MEDIAPIPE_DISABLE_GPU=0"
        "--copt=-DMEDIAPIPE_GPU_ENABLED"
    )
else
    echo "Building with CPU only..."
    TARGET=":libmediapipe_bridge.so"
    BUILD_FLAGS+=(
        "--define=MEDIAPIPE_DISABLE_GPU=1"
    )
fi

# Build the library
echo ""
echo "Running: bazel build ${BUILD_FLAGS[*]} $TARGET"
echo ""

bazel build "${BUILD_FLAGS[@]}" "$TARGET"

# Check result
if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Build successful!"
    echo ""
    echo "Output: bazel-bin/libmediapipe_bridge*.so"
    echo ""
    echo "To install:"
    echo "  sudo cp bazel-bin/libmediapipe_bridge*.so /usr/local/lib/"
    echo "  sudo ldconfig"
    echo ""
    echo "Or for local use:"
    echo "  mkdir -p ../pkg/mediapipe/bridge/lib"
    echo "  cp bazel-bin/libmediapipe_bridge*.so ../pkg/mediapipe/bridge/lib/"
else
    echo ""
    echo "❌ Build failed!"
    exit 1
fi
