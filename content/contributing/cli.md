---
title: "Contributing to LVT CLI"
source_repo: "https://github.com/livetemplate/lvt"
source_path: "CONTRIBUTING.md"
source_ref: "v0.1.7"
source_commit: "e132b0dbd81c773d3fedf7980b306cf76f50b2f6"
---

# Contributing to LVT CLI

Thank you for your interest in contributing to the LVT CLI!

## Development Setup

### Prerequisites

- Go 1.25 or higher
- Git
- (Optional) GoReleaser for releases

### Getting Started

```bash
# Clone the repository
git clone https://github.com/livetemplate/lvt.git
cd lvt

# Install dependencies
go mod download

# Install git hooks
./scripts/install-hooks.sh

# Run tests
go test ./...

# Build
go build -o lvt .
```

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### 2. Make Changes

- Write code following Go best practices
- Add tests for new functionality
- Update documentation as needed
- Ensure all tests pass

### 3. Local Development with Core Library

If you're making changes that depend on unreleased core library changes, you have two options:

#### Recommended: Go Workspace (Automatic)

The **easiest way** - Go automatically uses local modules without any `go.mod` changes:

```bash
# From parent directory containing all repos
cd ..
./setup-workspace.sh

# Now test with local core library
cd lvt
go test ./...  # Automatically uses ../livetemplate
```

The workspace setup is done once and affects all repositories. See the [core library CONTRIBUTING.md](/contributing/livetemplate#testing-core-changes-with-lvtexamples) for details.

#### Alternative: Manual Replace Directives

If you prefer manual control:

```bash
# Enable local development mode
./scripts/setup-local-dev.sh

# Test with local core
go test ./...

# Revert to published version
./scripts/setup-local-dev.sh --undo
```

**Directory structure for both methods:**
```
parent/
├── livetemplate/  (core library)
├── lvt/           (this repo)
└── examples/      (optional)
```

### 4. Test Your Changes

```bash
# Run all tests
go test ./...

# Run with timeout
go test ./... -timeout=120s

# Run specific package
go test ./internal/generator -v

# Build to verify
go build -o lvt .
```

### 5. Commit Your Changes

The repository has a pre-commit hook that will:
- Auto-format Go code
- Run linter (if golangci-lint is installed)
- Run all tests

```bash
git add .
git commit -m "feat: add new feature"
```

#### Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `test:` Adding or updating tests
- `refactor:` Code refactoring
- `perf:` Performance improvements
- `chore:` Build process or tooling changes

### 6. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub.

## Code Style

### Go Conventions

- Follow standard Go formatting (`go fmt`)
- Use meaningful variable names
- Keep functions focused and small
- Document exported functions and types
- Handle errors explicitly

### Project Structure

```
lvt/
├── main.go                 # Entry point
├── commands/               # CLI commands
├── internal/              # Internal packages
│   ├── generator/         # Code generators
│   ├── kits/             # Kit system
│   ├── config/           # Configuration
│   ├── validator/        # Validation
│   └── serve/            # Development server
├── testing/              # Testing utilities
├── e2e/                  # End-to-end tests
└── scripts/              # Build and release scripts
```

## Testing Guidelines

### Test Categories

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test package interactions
3. **E2E Tests**: Test complete workflows

### Writing Tests

```go
func TestGenerateResource(t *testing.T) {
    // Arrange
    gen := generator.New()

    // Act
    err := gen.GenerateResource("Post", fields)

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

### Test Coverage

- Aim for >70% code coverage
- Test happy paths and error cases
- Test edge cases

## Adding New Features

### 1. Generators

Add to `internal/generator/`:

```go
// internal/generator/myfeature.go
func (g *Generator) GenerateMyFeature(name string) error {
    // Implementation
}
```

### 2. Commands

Add to `commands/`:

```go
// commands/mycommand.go
var myCmd = &cobra.Command{
    Use:   "my",
    Short: "Description",
    Run:   runMy,
}

func runMy(cmd *cobra.Command, args []string) {
    // Implementation
}
```

### 3. Kits

Kits are located in `internal/kits/system/`.

To add a new kit:
1. Create directory: `internal/kits/system/mykit/`
2. Add `kit.yaml` manifest
3. Add CSS helpers
4. Add component templates
5. Add generator templates

## Versioning

LVT follows the core library's major.minor version:

- **Patch versions**: Independent, for LVT-specific fixes
- **Minor versions**: Match core library minor version
- **Major versions**: Match core library major version

## Release Process

Releases are automated via `scripts/release.sh`:

```bash
# Dry run
./scripts/release.sh --dry-run

# Actual release (maintainers only)
./scripts/release.sh
```

The script will:
1. Validate version against core library
2. Update VERSION file
3. Generate CHANGELOG
4. Run tests and build
5. Commit and tag
6. Push to GitHub
7. Run GoReleaser to build binaries and create release

## Core Library Coordination

When the core library changes:

1. **Protocol changes**: Update client code to handle new formats
2. **API changes**: Update generators and templates
3. **Breaking changes**: Coordinate version bump

## Documentation

### README.md

Update for:
- New commands
- New features
- Configuration changes
- Examples

### Code Comments

- Document exported functions
- Explain complex logic
- Keep comments up-to-date

## Getting Help

- **Questions**: [GitHub Discussions](https://github.com/livetemplate/lvt/discussions)
- **Bugs**: [GitHub Issues](https://github.com/livetemplate/lvt/issues)
- **Core Library**: [LiveTemplate Repo](https://github.com/livetemplate/livetemplate)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
