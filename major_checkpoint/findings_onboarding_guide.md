# Developer Onboarding Guide

## Prerequisites

- Go 1.24 or higher
- Git
- Make (optional, but recommended)

## Getting Started

```bash
# Clone the repository
git clone https://github.com/Les-El/hashi.git

# Install dependencies
go mod download

# Run tests
go test ./...

# Build the project
go build -o hashi ./cmd/hashi
```

## Project Architecture

- `cmd/hashi`: The main entry point and CLI command definitions.
- `internal/`: Private packages containing the core logic:
  - `checkpoint`: The major checkpoint analysis system.
  - `config`: Configuration parsing and flag management.
  - `hash`: Core hashing algorithms and file processing.
  - `conflict`: Flag conflict resolution logic.
- `docs/`: Comprehensive project documentation and ADRs.

## Coding Standards

1. **Testing**: All new features must include unit tests and, where appropriate, property-based tests.
2. **Documentation**: All exported functions and types must have descriptive Go documentation.
3. **Error Handling**: Use custom error types and avoid `panic` for expected error conditions.
4. **Formatting**: Always run `go fmt` before committing.
