# MediaPipe Integration Package

This package provides Go bindings to MediaPipe Holistic for real-time facial, hand, and pose tracking.

## Status

✅ **PRODUCTION-READY INTERFACE**

The Go API is complete and ready to use. The C++ implementation is in `../../cpp_core/`.

**To build:** See `../../cpp_core/QUICKSTART.md` for complete build instructions.

## Architecture

```
┌──────────────────────────┐
│   Go Package             │  processor.go (this package)
│   (CGO bindings)         │  ↓ calls via CGO
├──────────────────────────┤
│   C++ Library            │  ../../cpp_core/
│   libmediapipe_bridge.so │  - mediapipe_bridge.h (C API)
│                          │  - mediapipe_bridge.cc (C++ impl)
│   MediaPipe Holistic     │  - Bazel build system
└──────────────────────────┘
```

The C++ library is built **separately** with Bazel and linked at runtime.

## Usage

```go
import "github.com/MiFaceDEV/miface/pkg/mediapipe"

// Create processor with default config
config := mediapipe.DefaultConfig()
processor, err := mediapipe.NewMediaPipeProcessor(config)
if err != nil {
    log.Fatal(err)
}
defer processor.Close()

// Process a frame (RGB format)
frame := gocv.IMRead("image.jpg", gocv.IMReadColor)
data, err := processor.Process(frame)
if err != nil {
    log.Fatal(err)
}

// Access landmarks
if data.Face != nil {
    fmt.Printf("Detected %d face landmarks\n", len(data.Face.Landmarks))
}
```

## Configuration

```go
config := mediapipe.Config{
    ModelComplexity:        mediapipe.ComplexityFull, // 0=lite, 1=full, 2=heavy
    MinDetectionConfidence: 0.5,                      // [0.0, 1.0]
    MinTrackingConfidence:  0.5,                      // [0.0, 1.0]
    StaticImageMode:        false,                    // false = video tracking
    SmoothLandmarks:        true,                     // temporal smoothing
}
```

## Performance Tuning

- **ComplexityLite**: ~30-60 FPS, less accurate
- **ComplexityFull**: ~15-30 FPS, balanced (recommended)
- **ComplexityHeavy**: ~10-20 FPS, most accurate

For VTubing, use `ComplexityFull` with GPU acceleration.

## Building

The C++ library must be built separately:

```bash
cd ../../cpp_core
./build.sh      # CPU version
# OR
./build.sh gpu  # GPU version with CUDA
```

See `../../cpp_core/QUICKSTART.md` for complete setup instructions.

## Troubleshooting

### "cannot find -lmediapipe_bridge"
Build the C++ library first:
```bash
cd ../../cpp_core && ./build.sh
```

### "Processor not initialized"
MediaPipe failed to load. Check:
1. Is the shared library in the library path?
2. Are all MediaPipe dependencies installed?
3. Check error message from `NewMediaPipeProcessor()`

### Poor tracking quality
Adjust:
- `MinDetectionConfidence`: Lower = more false positives
- `MinTrackingConfidence`: Lower = better tracking in poor lighting
- `ModelComplexity`: Higher = more accurate but slower

### High CPU usage
- Enable GPU acceleration in the build
- Use `ComplexityLite` model
- Reduce camera resolution

## Next Steps

After building the C++ library:

1. **Add real MediaPipe graph config**
   - Edit `../../cpp_core/holistic_config.cc`
   - See `../../cpp_core/GRAPH_CONFIG_GUIDE.md`

2. **Test integration**
   ```bash
   go test -v ./pkg/mediapipe
   ```

3. **Implement blendshapes** (Section 4 of TODO.md)
   - Convert 468 face landmarks → 52 ARKit blendshapes
   - This is critical for VTuber facial expressions

## Related

- C++ Implementation: `../../cpp_core/`
- Build Guide: `../../cpp_core/QUICKSTART.md`
- Graph Config: `../../cpp_core/GRAPH_CONFIG_GUIDE.md`

## References

- [MediaPipe Holistic](https://google.github.io/mediapipe/solutions/holistic.html)
- [MediaPipe GitHub](https://github.com/google/mediapipe)
- [Face Mesh Model Card](https://google.github.io/mediapipe/solutions/face_mesh.html#model-card)
- [Hand Landmarks](https://google.github.io/mediapipe/solutions/hands.html#hand-landmark-model)
- [Pose Landmarks](https://google.github.io/mediapipe/solutions/pose.html#pose-landmark-model-blazepose-ghum-3d)
