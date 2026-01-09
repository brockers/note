# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Nature

This is a **minimalist command-line note-taking tool** written in Go. It provides a simple, opinionated interface for creating, organizing, and managing markdown notes with automatic date stamping and bash completion.

## Core Architecture

### Single Binary Design

- **Language**: Go (1.24.11+)
- **Main file**: `main.go` - contains all core functionality
- **Philosophy**: Unix philosophy - do one thing well
- **Dependencies**: Zero external dependencies, just Go standard library

### Key Features

- Automatic date stamping (`-yyyymmdd` format)
- Bash tab completion
- Simple two-question setup process
- Archive functionality (notes never deleted)
- Markdown-only format for universal access
- Configuration stored in `~/.note` file

## Development Commands

### Available Make Targets

```bash
# Build the binary
make build

# Run unit tests
make test

# Run integration tests  
make integration-test

# Run all tests
make test-all

# Install system-wide
make install

# Format code
make fmt

# Clean artifacts
make clean
```

### Test Commands

```bash
# Unit tests
go test -v

# Integration tests (requires setup)
./scripts/integration_test.sh

# Completion tests
./scripts/completion_test.sh
```

## Validation Commands

When working on this project, run these validation steps:

```bash
# Level 1: Build Check
go build -o note

# Level 2: Unit Tests
make test

# Level 3: Integration Tests
make integration-test

# Level 4: Completion Tests
make completion-test
```

## Project Structure

```
note/
├── main.go                    # Main application code
├── main_test.go              # Unit tests
├── go.mod                    # Go module definition
├── Makefile                  # Build automation
├── README.md                 # User documentation
├── SETUP.md                  # Setup instructions
├── scripts/                  # Test and utility scripts
│   ├── integration_test.sh       # E2E tests
│   ├── completion_test.sh        # Tab completion tests
│   └── setup_integration_test.sh # Test setup script
├── docs/                     # Documentation
│   └── note-cli-prd.md      # Product requirements
└── .claude/                  # Claude Code configuration
    ├── commands/             # Claude commands for development
    └── settings.local.json   # Tool permissions
```

## Development Guidelines

### Code Patterns

- Single-file architecture in `main.go`
- Struct-based configuration (`Config` type)
- ANSI color codes for terminal highlighting
- Comprehensive error handling
- File operations use `filepath` package for cross-platform compatibility

### Testing Strategy

- Unit tests in `main_test.go` cover core functions
- Integration tests simulate full user workflows
- Completion tests verify bash autocomplete functionality
- All tests designed to work in isolated environments

### Key Functions to Understand

- `setupNote()` - First-time configuration
- `listNotes()` - List/search notes functionality
- `archiveNote()` - Archive (soft delete) functionality
- `openNote()` - Open existing or create new notes
- `highlightSearchTerms()` - Terminal highlighting for search results

Remember: This is a focused CLI tool following Unix philosophy. Keep changes minimal, well-tested, and true to the simple note-taking workflow.
