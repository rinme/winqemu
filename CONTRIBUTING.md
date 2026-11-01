# Contributing to WinQEMU ⚡

Thank you for your interest in contributing to **WinQEMU**! We welcome bug reports, feature suggestions, and pull requests.

## 🛠️ Development Setup

### Prerequisites
- Go 1.22+
- QEMU / KVM (on Linux) or equivalent hypervisor tools
- `make`

### Building & Testing Locally
```bash
# Clone your fork
git clone https://github.com/rinme/winqemu.git
cd winqemu

# Run unit tests across all packages
make test

# Build executable
make build

# Launch the interactive TUI
./bin/winqemu
```

## 📋 Pull Request Guidelines

1. **Keep it focused**: One bug fix or feature per pull request.
2. **Add unit tests**: Ensure all changes have corresponding unit tests in `pkg/*` or `internal/*`.
3. **Verify tests pass**: Run `make test` before pushing.
4. **Follow Go conventions**: Run `go fmt ./...` and `go vet ./...`.

## 🐛 Reporting Bugs

If you find a bug or issue (e.g., driver loading issue, package detection error on a specific distro, or XML formatting issue):
- Open an issue using the [Bug Report](https://github.com/rinme/winqemu/issues/new?template=bug_report.yml) template.
- Include your host OS, distribution, and the Windows version you attempted to create.

## 💡 Feature Requests & Catalog Additions

To suggest a new Windows edition or performance tuning flag:
- Submit an issue using the [Feature Request](https://github.com/rinme/winqemu/issues/new?template=feature_request.yml) template.
- Or start a conversation in [GitHub Discussions](https://github.com/rinme/winqemu/discussions).
