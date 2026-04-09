# Building MiFace

MiFace is in a Rust rewrite. The build steps below focus on the new Rust core and will evolve as MediaPipe/OpenCV integration lands.

## Prerequisites

- Rust stable toolchain (install via rustup)
- Git

Optional (future): MediaPipe/OpenCV toolchains for C++ interop.

## Build

```bash
# Clone the repository
git clone https://github.com/MiFaceDEV/miface
cd miface

# Build the workspace
cargo build
```

## Test

```bash
cargo test
```

## Format And Lint

```bash
cargo fmt
cargo clippy --all-targets --all-features -- -D warnings
```

## Troubleshooting

### rustfmt or clippy missing

```bash
rustup component add rustfmt clippy
```

## Notes

As the rewrite progresses, this guide will include MediaPipe/OpenCV setup and runtime instructions.
