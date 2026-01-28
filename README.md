# note

[![Version](https://img.shields.io/badge/version-0.1.5-blue.svg)](https://github.com/brockers/note/releases/tag/v0.1.5)
[![License](https://img.shields.io/badge/license-GPL--3.0-green.svg)](https://www.gnu.org/licenses/gpl-3.0.en.html)
[![Tests](https://img.shields.io/badge/tests-211%20passing-brightgreen.svg)](#development)

A minimalist command-line note-taking tool written in Go. Take notes in your favorite editor with automatic date stamping, simple organization, and zero lock-in—just plain markdown files.

## Features

- **Automatic Date Stamping**: Notes get `-YYYYMMDD` appended to filenames
- **Tab Completion**: Bash, Zsh, and Fish support
- **Archive System**: Notes are never deleted, just archived
- **Full-Text Search**: Search note contents with `-s`
- **Flag Chaining**: Combine flags like `-al` or `-as`
- **Shell Aliases**: Optional shortcuts (`n`, `nls`, `nrm`)
- **Zero Dependencies**: Single static binary, no external libraries

## Quick Start

First run prompts for configuration:

```bash
$ note
What is your preferred text editor (vim): nvim
What directory are you saving notes in (~/Notes): ~/Dropbox/Notes
```

Reconfigure anytime with `note --config`. Settings stored in `~/.note`.

## Usage

### Create or Open a Note

```bash
note MyIdea                    # Creates MyIdea-20260128.md
note MyIdea-20260128.md        # Opens existing note
note My<TAB>                   # Tab completion finds matching notes
```

### List Notes

```bash
note -l                        # List all notes
note -l project                # Filter by pattern (case-insensitive)
note -al project               # Include archived notes
```

### Search Note Contents

```bash
note -s "important"            # Search text within notes
note -as "important"           # Search including archived
```

### Archive Notes

```bash
note -d OldNote                # Archive a note (moves to Archive/)
note -d Old*                   # Archive with wildcards
```

### Shell Aliases

```bash
note --alias                   # Install n, nls, nrm aliases
```

### Help

```bash
note -h                        # Quick help
note --help                    # Detailed help
note --version                 # Version info
```

## Installation

### From Binary

```bash
# Download from https://github.com/brockers/note/releases
chmod +x note
sudo mv note /usr/local/bin/
```

### From Source

```bash
git clone https://github.com/brockers/note.git
cd note
make build
make install    # or: cp note ~/bin/
```

### Enable Tab Completion

```bash
note --autocomplete            # Auto-detects and configures your shell
```

## Development

```bash
make build      # Build binary
make test       # Run all 211 tests
make fmt        # Format code
```

## Philosophy

* Just markdown files in folders. 
* No databases, no sync, no lock-in. 
* Use git, Dropbox, or any tool you prefer to sync. 
* Follows Unix philosophy: do one thing well.

## License

GPL-3.0 — See [LICENSE](LICENSE) for details.

## Links

- [Repository](https://github.com/brockers/note)
- [Releases](https://github.com/brockers/note/releases)
- [Issues](https://github.com/brockers/note/issues)
- [Release Notes](RELEASE.md)
