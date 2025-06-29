# Go Engineering Guidelines

## Core Principles
- Follow official Go documentation and standards from https://go.dev/doc/
- Adhere to Effective Go guidelines: https://go.dev/doc/effective_go
- Use Go Code Review Comments: https://github.com/golang/go/wiki/CodeReviewComments
- Apply Go Proverbs: https://go-proverbs.github.io/

## Best Practices

### Code Style
- Use `gofmt` for consistent formatting
- Follow standard Go naming conventions (mixedCaps, not underscores)
- Keep interfaces small and focused
- Prefer composition over inheritance
- Return early to reduce nesting

### Error Handling
- Always check errors immediately after function calls
- Use descriptive error messages with context
- Wrap errors with `fmt.Errorf` and `%w` verb for error chains
- Consider custom error types for domain-specific errors

### Testing
- Write table-driven tests when appropriate
- Use subtests with `t.Run()` for better organization
- Keep tests close to the code they test
- Use testify/assert sparingly - prefer standard library testing
- Aim for high test coverage but focus on meaningful tests

### Concurrency
- Don't communicate by sharing memory; share memory by communicating
- Use channels for synchronization
- Protect shared state with mutexes when channels aren't appropriate
- Always handle goroutine lifecycle (no goroutine leaks)

### Performance
- Measure before optimizing
- Use benchmarks to validate improvements
- Consider memory allocation patterns
- Use sync.Pool for frequently allocated objects

### Dependencies
- Use Go modules for dependency management
- Minimize external dependencies
- Vendor dependencies for reproducible builds when needed
- Keep go.mod and go.sum files up to date

### Documentation
- Write clear godoc comments for all exported types and functions
- Start comments with the name of the thing being described
- Include examples in documentation when helpful
- Use code examples in _test.go files

## Project Structure
- Follow standard Go project layout patterns
- Keep packages focused and cohesive
- Use internal/ for private application code
- Separate concerns appropriately (cmd/, pkg/, internal/)

## Tools to Use
- `go fmt` - format code
- `go vet` - examine code for suspicious constructs
- `golangci-lint` - comprehensive linting
- `go test -race` - detect race conditions
- `go mod tidy` - clean up dependencies

## References
- Official Go Documentation: https://go.dev/doc/
- Go by Example: https://gobyexample.com/
- Go Wiki: https://github.com/golang/go/wiki
- Standard library source code as reference implementation