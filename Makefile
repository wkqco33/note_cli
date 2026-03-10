.PHONY: build install clean test run fmt help

# The name of the resulting executable
BINARY_NAME=note

build:
	@echo "Building the application..."
	go build -o $(BINARY_NAME) main.go

install:
	@echo "Installing the application to your local Go bin path..."
	go install

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
	@echo "  make build   - Build the Note CLI binary as '$(BINARY_NAME)'"
	@echo "  make install - Install the binary to your Go bin path"
	@echo "  make clean   - Remove the built binary"
	@echo "  make test    - Run all unit tests"
	@echo "  make fmt     - Format all Go source files"
	@echo "  make run     - Build and run the application (shows help command)"
