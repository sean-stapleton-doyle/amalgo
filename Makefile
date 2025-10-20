BINARY_NAME=amalgo
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html
GO=go
GOFLAGS=
LDFLAGS=


.PHONY: all build install test test-verbose coverage coverage-html clean fmt vet lint help

all: fmt vet test build

help:
	@echo "Amalgo - Makefile Commands"
	@echo ""
	@echo "Available targets:"
	@grep -E '^

build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(GOFLAGS) -ldflags "$(LDFLAGS)" .
	@echo "Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

test:
	@echo "Running tests..."
	$(GO) test ./...
	@echo "Tests passed"

test-verbose:
	@echo "Running tests (verbose)..."
	$(GO) test -v ./...

test-short:
	@echo "Running short tests..."
	$(GO) test -short ./...

coverage:
	@echo "Running tests with coverage..."
	$(GO) test -coverprofile=$(COVERAGE_FILE) ./...
	@echo "Coverage report generated: $(COVERAGE_FILE)"
	@$(GO) tool cover -func=$(COVERAGE_FILE) | tail -n 1

coverage-html: coverage
	@echo "Generating HTML coverage report..."
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "HTML coverage report: $(COVERAGE_HTML)"
	@if command -v xdg-open > /dev/null; then \
		xdg-open $(COVERAGE_HTML); \
	elif command -v open > /dev/null; then \
		open $(COVERAGE_HTML); \
	else \
		echo "Open $(COVERAGE_HTML) manually in your browser"; \
	fi

fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Code formatted"

vet:
	@echo "Running go vet..."
	$(GO) vet ./...
	@echo "Vet checks passed"

lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
		echo "Lint checks passed"; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

tidy:
	@echo "Tidying go.mod..."
	$(GO) mod tidy
	$(GO) mod verify
	@echo "Dependencies tidied"

clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -f $(COVERAGE_FILE)
	@rm -f $(COVERAGE_HTML)
	@rm -rf dist/
	@echo "Cleaned"

deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod verify
	@echo "Dependencies downloaded"

check: fmt vet lint test
	@echo "All checks passed"

ci: fmt vet test coverage build
	@echo "CI pipeline complete"

run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME) -e .go -o example_output.md

bench:
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

profile-cpu:
	@echo "Running CPU profiling..."
	$(GO) test -cpuprofile=cpu.prof ./...
	@echo "View profile with: go tool pprof cpu.prof"

profile-mem:
	@echo "Running memory profiling..."
	$(GO) test -memprofile=mem.prof ./...
	@echo "View profile with: go tool pprof mem.prof"

watch:
	@if command -v entr > /dev/null; then \
		echo "Watching for changes..."; \
		find . -name '*.go' | entr -c make test; \
	else \
		echo "entr not installed. Install with your package manager"; \
		echo "  macOS: brew install entr"; \
		echo "  Linux: apt install entr / yum install entr"; \
	fi

info:
	@echo "$Build Information:"
	@echo "  Go version:    $(shell $(GO) version)"
	@echo "  Binary name:   $(BINARY_NAME)"
	@echo "  GOPATH:        $(shell go env GOPATH)"
	@echo "  GOROOT:        $(shell go env GOROOT)"
	@echo "  Module:        $(shell head -n 1 go.mod | cut -d' ' -f2)"