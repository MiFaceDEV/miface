# Building MiFace

## Prerequisites

### OpenCV Installation

MiFace requires OpenCV 4.x with development headers. The camera capture module uses GoCV, which is a CGO wrapper around OpenCV.

#### Ubuntu/Debian
```bash
# Install OpenCV and dependencies
sudo apt-get update
sudo apt-get install -y \
    libopencv-dev \
    libopencv-core-dev \
    libopencv-highgui-dev \
    libopencv-videoio-dev \
    libopencv-imgproc-dev \
    pkg-config

# Verify installation
pkg-config --modversion opencv4
```

#### Arch Linux
```bash
sudo pacman -S opencv vtk hdf5
```

#### Fedora
```bash
sudo dnf install opencv opencv-devel
```

#### macOS
```bash
brew install opencv
```

### Go Requirements

- Go 1.24.10 or later
- CGO enabled (required for OpenCV bindings)

## Building

### Standard Build
```bash
# Clone the repository
git clone https://github.com/MiFaceDEV/miface
cd miface

# Download dependencies
go mod download

# Build the CLI
go build -o miface ./cmd/miface

# Run
./miface -help
```

### Build Tags

The camera module uses the `cgo` build tag. It's automatically included when CGO is enabled.

### Troubleshooting

#### "could not import gocv.io/x/gocv"
Make sure OpenCV is installed and `pkg-config` can find it:
```bash
pkg-config --cflags --libs opencv4
```

#### Missing VTK/HDF5 Libraries
On Arch Linux and some distributions, you may need to install additional dependencies:
```bash
sudo pacman -S vtk hdf5  # Arch
sudo apt-get install libvtk9-dev libhdf5-dev  # Ubuntu
```

#### CGO Errors
Ensure CGO is enabled:
```bash
export CGO_ENABLED=1
go env CGO_ENABLED  # Should output: 1
```

#### Camera Permission Denied
On Linux, your user needs access to video devices:
```bash
sudo usermod -aG video $USER
# Log out and log back in for changes to take effect
```

## Running Tests

```bash
# Run all tests (requires camera hardware)
go test ./...

# Run tests without camera integration
go test -short ./...

# Run with verbose output
go test -v ./...
```

## Cross-Platform Notes

### Linux
- Uses V4L2 through OpenCV
- Tested on Ubuntu 22.04+, Arch Linux, Fedora 38+

### macOS
- Uses AVFoundation through OpenCV
- Requires macOS 10.13+

### Windows
- Uses DirectShow through OpenCV
- Requires Visual Studio Build Tools for CGO
- See [GoCV Windows Installation](https://gocv.io/getting-started/windows/)

## Docker Build (Alternative)

For a reproducible build environment:

```dockerfile
FROM golang:1.24-bullseye

RUN apt-get update && apt-get install -y \
    libopencv-dev \
    libopencv-core-dev \
    libopencv-highgui-dev \
    libopencv-videoio-dev \
    libopencv-imgproc-dev \
    pkg-config

WORKDIR /app
COPY . .
RUN go build -o miface ./cmd/miface

CMD ["./miface"]
```

Build and run:
```bash
docker build -t miface .
docker run --rm --device=/dev/video0 miface
```

## Performance Tips

- **GPU Acceleration**: Install CUDA-enabled OpenCV for GPU support
- **Optimized Build**: Use `-tags opencv_static` for static linking
- **Debugging**: Set `CGO_CFLAGS="-g"` for debug symbols

## Next Steps

After successful build:
1. Run `./miface -help` to see available options
2. Test camera capture: `./miface -verbose`
3. Configure VMC output: `./miface -vmc-port 39539`
4. Load a VRM model: `./miface -vrm model.vrm`

For development workflow, see `TODO.md` for the implementation roadmap.
