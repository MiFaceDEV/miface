# MediaPipe Bridge - Quick Start Guide

## Overview

This standalone C++ library wraps MediaPipe Holistic and exposes a simple C API for Go CGO integration.

## Setup (One-time)

### 1. Install Bazel

```bash
# Ubuntu/Debian
sudo apt install apt-transport-https curl gnupg
curl https://bazel.build/bazel-release.pub.gpg | sudo apt-key add -
echo "deb [arch=amd64] https://storage.googleapis.com/bazel-apt stable jdk1.8" | \
    sudo tee /etc/apt/sources.list.d/bazel.list
sudo apt update && sudo apt install bazel

# macOS
brew install bazel

# Verify installation
bazel --version
```

### 2. Get MediaPipe

**Option A: Local clone (Recommended for development)**
```bash
cd ..  # Go to miface parent directory
git clone https://github.com/google/mediapipe.git
cd miface/cpp_core
# Update WORKSPACE to use local_repository
```

**Option B: Let Bazel fetch it (Easier, slower first build)**
```bash
# Already configured in WORKSPACE - Bazel will download automatically
```

### 3. Install Dependencies

```bash
# Ubuntu/Debian
sudo apt install -y \
    libopencv-dev \
    libopencv-contrib-dev \
    protobuf-compiler \
    libprotobuf-dev

# macOS
brew install opencv protobuf
```

## Building

### Quick Build (CPU)

```bash
cd cpp_core
./build.sh
```

This creates: `bazel-bin/libmediapipe_bridge.so`

### GPU Build (NVIDIA)

```bash
./build.sh gpu
```

Requires CUDA toolkit installed.

### Manual Build

```bash
# CPU version
bazel build -c opt --define=MEDIAPIPE_DISABLE_GPU=1 :libmediapipe_bridge.so

# GPU version
bazel build -c opt :libmediapipe_bridge_gpu.so

# Test binary
bazel build :bridge_test
```

## Testing

```bash
# Build and run test
bazel build :bridge_test
./bazel-bin/bridge_test

# Expected output:
# MediaPipe Bridge Test
# =====================
# Version: MediaPipe Bridge v1.0.0
# ...
# ✅ Test completed successfully!
```

## Installation

### System-wide

```bash
sudo cp bazel-bin/libmediapipe_bridge.so /usr/local/lib/
sudo ldconfig
```

### Local (for Go CGO)

```bash
mkdir -p ../pkg/mediapipe/bridge/lib
cp bazel-bin/libmediapipe_bridge.so ../pkg/mediapipe/bridge/lib/
```

## Integration with Go

Update `pkg/mediapipe/processor.go`:

```go
/*
#cgo LDFLAGS: -L${SRCDIR}/../../cpp_core/bazel-bin -lmediapipe_bridge
#cgo LDFLAGS: -Wl,-rpath,${SRCDIR}/../../cpp_core/bazel-bin
#include "../../cpp_core/mediapipe_bridge.h"
*/
import "C"
```

Then build your Go code:

```bash
cd ..
go build ./cmd/miface
```

## Troubleshooting

### "Cannot find mediapipe"

Check WORKSPACE file points to correct MediaPipe location.

### "Bazel version too old"

Update Bazel: `sudo apt update && sudo apt upgrade bazel`

### "opencv not found"

Install OpenCV development packages (see step 3 above).

### Build is extremely slow

First build compiles all of MediaPipe (~20-30 min). Subsequent builds are fast (~1-2 min).

Use `bazel build --jobs=4` to limit CPU usage.

### "undefined reference to cv::..."

OpenCV linking issue. Add to BUILD file:
```python
linkopts = ["-lopencv_core", "-lopencv_imgproc"]
```

## File Reference

- `mediapipe_bridge.h` - C API header (include in Go)
- `mediapipe_bridge.cc` - Main implementation
- `holistic_config.cc` - MediaPipe graph config
- `BUILD` - Bazel build rules
- `WORKSPACE` - Bazel dependencies
- `build.sh` - Convenience build script
- `bridge_test.cc` - Test program

## Next Steps

1. ✅ Build the library successfully
2. ✅ Run `bridge_test` to verify it works
3. Update Go CGO directives in `pkg/mediapipe/processor.go`
4. Test Go integration with: `go test ./pkg/mediapipe`
5. Implement blendshape calculation (Section 4 of TODO.md)

## Performance Tips

- Use GPU build for real-time tracking (>30 FPS)
- CPU build works well for medium resolution (640x480 @ 20-30 FPS)
- Adjust `model_complexity` in config (0=fast, 2=accurate)
- Lower `min_detection_confidence` if tracking drops frequently

## Support

- [MediaPipe Documentation](https://google.github.io/mediapipe/)
- [Bazel Documentation](https://bazel.build/docs)
- [MiFace GitHub Issues](https://github.com/MiFaceDEV/miface/issues)
