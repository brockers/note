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
	"flag"
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

	// Parse flags
	var (
		listFlag        = flag.Bool("ls", false, "List all current notes")
		listFlagAlt     = flag.Bool("l", false, "List all current notes (short form)")
		searchFlag      = flag.String("s", "", "Full-text search in notes")
		archiveFlag     = flag.Bool("a", false, "List/search all notes including archived")
		removeFlag      = flag.String("rm", "", "Archive matching notes")
		configFlag      = flag.Bool("config", false, "Run setup/reconfigure")
		autocompleteFlag = flag.Bool("autocomplete", false, "Setup/update command line autocompletion")
		helpFlag        = flag.Bool("help", false, "Show help")
		helpFlagAlt     = flag.Bool("h", false, "Show help (short form)")
	)
	flag.Parse()

	// Handle help
	if *helpFlag || *helpFlagAlt {
		printHelp()
		return
	}

	// Handle config
	if *configFlag {
		runSetup()
		return
	}

	// Handle autocomplete setup
	if *autocompleteFlag {
		runAutocompleteSetup()
		return
	}

	// Handle listing
	if *listFlag || *listFlagAlt {
		pattern := ""
		if flag.NArg() > 0 {
			// Join all arguments to handle spaces in search patterns
			noteArgs := flag.Args()
			pattern = strings.Join(noteArgs, " ")
		}
		listNotes(config, pattern, false)
		return
	}

	// Handle archive listing
	if *archiveFlag {
		pattern := ""
		if flag.NArg() > 0 {
			// Join all arguments to handle spaces in search patterns
			noteArgs := flag.Args()
			pattern = strings.Join(noteArgs, " ")
		}
		listNotes(config, pattern, true)
		return
	}

	// Handle full-text search
	if *searchFlag != "" {
		searchNotes(config, *searchFlag, false)
		return
	}

	// Handle archive/remove
	if *removeFlag != "" {
		archiveNotes(config, *removeFlag)
		return
	}

	// Handle note creation/opening
	if flag.NArg() == 0 {
		// No arguments, just run note without args (could open today's journal or show help)
		printHelp()
		return
	}

	// Join all arguments to handle spaces in note names
	noteArgs := flag.Args()
	noteName := strings.Join(noteArgs, " ")
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
	archiveDir := filepath.Join(config.NotesDir, "Archive")
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating archive directory: %v\n", err)
		os.Exit(1)
	}

	// Ask about command line completion
	setupCompletion(reader)

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

func setupCompletion(reader *bufio.Reader) {
	// Check if completion is already set up
	if isCompletionAlreadySetup() {
		return
	}

	fmt.Println()
	fmt.Print("Would you like to set up command line completion for note? (y/N): ")
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))
	
	if response != "y" && response != "yes" {
		fmt.Println("Skipping completion setup. You can run 'note --config' later to set it up.")
		return
	}

	shell := detectShell()
	if shell == "" {
		fmt.Println("Could not detect shell type. Skipping completion setup.")
		return
	}

	switch shell {
	case "bash":
		setupBashCompletion()
	case "zsh":
		setupZshCompletion()
	case "fish":
		setupFishCompletion()
	default:
		fmt.Printf("Shell '%s' not supported for completion. Supported shells: bash, zsh, fish\n", shell)
	}
}

func isCompletionAlreadySetup() bool {
	shell := detectShell()
	if shell == "" {
		return false
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	switch shell {
	case "bash":
		// Check if ~/.note.bash exists and is sourced in shell config
		bashCompletionFile := filepath.Join(homeDir, ".note.bash")
		if _, err := os.Stat(bashCompletionFile); err == nil {
			// Check .bashrc or .bash_profile for note completion
			bashFiles := []string{".bashrc", ".bash_profile", ".profile"}
			for _, file := range bashFiles {
				if checkFileForCompletionSource(filepath.Join(homeDir, file)) {
					return true
				}
			}
		}
	case "zsh":
		// Check if ~/.note.zsh exists and is sourced in .zshrc
		zshCompletionFile := filepath.Join(homeDir, ".note.zsh")
		if _, err := os.Stat(zshCompletionFile); err == nil {
			if checkFileForCompletionSource(filepath.Join(homeDir, ".zshrc")) {
				return true
			}
		}
	case "fish":
		// Check fish completion directory
		fishCompletionDir := filepath.Join(homeDir, ".config", "fish", "completions")
		noteCompletionFile := filepath.Join(fishCompletionDir, "note.fish")
		if _, err := os.Stat(noteCompletionFile); err == nil {
			return true
		}
	}

	return false
}

func checkFileForCompletionSource(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, ".note.bash") || strings.Contains(line, ".note.zsh") || 
		   (strings.Contains(line, "note") && (strings.Contains(line, "complete") || strings.Contains(line, "completion"))) {
			return true
		}
	}
	return false
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return ""
	}

	// Extract shell name from path
	shellName := filepath.Base(shell)
	
	// Map common shell variants
	switch shellName {
	case "bash":
		return "bash"
	case "zsh":
		return "zsh"
	case "fish":
		return "fish"
	default:
		return shellName
	}
}

