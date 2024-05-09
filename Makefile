# Set default target
.DEFAULT_GOAL := build

# Define variables
BINARY_NAME := instructor
BINARY_DIR := bin
GO_SOURCES := $(shell find . -name '*.go')
GO_LINT_TOOLS := $(shell go env GOPATH)/bin/golangci-lint

# Build target
build: $(GO_SOURCES)
	mkdir -p $(BINARY_DIR)
	go build -o $(BINARY_DIR)/$(BINARY_NAME)

# Install dependencies
install-deps:
	go get -u golang.org/x/lint/golint

# Clean target
.PHONY: clean
clean:
	rm -rf $(BINARY_DIR)

# Format target
.PHONY: fmt
fmt:
	go fmt ./...

# Lint target
.PHONY: lint
lint: install-deps
	$(GO_LINT_TOOLS) run

# Test target
.PHONY: test
test:
	go test ./...

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build       Build the binary (default)"
	@echo "  clean       Remove the binary"
	@echo "  fmt         Format the source code"
	@echo "  lint        Run linter checks"
	@echo "  test        Run tests"
	@echo "  help        Show this help message"
