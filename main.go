/*
Copyright (C) 2025  Note CLI Contributors

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Config struct {
	Editor   string
	NotesDir string
}

// ANSI color codes for terminal highlighting
const (
	ColorRed   = "\033[31m"
	ColorReset = "\033[0m"
)

// isOutputToTerminal checks if stdout is a terminal (not piped)
func isOutputToTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// highlightTerm highlights the search term in the text with red color
func highlightTerm(text, term string) string {
	if term == "" || !isOutputToTerminal() {
		return text
	}
	
	// Case-insensitive highlighting
	lowerText := strings.ToLower(text)
	lowerTerm := strings.ToLower(term)
	
	// Find all occurrences and highlight them
	result := text
	startPos := 0
	for {
		pos := strings.Index(lowerText[startPos:], lowerTerm)
		if pos == -1 {
			break
		}
		
		actualPos := startPos + pos
		
		// Bounds checking to prevent panic
		if actualPos+len(term) > len(result) {
			break
		}
		
		// Preserve original case in the highlight
		originalTerm := result[actualPos : actualPos+len(term)]
		highlighted := ColorRed + originalTerm + ColorReset
		
		result = result[:actualPos] + highlighted + result[actualPos+len(term):]
		
		// Adjust positions accounting for added color codes
		colorCodeLength := len(ColorRed) + len(ColorReset)
		startPos = actualPos + len(term) + colorCodeLength
		
		// Update lowerText to match result changes
		lowerText = strings.ToLower(result)
	}
	
	return result
}

func main() {
	config, firstTimeSetup := loadOrCreateConfig()

	// If first-time setup was just completed, exit gracefully
	if firstTimeSetup {
		return
	}

	// Parse custom flags with Unix-like behavior
	flags, args := parseFlags(os.Args[1:])

	// Handle help
	if flags.Help {
		printHelp()
		return
	}

	// Handle config
	if flags.Config {
		runSetup()
		// Explicitly exit after config to prevent any further execution
		os.Exit(0)
	}

	// Handle autocomplete setup
	if flags.Autocomplete {
		RunAutocompleteSetup()
		return
	}

	// Handle alias setup
	if flags.Alias {
		RunAliasSetup()
		return
	}

	// Handle combined archive + list or search
	if flags.Archive && flags.List {
		pattern := ""
		if len(args) > 0 {
			pattern = strings.Join(args, " ")
		}
		listNotes(config, pattern, true)
		return
	}

	// Handle combined archive + search
	if flags.Archive && flags.Search != "" {
		searchNotes(config, flags.Search, true)
		return
	}

	// Handle listing
	if flags.List {
		pattern := ""
		if len(args) > 0 {
			pattern = strings.Join(args, " ")
		}
		listNotes(config, pattern, false)
		return
	}

	// Handle archive listing only
	if flags.Archive {
		pattern := ""
		if len(args) > 0 {
			pattern = strings.Join(args, " ")
		}
		listNotes(config, pattern, true)
		return
	}

	// Handle full-text search
	if flags.Search != "" {
		searchNotes(config, flags.Search, false)
		return
	}

	// Handle archive/delete
	if flags.Delete != "" {
		archiveNotes(config, flags.Delete)
		return
	}

	// Handle note creation/opening
	if len(args) == 0 {
		// No arguments, just run note without args (could open today's journal or show help)
		printHelp()
		return
	}

	// Join all arguments to handle spaces in note names
	noteName := strings.Join(args, " ")
	openOrCreateNote(config, noteName)
}

func loadOrCreateConfig() (Config, bool) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		os.Exit(1)
	}

	configPath := filepath.Join(homeDir, ".note")

	// Check if config exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// First run, create config
		return runSetup(), true
	}

	// Load existing config
	file, err := os.Open(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening config: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	config := Config{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "editor":
			config.Editor = value
		case "notesdir":
			config.NotesDir = expandPath(value)
		}
	}

	if config.Editor == "" || config.NotesDir == "" {
		fmt.Println("Invalid config file. Running setup...")
		return runSetup(), false
	}

	return config, false
}

func runSetup() Config {
	reader := bufio.NewReader(os.Stdin)
	config := Config{}

	// Get current values if they exist
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".note")
	if file, err := os.Open(configPath); err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				switch strings.TrimSpace(parts[0]) {
				case "editor":
					config.Editor = strings.TrimSpace(parts[1])
				case "notesdir":
					config.NotesDir = expandPath(strings.TrimSpace(parts[1]))
				}
			}
		}
		file.Close()
	}

	// Ask for editor
	defaultEditor := config.Editor
	if defaultEditor == "" {
		defaultEditor = os.Getenv("EDITOR")
		if defaultEditor == "" {
			defaultEditor = "vim"
		}
	}

	fmt.Printf("What is your preferred text editor (%s): ", defaultEditor)
	editor, _ := reader.ReadString('\n')
	editor = strings.TrimSpace(editor)
	if editor == "" {
		editor = defaultEditor
	}

	// Validate editor exists
	if _, err := exec.LookPath(editor); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Editor '%s' not found in PATH\n", editor)
		fmt.Fprintf(os.Stderr, "  Try: 'vim', 'nano', or install your preferred editor\n")
		os.Exit(1)
	}

	fmt.Printf("Setting %s as default text editor...\n", editor)
	config.Editor = editor

	// Ask for notes directory
	defaultDir := config.NotesDir
	if defaultDir == "" {
		defaultDir = "~/Notes"
	}

	fmt.Printf("Where are you saving your notes (%s): ", defaultDir)
	notesDir, _ := reader.ReadString('\n')
	notesDir = strings.TrimSpace(notesDir)
	if notesDir == "" {
		notesDir = defaultDir
	}

	notesDir = expandPath(notesDir)
	fmt.Printf("Setting your notes location to %s ...\n", notesDir)
	config.NotesDir = notesDir

	// Create directory if it doesn't exist
	if err := os.MkdirAll(config.NotesDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating notes directory: %v\n", err)
		os.Exit(1)
	}

	// Create Archive directory
	archiveDir := getArchiveDir(config.NotesDir)
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating archive directory: %v\n", err)
		os.Exit(1)
	}

	// Ask about command line completion
	SetupCompletion(reader)

	// Ask about shell aliases
	setupAliases(reader)

	// Save config
	saveConfig(config)
	return config
}

func saveConfig(config Config) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		os.Exit(1)
	}

	configPath := filepath.Join(homeDir, ".note")
	file, err := os.Create(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Convert absolute path back to ~ notation for config file
	notesDir := config.NotesDir
	if strings.HasPrefix(notesDir, homeDir) {
		notesDir = "~" + strings.TrimPrefix(notesDir, homeDir)
	}

	fmt.Fprintf(file, "editor=%s\n", config.Editor)
	fmt.Fprintf(file, "notesdir=%s\n", notesDir)
}


func setupAliases(reader *bufio.Reader) {
	// Check if aliases are already set up
	if areAliasesAlreadySetup() {
		return
	}

	fmt.Println()
	fmt.Print("Would you like to set up shell aliases (n -> note, nls -> note -l, nrm -> note -d)? (y/N): ")
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))
	
	if response != "y" && response != "yes" {
		fmt.Println("Skipping alias setup. You can run 'note --config' later to set them up.")
		return
	}

	shell := detectShell()
	if shell == "" {
		fmt.Println("Could not detect shell type. Skipping alias setup.")
		return
	}

	switch shell {
	case "bash":
		setupBashAliases()
	case "zsh":
		setupZshAliases()
	case "fish":
		setupFishAliases()
	default:
		fmt.Printf("Shell '%s' not supported for aliases. Supported shells: bash, zsh, fish\n", shell)
	}
}

func areAliasesAlreadySetup() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	shell := detectShell()
	switch shell {
	case "bash":
		bashrc := filepath.Join(homeDir, ".bashrc")
		if content, err := os.ReadFile(bashrc); err == nil {
			contentStr := string(content)
			return strings.Contains(contentStr, "alias n=") && strings.Contains(contentStr, "alias nls=") && strings.Contains(contentStr, "alias nrm=")
		}
	case "zsh":
		zshrc := filepath.Join(homeDir, ".zshrc")
		if content, err := os.ReadFile(zshrc); err == nil {
			contentStr := string(content)
			return strings.Contains(contentStr, "alias n=") && strings.Contains(contentStr, "alias nls=") && strings.Contains(contentStr, "alias nrm=")
		}
	case "fish":
		fishConfigDir := filepath.Join(homeDir, ".config", "fish", "config.fish")
		if content, err := os.ReadFile(fishConfigDir); err == nil {
			contentStr := string(content)
			return strings.Contains(contentStr, "alias n ") && strings.Contains(contentStr, "alias nls ") && strings.Contains(contentStr, "alias nrm ")
		}
	}
	return false
}

func setupBashAliases() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		return
	}

	bashrcPath := filepath.Join(homeDir, ".bashrc")
	
	// Get the full path to the note binary
	notePath, err := os.Executable()
	if err != nil {
		// Fallback to checking PATH
		notePath, err = exec.LookPath("note")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not determine note command path: %v\n", err)
			return
		}
	}

	aliasLines := fmt.Sprintf("\n# note command aliases\nalias n='%s'\nalias nls='%s -l'\nalias nrm='%s -d'\n", notePath, notePath, notePath)

	// Append to .bashrc
	file, err := os.OpenFile(bashrcPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening .bashrc: %v\n", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(aliasLines); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing aliases to .bashrc: %v\n", err)
		return
	}

	fmt.Printf("✓ Bash aliases setup complete!\n")
	fmt.Printf("  Added 'n', 'nls', and 'nrm' aliases to %s\n", bashrcPath)
	fmt.Printf("  Run 'source ~/.bashrc' or restart your shell to activate aliases\n")
}

func setupZshAliases() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		return
	}

	zshrcPath := filepath.Join(homeDir, ".zshrc")
	
	// Get the full path to the note binary
	notePath, err := os.Executable()
	if err != nil {
		// Fallback to checking PATH
		notePath, err = exec.LookPath("note")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not determine note command path: %v\n", err)
			return
		}
	}

	aliasLines := fmt.Sprintf("\n# note command aliases\nalias n='%s'\nalias nls='%s -l'\nalias nrm='%s -d'\n", notePath, notePath, notePath)

	// Append to .zshrc
	file, err := os.OpenFile(zshrcPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening .zshrc: %v\n", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(aliasLines); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing aliases to .zshrc: %v\n", err)
		return
	}

	fmt.Printf("✓ Zsh aliases setup complete!\n")
	fmt.Printf("  Added 'n', 'nls', and 'nrm' aliases to %s\n", zshrcPath)
	fmt.Printf("  Run 'source ~/.zshrc' or restart your shell to activate aliases\n")
}

func setupFishAliases() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		return
	}

	// Create fish config directory if it doesn't exist
	fishConfigDir := filepath.Join(homeDir, ".config", "fish")
	if err := os.MkdirAll(fishConfigDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating fish config directory: %v\n", err)
		return
	}

	fishConfigPath := filepath.Join(fishConfigDir, "config.fish")
	
	// Get the full path to the note binary
	notePath, err := os.Executable()
	if err != nil {
		// Fallback to checking PATH
		notePath, err = exec.LookPath("note")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not determine note command path: %v\n", err)
			return
		}
	}

	aliasLines := fmt.Sprintf("\n# note command aliases\nalias n '%s'\nalias nls '%s -l'\nalias nrm '%s -d'\n", notePath, notePath, notePath)

	// Append to config.fish
	file, err := os.OpenFile(fishConfigPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening fish config: %v\n", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(aliasLines); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing aliases to fish config: %v\n", err)
		return
	}

	fmt.Printf("✓ Fish aliases setup complete!\n")
	fmt.Printf("  Added 'n', 'nls', and 'nrm' aliases to %s\n", fishConfigPath)
	fmt.Printf("  Restart your shell to activate aliases\n")
}

func expandPath(path string) string {
	// Handle tilde expansion first
	if strings.HasPrefix(path, "~/") {
		homeDir, _ := os.UserHomeDir()
		path = filepath.Join(homeDir, path[2:])
	}
	
	// Resolve symbolic links to get the actual path
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		// If we can't resolve symlinks, return the original path
		// This handles cases where the path doesn't exist yet or other errors
		return path
	}
	
	return resolvedPath
}

func openOrCreateNote(config Config, noteName string) {
	// Check if it's a specific file with .md extension
	if strings.HasSuffix(noteName, ".md") {
		// Open specific file
		notePath := filepath.Join(config.NotesDir, noteName)
		openInEditor(config.Editor, notePath)
		return
	}

	// Check if there's an exact match for noteName.md (existing file)
	// This handles cases like 'roloText-Meeting-Notes-20240426' which should open 'roloText-Meeting-Notes-20240426.md'
	exactFileName := noteName + ".md"
	exactPath := filepath.Join(config.NotesDir, exactFileName)
	if _, err := os.Stat(exactPath); err == nil {
		// Exact file exists, open it
		openInEditor(config.Editor, exactPath)
		return
	}

	// Generate today's date for new file
	today := time.Now().Format("20060102")
	// Replace spaces with underscores for filename
	cleanNoteName := strings.ReplaceAll(noteName, " ", "_")
	filename := fmt.Sprintf("%s-%s.md", cleanNoteName, today)
	notePath := filepath.Join(config.NotesDir, filename)

	// Check if note already exists for today
	if _, err := os.Stat(notePath); err == nil {
		// Note exists, open it
		openInEditor(config.Editor, notePath)
		return
	}

	// Check for similar notes (for tab completion hint)
	matches := findMatchingNotes(config.NotesDir, noteName, false)
	if len(matches) > 0 && len(matches) <= 5 {
		fmt.Println("Similar notes found:")
		for _, match := range matches {
			fmt.Printf("  %s\n", match)
		}
		fmt.Println()
	}

	// Create new note with today's date
	openInEditor(config.Editor, notePath)
}

func openInEditor(editor, filepath string) {
	cmd := exec.Command(editor, filepath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error opening editor: %v\n", err)
		os.Exit(1)
	}
}

// getArchiveDir returns the path to the archive directory, checking for both "Archive" and "archive"
func getArchiveDir(notesDir string) string {
	// Check for "Archive" first (preferred)
	archiveDir := filepath.Join(notesDir, "Archive")
	if _, err := os.Stat(archiveDir); err == nil {
		return archiveDir
	}
	
	// Check for "archive" (lowercase)
	archiveDir = filepath.Join(notesDir, "archive")
	if _, err := os.Stat(archiveDir); err == nil {
		return archiveDir
	}
	
	// Default to "Archive" if neither exists (for new creation)
	return filepath.Join(notesDir, "Archive")
}

func listNotes(config Config, pattern string, includeArchived bool) {
	dirs := []string{config.NotesDir}
	var archiveDirName string
	if includeArchived {
		archiveDir := getArchiveDir(config.NotesDir)
		dirs = append(dirs, archiveDir)
		archiveDirName = filepath.Base(archiveDir)
	}

	var allNotes []string
	for _, dir := range dirs {
		notes := findMatchingNotes(dir, pattern, false)
		if includeArchived && dir != config.NotesDir {
			// Prefix archived notes for clarity
			for i, note := range notes {
				notes[i] = archiveDirName + "/" + note
			}
		}
		allNotes = append(allNotes, notes...)
	}

	// Sort by modification time (newest first) or alphabetically
	sort.Strings(allNotes)

	for _, note := range allNotes {
		// Apply highlighting if pattern is provided and output is to terminal
		if pattern != "" {
			fmt.Println(highlightTerm(note, pattern))
		} else {
			fmt.Println(note)
		}
	}
}

func findMatchingNotes(dir, pattern string, includeSubdirs bool) []string {
	var notes []string

	// Walk the directory
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip Archive directory unless we want subdirs
		if !includeSubdirs && info.IsDir() && path != dir {
			return filepath.SkipDir
		}

		// Only look for .md files
		if !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		// Skip if in Archive subdirectory (unless we want subdirs)
		relPath, _ := filepath.Rel(dir, path)
		if !includeSubdirs && strings.Contains(relPath, string(os.PathSeparator)) {
			return nil
		}

		// Match pattern (case-insensitive)
		// Support both glob patterns and substring matching
		if pattern == "" {
			notes = append(notes, info.Name())
		} else {
			// First try glob pattern matching
			matched, err := filepath.Match(strings.ToLower(pattern), strings.ToLower(info.Name()))
			if err == nil && matched {
				notes = append(notes, info.Name())
			} else if strings.Contains(strings.ToLower(info.Name()), strings.ToLower(pattern)) {
				// Fall back to substring matching if not a valid glob or no match
				notes = append(notes, info.Name())
			}
		}

		return nil
	})

	return notes
}

func searchNotes(config Config, searchTerm string, includeArchived bool) {
	dirs := []string{config.NotesDir}
	if includeArchived {
		archiveDir := getArchiveDir(config.NotesDir)
		dirs = append(dirs, archiveDir)
	}

	fmt.Printf("Searching for '%s'...\n\n", searchTerm)

	for _, dir := range dirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			// Skip directories except Archive
			if info.IsDir() {
				return nil
			}

			// Only search .md files
			if !strings.HasSuffix(info.Name(), ".md") {
				return nil
			}

			// Read file and search
			file, err := os.Open(path)
			if err != nil {
				return nil
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			lineNum := 0
			found := false
			var matches []string

			for scanner.Scan() {
				lineNum++
				line := scanner.Text()
				if strings.Contains(strings.ToLower(line), strings.ToLower(searchTerm)) {
					if !found {
						relPath, _ := filepath.Rel(config.NotesDir, path)
						fmt.Printf("%s:\n", relPath)
						found = true
					}
					matches = append(matches, fmt.Sprintf("  %d: %s", lineNum, line))
					// Limit matches per file
					if len(matches) >= 3 {
						matches = append(matches, "  ...")
						break
					}
				}
			}

			if found {
				for _, match := range matches {
					fmt.Println(match)
				}
				fmt.Println()
			}

			return nil
		})
	}
}

func archiveNotes(config Config, pattern string) {
	notes := findMatchingNotes(config.NotesDir, pattern, false)
	
	if len(notes) == 0 {
		fmt.Printf("No notes found matching '%s'\n", pattern)
		return
	}

	archiveDir := getArchiveDir(config.NotesDir)
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating archive directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Archiving:")
	for _, note := range notes {
		fmt.Printf("  %s\n", note)
		srcPath := filepath.Join(config.NotesDir, note)
		dstPath := filepath.Join(archiveDir, note)
		
		// Move file
		if err := os.Rename(srcPath, dstPath); err != nil {
			// Try copy and delete if rename fails (cross-device)
			if err := copyFile(srcPath, dstPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error archiving %s: %v\n", note, err)
				continue
			}
			os.Remove(srcPath)
		}
	}
}

// ParsedFlags represents parsed command line flags
type ParsedFlags struct {
	List         bool
	Search       string
	Archive      bool
	Delete       string
	Config       bool
	Autocomplete bool
	Alias        bool
	Help         bool
}

// parseFlags implements Unix-like flag parsing with support for flag chaining
func parseFlags(args []string) (*ParsedFlags, []string) {
	flags := &ParsedFlags{}
	var remainingArgs []string
	
	for i := 0; i < len(args); i++ {
		arg := args[i]
		
		if arg == "--help" {
			flags.Help = true
		} else if arg == "--config" {
			flags.Config = true
		} else if arg == "--autocomplete" {
			flags.Autocomplete = true
		} else if arg == "--alias" {
			flags.Alias = true
		} else if strings.HasPrefix(arg, "--") {
			// Unknown long flag, treat as regular argument
			remainingArgs = append(remainingArgs, arg)
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			// Handle short flags and flag chaining
			flagChars := arg[1:] // Remove the '-' prefix
			
			for j, char := range flagChars {
				switch char {
				case 'h':
					flags.Help = true
				case 'l':
					flags.List = true
				case 'a':
					flags.Archive = true
				case 's':
					// -s requires an argument
					if j == len(flagChars)-1 {
						// -s is the last flag in the chain, next arg is the search term
						if i+1 < len(args) {
							i++
							flags.Search = args[i]
						} else {
							fmt.Fprintf(os.Stderr, "Error: -s flag requires a search term\n")
							os.Exit(1)
						}
					} else {
						fmt.Fprintf(os.Stderr, "Error: -s flag must be the last in a flag chain\n")
						os.Exit(1)
					}
				case 'd':
					// -d requires an argument
					if j == len(flagChars)-1 {
						// -d is the last flag in the chain, next arg is the pattern
						if i+1 < len(args) {
							i++
							flags.Delete = args[i]
						} else {
							fmt.Fprintf(os.Stderr, "Error: -d flag requires a pattern\n")
							os.Exit(1)
						}
					} else {
						fmt.Fprintf(os.Stderr, "Error: -d flag must be the last in a flag chain\n")
						os.Exit(1)
					}
				default:
					fmt.Fprintf(os.Stderr, "Error: unknown flag -%c\n", char)
					os.Exit(1)
				}
			}
		} else {
			// Regular argument
			remainingArgs = append(remainingArgs, arg)
		}
	}
	
	return flags, remainingArgs
}

// RunAliasSetup handles the standalone alias setup flow
func RunAliasSetup() {
	fmt.Println("note - Shell Alias Setup")
	fmt.Println()
	fmt.Println("This will set up convenient shell aliases:")
	fmt.Println("• n -> note")
	fmt.Println("• nls -> note -l")
	fmt.Println("• nrm -> note -d")
	fmt.Println()
	
	// Check if aliases are already set up
	if areAliasesAlreadySetup() {
		fmt.Println("Aliases are already set up!")
		return
	}
	
	reader := bufio.NewReader(os.Stdin)
	
	// Use the existing setupAliases function for the core logic
	setupAliases(reader)
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func printHelp() {
	fmt.Println(`note - A minimalist CLI note-taking tool

USAGE:
  note [name]              Create/open note with automatic dating
  note [name-date.md]      Open specific dated note
  note [OPTIONS] [args...]

OPTIONS:

  -l [pattern]             List notes (optionally matching pattern)
  -s <term>                Full-text search in notes
  -d <pattern>             Delete/archive matching notes
  -a [pattern]             Include archived notes in list/search
  -h                       Show this help message

  --help                   Show this help message
  --config                 Run setup/reconfigure
  --autocomplete           Setup/update command line autocompletion
  --alias                  Setup shell aliases (n, nls, nrm)

FLAG CHAINING:
  Single-character flags can be combined:
  -al [pattern]            List all notes (including archived)
  -as <term>               Search all notes (including archived)
  -la [pattern]            Same as -al

EXAMPLES:
  note meeting             Creates meeting-20260109.md
  note project-ideas       Creates project-ideas-20260109.md
  note -l                  List all current notes
  note -l project          List notes containing "project"
  note -s "todo"           Search for "todo" in current notes
  note -as "todo"          Search for "todo" in all notes (including archived)
  note -d old-*            Archive notes starting with "old-"
  note -a                  List all notes including archived

ALIASES:
  After running 'note --alias', you can use:
  n                        Same as 'note'
  nls                      Same as 'note -l'
  nrm                      Same as 'note -d'

CONFIGURATION:
  Settings are stored in ~/.note
  Use 'note --config' to reconfigure

LICENSE:
  This program is free software licensed under GPL-3.0.
  See <https://www.gnu.org/licenses/> for details.

For more information, see: https://github.com/bobby/note`)
}