func setupBashCompletion() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		return
	}

	// Write the embedded completion script to ~/.note.bash
	completionScriptPath := filepath.Join(homeDir, ".note.bash")
	bashCompletionScript := `#!/bin/bash

_note_complete() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local prev="${COMP_WORDS[COMP_CWORD-1]}"
    
    # If we're on the first argument
    if [[ ${COMP_CWORD} -eq 1 ]]; then
        # If user starts typing a dash, offer flags
        if [[ "$cur" == -* ]]; then
            local flags="-ls -l -s -a -rm --config --autocomplete --help -h"
            COMPREPLY=($(compgen -W "$flags" -- "${cur}"))
        else
            # Otherwise, prioritize note names
            if [[ -f ~/.note ]]; then
                local notesdir=$(grep "^notesdir=" ~/.note | cut -d= -f2 | sed "s|~|$HOME|")
                if [[ -d "$notesdir" ]]; then
                    # Get all .md files and remove the .md extension for easier completion
                    local notes=$(find "$notesdir" -maxdepth 1 -name "*.md" -type f -exec basename {} .md \; 2>/dev/null | sort | tr '\n' ' ')
                    # Use case-insensitive matching by converting both to lowercase
                    local cur_lower=$(echo "$cur" | tr '[:upper:]' '[:lower:]')
                    COMPREPLY=()
                    for note in $notes; do
                        local note_lower=$(echo "$note" | tr '[:upper:]' '[:lower:]')
                        if [[ "$note_lower" == "$cur_lower"* ]]; then
                            COMPREPLY+=("$note")
                        fi
                    done
                fi
            fi
        fi
    # If previous was -ls, -l, -a, or -rm, offer note names
    elif [[ "$prev" == "-ls" || "$prev" == "-l" || "$prev" == "-a" || "$prev" == "-rm" ]]; then
        if [[ -f ~/.note ]]; then
            local notesdir=$(grep "^notesdir=" ~/.note | cut -d= -f2 | sed "s|~|$HOME|")
            if [[ -d "$notesdir" ]]; then
                local notes=$(find "$notesdir" -maxdepth 1 -name "*.md" -type f -exec basename {} .md \; 2>/dev/null | sort | tr '\n' ' ')
                # Use case-insensitive matching by converting both to lowercase
                local cur_lower=$(echo "$cur" | tr '[:upper:]' '[:lower:]')
                COMPREPLY=()
                for note in $notes; do
                    local note_lower=$(echo "$note" | tr '[:upper:]' '[:lower:]')
                    if [[ "$note_lower" == "$cur_lower"* ]]; then
                        COMPREPLY+=("$note")
                    fi
                done
            fi
        fi
    fi
}

complete -F _note_complete note
`

	if err := os.WriteFile(completionScriptPath, []byte(bashCompletionScript), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing completion script: %v\n", err)
		return
	}

	// Add source line to .bashrc
	bashrcPath := filepath.Join(homeDir, ".bashrc")
	sourceLine := fmt.Sprintf("\n# note command completion\nsource %s\n", completionScriptPath)
	
	file, err := os.OpenFile(bashrcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening .bashrc: %v\n", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(sourceLine); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to .bashrc: %v\n", err)
		return
	}

	fmt.Printf("✓ Bash completion setup complete!\n")
	fmt.Printf("  Created completion script at %s\n", completionScriptPath)
	fmt.Printf("  Added source line to %s\n", bashrcPath)
	fmt.Printf("  Restart your shell or run: source %s\n", bashrcPath)
}

