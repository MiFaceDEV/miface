# MiFace Development TODO

This document tracks the implementation status and remaining work for MiFace, a real-time facial and upper-body tracking library for VTubers.

**Legend:**
- `[x]` - Implemented
- `[ ]` - Not yet implemented
- `[~]` - Partially implemented

---

## 1. Project Scaffolding & Architecture

### Core Structure
- [x] Library-first design with `pkg/miface` package
- [x] CLI wrapper in `cmd/miface`
- [x] Configuration system with TOML support
- [x] Interface-based design for pluggable components
- [x] Concurrent-safe tracker architecture
- [x] Subscriber pattern for data streaming

### Data Structures
- [x] `Point3D`, `Landmark`, `Quaternion` types defined
- [x] `TrackingData` structure with face, hands, and pose
- [x] `FaceData` with 468 landmarks and blend shapes map
- [x] `HandData` with 21 landmarks per hand
- [x] `PoseData` with 33 pose landmarks
- [x] `TrackerState` enum (Idle, Running, Stopped, Closed)

---

## 2. Camera Capture

### Webcam Integration
- [x] `CameraSource` interface defined
- [x] **OpenCV/GoCV webcam capture implementation**
  - [x] Install and integrate `gocv.io/x/gocv` (OpenCV bindings)
  - [x] Implement `OpenCVCamera` struct implementing `CameraSource`
  - [x] Support for V4L2 on Linux (explicit V4L2 backend, no GStreamer)
  - [x] Device enumeration and selection by ID
  - [x] Resolution and FPS configuration
  - [x] MJPEG codec configuration for USB webcam compatibility
  - [x] Error handling for camera access permissions
  - [x] Graceful fallback if camera unavailable
  - [x] **Horizontal flip (mirror mode) support for VTubing**
  - [x] BGR to RGB color conversion
  - [x] Integration in CLI (`cmd/miface/main.go`)
  - [x] Unit tests and benchmarks
- [ ] **Alternative: Pure Go capture using `github.com/blackjack/webcam`**
  - [ ] Evaluate performance vs OpenCV
  - [ ] MJPEG/YUYV format support
  - [ ] Frame format conversion to RGB24
- [~] Mock camera for testing (tracker generates stub data, but no formal mock)
- [~] Camera capabilities detection (basic enumeration implemented)
- [ ] Auto-exposure and white balance configuration

---

## 3. MediaPipe Holistic Integration

### MediaPipe Backend
- [x] `Processor` interface defined
- [ ] **Choose MediaPipe integration approach:**
  - **Option A: CGO bindings to MediaPipe C++ API**
    - [ ] Install MediaPipe from source or prebuilt binaries
    - [ ] Create CGO wrapper for Holistic pipeline
    - [ ] Memory management between Go and C++
    - [ ] Thread safety considerations
  - **Option B: Use `github.com/google/mediapipe` with Go bindings**
    - [ ] Evaluate existing Go wrappers (if any)
    - [ ] May require custom bindings
  - **Option C: Python bridge via subprocess/RPC**
    - [ ] Implement protocol buffer or JSON-based communication
    - [ ] Latency concerns for real-time tracking
- [ ] **Implement `MediaPipeProcessor` struct**
  - [ ] Initialize Holistic model with static image mode = false
  - [ ] Configure model complexity (0=lite, 1=full, 2=heavy)
  - [ ] Min detection confidence threshold
  - [ ] Min tracking confidence threshold
- [ ] **Process RGB frames to get landmarks:**
  - [ ] Face mesh: 468 landmarks
  - [ ] Left hand: 21 landmarks
  - [ ] Right hand: 21 landmarks
  - [ ] Pose: 33 landmarks (focus on upper body: 0-16)
- [ ] Convert MediaPipe normalized coordinates to 3D world coordinates
- [ ] Handle landmark visibility/presence scores
- [ ] GPU acceleration support (CUDA/OpenGL/Metal)
- [ ] Model caching and warm-up on initialization

### Coordinate System
- [ ] Map MediaPipe's normalized [0,1] XY to world space
- [ ] Scale Z depth values appropriately
- [ ] Coordinate system transformation (MediaPipe → VMC coordinate space)
- [ ] Handle camera intrinsics for proper 3D projection

---

## 4. ARKit Blendshapes (Critical Feature)

### Blendshape Generation
- [ ] **Implement landmark-to-blendshape conversion algorithm**
  - This is the MOST CRITICAL missing piece for face tracking
