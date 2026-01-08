.PHONY: build test clean install uninstall

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
	./integration_test.sh

# Run all tests
test-all: test integration-test

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