func setupZshCompletion() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		return
	}

	// Write the embedded completion script to ~/.note.zsh
	completionScriptPath := filepath.Join(homeDir, ".note.zsh")
	zshCompletionScript := `#!/bin/zsh

_note_complete() {
    local cur="${words[CURRENT]}"
    local prev="${words[CURRENT-1]}"
    
    # If we're on the first argument
    if [[ $CURRENT -eq 2 ]]; then
        # If user starts typing a dash, offer flags
        if [[ "$cur" == -* ]]; then
            local flags=("-ls" "-l" "-s" "-a" "-rm" "--config" "--autocomplete" "--help" "-h")
            compadd -a flags
        else
            # Otherwise, prioritize note names
            local notes=()
            if [[ -f ~/.note ]]; then
                local notesdir=$(grep "^notesdir=" ~/.note | cut -d= -f2 | sed "s|~|$HOME|")
                if [[ -d "$notesdir" ]]; then
                    # Get all .md files and remove the .md extension for easier completion
                    local all_notes=(${(f)"$(find "$notesdir" -maxdepth 1 -name "*.md" -type f -exec basename {} .md \; 2>/dev/null | sort)"})
                    # Filter case-insensitively
                    local cur_lower="${cur:l}"
                    for note in $all_notes; do
                        if [[ "${note:l}" == ${cur_lower}* ]]; then
                            notes+=("$note")
                        fi
                    done
                fi
            fi
            compadd -a notes
        fi
        
    # If previous was -ls, -l, -a, or -rm, offer note names
    elif [[ "$prev" == "-ls" || "$prev" == "-l" || "$prev" == "-a" || "$prev" == "-rm" ]]; then
        if [[ -f ~/.note ]]; then
            local notesdir=$(grep "^notesdir=" ~/.note | cut -d= -f2 | sed "s|~|$HOME|")
            if [[ -d "$notesdir" ]]; then
                local all_notes=(${(f)"$(find "$notesdir" -maxdepth 1 -name "*.md" -type f -exec basename {} .md \; 2>/dev/null | sort)"})
                # Filter case-insensitively
                local notes=()
                local cur_lower="${cur:l}"
                for note in $all_notes; do
                    if [[ "${note:l}" == ${cur_lower}* ]]; then
                        notes+=("$note")
                    fi
                done
                compadd -a notes
            fi
        fi
    fi
}

compdef _note_complete note
`

	if err := os.WriteFile(completionScriptPath, []byte(zshCompletionScript), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing completion script: %v\n", err)
		return
	}

	// Add source line to .zshrc
	zshrcPath := filepath.Join(homeDir, ".zshrc")
	sourceLine := fmt.Sprintf("\n# note command completion\nautoload -U +X compinit && compinit\nsource %s\n", completionScriptPath)
	
	file, err := os.OpenFile(zshrcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening .zshrc: %v\n", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(sourceLine); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to .zshrc: %v\n", err)
		return
	}

	fmt.Printf("✓ Zsh completion setup complete!\n")
	fmt.Printf("  Created completion script at %s\n", completionScriptPath)
	fmt.Printf("  Added source line to %s\n", zshrcPath)
	fmt.Printf("  Restart your shell or run: source %s\n", zshrcPath)
}

func setupFishCompletion() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		return
	}

	// Create fish completion directory if it doesn't exist
	fishCompletionDir := filepath.Join(homeDir, ".config", "fish", "completions")
	if err := os.MkdirAll(fishCompletionDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating fish completion directory: %v\n", err)
		return
	}

	// Create a simple fish completion script
	fishCompletionScript := `# note command completion for fish
complete -c note -f
complete -c note -s l -s ls -d "List notes"
complete -c note -s s -d "Search notes" -r
complete -c note -s a -d "Include archived notes"
complete -c note -s rm -d "Archive notes" -r
complete -c note -l config -d "Run setup/reconfigure"
complete -c note -l autocomplete -d "Setup/update command line autocompletion"
complete -c note -s h -l help -d "Show help"

# Complete with existing note names for main argument
complete -c note -n '__fish_is_first_token' -a '(if test -f ~/.note; set notesdir (grep "^notesdir=" ~/.note | cut -d= -f2 | sed "s|~|$HOME|"); if test -d "$notesdir"; find "$notesdir" -maxdepth 1 -name "*.md" -type f -exec basename {} .md \\; 2>/dev/null | sort; end; end)'
`

	noteCompletionFile := filepath.Join(fishCompletionDir, "note.fish")
	if err := os.WriteFile(noteCompletionFile, []byte(fishCompletionScript), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing fish completion script: %v\n", err)
		return
	}

	fmt.Printf("✓ Fish completion setup complete!\n")
	fmt.Printf("  Created completion file at %s\n", noteCompletionFile)
	fmt.Printf("  Restart your shell to activate completions\n")
}

