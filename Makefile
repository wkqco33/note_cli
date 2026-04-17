.PHONY: build install uninstall clean test run fmt help

# The name of the resulting executable
BINARY_NAME=note_cli

# Version info from git tag
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# LDFLAGS for injecting build info
LDFLAGS=-ldflags "-X 'note_cli/config/buildinfo.Version=${VERSION}'"

# Installation prefix
PREFIX ?= /usr/local

build:
	@echo "Building the application..."
	go build $(LDFLAGS) -o $(BINARY_NAME) main.go

install:
	@echo "Installing the application to $(PREFIX)/bin..."
	mkdir -p $(PREFIX)/bin
	install -m 755 $(BINARY_NAME) $(PREFIX)/bin/

uninstall:
	@echo "Uninstalling the application from $(PREFIX)/bin..."
	rm -f $(PREFIX)/bin/$(BINARY_NAME)

clean:
	@echo "Cleaning up..."
	go clean
	rm -f $(BINARY_NAME)

test:
	@echo "Running tests..."
	go test ./... -v

fmt:
	@echo "Formatting code..."
	go fmt ./...

run: build
	@echo "Running the application (showing help by default)..."
	./$(BINARY_NAME) --help

help:
	@echo "Available commands:"
	@echo "  make build     - Build the Note CLI binary as '$(BINARY_NAME)'"
	@echo "  make install   - Install the binary to \$$(PREFIX)/bin (default: /usr/local/bin)"
	@echo "  make uninstall - Uninstall the binary from \$$(PREFIX)/bin"
	@echo "  make clean     - Remove the built binary"
	@echo "  make test      - Run all unit tests"
	@echo "  make fmt       - Format all Go source files"
	@echo "  make run       - Build and run the application (shows help command)"
