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
	"os"
	"path/filepath"
	"strings"
)

// SetupCompletion handles the interactive completion setup prompt
func SetupCompletion(reader *bufio.Reader) {
	// Check if completion is already set up
	if IsCompletionAlreadySetup() {
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
		SetupBashCompletion()
	case "zsh":
		SetupZshCompletion()
	case "fish":
		SetupFishCompletion()
	default:
		fmt.Printf("Shell '%s' not supported for completion. Supported shells: bash, zsh, fish\n", shell)
	}
}

// IsCompletionAlreadySetup checks if command line completion is already configured
func IsCompletionAlreadySetup() bool {
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
				if CheckFileForCompletionSource(filepath.Join(homeDir, file)) {
					return true
				}
			}
		}
	case "zsh":
		zshCompletionFile := filepath.Join(homeDir, ".note.zsh")
		if _, err := os.Stat(zshCompletionFile); err == nil {
			if CheckFileForCompletionSource(filepath.Join(homeDir, ".zshrc")) {
				return true
			}
		}
	case "fish":
		// Check fish completion directory
		fishCompletionDir := filepath.Join(homeDir, ".config", "fish", "completions")
		fishCompletionFile := filepath.Join(fishCompletionDir, "note.fish")
		_, err := os.Stat(fishCompletionFile)
		return err == nil
	}
	return false
}

// CheckFileForCompletionSource checks if a file sources note completion
func CheckFileForCompletionSource(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if (strings.Contains(line, "~/.note.bash") || strings.Contains(line, "~/.note.zsh")) &&
		   (strings.Contains(line, "source") || strings.Contains(line, ".")) ||
		   (strings.Contains(line, "note") && (strings.Contains(line, "complete") || strings.Contains(line, "completion"))) {
			return true
		}
	}
	return false
}

// SetupBashCompletion sets up bash command completion
func SetupBashCompletion() {
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
		fmt.Fprintf(os.Stderr, "Error writing bash completion script: %v\n", err)
		return
	}

	// Add source line to .bashrc
	bashrc := filepath.Join(homeDir, ".bashrc")
	sourceLine := fmt.Sprintf("\n# note command completion\nsource ~/.note.bash\n")
	
	file, err := os.OpenFile(bashrc, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening .bashrc: %v\n", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(sourceLine); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating .bashrc: %v\n", err)
		return
	}

	fmt.Printf("✓ Bash completion setup complete!\n")
	fmt.Printf("  Created completion script at %s\n", completionScriptPath)
	fmt.Printf("  Updated %s to source completion\n", bashrc)
	fmt.Printf("  Run 'source ~/.bashrc' or restart your shell to activate completions\n")
}

// SetupZshCompletion sets up zsh command completion
func SetupZshCompletion() {
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

// SetupFishCompletion sets up fish command completion
func SetupFishCompletion() {
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

// RunAutocompleteSetup handles the main autocomplete setup flow
func RunAutocompleteSetup() {
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
	CleanupExistingCompletion(shell)

	// Set up completion for the detected shell
	fmt.Printf("Setting up %s completion...\n", shell)
	switch shell {
	case "bash":
		SetupBashCompletion()
	case "zsh":
		SetupZshCompletion()
	case "fish":
		SetupFishCompletion()
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

// CleanupExistingCompletion removes existing completion setup for the specified shell
func CleanupExistingCompletion(shell string) {
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

// cleanupShellConfig removes note completion lines from shell config files
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

// detectShell detects the current shell from environment variables
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