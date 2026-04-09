# Contributing to MiFace

Thanks for contributing. MiFace is in a Rust rewrite, so expect rapid changes and moving targets. Short, focused PRs are preferred.

## Development Setup

### Prerequisites

- Rust stable toolchain (rustup recommended)
- Git

Some parts will later require MediaPipe/OpenCV, but the rewrite starts with pure Rust modules.

### Getting Started

1. Fork the repository on GitHub
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/miface.git
   cd miface
   ```
3. Add upstream remote:
   ```bash
   git remote add upstream https://github.com/MiFaceDEV/miface.git
   ```
4. Build and test:
   ```bash
   cargo build
   cargo test
   ```

## Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-number-description
```

### 2. Make Your Changes

- Write clean, idiomatic Rust
- Keep modules small and testable
- Update documentation when behavior changes

### 3. Format, Lint, Test

```bash
cargo fmt
cargo clippy --all-targets --all-features -- -D warnings
cargo test
```

### 4. Commit Your Changes

```bash
git add .
git commit -m "Add blendshape mapping baseline"
```

Commit message format:
```
Brief summary (50 chars or less)

More detail if needed. Explain what and why, not how.
Fixes #123
```

### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

PRs should include:
- Clear title and summary
- Tests added or updated (when relevant)
- Notes about behavior changes

## Code Style Guidelines

- `rustfmt` is required
- `clippy` must pass with `-D warnings`
- Prefer explicit types at public API boundaries
- Avoid allocations in hot paths unless justified
- Document non-obvious math and coordinate transforms

## Tests

- Use unit tests for geometry and blendshape math
- Add fixture-based tests for landmark pipelines when available
- Keep tests deterministic

## License

By contributing to MiFace, you agree that your contributions will be licensed under the AGPL-3.0 License.
