set dotenv-load := false

# Default recipe: list available recipes
default:
    @just --list

# Run CLI in human mode (default)
run *args:
    go run ./cmd/go-scaffold {{ args }}

# Run CLI in agent-facing mode
run-ai *args:
    AGENT=1 go run ./cmd/go-scaffold {{ args }}

# Run test suite
test: setup
    go test -v ./...

# Build binary into bin/
build:
    mkdir -p bin
    go build -o bin/go-scaffold ./cmd/go-scaffold

# Install binary to GOPATH bin directory
install:
    go install ./cmd/go-scaffold

# Run static analysis and vet
lint:
    go vet ./...

# Alias for lint
vet: lint

# Format Go code
fmt:
    go fmt ./...

# Run static analysis and test suite in sequence
check: setup
    go vet ./...
    go test -v ./...

# Clean build and coverage outputs
clean:
    rm -rf bin coverage.out dist

# Download project dependencies for local development
setup:
    go mod download
