# note

[![Version](https://img.shields.io/badge/version-0.1.5-blue.svg)](https://github.com/brockers/note/releases/tag/v0.1.5)
[![License](https://img.shields.io/badge/license-GPL--3.0-green.svg)](https://www.gnu.org/licenses/gpl-3.0.en.html)
[![Tests](https://img.shields.io/badge/tests-173%20passing-brightgreen.svg)](#testing)

A minimalist, opinionated command-line note-taking tool written in Go. Take notes in your favorite text editor with automatic date stamping, simple organization, and zero lock-in - just plain markdown files.

## Features

- **Automatic Date Stamping**: All notes get `-YYYYMMDD` added to the end of their names
- **Tab Completion**: Type `note SomeOldNote` and press TAB to see all matching notes (Bash, Zsh, Fish)
- **Two-Question Setup**: Simple configuration when first running the app
- **Archive System**: Notes are never deleted, just archived and can be listed/searched with the `-a` flag
- **Flag Chaining**: Combine flags like `-al` for listing all notes including archived
- **Full-Text Search**: Search note contents with the `-s` flag
- **Universal Markdown**: All notes saved as `.md` files for maximum compatibility
- **Shell Aliases**: Optional convenient shortcuts (`n`, `nls`, `nrm`)
- **Well Documented**: Standard Unix help via `-h` or `--help` options
- **Thoroughly Tested**: 173 automated tests covering unit, integration, completion, and setup scenarios

## Example Usage

### Setup

If it's the first time you run note, it will ask you for your preferred editor and where you want to save your notes.

```bash
note
What is your preferred text editor (vim): nvim-qt
Setting nvim-qt as default text editor...
Where are you saving your notes (~/Notes): ~/Dropbox/Reference/Notes/
Setting your notes location to ~/Dropbox/Reference/Notes/ ...
```

After which, note will save these configuration settings to your `~/.note` file. You can also edit these settings by re-running the setup with:

```bash
note --config
What is your preferred text editor (nvim-qt): nvim
Setting nvim as default text editor...
Where are you saving your notes (/Dropbox/Reference/Notes/): ~/Documents/Notes/
Setting your notes location to ~/Documents/Notes/ ...
```

### Create a New Note

```bash
note Thoughts_on_AI
```

Note will open your preferred text editor with a file named `Thoughts_on_AI-20260112.md` where the date is automatically added in `YYYYMMDD` format.

### Open an Existing Note

```bash
note Thoughts_on_AI-20260407.md
```

Note will open the specified note in your preferred text editor.

You can also use tab completion to find your note by starting to type the note name and pressing TAB twice:

```bash
note Thou<TAB><TAB>
# Shows: Thoughts_on_AI-20260112.md  Thoughts_on_AI-20260407.md
```

### List Notes

List all of your current notes:

```bash
note -l

  Chess-Notes-20230714.md
  civ5-notes-20200101.md
  Day-Counting-Billing-Workflow-20211006.md
  Die_backpacker_die-20230507.md
  Emily_Interview_Notes-20210224.md
  Engineering-Infrastructure-Design-20210216.md
  Experience_vs_Achievement-20240511.md
  Family-Notes-20220314.md
  Fitness-after-45-Plan-20241109.md
```

Filter by filename pattern (case-insensitive):

```bash
note -l 2021

  Emily_Interview_Notes-20210224.md
  Engineering-Infrastructure-Design-20210216.md
```

### Search Note Contents

Search for text within your notes using full-text search:

```bash
note -s "important idea"

  Project-Ideas-20260110.md:
    15: This is an important idea for the future

  Meeting-Notes-20260108.md:
    23: Discussed important idea with team
```
### Archive Notes

Archive notes to keep your workspace clean. Archived notes move to `<notespath>/Archive/` and won't appear in normal `-l` searches.

Archive a single note:

```bash
note -d Chess-Notes-20230714  # The .md extension is optional
Archiving:
  Chess-Notes-20230714
```

Archive multiple notes with wildcards:

```bash
note -d E*
Archiving:
  Emily_Interview_Notes-20210224.md
  Engineering-Infrastructure-Design-20210216.md
  Experience_vs_Achievement-20240511.md
```

### Working with Archived Notes

List all notes including archived:

```bash
note -a

  Chess-Notes-20230714.md
  civ5-notes-20200101.md
  Day-Counting-Billing-Workflow-20211006.md
  Die_backpacker_die-20230507.md
  Emily_Interview_Notes-20210224.md
  Engineering-Infrastructure-Design-20210216.md
  Experience_vs_Achievement-20240511.md
  Family-Notes-20220314.md
  Fitness-after-45-Plan-20241109.md
```

Search all notes including archived (case-insensitive):

```bash
note -a Notes

  Chess-Notes-20230714.md
  civ5-notes-20200101.md
  Emily_Interview_Notes-20210224.md
  Family-Notes-20220314.md
```

### Flag Chaining

Combine flags for powerful operations:

```bash
note -al project        # List all notes (including archived) matching "project"
note -as "bug fix"      # Search all notes (including archived) for "bug fix"
note -la meeting        # Same as -al (order doesn't matter)
```

### Shell Aliases (Optional)

Set up convenient aliases for common commands:

```bash
note --alias
```

This creates:
- `n` - Same as `note`
- `nls` - Same as `note -l`
- `nrm` - Same as `note -d`

### Help

Get help anytime with:

```bash
note -h         # Short help
note --help     # Detailed help
note --version  # Show version information
```

## Installation

### From Release Binary

Download the latest release from [GitHub Releases](https://github.com/brockers/note/releases):

```bash
# Download the release binary (replace with latest version)
wget https://github.com/brockers/note/releases/download/v0.1.5/note

# Make it executable
chmod +x note

# Move to your PATH
sudo mv note /usr/local/bin/

# Verify installation
note --version
```

### From Source

Requirements:
- Go 1.21 or later

```bash
# Clone the repository
git clone https://github.com/brockers/note.git
cd note

# Build the binary
make build

# Install system-wide (optional, requires sudo)
make install

# Or copy manually to your PATH
cp note ~/bin/  # or wherever you keep personal binaries
```

### Enable Tab Completion

After installation, enable tab completion for your shell:

```bash
# Automatic setup (recommended)
note --autocomplete

# This will configure completion for your detected shell (Bash, Zsh, or Fish)
```

Manual setup if needed:

```bash
# Bash
echo 'source <(note --autocomplete bash)' >> ~/.bashrc

# Zsh
echo 'source <(note --autocomplete zsh)' >> ~/.zshrc

# Fish
note --autocomplete fish > ~/.config/fish/completions/note.fish
```

## Development

### Building

```bash
# Build the binary
make build

# Build release version with version info
make release

# Format code
make fmt

# Run static analysis
make vet

# Clean build artifacts
make clean
```

### Testing

`note` has a comprehensive test suite with 173 automated tests:

```bash
# Run unit tests
make test

# Run integration tests
make integration-test

# Run completion tests
make completion-test

# Run setup tests
make setup-test

# Run all tests
make test-all
```

Test coverage includes:
- **Unit Tests** (51 tests): Core functionality, path handling, search, configuration
- **Integration Tests** (51 tests): End-to-end workflows, archiving, listing, searching
- **Completion Tests** (21 tests): Tab completion for Bash, Zsh, Fish
- **Setup Tests** (50 tests): First-run setup, configuration, shell integration

### Release Process

To create a new release:

```bash
# Run the release command (handles testing, versioning, building, and tagging)
# Patch version bump (0.1.4 -> 0.1.5)
make bump && make release && git push origin <TAG>

# Or use the automated release workflow
# /development:release [patch|minor|major]
```

## Example Workflows

### Daily Journaling

```bash
note journal            # Creates journal-20260112.md today
note journal            # Creates journal-20260113.md tomorrow
note -l journal         # List all journal entries
```

### Project Notes

```bash
note project-alpha-design
note project-alpha-meeting
note -l project-alpha   # List all project-alpha notes
note -s "action items"  # Search for action items across notes
```

### Research Workflow

```bash
note research-topic
note -s "important finding"
note -al research       # List all research notes, including archived
note -d research-*      # Archive research notes when done
```

### Integration with Unix Tools

```bash
# Count your notes
note -l | wc -l

# Find TODOs from 2026
note -l 2026 | xargs grep "TODO"

# Version control your notes
cd ~/Notes && git init && git add . && git commit -m "Initial notes"

# Sync with Dropbox
note --config  # Point to ~/Dropbox/Notes
```

## Philosophy

`note` follows the Unix philosophy: do one thing well and compose with other tools. It's intentionally minimal and opinionated to provide a frictionless note-taking experience for terminal users.

- **No databases**: Just markdown files in folders
- **No sync built-in**: Use git, Dropbox, or any sync tool you prefer
- **No tags or categories**: Use your filesystem and grep
- **No dependencies**: Single static binary
- **No lock-in**: Your notes are just text files

## Support & Contributing

- **Repository**: https://github.com/brockers/note
- **Issues**: https://github.com/brockers/note/issues
- **Releases**: https://github.com/brockers/note/releases
- **Documentation**: See [RELEASE.md](RELEASE.md) for release notes and [SETUP.md](SETUP.md) for setup details

Contributions are welcome! Please feel free to submit issues or pull requests.

## Related Projects

- [mark](https://github.com/brockers/mark): An opinionated go commandline unix utility to create bookmarks for easy access to other folders.

## License

This program is free software licensed under GPL-3.0.
See https://www.gnu.org/licenses/ for details.

---

**Version 0.1.5** | Built with Go | [View Release Notes](RELEASE.md)
