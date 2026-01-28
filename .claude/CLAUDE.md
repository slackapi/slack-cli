# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The Slack CLI is a command-line interface for building apps on the Slack Platform. Written in Go, it provides developers with tools to create, run, deploy, and manage Slack apps locally and remotely.

## Development Commands

### Building
```bash
make build              # Build the CLI (includes linting and cleaning)
make build-ci           # Build for CI (skips lint and tests)
./bin/slack --version   # Run the compiled binary
```

### Testing
```bash
make test                                          # Run all unit tests
make test testdir=cmd/auth testname=TestLoginCommand  # Run specific test
make coverage                                      # View test coverage report
```

### Linting
```bash
make lint               # Run golangci-lint
golangci-lint --version # Verify linter version
```

### Other Commands
```bash
make init               # Initialize project (fetch tags, dependencies)
make clean              # Remove build artifacts (bin/, dist/)
slack docgen ./docs/reference  # Generate command documentation
```

### Running Tests in Development
For targeted test runs during development:
```bash
go test -v ./cmd/version -run TestVersionCommand  # Run specific test directly
go test -ldflags="$(LDFLAGS)" -v ./cmd/auth -race -covermode=atomic -coverprofile=coverage.out
```

Note: When running tests directly with `go test`, use the LDFLAGS from the Makefile to inject version information.

## Architecture

### Project Structure

```
cmd/              Commands (user interface layer)
  ├── auth/       Authentication commands
  ├── app/        App management commands
  ├── platform/   Platform commands (deploy, run, activity)
  ├── project/    Project commands (create, init, samples)
  └── root.go     Root command initialization and alias mapping

internal/         Private packages (implementation details)
  ├── api/        Slack API client and HTTP communication
  ├── app/        App manifest and client logic
  ├── auth/       Authentication handling
  ├── config/     Configuration management
  ├── hooks/      Hook system for lifecycle events
  ├── iostreams/  I/O handling (stdin, stdout, stderr)
  ├── shared/     Shared client factory pattern
  └── experiment/ Feature flag system

main.go           Entry point with tracing and context setup
```

### Key Architectural Patterns

**Command Aliases**: Many commands have shortcut aliases defined in `cmd/root.go`'s `AliasMap`:
- `slack login` → `slack auth login`
- `slack deploy` → `slack platform deploy`
- `slack create` → `slack project create`

**Client Factory Pattern**: `internal/shared/clients.go` provides a `ClientFactory` that manages shared clients and configurations across commands:
- `API()` - Slack API client
- `Auth()` - Authentication client
- `AppClient()` - App management client
- `Config` - Configuration state
- `IO` - I/O streams
- `Fs` - File system (afero)

Commands receive the `ClientFactory` and use it to access functionality.

**Context-Based State**: The codebase uses `context.Context` extensively to pass runtime state (session IDs, trace IDs, versions) through the call stack via `internal/slackcontext`.

**Tracing**: OpenTracing (Jaeger) is initialized in `main.go` for observability.

**Hook System**: `internal/hooks/` implements a lifecycle hook system that allows SDK projects to inject custom behavior at key points.

**Experiment System**: Features can be gated behind experiments defined in `internal/experiment/experiment.go`:
- Add new experiments as constants
- Register in `AllExperiments` slice
- Enable permanently via `EnabledExperiments`
- Use `--experiment` flag to toggle

### Command Structure

Commands follow this pattern:
1. Define in `cmd/<category>/<command>.go`
2. Create a Cobra command with use/short/long descriptions
3. Add flags specific to that command
4. Implement `RunE` function that receives clients
5. Add unit tests in `*_test.go` alongside

Example command structure:
```go
func NewExampleCommand(clients *shared.ClientFactory) *cobra.Command {
    return &cobra.Command{
        Use:   "example",
        Short: "Brief description",
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := cmd.Context()
            // Command implementation using clients
            return nil
        },
    }
}
```

### Testing Philosophy

