# note

A minimalist, opinionated command-line note-taking tool written in Go. Take notes in your favorite text editor with automatic date stamping, simple organization, and zero lock-in - just plain markdown files.

note includes:

- **Automatic Time Stamping**: All note notes get `-yyyymmdd` added to the end of their names
- **Auto-compete**: Typing `note SomeOldNote` and tabbing twice will show all notes that being with SomeOldNote
- **Easy Configutration**: 2 question setup when first running the app
- **Archiving**: Notes are never deleted by note, they are archived in a folder in your setup folder and can be listed/searched with the `-a` command
- **Always in MD format**: For universal access, even when not in note.
- **Well Documented**: Standard unix help via `-h` or `--help` options
- **Fully Tested**: Full setup of unit and e2e tests to make sure the program updates always work and never accidently delete notes.

## Example Usage

### Setup

If it is the first time you run note, it will ask you for your prefered editor, and where you put your notes.

```bash
note
What is your preferred text editor (vim): nvim-qt
Setting nvim-qt as default text editor...
Where are you saving your notes (~/Notes): ~/Dropbox/Reference/Notes/
Setting your notes location to ~/Dropbox/Reference/Notes/ ...
```

After which, note will save these configuration settings to your ~/note file. You can also edit these settings by re-running the setup with:

```bash
note --config
What is your preferred text editor (nvim-qt): nvim
Setting nvim as default text editor...
Where are you saving your notes (/Dropbox/Reference/Notes/): ~/Documents/Notes/
Setting your notes location to ~/Documents/Notes/ ...
```

### Create a new Note

```bash
note Thoughts_on_Ai
```

Note will open your prefered text editor with an unsaved file name of `Thoughts_on_Ai-<currentdate>.md` where `<currentdate>` is in the format yyyymmdd.

### Open an old Note

```bash
note Thoughts_on_Ai-20260407.md
```

Note will open the note `Thoughts_on_Ai-20260407.md` in your preferred text editor.

Additionally, you can use note's autocomplete feature to find your file by starting to type the note name and double tapping TAB.

### List/Search Notes

List all of your current notes with 

```bash
note -ls # Or alternatively 'note -l'

  Chess-Notes-20230714.md
  civ5-notes-20200101.md
  Day-Counting-Billing-Workflow-20211006.md
  Die_backpacker_die-20230507.md
  Emily_Interview_Notes-20210224.md
  Engineering-Infrastructure-Design-20210216.md
  Experience_vs_Achivement-20240511.md
  Family-Notes-20220314.md
  Fitness-after-45-Plan-20241109.md
```

You can search (case insensive) your current notes with:

```bash
note -ls 2021

  Emily_Interview_Notes-20210224.md
  Engineering-Infrastructure-Design-20210216.md
```
### Archive Note(s)

You can archive notes. This makes them not current by moving them into your '<notespath>/Archive/' where '<notespath>' is the location you specified during setup. Archived notes will not appear in normal '-l' and '-al' search results.

```bash
note -rm Chess-Notes-20230714 #The md is optional
Archiving:
  Chess-Notes-20230714
```

or for multiple notes

```bash
note -rm E*
Archiving:
  Emily_Interview_Notes-20210224.md
  Engineering-Infrastructure-Design-20210216.md
  Experience_vs_Achivement.md
```

You can list ALL your notes (including archived notes) simply by:

```bash
note-a 

  Chess-Notes-20230714.md
  civ5-notes-20200101.md
  Day-Counting-Billing-Workflow-20211006.md
  Die_backpacker_die-20230507.md
  Emily_Interview_Notes-20210224.md
  Engineering-Infrastructure-Design-20210216.md
  Experience_vs_Achivement-20240511.md
  Family-Notes-20220314.md
  Fitness-after-45-Plan-20241109.md
```

You can search (case insensive) ALL your notes (including archived notes) by doing:

```bash
note -a Notes

  Chess-Notes-20230714.md
  civ5-notes-20200101.md
  Emily_Interview_Notes-20210224.md
  Family-Notes-20220314.md
```

### Help

You can always get help via: `note -h` or `note --help`

## Installation

### From Source

Requirements:
- Go 1.21 or later

```bash
# Clone the repository
git clone https://github.com/bobby/note.git
cd note

# Build the binary
make build
# Or directly with go:
# go build -o note

# Install system-wide (optional, requires sudo)
make install

# Or copy manually to your PATH:
# cp note ~/bin/  # or wherever you keep personal binaries
```

### Bash Completion

To enable tab completion for bash:

```bash
# If installed via make install, completions are already in place
# Otherwise, source the completion script:
source completions/bash/note
```

## Development

```bash
# Run tests
make test

# Run integration tests
make integration-test

# Run all tests
make test-all

# Format code
make fmt

# Clean build artifacts
make clean
```

## Philosophy

`note` follows the Unix philosophy: do one thing well and compose with other tools. It's intentionally minimal and opinionated to provide a frictionless note-taking experience for terminal users.

- **No databases**: Just markdown files in folders
- **No sync built-in**: Use git, Dropbox, or any sync tool you prefer  
- **No tags or categories**: Use your filesystem and grep
- **No dependencies**: Single static binary
- **No lock-in**: Your notes are just text files
