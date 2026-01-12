.PHONY: build release test clean install uninstall help bump bump-major bump-minor bump-patch version-current

# Default target - build the project
all: build

# Display available targets
help:
	@echo "Available make targets:"
	@echo "  build           - Build the note binary"
	@echo "  release         - Build release binary with version from current git tag"
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
	@echo ""
	@echo "Version management:"
	@echo "  version-current - Show current version from git tags"
	@echo "  bump            - Bump patch version (default, same as bump-patch)"
	@echo "  bump-major      - Bump major version (X.0.0) and create git tag"
	@echo "  bump-minor      - Bump minor version (x.X.0) and create git tag"
	@echo "  bump-patch      - Bump patch version (x.x.X) and create git tag"

# Build variables
BINARY_NAME=note
INSTALL_PATH=/usr/local/bin

# Build the application
build:
	go build -o $(BINARY_NAME)

# Build release with version information
release:
	@CURRENT_TAG=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	VERSION=$${CURRENT_TAG#v}; \
	echo "Building release version: $$VERSION"; \
	go build -ldflags "-X 'main.Version=$$VERSION' -X 'main.CommitSHA=$(shell git rev-parse --short HEAD)' -X 'main.BuildDate=$(shell date -u +'%Y-%m-%d_%H:%M:%SUTC')'" -o $(BINARY_NAME)

# Version management targets
version-current:
	@git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"

# Default bump command (patch version)
bump: bump-patch

bump-major:
	@echo "Bumping MAJOR version..."
	@CURRENT=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	CURRENT=$${CURRENT#v}; \
	MAJOR=$$(echo $$CURRENT | cut -d. -f1); \
	NEW_VERSION="v$$((MAJOR + 1)).0.0"; \
	echo "Current version: v$$CURRENT"; \
	echo "New version: $$NEW_VERSION"; \
	git tag -a $$NEW_VERSION -m "Release $$NEW_VERSION"; \
	echo "Created tag $$NEW_VERSION"

bump-minor:
	@echo "Bumping MINOR version..."
	@CURRENT=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	CURRENT=$${CURRENT#v}; \
	MAJOR=$$(echo $$CURRENT | cut -d. -f1); \
	MINOR=$$(echo $$CURRENT | cut -d. -f2); \
	NEW_VERSION="v$$MAJOR.$$((MINOR + 1)).0"; \
	echo "Current version: v$$CURRENT"; \
	echo "New version: $$NEW_VERSION"; \
	git tag -a $$NEW_VERSION -m "Release $$NEW_VERSION"; \
	echo "Created tag $$NEW_VERSION"

bump-patch:
	@echo "Bumping PATCH version..."
	@CURRENT=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	CURRENT=$${CURRENT#v}; \
	MAJOR=$$(echo $$CURRENT | cut -d. -f1); \
	MINOR=$$(echo $$CURRENT | cut -d. -f2); \
	PATCH=$$(echo $$CURRENT | cut -d. -f3); \
	NEW_VERSION="v$$MAJOR.$$MINOR.$$((PATCH + 1))"; \
	echo "Current version: v$$CURRENT"; \
	echo "New version: $$NEW_VERSION"; \
	git tag -a $$NEW_VERSION -m "Release $$NEW_VERSION"; \
	echo "Created tag $$NEW_VERSION"

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