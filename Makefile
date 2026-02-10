.PHONY: lint test clean

# Lint runs custom namedreturns linter followed by golangci-lint
lint:
	@echo "Running namedreturns linter..."
	namedreturns ./...
	@echo "Running golangci-lint..."
	golangci-lint run

# Test runs all unit tests
test:
	go test -v -race -cover -count=1 ./...

# Test-all runs unit tests and lint
test-all: test lint

# Clean removes build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf dist/
