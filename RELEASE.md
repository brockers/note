# Release Notes

## Version 0.1.6

**Release Date:** January 28, 2026

### What's New

#### Features
- **Archive-aware tab completion**: Tab completion now respects the `-a` flag for archive operations, improving workflow when working with archived notes (cded6fd)
- **`--configure` flag alias**: Added `--configure` as an alias for `--config` flag for improved discoverability (e79f42a)
- **Automatic release notes generation**: Release workflow now automatically generates categorized release notes from git commits (afccb59)

#### Bug Fixes
- **Setup file path validation**: Setup now properly rejects file paths (vs directories) when prompting for notes directory, with clear user feedback (f829154)
- **Shell alias completion**: Fixed tab completion for shell aliases `n`, `nls`, and `nrm` (762ac0d)

#### Refactoring
- **Centralized shell configuration**: Implemented centralized shell configuration files for cleaner shell integration (104c405)

#### Documentation
- **Git workflow rules**: Added git workflow rules to CLAUDE.md to prevent unauthorized staging (f3d9dac)
- **Related projects**: Added related projects section to README with mark bookmark utility (0aa0f87)
- **v0.1.5 documentation**: Added comprehensive documentation for v0.1.5 first official release (289e2df)

### Test Coverage
- **165+ automated tests** across unit, integration, completion, and setup test suites
- New tests for file path rejection during setup
- Enhanced completion test infrastructure

---

## 🎉 Version 0.1.5 - First Official Release

**Release Date:** January 12, 2026
**Build Date:** 2026-01-12_19:45:07UTC
**Commit SHA:** 5a123d0

This marks the **first official release** of `note` - a minimalist command-line note-taking tool that embraces the Unix philosophy: do one thing well and compose with other tools.

---

## 📦 What is `note`?

`note` is a minimalist, opinionated command-line tool designed for developers, sysadmins, and anyone who lives in the terminal and wants frictionless note-taking without leaving their workflow.

### Core Philosophy

- **No databases** - Just markdown files in folders
- **No sync built-in** - Use git, Dropbox, or any sync tool you prefer
- **No tags or categories** - Use your filesystem and grep
- **No dependencies** - Single static binary written in Go
- **No lock-in** - Your notes are just text files

---

## ✨ Key Features

### 🎯 Automatic Date Stamping
Every note automatically gets `-YYYYMMDD` appended to its filename, creating a chronological archive of your thoughts without any manual effort.

```bash
note meeting            # Creates meeting-20260112.md
note project-ideas      # Creates project-ideas-20260112.md
```

### ⚡ Two-Question Setup
No complex configuration files or endless options. Just two questions:
1. What's your preferred text editor?
2. Where do you want to save your notes?

Done. You're ready to take notes.

### 🔍 Smart Search & Listing
Find your notes quickly with built-in search:

```bash
note -l                 # List all current notes
note -l project         # List notes containing "project"
note -s "todo"          # Full-text search across all notes
note -as "idea"         # Search including archived notes
```

### 📁 Archive System (Never Delete)
Notes are precious. `note` never deletes - it archives:

```bash
note -d old-*           # Archive notes starting with "old-"
note -a                 # List all notes including archived
note -al meeting        # List all meeting notes, even archived ones
```

### ⌨️ Tab Completion
Bash, Zsh, and Fish shell completions make finding and opening notes as fast as typing a few characters and hitting TAB.

```bash
note Life<TAB><TAB>     # Shows all notes starting with "Life"
```

### 🔗 Flag Chaining
Combine flags for powerful operations:

```bash
note -al project        # List all project notes (including archived)
note -as "bug"          # Search all notes for "bug" (including archived)
```

### 📝 Universal Markdown Format
All notes are saved as plain markdown files. Use any editor, sync with any service, read on any device. Your notes will outlive any app.

### 🚀 Shell Aliases
Optional convenient aliases for common operations:

```bash
n meeting               # Same as 'note meeting'
nls project             # Same as 'note -l project'
nrm old-*               # Same as 'note -d old-*'
```

---

## 🧪 Release Validation

This release has been thoroughly tested with:

### Unit Tests
✅ **51 tests passed** covering:
- Path expansion and symlink handling
- Note matching and filename generation
- Configuration management
- Search highlighting
- Flag parsing and chaining
- Edge case handling

### Integration Tests
✅ **51 tests passed** covering:
- Complete setup workflow
- Note creation and management
- List and search operations
- Archive functionality
- Bulk operations
- Special character handling
- Symlink directory support
- Multi-match scenarios
- Flag chaining combinations