func runAutocompleteSetup() {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Println("note - Command Line Autocompletion Setup")
	fmt.Println()
	fmt.Println("This will set up tab completion for the note command, allowing you to:")
	fmt.Println("• Tab-complete note names")
	fmt.Println("• Tab-complete command flags")
	fmt.Println("• Get context-aware completions")
	fmt.Println()
	fmt.Print("Would you like to set up autocompletion? (y/N): ")
	
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))
	
	if response != "y" && response != "yes" {
		fmt.Println("Autocompletion setup cancelled.")
		return
	}

	shell := detectShell()
	if shell == "" {
		fmt.Println("Could not detect shell type. Skipping completion setup.")
		fmt.Println("Supported shells: bash, zsh, fish")
		return
	}

	fmt.Printf("Detected shell: %s\n", shell)
	fmt.Println()

	// Clean up any existing completion setup
	fmt.Println("Cleaning up any existing completion setup...")
	cleanupExistingCompletion(shell)

	// Set up completion for the detected shell
	fmt.Printf("Setting up %s completion...\n", shell)
	switch shell {
	case "bash":
		setupBashCompletion()
	case "zsh":
		setupZshCompletion()
	case "fish":
		setupFishCompletion()
	default:
		fmt.Printf("Shell '%s' not supported for completion. Supported shells: bash, zsh, fish\n", shell)
		return
	}

	fmt.Println()
	fmt.Println("✓ Autocompletion setup complete!")
	fmt.Println("  To activate, run one of:")
	
	homeDir, _ := os.UserHomeDir()
	switch shell {
	case "bash":
		fmt.Printf("    source ~/.bashrc\n")
		fmt.Printf("    source %s\n", filepath.Join(homeDir, ".note.bash"))
	case "zsh":
		fmt.Printf("    source ~/.zshrc\n")
		fmt.Printf("    source %s\n", filepath.Join(homeDir, ".note.zsh"))
	case "fish":
		fmt.Println("    (restart your shell)")
	}
	fmt.Println("  Or simply restart your shell")
}

func cleanupExistingCompletion(shell string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	switch shell {
	case "bash":
		// Remove existing .note.bash file
		bashCompletionFile := filepath.Join(homeDir, ".note.bash")
		os.Remove(bashCompletionFile)
		
		// Clean up shell config files
		cleanupShellConfig(filepath.Join(homeDir, ".bashrc"))
		cleanupShellConfig(filepath.Join(homeDir, ".bash_profile"))
		cleanupShellConfig(filepath.Join(homeDir, ".profile"))
		
	case "zsh":
		// Remove existing .note.zsh file
		zshCompletionFile := filepath.Join(homeDir, ".note.zsh")
		os.Remove(zshCompletionFile)
		
		// Clean up .zshrc
		cleanupShellConfig(filepath.Join(homeDir, ".zshrc"))
		
	case "fish":
		// Remove existing fish completion file
		fishCompletionDir := filepath.Join(homeDir, ".config", "fish", "completions")
		noteCompletionFile := filepath.Join(fishCompletionDir, "note.fish")
		os.Remove(noteCompletionFile)
	}
}

func cleanupShellConfig(configFile string) {
	// Read the file
	file, err := os.Open(configFile)
	if err != nil {
		return
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	skipNext := false
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// Skip lines that contain note completion references
		if strings.Contains(line, "# note command completion") {
			skipNext = true
			continue
		}
		
		if skipNext && (strings.Contains(line, ".note.bash") || 
			strings.Contains(line, ".note.zsh") || 
			strings.Contains(line, "completions/bash/note") ||
			(strings.Contains(line, "note") && strings.Contains(line, "source"))) {
			skipNext = false
			continue
		}
		
		if skipNext && strings.TrimSpace(line) == "" {
			continue
		}
		
		skipNext = false
		lines = append(lines, line)
	}

	// Write the cleaned file back
	outFile, err := os.Create(configFile)
	if err != nil {
		return
	}
	defer outFile.Close()

	for _, line := range lines {
		fmt.Fprintln(outFile, line)
	}
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

func listNotes(config Config, pattern string, includeArchived bool) {
	dirs := []string{config.NotesDir}
	if includeArchived {
		archiveDir := filepath.Join(config.NotesDir, "Archive")
		dirs = append(dirs, archiveDir)
	}

	var allNotes []string
	for _, dir := range dirs {
		notes := findMatchingNotes(dir, pattern, true)
		if includeArchived && dir != config.NotesDir {
			// Prefix archived notes for clarity
			for i, note := range notes {
				notes[i] = "Archive/" + note
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
		archiveDir := filepath.Join(config.NotesDir, "Archive")
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

	archiveDir := filepath.Join(config.NotesDir, "Archive")
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
  note [OPTIONS]

OPTIONS:

  -ls, -l [pattern]        List notes (optionally matching pattern)
  -s [term]                Full-text search in notes
  -rm [pattern]            Archive matching notes
  -a [pattern]             List/search all notes including archived

  --help, -h               Show this help message
  --config                 Run setup/reconfigure
  --autocomplete           Setup/update command line autocompletion

EXAMPLES:
  note meeting             Creates meeting-20260108.md
  note project-ideas       Creates project-ideas-20260108.md
  note -ls                 List all current notes
  note -ls project         List notes containing "project"
  note -s "todo"           Search for "todo" in all notes
  note -rm old-*           Archive notes starting with "old-"
  note -a                  List all notes including archived

CONFIGURATION:
  Settings are stored in ~/.note
  Use 'note --config' to reconfigure

LICENSE:
  This program is free software licensed under GPL-3.0.
  See <https://www.gnu.org/licenses/> for details.

For more information, see: https://github.com/bobby/note`)
}