- [ ] **Map 468 MediaPipe face mesh landmarks → 52 ARKit blendshapes**
  - Reference: [ARKit Blendshape Specification](https://arkit-face-blendshapes.com/)
  
#### Blendshape Categories (52 total):

**Eye Blendshapes (16):**
- [ ] `eyeBlinkLeft` - Left eye closure (landmarks 159, 145, etc.)
- [ ] `eyeBlinkRight` - Right eye closure (landmarks 386, 374, etc.)
- [ ] `eyeLookDownLeft` - Left eye looking down
- [ ] `eyeLookDownRight` - Right eye looking down
- [ ] `eyeLookInLeft` - Left eye looking inward
- [ ] `eyeLookInRight` - Right eye looking inward
- [ ] `eyeLookOutLeft` - Left eye looking outward
- [ ] `eyeLookOutRight` - Right eye looking outward
- [ ] `eyeLookUpLeft` - Left eye looking up
- [ ] `eyeLookUpRight` - Right eye looking up
- [ ] `eyeSquintLeft` - Left eye squint
- [ ] `eyeSquintRight` - Right eye squint
- [ ] `eyeWideLeft` - Left eye wide open
- [ ] `eyeWideRight` - Right eye wide open

**Jaw Blendshapes (4):**
- [ ] `jawOpen` - Mouth open (vertical distance between lips)
- [ ] `jawForward` - Jaw pushed forward
- [ ] `jawLeft` - Jaw moved to left
- [ ] `jawRight` - Jaw moved to right

**Mouth Blendshapes (20):**
- [ ] `mouthClose` - Lips pressed together
- [ ] `mouthFunnel` - Lips funneled (O-shape)
- [ ] `mouthPucker` - Lips puckered (kiss shape)
- [ ] `mouthSmileLeft` - Left side smile
- [ ] `mouthSmileRight` - Right side smile
- [ ] `mouthFrownLeft` - Left side frown
- [ ] `mouthFrownRight` - Right side frown
- [ ] `mouthDimpleLeft` - Left dimple
- [ ] `mouthDimpleRight` - Right dimple
- [ ] `mouthStretchLeft` - Left mouth stretch
- [ ] `mouthStretchRight` - Right mouth stretch
- [ ] `mouthRollLower` - Lower lip rolled in
- [ ] `mouthRollUpper` - Upper lip rolled in
- [ ] `mouthShrugLower` - Lower lip shrug
- [ ] `mouthShrugUpper` - Upper lip shrug
- [ ] `mouthPressLeft` - Left lip press
- [ ] `mouthPressRight` - Right lip press
- [ ] `mouthLowerDownLeft` - Lower left lip down
- [ ] `mouthLowerDownRight` - Lower right lip down
- [ ] `mouthUpperUpLeft` - Upper left lip up
- [ ] `mouthUpperUpRight` - Upper right lip up

**Nose Blendshapes (2):**
- [ ] `noseSneerLeft` - Left nose sneer
- [ ] `noseSneerRight` - Right nose sneer

**Cheek Blendshapes (4):**
- [ ] `cheekPuff` - Cheeks puffed out
- [ ] `cheekSquintLeft` - Left cheek raised (squint)
- [ ] `cheekSquintRight` - Right cheek raised (squint)

**Brow Blendshapes (6):**
- [ ] `browDownLeft` - Left brow down (angry)
- [ ] `browDownRight` - Right brow down
- [ ] `browInnerUp` - Inner brows up (sad/worried)
- [ ] `browOuterUpLeft` - Left outer brow up (surprised)
- [ ] `browOuterUpRight` - Right outer brow up

**Tongue Blendshape (1):**
- [ ] `tongueOut` - Tongue sticking out

#### Implementation Strategy:
- [ ] Create `blendshape.go` module
- [ ] Define landmark index groups for each blendshape
- [ ] Implement geometric calculations:
  - [ ] Distance-based (e.g., lip distance for `jawOpen`)
  - [ ] Angle-based (e.g., brow angle for `browInnerUp`)
  - [ ] Ratio-based (e.g., eye aspect ratio for blinks)
- [ ] Normalize blendshape values to [0.0, 1.0] range
- [ ] Add calibration system for neutral face baseline
- [ ] Test with various facial expressions
- [ ] Document landmark indices used for each blendshape

### Reference Materials
- [ ] Study MediaPipe Face Mesh topology
- [ ] Review ARKit documentation and sample videos
- [ ] Analyze VSeeFace/VNyan expected blendshape names
- [ ] Create test dataset with labeled expressions

---

## 5. Head and Body Pose Estimation

### Head Rotation
- [ ] **Implement head rotation estimation from face landmarks**
  - [ ] Use solvePnP or similar algorithm
  - [ ] Map to Quaternion for VMC protocol
  - [ ] Euler angle extraction (pitch, yaw, roll)
  - [ ] Smooth rotation with quaternion SLERP
- [x] Head rotation field exists in `FaceData`
- [ ] Convert MediaPipe pose landmarks to rotation

### Head Position
- [ ] Estimate head translation (X, Y, Z) from face size and position
- [ ] Depth estimation from landmark dispersion
- [ ] Calibrate to room-scale coordinates
- [x] Head position field exists in `FaceData`

### Body Pose
- [ ] Map MediaPipe pose landmarks (0-16) to VMC bones:
  - [ ] Spine/chest rotation (landmarks 11, 12, 23, 24)
  - [ ] Shoulder positions (landmarks 11, 12)
  - [ ] Neck position and rotation
- [ ] Upper body bone chain IK solving
- [ ] Shoulder rotation from arm pose

---

## 6. Hand Tracking

### Hand Landmarks
- [x] `HandData` structure defined with 21 landmarks
- [ ] **Connect MediaPipe hand landmarks to data structure**
- [ ] Left/right hand detection and assignment
- [ ] Hand confidence filtering (ignore low-confidence detections)

### Hand Bone Mapping
- [x] VMC hand bone names mapped in `sendHandBones()`
- [ ] **Calculate proper finger bone rotations from landmarks**
  - Currently only sends positions with identity quaternions
  - [ ] Compute finger joint angles
  - [ ] Convert angles to quaternions
  - [ ] Apply to thumb, index, middle, ring, pinky chains
- [ ] Hand pose estimation (open, closed, pointing, etc.)
- [ ] Gesture recognition (optional enhancement)

---

## 7. Kalman Filtering & Smoothing

### Current Implementation
- [x] `KalmanFilter` 1D implementation
- [x] `KalmanFilter3D` for 3D points
- [x] `LandmarkSmoother` for landmark arrays
- [x] Smoothing factor configuration (0.0-1.0)
- [x] Unit tests for Kalman filters

### Integration
- [ ] **Wire Kalman filters into tracking pipeline**
  - [ ] Apply to face landmarks before blendshape calculation
  - [ ] Apply to hand landmarks before bone calculation
  - [ ] Apply to pose landmarks before rotation estimation
  - [ ] Apply to head rotation quaternion (use SLERP, not linear Kalman)
- [ ] Adaptive smoothing based on movement velocity
- [ ] Separate smoothing factors for different body parts
- [ ] Predictive filtering for reduced latency

### Advanced Smoothing
- [ ] Implement quaternion-specific smoothing (SLERP-based)
- [ ] Velocity-based prediction for head rotation
- [ ] Outlier detection and rejection
- [ ] One-Euro filter as alternative to Kalman

---

## 8. VMC Protocol & OSC Sender

### Current Implementation
- [x] `VMCSender` struct with UDP connection
- [x] OSC message building (`buildOSCMessage`)
- [x] `/VMC/Ext/Bone/Pos` for head bone
- [x] `/VMC/Ext/Blend/Val` for blend shapes
- [x] `/VMC/Ext/Blend/Apply` signal
- [x] Hand bone sending (position only)
- [x] Unit tests for OSC encoding

### Missing Features
- [ ] **Send correct quaternion rotations for all bones**
  - Currently sends identity quaternions for hands
- [ ] **Add more VMC bone types:**
  - [ ] Neck
  - [ ] Spine/Chest
  - [ ] Left/Right Shoulder
  - [ ] Left/Right Upper Arm
  - [ ] Left/Right Lower Arm
- [ ] Time synchronization (`/VMC/Ext/T`)
- [ ] Root transform (`/VMC/Ext/Root/Pos`)
- [ ] Device info (`/VMC/Ext/Hmd/Pos`, `/VMC/Ext/Con/Pos`)
- [ ] Validate against VMC 2.x specification
- [ ] Test with VSeeFace, VNyan, VMagicMirror
- [ ] Configurable bone name mapping (different apps use different conventions)

### OSC Layer
- [x] Basic OSC encoding (strings, floats, ints)
- [ ] OSC bundles for atomic updates
- [ ] OSC timetags for synchronization
- [ ] Error handling for network failures

---

## 9. VRM Skeleton & Calibration

### Current Implementation
- [x] `VRMSkeleton` parsing from glTF/VRM files
- [x] Support for VRM 0.x and 1.0 formats
- [x] Extract bone hierarchy and transforms
- [x] Calculate body proportions (arm span, height, head size)
- [x] `GetProportions()` method
- [x] `GetBonePosition()` method
- [x] Unit tests for VRM parsing

### Missing Features
- [ ] **Use VRM proportions to scale tracking data**
  - [ ] Calibrate hand tracking to model arm length
  - [ ] Scale head position to model head size
  - [ ] Adjust shoulder width for pose tracking
- [ ] Support VRM spring bones (optional, for physics simulation)
- [ ] T-pose detection and calibration
- [ ] User-triggered recalibration command
- [ ] Save/load calibration profiles

---

## 10. Configuration & CLI

### Current Implementation
- [x] TOML configuration file support
- [x] Camera settings (device ID, resolution, FPS)
- [x] Tracking toggles (face, hands, pose)
- [x] Smoothing factor configuration
- [x] VMC sender settings (address, port)
- [x] Command-line flag overrides
- [x] Default configuration fallback
- [x] Config validation
- [x] Unit tests for config loading

### Enhancements
- [ ] Config hot-reload without restart
- [ ] Runtime adjustment of smoothing via OSC or web API
- [ ] Profile system (low-latency, high-quality, balanced)
- [ ] MediaPipe model complexity setting
- [ ] GPU device selection
- [ ] Multi-camera support

---

## 11. Testing

### Current Test Coverage
- [x] Config loading and validation tests
- [x] Tracker state machine tests
- [x] Kalman filter tests
- [x] OSC message encoding tests
- [x] VRM parsing tests
- [x] Mock component integration tests

### Missing Tests
- [ ] **Integration test with real MediaPipe processor**
  - Requires MediaPipe integration to be implemented first
- [ ] **Integration test with real camera capture**
- [ ] **Blendshape calculation tests**
  - [ ] Test each blendshape with known landmark positions
  - [ ] Validate output ranges [0.0, 1.0]
- [ ] **End-to-end test:**
  - [ ] Feed pre-recorded video frames
  - [ ] Verify OSC output matches expected values
- [ ] Performance benchmarks
  - [ ] Measure frames per second
  - [ ] Latency from capture to OSC send
  - [ ] CPU/memory profiling
- [ ] Stress tests (long-running stability)
- [ ] Test on different hardware (x86, ARM/Raspberry Pi)

---

## 12. Documentation

### Current Documentation
- [x] README with library and CLI usage
- [x] Package-level godoc
- [x] Inline code comments
- [x] Configuration file example

### Missing Documentation
- [ ] **Architecture diagram (capture → process → filter → send)**
- [ ] **Blendshape mapping reference table**
- [ ] **Coordinate system documentation**
- [ ] **VMC protocol message examples**
- [ ] Troubleshooting guide
- [ ] Performance tuning guide
- [ ] Comparison with other tracking solutions
- [ ] Video tutorials or GIFs
- [ ] Contributing guide
- [ ] Release notes / changelog

---

## 13. Performance Optimization

### Current State
- [x] Concurrent-safe design with goroutines
- [x] Non-blocking subscriber channels (drop frames if slow)
- [x] Minimal allocations in hot paths (mostly)

### Future Optimizations
- [ ] **GPU acceleration for MediaPipe**
  - [ ] CUDA support (NVIDIA)
  - [ ] OpenGL/Vulkan (cross-platform)
  - [ ] Metal (macOS)
- [ ] Profile and optimize blendshape calculations
- [ ] SIMD vectorization for landmark processing
- [ ] Object pooling for `TrackingData` structs
- [ ] Reduce GC pressure in tight loops
- [ ] Multi-threaded processing pipeline
- [ ] Frame skipping under high load
- [ ] Benchmark different Kalman filter implementations
- [ ] Consider Rust/C interop for critical paths

---

## 14. Cross-Platform Support

### Current Support
- [x] Linux (primary target)
- [x] Platform-agnostic architecture (interfaces)

### Platform-Specific Work
- [ ] **Linux:**
  - [ ] V4L2 camera support (via OpenCV or `webcam` lib)
  - [ ] Test on Ubuntu, Fedora, Arch
  - [ ] AppImage or Snap packaging
- [ ] **macOS:**
  - [ ] AVFoundation camera support
  - [ ] Metal GPU acceleration
  - [ ] Homebrew distribution
- [ ] **Windows:**
  - [ ] DirectShow/Media Foundation camera support
  - [ ] DirectX GPU acceleration
  - [ ] NSIS installer or MSI package
- [ ] **ARM/Raspberry Pi:**
  - [ ] Camera module support
  - [ ] Optimize for limited CPU/GPU
  - [ ] Pre-built binaries

---

## 15. Error Handling & Robustness

### Current State
- [x] Basic error propagation with wrapped errors
- [x] Graceful shutdown on signals (SIGINT, SIGTERM)
- [x] State machine prevents invalid operations

### Improvements
- [ ] Retry logic for camera initialization
- [ ] Automatic recovery from camera disconnection
- [ ] Fallback to lower resolution on performance issues
- [ ] Logging levels (debug, info, warn, error)
- [ ] Structured logging with `log/slog` or `zerolog`
- [ ] Telemetry/metrics (optional, Prometheus/OpenTelemetry)
- [ ] Crash reporting and diagnostics
- [ ] Watchdog timer for frozen tracking loop

---

## 16. Advanced Features (Future)

### Possible Enhancements
- [ ] **Expression presets and recording**
  - [ ] Record and playback facial expressions
  - [ ] Morph between presets
- [ ] **Multiplayer/multi-user support**
  - [ ] Track multiple faces in frame
  - [ ] Assign to different VMC ports
- [ ] **Body tracking extension**
  - [ ] Full-body pose (33 landmarks)
  - [ ] Lower body bones (hips, legs, feet)
- [ ] **Lighting estimation**
  - [ ] Send lighting data for realistic avatar rendering
- [ ] **Background removal**
  - [ ] Integrate MediaPipe Selfie Segmentation
  - [ ] Green screen replacement
- [ ] **AR effects and overlays**
  - [ ] Face filters, masks, accessories
- [ ] **Web UI for configuration and monitoring**
  - [ ] Real-time visualization of landmarks
  - [ ] Calibration wizard
  - [ ] Network discovery for VMC targets
- [ ] **Mobile support (Android/iOS)**
  - [ ] Use gomobile for cross-compilation
  - [ ] Native camera and MediaPipe integration
- [ ] **VR/OpenXR integration**
  - [ ] Send to OpenXR runtimes
  - [ ] Support for VR HMD tracking

---

## Priority Roadmap

### Phase 1: Core Functionality (MVP)
1. **Implement camera capture** (OpenCV/GoCV)
2. **Integrate MediaPipe Holistic** (CGO bindings)
3. **Implement ARKit blendshape conversion** (CRITICAL)
4. **Wire Kalman filtering into pipeline**
5. **Test end-to-end with VSeeFace**

### Phase 2: Polish & Testing
6. Add integration tests with real data
7. Performance profiling and optimization
8. Head rotation estimation (solvePnP)
9. Hand bone rotation calculations
10. Comprehensive documentation

### Phase 3: Production Ready
11. Cross-platform support (Windows, macOS)
12. Error handling and recovery
13. VRM calibration integration
14. Packaging and distribution
15. Community feedback and iteration

### Phase 4: Advanced Features
16. GPU acceleration
17. Web UI for configuration
18. Expression recording
19. Multi-user support
20. Mobile ports

---

## Development Notes

### Dependencies to Install
- `gocv.io/x/gocv` - OpenCV bindings for Go (camera + image processing)
- MediaPipe C++ library or Go bindings (TBD: need to choose approach)
- `github.com/BurntSushi/toml` - Already included ✓

### Build Requirements
- Go 1.24.10+ ✓
- OpenCV 4.x (for GoCV)
- MediaPipe (building from source may be required)
- C++ compiler (gcc/clang) for CGO
- pkg-config for library detection

### Testing Setup
- Test VRM models for calibration
- Pre-recorded video clips with known expressions
- OSC monitoring tool (e.g., `oscdump`, Protokol)
- VSeeFace or VNyan for end-to-end validation

---

## Contributing

When implementing features from this TODO:
1. Create feature branch from `main`
2. Implement feature with unit tests
3. Update this TODO.md (change `[ ]` to `[x]`)
4. Update relevant documentation
5. Submit PR with description referencing TODO item

---

**Last Updated:** 2025-11-26
**Version:** 0.1.0-dev
