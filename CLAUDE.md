# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Current Version**: v0.1.5 (First Official Release)
**Release Date**: January 12, 2026
**Status**: Stable, Production Ready

This is a **minimalist command-line note-taking tool** written in Go. It provides a simple, opinionated interface for creating, organizing, and managing markdown notes with automatic date stamping, full-text search, and multi-shell completion support (Bash, Zsh, Fish).

## Core Architecture

### Single Binary Design

- **Language**: Go (1.24.11+)
- **Main file**: `main.go` - contains all core functionality
- **Philosophy**: Unix philosophy - do one thing well
- **Dependencies**: Zero external dependencies, just Go standard library

### Key Features

- **Automatic date stamping** (`-YYYYMMDD` format appended to all note filenames)
- **Multi-shell tab completion** (Bash, Zsh, Fish with `--autocomplete` flag)
- **Two-question setup** process (editor preference and notes directory)
- **Archive system** (notes never deleted, just moved to Archive/ subdirectory)
- **Full-text search** (`-s` flag for searching note contents)
- **Flag chaining** (`-al`, `-as` for combining archive + list/search operations)
- **Shell aliases** (`n`, `nls`, `nrm` via `--alias` command)
- **Universal markdown** format (all notes are `.md` files)
- **Zero dependencies** (single static binary, no external libraries)
- **Configuration** stored in `~/.note` file
- **Version tracking** built into release binaries (`--version` flag)

## Development Commands

### Available Make Targets

```bash
# Build Commands
make build          # Build the binary
make release        # Build release binary with version info
make install        # Install system-wide (requires sudo)
make clean          # Clean build artifacts

# Testing Commands
make test                 # Run unit tests
make integration-test     # Run integration tests
make completion-test      # Run tab completion tests
make setup-test          # Run setup/configuration tests
make test-all            # Run all test suites (173 tests total)

# Code Quality Commands
make fmt            # Format Go code
make vet            # Run Go static analysis

# Release Commands
make bump           # Bump patch version (0.1.4 -> 0.1.5)
make bump-minor     # Bump minor version (0.1.5 -> 0.2.0)
make bump-major     # Bump major version (0.1.5 -> 1.0.0)
```

### Test Commands

```bash
# Unit tests (51 tests)
go test -v

# Integration tests (51 tests)
./scripts/integration_test.sh

# Completion tests (21 tests)
./scripts/completion_test.sh

# Setup tests (50 tests)
./scripts/setup_integration_test.sh

# All tests (173 tests total)
make test-all
```

## Validation Commands

When working on this project, **ALWAYS** run these validation steps before committing:

```bash
# Level 0: Clean and Check (CRITICAL)
make clean                  # Remove build artifacts
make vet                    # Static analysis
make fmt                    # Format code
git diff --exit-code        # Verify fmt made no changes

# Level 1: Build Check
make build                  # Build the binary
./note --version            # Verify build

# Level 2: Unit Tests (51 tests)
make test

# Level 3: Integration Tests (51 tests)
make integration-test

# Level 4: Completion Tests (21 tests)
make completion-test

# Level 5: Setup Tests (50 tests)
make setup-test

# Level 6: All Tests (173 tests - REQUIRED before release)
make test-all
```

**Important**: All 173 tests MUST pass before any release. If `make fmt` changes files, commit those changes before proceeding.

## Project Structure

```
note/
├── main.go                       # Main application code (single-file architecture)
├── main_test.go                  # Unit tests (51 tests)
├── go.mod                        # Go module definition
├── Makefile                      # Build automation and release management
├── README.md                     # User documentation (updated with v0.1.5 info)
├── RELEASE.md                    # Release notes and version history
├── SETUP.md                      # Setup instructions
├── CLAUDE.md                     # This file - guidance for Claude Code
├── LICENSE                       # GPL-3.0 license
├── scripts/                      # Test and utility scripts
│   ├── integration_test.sh           # Integration tests (51 tests)
│   ├── completion_test.sh            # Tab completion tests (21 tests)
│   └── setup_integration_test.sh     # Setup tests (50 tests)
├── docs/                         # Documentation
│   └── note-cli-prd.md          # Product requirements document
└── .claude/                      # Claude Code configuration
    ├── commands/                 # Custom Claude commands
    │   └── development:release   # Automated release workflow
    └── settings.local.json       # Tool permissions
```

## Development Guidelines

### Code Patterns

- Single-file architecture in `main.go`
- Struct-based configuration (`Config` type)
- ANSI color codes for terminal highlighting
- Comprehensive error handling
- File operations use `filepath` package for cross-platform compatibility

### Testing Strategy

The project has **173 automated tests** across four test suites:

1. **Unit Tests** (51 tests in `main_test.go`)
   - Core functionality, path handling, search, configuration
   - Filename generation and matching
   - Flag parsing and chaining
   - Edge case handling