### Completion Tests
✅ **21 tests passed** covering:
- Partial note name matching
- Case-insensitive completion
- Flag completion
- Completion after flags
- Empty input handling

### Setup Tests
✅ **50 tests passed** covering:
- First-run setup flow
- Configuration management
- Shell detection and alias setup
- Autocomplete installation
- Directory operations
- Error handling
- Migration scenarios
- Performance benchmarks

**Total: 173 automated tests passed**

---

## 📥 Installation

### From Release Binary

```bash
# Download the release binary
wget https://github.com/brockers/note/releases/download/v0.1.5/note

# Make it executable
chmod +x note

# Move to your PATH
sudo mv note /usr/local/bin/
```

### From Source

```bash
# Clone the repository
git clone https://github.com/brockers/note.git
cd note

# Checkout the release tag
git checkout v0.1.5

# Build
make build

# Install (optional, requires sudo)
make install
```

### Requirements
- Go 1.21 or later (if building from source)
- Your favorite text editor (vim, nano, emacs, vscode, etc.)
- Bash, Zsh, or Fish shell (for completions)

---

## 🚀 Quick Start

1. **Run note for the first time:**
   ```bash
   note
   ```

2. **Answer two questions:**
   - What is your preferred text editor? (e.g., `vim`, `nano`, `code`)
   - Where are you saving your notes? (e.g., `~/Notes`, `~/Dropbox/Notes`)

3. **Start taking notes:**
   ```bash
   note ideas              # Create a new note
   note -l                 # List your notes
   note -s "keyword"       # Search your notes
   ```

4. **Set up completions (optional):**
   ```bash
   note --autocomplete     # Enable tab completion
   note --alias            # Set up convenient aliases (n, nls, nrm)
   ```

---

## 📊 What's New in v0.1.5

This first official release includes:

- ✅ Complete note-taking workflow (create, list, search, archive)
- ✅ Automatic date stamping
- ✅ Tab completion for Bash, Zsh, and Fish
- ✅ Flag chaining support (-al, -as, etc.)
- ✅ Archive system with -a flag
- ✅ Full-text search with highlighting
- ✅ Comprehensive test coverage (173 tests)
- ✅ Shell alias support
- ✅ Symlink directory support
- ✅ Version information flag (-v, --version)
- ✅ Automated release workflow

---

## 🔧 Configuration

Settings are stored in `~/.note`:

```bash
editor=nvim
notesdir=~/Documents/Notes
```

Reconfigure anytime with:
```bash
note --config
```

---

## 📖 Example Workflows

### Daily Journaling
```bash
note journal            # Creates journal-20260112.md today
note journal            # Creates journal-20260113.md tomorrow
```

### Project Notes
```bash
note project-alpha-design
note project-alpha-meeting
note -l project-alpha   # List all project-alpha notes
```

### Research Workflow
```bash
note research-topic
note -s "important finding"
note -d research-*      # Archive when done
```

### Integration with Unix Tools
```bash
note -l | wc -l                          # Count your notes
note -l 2026 | xargs grep "TODO"         # Find TODOs from 2026
cd ~/Notes && git init                   # Version control your notes
```

---

## 🐛 Known Issues

None reported. This is a stable first release.

---

## 🔮 Future Roadmap

While `note` is intentionally minimal, potential future enhancements under consideration:

- Note restoration from archive
- Customizable date formats
- FZF integration for fuzzy finding
- Additional shell completions
- Performance optimizations for 10k+ notes

**Philosophy:** We say "no" more than "yes" to keep `note` focused and simple.

---

## 📜 License

This program is free software licensed under GPL-3.0.
See https://www.gnu.org/licenses/ for details.

---

## 🙏 Acknowledgments

Built with:
- Go 1.24.11
- Unix philosophy
- Love for the command line

---

## 📞 Support & Contributing

- **Repository:** https://github.com/brockers/note
- **Issues:** https://github.com/brockers/note/issues
- **Documentation:** See README.md and SETUP.md

---

## 🎯 Version Information

```
Version:     0.1.5
Build Date:  2026-01-12_19:45:07UTC
Commit SHA:  5a123d0
Git Tag:     v0.1.5
```

**Verify your installation:**
```bash
note --version
```

---

**Happy note-taking! 📝**

*Remember: The best tool is the one you actually use.*
