.PHONY: build test clean install uninstall help

# Default target - build the project
all: build

# Display available targets
help:
	@echo "Available make targets:"
	@echo "  build           - Build the note binary"
	@echo "  test            - Run Go unit tests"
	@echo "  integration-test - Run integration tests"
	@echo "  completion-test - Run completion functionality tests"
	@echo "  setup-test      - Run setup/config integration tests"
	@echo "  setup-test-ci   - Run setup tests (non-failing for CI)"
	@echo "  test-all        - Run all tests"
	@echo "  test-all-ci     - Run all tests (non-failing for CI)"
	@echo "  install         - Install note system-wide (requires sudo)"
	@echo "  uninstall       - Remove note from system"
	@echo "  clean           - Clean build artifacts"
	@echo "  dev             - Build and show help"
	@echo "  fmt             - Format Go code"
	@echo "  vet             - Run go vet"

# Build variables
BINARY_NAME=note
INSTALL_PATH=/usr/local/bin

# Build the application
build:
	go build -o $(BINARY_NAME)

# Run unit tests
test:
	go test -v ./...

# Run integration tests
integration-test: build
	./scripts/integration_test.sh

# Run completion tests
completion-test: build
	./scripts/completion_test.sh

# Run setup integration tests
setup-test: build
	./scripts/setup_integration_test.sh

# Run setup integration tests (non-failing for CI)
setup-test-ci: build
	-./scripts/setup_integration_test.sh

# Run all tests
test-all: test integration-test completion-test setup-test-ci

# Run all tests (non-failing for CI) 
test-all-ci: test integration-test completion-test setup-test-ci

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)

# Install the binary system-wide (requires sudo)
install: build
	sudo cp $(BINARY_NAME) $(INSTALL_PATH)/
	@echo "Installing bash completions..."
	@if [ -d /etc/bash_completion.d ]; then \
		sudo cp completions/bash/note /etc/bash_completion.d/; \
		echo "Bash completions installed to /etc/bash_completion.d/"; \
	else \
		echo "Bash completion directory not found. Please manually source completions/bash/note"; \
	fi
	@echo "note installed to $(INSTALL_PATH)/$(BINARY_NAME)"

# Uninstall the binary
uninstall:
	sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	sudo rm -f /etc/bash_completion.d/note
	@echo "note uninstalled"

# Development - build and run with example
dev: build
	./$(BINARY_NAME) --help

# Format the code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...