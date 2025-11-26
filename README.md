> ⚠️ This project is in beta, the core was made with AI (AI isnt able to do everything and only do repetitive tasks) and is being improved and refactored continuously by humans (mainly me, MiguVT)

# MiFace

Real-time facial and upper body tracking library for VTubers. MediaPipe Holistic + Kalman filtering + VMC/OSC sender. Go library with CLI.

## Features

- 🎭 **Face Tracking** - 468 face mesh landmarks with blend shapes
- ✋ **Hand Tracking** - Left and right hand landmark detection
- 🏃 **Pose Tracking** - Upper body pose estimation
- 🎯 **Kalman Smoothing** - Reduces jitter while maintaining responsiveness
- 📡 **VMC Protocol** - Standard protocol for VTuber applications (uses OSC)
- 🦴 **VRM Calibration** - Load VRM models for bone proportion calibration
- ⚡ **Concurrent-Safe** - Thread-safe design for real-time performance

## Installation

### As a Library

```bash
go get github.com/MiFaceDEV/miface
```

### CLI Tool

```bash
go install github.com/MiFaceDEV/miface/cmd/miface@latest
```

## Quick Start

### Library Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/MiFaceDEV/miface/internal/config"
    "github.com/MiFaceDEV/miface/pkg/miface"
)

func main() {
    // Create tracker with default config
    tracker, err := miface.NewTracker(nil)
    if err != nil {
        log.Fatal(err)
    }
    defer tracker.Close()

    // Subscribe to tracking data
    dataCh := tracker.Subscribe()

    // Start tracking
    if err := tracker.Start(); err != nil {
        log.Fatal(err)
    }

    // Process tracking data
    for data := range dataCh {
        if data.Face != nil {
            fmt.Printf("Frame %d: %d face landmarks\n", 
                data.FrameNumber, len(data.Face.Landmarks))
        }
    }
}
```

### With Custom Configuration

```go
cfg, _ := config.Load("config.toml")
tracker, err := miface.NewTracker(cfg)
```

### With VMC Output

```go
tracker, _ := miface.NewTracker(nil)

// Set up VMC sender for VTuber applications
vmcSender, _ := miface.NewVMCSender("127.0.0.1", 39539)
tracker.SetVMCSender(vmcSender)

tracker.Start()
```

### VRM Calibration

Load a VRM file to extract bone proportions for accurate tracking mapping:

```go
// Load VRM skeleton (bones only, no meshes/textures)
skeleton, err := miface.LoadVRMSkeleton("model.vrm")
if err != nil {
    log.Fatal(err)
}

// Get bone proportions for calibration
props := skeleton.GetProportions()
fmt.Printf("Arm length: %.3f\n", props.ArmLength)
fmt.Printf("Shoulder width: %.3f\n", props.ShoulderWidth)

// List available humanoid bones
bones := skeleton.ListHumanBones()
for _, bone := range bones {
    pos, _ := skeleton.GetBonePosition(bone)
    fmt.Printf("%s: (%.2f, %.2f, %.2f)\n", bone, pos.X, pos.Y, pos.Z)
}
```

### CLI Usage

```bash
# Run with default settings
miface

# Show camera preview window (debug mode)
miface -preview

# Use custom configuration
miface -config config.toml

# Override VMC settings
miface -vmc-addr 192.168.1.100 -vmc-port 39540

# Calibrate with VRM model
miface -vrm model.vrm -verbose

# Show version
miface -version

# Show help
miface -help
```

## Development

### Building

```bash
# Using Makefile
make build          # Build binary to bin/miface
make test           # Run all tests
make test-coverage  # Run tests with coverage report
make bench          # Run benchmarks
make install        # Install to $GOPATH/bin
make run-preview    # Build and run with preview window

# Manual build
go build -o miface ./cmd/miface
go test ./...
```

See [BUILDING.md](BUILDING.md) for platform-specific build instructions and dependencies.

## Configuration

Create a `config.toml` file:

```toml
[camera]
device_id = 0
width = 1280
height = 720
fps = 30

[tracking]
enable_face = true
enable_hands = true
enable_pose = true
smoothing_factor = 0.5  # 0.0 = max smoothing, 1.0 = no smoothing

[vmc]
enabled = true
address = "127.0.0.1"
port = 39539
```

## Architecture

MiFace follows a library-first design for maximum reusability:

```
pkg/miface/          # Core library
├── tracker.go       # Main tracker coordinator
├── kalman.go        # Kalman filter for smoothing
├── sender.go        # VMC protocol sender (uses OSC)
└── vrm.go           # VRM bone/skeleton parser

cmd/miface/          # CLI wrapper
└── main.go

internal/config/     # TOML configuration
└── config.go
```

### Key Components

- **Tracker**: Main coordinator managing capture, tracking, and output
- **CameraSource**: Interface for webcam capture backends (pluggable)
- **Processor**: Interface for landmark detection (MediaPipe integration)
- **KalmanFilter**: Smoothing filter for landmark stabilization
- **VMCSender**: Protocol sender for VTuber applications
- **VRMSkeleton**: Bone proportion extraction from VRM files

## VMC Protocol

MiFace sends tracking data using the VMC (Virtual Motion Capture) protocol over OSC:

- `/VMC/Ext/Bone/Pos` - Bone positions and rotations
- `/VMC/Ext/Blend/Val` - Facial blend shape values
- `/VMC/Ext/Blend/Apply` - Blend shape apply signal

Compatible with:
- VSeeFace
- VMagicMirror
- VirtualMotionCapture
- Other VMC-compatible applications

## License

AGPL-3.0 - See [LICENSE](LICENSE) for details.

