# MiFace Rust Rewrite TODO

This document tracks the Rust-first implementation for MiFace.

**Legend:**
- `[x]` - Implemented
- `[ ]` - Not yet implemented
- `[~]` - Partially implemented

---

## 1. Project Setup

- [ ] Create Cargo workspace layout
- [ ] Define crate boundaries (core, io, cli, osc, math)
- [ ] Add minimal CI checks (fmt/clippy/test)
- [ ] Add initial versioning and release tags

## 2. Core Types And Math

- [ ] Define core types: `Point3`, `Landmark`, `Quaternion`, `TrackingFrame`
- [ ] Coordinate transforms and normalization helpers
- [ ] Time and frame indexing utilities

## 3. Data Pipeline

- [ ] Frame input trait for camera/image sources
- [ ] Processor trait for landmark extraction
- [ ] Tracking loop with backpressure-safe channels
- [ ] Basic metrics for latency and FPS

## 4. MediaPipe Bridge (C++)

- [ ] Decide bridge style (`cxx` or `autocxx`)
- [ ] Define FFI boundary types
- [ ] Implement MediaPipe Holistic runner
- [ ] Map results to Rust structs
- [ ] Build scripts and dev setup docs

## 5. ARKit Blendshapes (Critical)

- [ ] Define blendshape schema (52 ARKit names)
- [ ] Implement eye, brow, mouth, jaw, cheek metrics
- [ ] Normalize to $[0,1]$ and clamp
- [ ] Add neutral-face calibration layer
- [ ] Add unit tests for each blendshape

## 6. Head And Body Pose

- [ ] Head rotation estimation (PnP or equivalent)
- [ ] Head position estimation
- [ ] Upper body pose mapping to VMC bones

## 7. Hand Tracking

- [ ] Assign left/right hands
- [ ] Compute finger joint rotations
- [ ] Convert to quaternions

## 8. Smoothing

- [ ] Kalman or One-Euro filter modules
- [ ] Quaternion smoothing via SLERP
- [ ] Per-part smoothing configuration

## 9. VMC / OSC Output

- [ ] OSC encoding layer
- [ ] VMC bone and blendshape sending
- [ ] Time sync messages
- [ ] Integration tests with sample frames

## 10. CLI And Config

- [ ] TOML config schema
- [ ] CLI commands (run, preview, record)
- [ ] Config validation

## 11. Testing And Benchmarks

- [ ] Unit tests for math and blendshapes
- [ ] Integration test with prerecorded frames
- [ ] Benchmarks for FPS and latency

## 12. Documentation

- [ ] Architecture diagram
- [ ] Blendshape reference table
- [ ] Coordinate system guide
- [ ] VMC message examples

## 13. Performance And Reliability

- [ ] GPU acceleration (CUDA, Metal, Vulkan/OpenGL)
- [ ] SIMD vectorization for landmark processing
- [ ] Multi-threaded processing pipeline
- [ ] Frame skipping under high load
- [ ] Logging levels (debug/info/warn/error)
- [ ] Recovery from camera disconnects

## 14. Cross-Platform Support

- [ ] Linux camera backend (V4L2 or OpenCV)
- [ ] macOS camera backend (AVFoundation)
- [ ] Windows camera backend (Media Foundation)
- [ ] ARM/Raspberry Pi builds

## 15. Advanced Features (Future)

- [ ] Expression presets and recording
- [ ] Multi-user tracking
- [ ] Full-body pose extension
- [ ] Background removal
- [ ] Web UI for configuration
- [ ] Mobile support
- [ ] VR/OpenXR integration

---

**Last Updated:** 2026-04-09
**Version:** 0.2.0-dev
