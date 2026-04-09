# MiFace

Real-time facial and upper-body tracking for VTubers, rewritten in Rust for low-latency, predictable performance, and clean C++ interoperability.

## Status

MiFace is currently in a Rust rewrite. Expect breaking changes and incomplete functionality while the new core is built.

## Goals

- Predictable frame times (no GC pauses)
- High-performance C++ interop for MediaPipe/OpenCV
- Clean, testable core for face, hand, and pose tracking
- First-class VMC/OSC output support

## What To Expect Soon

- Face mesh ingestion and landmark pipeline
- ARKit blendshape mapping from MediaPipe face mesh
- VMC/OSC sender
- CLI and config system

## Build And Run

See [BUILDING.md](BUILDING.md) for the current Rust build instructions and system dependencies.

## Contributing

Start with [CONTRIBUTING.md](CONTRIBUTING.md). The rewrite is moving fast, so short, focused PRs are preferred.

## License

AGPL-3.0 - See [LICENSE](LICENSE) for details.

