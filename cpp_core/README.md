# C++ Core - MediaPipe Integration Library

This directory contains the standalone C++ shared library that wraps MediaPipe Holistic.
It's built separately with Bazel and linked by Go via CGO.

## Structure

```
cpp_core/
├── BUILD                    # Bazel build configuration
├── WORKSPACE               # Bazel workspace setup
├── mediapipe_bridge.h      # C API header
├── mediapipe_bridge.cc     # C++ implementation
├── holistic_config.cc      # MediaPipe graph configuration
├── build.sh                # Build script
└── README.md               # This file
```

## Building

### Prerequisites

```bash
# Install Bazel
sudo apt install bazel  # Ubuntu
brew install bazel      # macOS

# Install MediaPipe dependencies
sudo apt install libopencv-dev
```

### Build Commands

```bash
# Build CPU version (default)
./build.sh

# Build GPU version (CUDA)
./build.sh gpu

# Clean build
bazel clean --expunge
```

Output: `bazel-bin/libmediapipe_bridge.so` (or `.dylib` on macOS)

### Install

```bash
# Copy to system library path
sudo cp bazel-bin/libmediapipe_bridge.so /usr/local/lib/
sudo ldconfig

# Or use local path (set in Go CGO LDFLAGS)
mkdir -p ../pkg/mediapipe/bridge/lib
cp bazel-bin/libmediapipe_bridge.so ../pkg/mediapipe/bridge/lib/
```

## Testing

```bash
# Build test binary
bazel build :bridge_test

# Run test
./bazel-bin/bridge_test
```

## Integration with Go

Once built, update `pkg/mediapipe/processor.go`:

```go
/*
#cgo LDFLAGS: -L${SRCDIR}/../../cpp_core/bazel-bin -lmediapipe_bridge
#include "../../cpp_core/mediapipe_bridge.h"
*/
import "C"
```

## Troubleshooting

### "Cannot find MediaPipe headers"
Clone MediaPipe in a sibling directory or update WORKSPACE path.

### "Undefined reference to cv::..."
OpenCV linking issue. Ensure `libopencv_core.so` is findable.

### Build is slow
First build takes 10-30 minutes. Subsequent builds are incremental.
