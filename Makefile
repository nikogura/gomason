.PHONY: lint test test-all build build-all clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

PLATFORMS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

# Lint runs the custom namedreturns linter followed by golangci-lint.
# namedreturns runs first: if it fails, there is no point running the rest.
lint:
	@echo "Running namedreturns linter..."
	namedreturns ./...
	@echo "Running golangci-lint..."
	golangci-lint run

# Test runs all unit tests.
test:
	go test -v -race -cover -count=1 ./...

# Test-all runs unit tests and lint.
test-all: test lint

# Build builds gomason for the host platform.
build:
	@echo "Building gomason ($(VERSION))..."
	CGO_ENABLED=0 go build -o bin/gomason .

# Build-all builds gomason for every released platform.
build-all:
	@echo "Building gomason for all platforms ($(VERSION))..."
	@for pair in $(PLATFORMS); do \
		OS=$$(echo $$pair | cut -d/ -f1); \
		ARCH=$$(echo $$pair | cut -d/ -f2); \
		echo "  $$OS/$$ARCH..."; \
		mkdir -p dist/$$OS/$$ARCH; \
		CGO_ENABLED=0 GOOS=$$OS GOARCH=$$ARCH \
			go build -o dist/$$OS/$$ARCH/gomason . || exit 1; \
	done
	@echo "Binaries built in dist/"

# Clean removes build artifacts.
clean:
	@echo "Cleaning build artifacts..."
	rm -rf dist/ bin/ coverage.txt
