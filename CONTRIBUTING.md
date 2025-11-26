# Contributing to MiFace

Thank you for your interest in contributing to MiFace! This document provides guidelines and instructions for contributing.

## Development Setup

### Prerequisites

- Go 1.21 or later
- OpenCV 4.x with development headers
- GCC/Clang for CGO
- Git

See [BUILDING.md](BUILDING.md) for detailed platform-specific setup instructions.

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
4. Install dependencies:
   ```bash
   make deps
   ```
5. Build and test:
   ```bash
   make build
   make test
   ```

## Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-number-description
```

### 2. Make Your Changes

- Write clean, idiomatic Go code
- Follow the existing code style
- Add tests for new functionality
- Update documentation as needed

### 3. Test Your Changes

```bash
# Run all tests
make test

# Run tests with race detector
make test

# Run specific tests
go test -v -run TestName ./pkg/miface/

# Check test coverage
make test-coverage
```

### 4. Format and Lint

```bash
# Format code
make fmt

# Run go vet
make vet

# Run linter (requires golangci-lint)
make lint
```

### 5. Commit Your Changes

Use clear, descriptive commit messages:

```bash
git add .
git commit -m "Add support for XYZ feature"
```

Good commit message format:
```
Brief summary (50 chars or less)

More detailed explanation if needed. Wrap at 72 characters.
Explain what and why, not how.

Fixes #123
```

### 6. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub with:
- Clear title describing the change
- Description explaining what and why
- Link to related issues (if any)
- Screenshots/examples if applicable

## Code Style Guidelines

### General

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting (run `make fmt`)
- Keep functions small and focused
- Write self-documenting code with clear variable names
- Add comments for complex logic

### Documentation

- All exported functions/types must have godoc comments
- Comments should start with the function/type name
- Use complete sentences
- Example:
  ```go
  // NewTracker creates a new tracker with the given configuration.
  // If cfg is nil, default configuration is used.
  func NewTracker(cfg *config.Config) (*Tracker, error) {
  ```

### Error Handling

- Return errors, don't panic (except in init or fatal setup)
- Wrap errors with context using `fmt.Errorf`
- Use sentinel errors for expected error conditions
- Example:
  ```go
  if err != nil {
      return fmt.Errorf("failed to open camera: %w", err)
  }
  ```

### Testing

- Write table-driven tests when appropriate
- Use meaningful test names: `TestFunctionName_Scenario`
- Test both success and failure cases
- Use subtests for multiple scenarios
- Example:
  ```go
  func TestTracker_Start(t *testing.T) {
      t.Run("success", func(t *testing.T) { ... })
      t.Run("already_running", func(t *testing.T) { ... })
  }
  ```

### Concurrency

- Use channels for communication between goroutines
- Protect shared state with mutexes
- Document goroutine lifetimes and ownership
- Avoid goroutine leaks (ensure all goroutines exit)

## Project Structure

```
miface/
├── cmd/miface/          # CLI application
│   └── main.go
├── pkg/miface/          # Public library API
│   ├── tracker.go       # Main tracker
│   ├── camera_gocv.go   # Camera implementation
│   ├── kalman.go        # Kalman filtering
│   ├── sender.go        # VMC protocol
│   └── vrm.go           # VRM parsing
├── internal/config/     # Internal config package
│   └── config.go
└── TODO.md              # Development roadmap
```

## Implementing TODO Items

When working on items from [TODO.md](TODO.md):

1. Check if the item is already claimed in an open PR
2. Create an issue describing your approach (for large features)
3. Implement the feature with tests
4. Update TODO.md (change `[ ]` to `[x]`)
5. Submit PR referencing the TODO item and any related issues

## Common Development Tasks

### Adding a New Camera Backend

1. Implement the `CameraSource` interface in `pkg/miface/`
2. Add tests in `*_test.go`
3. Update documentation

### Adding Protocol Support

1. Implement the `Sender` interface in `pkg/miface/`
2. Add protocol-specific message building
3. Add configuration options in `internal/config/`
4. Update README with usage examples

### Adding a Processor Backend

1. Implement the `Processor` interface
2. Handle landmark detection and conversion
3. Add benchmark tests for performance
4. Document performance characteristics

## Pull Request Checklist

Before submitting your PR, ensure:

- [ ] Code compiles without errors
- [ ] All tests pass (`make test`)
- [ ] Code is formatted (`make fmt`)
- [ ] No linter warnings (`make vet`)
- [ ] New code has tests
- [ ] Documentation is updated
- [ ] TODO.md updated if implementing tracked item
- [ ] Commit messages are clear
- [ ] PR description explains changes

## Questions or Issues?

- Check existing issues on GitHub
- Create a new issue for bugs or feature requests
- For discussions, use GitHub Discussions
- Be respectful and constructive

## License

By contributing to MiFace, you agree that your contributions will be licensed under the AGPL-3.0 License.

## Code of Conduct

Be kind, be respectful, be constructive. We're all here to build something cool together.
