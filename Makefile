.PHONY: build test coverage clean run lint

# Build the binary
build:
	go build -o build/gocamelpack .

# Run tests
test:
	go test ./...

# Run tests with coverage
coverage:
	go test -coverprofile=build/cover.out ./...
	go tool cover -html=build/cover.out -o build/coverage.html

# Clean build artifacts
clean:
	rm -rf build/

# Run the application
run:
	go run .

# Run linting
lint:
	go vet ./...
	golangci-lint run