2. **Integration Tests** (51 tests in `scripts/integration_test.sh`)
   - End-to-end user workflows
   - Note creation, listing, searching, archiving
   - Bulk operations and wildcards
   - Special character handling
   - Symlink directory support

3. **Completion Tests** (21 tests in `scripts/completion_test.sh`)
   - Tab completion for Bash, Zsh, Fish
   - Partial matching and case-insensitive completion
   - Flag completion
   - Edge cases (empty input, no matches)

4. **Setup Tests** (50 tests in `scripts/setup_integration_test.sh`)
   - First-run setup flow
   - Configuration management
   - Shell detection and alias setup
   - Autocomplete installation
   - Error handling and edge cases

All tests are designed to work in isolated environments and must pass before any release.

### Key Functions to Understand

- `setupNote()` - First-time configuration and reconfiguration
- `listNotes()` - List/search notes functionality with pattern matching
- `archiveNote()` - Archive (soft delete) functionality
- `openNote()` - Open existing or create new notes
- `searchNotes()` - Full-text search with content highlighting
- `highlightSearchTerms()` - Terminal highlighting for search results
- `parseFlags()` - Command-line flag parsing with chaining support
- `setupAliases()` - Shell alias installation
- `setupAutocomplete()` - Multi-shell completion setup

## Release Process

### Automated Release Workflow

Use the `/development:release` command for fully automated releases:

```bash
# Patch version bump (0.1.5 -> 0.1.6)
/development:release
/development:release patch

# Minor version bump (0.1.5 -> 0.2.0)
/development:release minor

# Major version bump (0.1.5 -> 1.0.0)
/development:release major
```

### What the Release Command Does

1. **Pre-release validation**:
   - Runs `make clean` to remove artifacts
   - Runs `make vet` for static analysis
   - Runs `make fmt` to format code
   - Verifies fmt made no changes (stops if it did)
   - Runs `make test-all` (all 173 tests must pass)

2. **Commit staged changes** (if any exist)

3. **Generate and commit release notes**:
   - Analyzes commits since last release tag
   - Categorizes changes by type (feature, bug, docs, etc.)
   - Updates RELEASE.md with new version's release notes
   - Commits RELEASE.md changes

4. **Version bump**: Creates git tag (e.g., v0.1.5)

5. **Build release binary**: Injects version, build date, commit SHA

6. **Validate binary**:
   - Checks version matches tag
   - Verifies version is NOT "dev" or "0.0.0"
   - Confirms build date and commit SHA are valid

7. **Push tag to origin**: Only after successful validation

### Manual Release (if needed)

```bash
# 1. Clean and validate
make clean && make vet && make fmt
git diff --exit-code  # Verify no fmt changes
make test-all         # All 173 tests must pass

# 2. Generate and commit release notes
# Manually update RELEASE.md with new version's release notes
# Include commits since last tag, categorized by type
git add RELEASE.md
git commit -m "docs(release): add release notes for v0.1.6"

# 3. Bump version and tag
make bump             # Creates tag (e.g., v0.1.6)

# 4. Build release
make release          # Builds with version info

# 5. Validate binary
./note --version      # Verify version, date, SHA

# 6. Push tag
git push origin v0.1.6
```

### Version Numbering

- **Patch** (0.1.X): Bug fixes, minor improvements, documentation updates
- **Minor** (0.X.0): New features, significant improvements (backward compatible)
- **Major** (X.0.0): Breaking changes, major rewrites

### Release Checklist

- [ ] All 173 tests passing
- [ ] Code formatted (`make fmt`)
- [ ] No static analysis warnings (`make vet`)
- [ ] RELEASE.md updated with new version's release notes (automated in `/development:release`)
- [ ] Release notes categorize commits by type (feature, bug, docs, etc.)
- [ ] RELEASE.md changes committed before tag creation
- [ ] Binary validated (version, date, SHA correct)
- [ ] Tag pushed to GitHub
- [ ] GitHub release created (optional)

## Development Philosophy

Remember: This is a focused CLI tool following Unix philosophy. Keep changes minimal, well-tested, and true to the simple note-taking workflow.

### Principles

- **Simplicity over features**: Only add what's truly needed
- **Test everything**: All changes require tests
- **No external dependencies**: Keep using only Go standard library
- **Cross-platform**: Test on Linux, macOS, BSD
- **Backward compatibility**: Don't break existing workflows
- **Documentation first**: Update docs before committing

### When Making Changes

1. **Read the code first**: Understand existing patterns
2. **Write tests first**: Test-driven development
3. **Keep it simple**: Resist feature creep
4. **Run all tests**: Use `make test-all` before committing
5. **Update docs**: README.md, RELEASE.md, and this file
6. **Follow conventions**: Match existing code style