- Unit tests live alongside code with `_test.go` suffix
- Mock implementations use `_mock.go` suffix
- Test code uses syntax compatible with the minimum supported Go version (see `go.mod`)
- The codebase uses `testify` for assertions and `testify/mock` for mocking
- Mock the `ClientFactory` and its dependencies for testing
- Always mock file system operations using `afero.Fs` to enable testability

### Table-Driven Test Conventions

**Preferred: Slice pattern** - uses `tc` for test case variable:
```go
tests := []struct {
    name     string
    input    string
    expected string
}{...}
for _, tc := range tests {
    t.Run(tc.name, func(t *testing.T) {
        // use tc.field
    })
}
```

**Legacy: Map pattern** - uses `tt` for test case variable (do not use for new tests):
```go
tests := map[string]struct {
    input    string
    expected string
}{...}
for name, tt := range tests {
    t.Run(name, func(t *testing.T) {
        // use tt.field
    })
}
```

## Version Management

Versions use semantic versioning with git tags (format: `v*.*.*`).

Version is generated dynamically using `git describe` and injected at build time:
```bash
LDFLAGS=-X 'github.com/slackapi/slack-cli/internal/pkg/version.Version=`git describe --tags --match 'v*.*.*'`'
```

**Versioning Rules**:
- `semver:patch` - Bug fixes and changes behind experiment flags
- `semver:minor` - New features (once experiments are removed)
- `semver:major` - Breaking changes

## Deprecation Process

When deprecating features, commands, or flags:

1. **Commands**: Add `Deprecated` attribute, optionally set `Hidden: true`
2. **Flags**: Print deprecation warning on use, hide with `.Hidden` attribute
3. **Public functionality**: Add comment `// DEPRECATED(semver:major): <description and migration path>`
4. **Internal functionality**: Add comment `// DEPRECATED: <description>`

## Important Configuration Files

- `.golangci.yml` - Linter configuration with custom initialisms (CLI, API, SDK, etc.) and staticcheck rules
- `.goreleaser.yml` - Release build configuration for multi-platform binaries
- `Makefile` - Common development tasks and build scripts
- `go.mod` - Go module dependencies and minimum Go version (see `go.mod` for current version)
- `.circleci/config.yml` - CircleCI workflows for CI/CD pipeline
- `.github/workflows/` - GitHub Actions for automated testing and releases

## Commit Message Format

When creating commits and PRs, append to the end of the commit message:
```
Co-Authored-By: Claude <svc-devxp-claude@slack-corp.com>
```

Use conventional commit format (feat:, fix:, chore:, etc.) for commit titles.

## Working with the Codebase

### Adding a New Command

1. Create `cmd/<category>/<command>.go`
2. Implement command using `NewXCommand(clients *shared.ClientFactory) *cobra.Command`
3. Register in `cmd/root.go` `Init()` function
4. Add tests in `cmd/<category>/<command>_test.go`
5. Run `slack docgen ./docs/reference` to generate docs
6. Consider adding command alias in `AliasMap` if appropriate

### Adding New Dependencies

1. Update `go.mod` with the new module version
2. Run `go mod tidy` to update `go.sum`
3. Use `go mod graph | grep <module>` to inspect dependency tree

### Understanding API Calls

All Slack API interactions go through `internal/api/`:
- `client.go` - HTTP client setup
- `app.go` - App management API calls
- `auth.go` - Authentication API calls
- `datastore.go` - Datastore API calls
- Each file has corresponding `*_test.go` with mocks

### File System Operations

Always use `clients.Fs` (afero.Fs) instead of direct `os` calls to enable testing and mocking.

## Development Notes

- The CLI binary in development builds is at `./bin/slack`
- Official releases use `/usr/local/bin/slack`
- Set `SLACK_DISABLE_TELEMETRY="true"` to disable telemetry during development
- View experiments with `slack --help --verbose`
- Build artifacts are in `bin/` (local builds) and `dist/` (release builds)
- The `make build` command runs linting automatically before building
- When testing locally, always use `./bin/slack` to avoid conflicts with globally installed versions
