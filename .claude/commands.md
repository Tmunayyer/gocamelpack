# CLAUDE.md - Development Commands

This file contains important development commands for the gocamelpack project.

## Available Make Commands

### Build
```bash
make build
```
Builds the binary to `build/gocamelpack`

### Test
```bash
make test
```
Runs all tests

### Coverage
```bash
make coverage
```
Runs tests with coverage and generates HTML coverage report at `build/coverage.html`

### Lint
```bash
make lint
```
Runs Go vet and golangci-lint for code quality checks

### Run
```bash
make run
```
Runs the application directly

### Clean
```bash
make clean
```
Removes build artifacts from the `build/` directory

## Development Workflow

When making changes:
1. Run `make lint` to check code quality
2. Run `make test` to ensure tests pass
3. Run `make build` to verify the build succeeds