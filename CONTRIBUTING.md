# Contributing to SitePod

Thanks for your interest in contributing! This guide will help you get started.

## Development Setup

### Prerequisites

- **Go** 1.21+ (server)
- **Rust** stable (CLI)
- **Make** (build orchestration)
- **Docker** (optional, for container builds)

### Quick Start

```bash
# Clone the repo
git clone https://github.com/sitepod-dev/sitepod.git
cd sitepod

# Build everything (server + CLI)
make quick-start

# Start the server (http://localhost:8080)
make run

# In another terminal, test with the CLI
./bin/sitepod login --endpoint http://localhost:8080
cd examples/simple-site
../../bin/sitepod deploy
```

See [CLAUDE.md](./CLAUDE.md) for the full development reference, including all Make commands, architecture details, and environment variables.

## Reporting Issues

- Search [existing issues](https://github.com/sitepod-dev/sitepod/issues) first to avoid duplicates.
- Use a clear, descriptive title.
- Include steps to reproduce, expected vs actual behavior, and your environment (OS, Go/Rust version).

## Pull Requests

1. Fork the repo and create a branch from `main`.
2. Make your changes with clear, focused commits.
3. Ensure all tests pass (`make test`).
4. Ensure linting passes (`make lint`).
5. Open a PR with a description of **what** and **why**.

Keep PRs small and focused — one feature or fix per PR is ideal.

## Code Style

### Go (Server)

- Follow standard Go conventions (`gofmt`, `go vet`).
- Run `make lint-server` (uses `golangci-lint`).
- Write table-driven tests where appropriate.

### Rust (CLI)

- Run `cargo fmt` before committing.
- Run `cargo clippy -- -D warnings` — no warnings allowed.
- Run `make lint-cli` to check both formatting and clippy.

## Testing

All changes should include tests when applicable:

```bash
make test           # Run all tests (Go + Rust)
make test-server    # Go tests only
make test-cli       # Rust tests only
./test-e2e.sh       # End-to-end tests
```

CI runs the full test suite and linters on every PR — make sure they pass before requesting review.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](./LICENSE).
