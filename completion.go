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
	"os/exec"
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

	// First check centralized config
	_, hasCompletion := GetCentralizedConfigStatus(shell)
	if hasCompletion {
		return true
	}

	// Fall back to checking legacy locations for backward compatibility
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
		// Check for centralized config
		if strings.Contains(line, BashCentralizedConfig) || strings.Contains(line, ZshCentralizedConfig) ||
			strings.Contains(line, FishCentralizedConfig) {
			return true
		}
		// Check for legacy config
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

	// Get current alias status to preserve it
	hasAliases, _ := GetCentralizedConfigStatus("bash")

	// Write centralized config with completion enabled
	if err := WriteCentralizedConfig("bash", hasAliases, true); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing centralized config: %v\n", err)
		return
	}

	// Ensure source line exists in .bashrc
	if err := EnsureSourceLine("bash"); err != nil {
		fmt.Fprintf(os.Stderr, "Error adding source line: %v\n", err)
		return
	}

	// Clean up legacy config
	CleanupLegacyConfig("bash")

	configPath := filepath.Join(homeDir, BashCentralizedConfig)
	fmt.Printf("✓ Bash completion setup complete!\n")
	fmt.Printf("  Created centralized config at %s\n", configPath)
	fmt.Printf("  Run 'source ~/.bashrc' or restart your shell to activate completions\n")
}

// SetupZshCompletion sets up zsh command completion
func SetupZshCompletion() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		return
	}

	// Get current alias status to preserve it
	hasAliases, _ := GetCentralizedConfigStatus("zsh")

	// Write centralized config with completion enabled
	if err := WriteCentralizedConfig("zsh", hasAliases, true); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing centralized config: %v\n", err)
		return
	}

	// Ensure source line exists in .zshrc
	if err := EnsureSourceLine("zsh"); err != nil {
		fmt.Fprintf(os.Stderr, "Error adding source line: %v\n", err)
		return
	}

	// Clean up legacy config
	CleanupLegacyConfig("zsh")

	configPath := filepath.Join(homeDir, ZshCentralizedConfig)
	fmt.Printf("✓ Zsh completion setup complete!\n")
	fmt.Printf("  Created centralized config at %s\n", configPath)
	fmt.Printf("  Restart your shell or run: source ~/.zshrc\n")
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
# Main command
complete -c note -f
complete -c note -s l -d "List notes"
complete -c note -s s -d "Search notes" -r
complete -c note -s a -d "Include archived notes"
complete -c note -s d -d "Archive notes" -r
complete -c note -l config -d "Run setup/reconfigure"
complete -c note -l autocomplete -d "Setup/update command line autocompletion"
complete -c note -l alias -d "Setup shell aliases"
complete -c note -s v -l version -d "Show version"
complete -c note -s h -l help -d "Show help"

# Complete with existing note names for main argument
complete -c note -n '__fish_is_first_token' -a '(if test -f ~/.note; set notesdir (grep "^notesdir=" ~/.note | cut -d= -f2 | sed "s|~|$HOME|"); if test -d "$notesdir"; find "$notesdir" -maxdepth 1 -name "*.md" -type f -exec basename {} .md \\; 2>/dev/null | sort; end; end)'

# Alias: n (same as note)
complete -c n -f
complete -c n -s l -d "List notes"
complete -c n -s s -d "Search notes" -r
complete -c n -s a -d "Include archived notes"
complete -c n -s d -d "Archive notes" -r
complete -c n -l config -d "Run setup/reconfigure"
complete -c n -l autocomplete -d "Setup/update command line autocompletion"
complete -c n -l alias -d "Setup shell aliases"
complete -c n -s v -l version -d "Show version"
complete -c n -s h -l help -d "Show help"
complete -c n -n '__fish_is_first_token' -a '(if test -f ~/.note; set notesdir (grep "^notesdir=" ~/.note | cut -d= -f2 | sed "s|~|$HOME|"); if test -d "$notesdir"; find "$notesdir" -maxdepth 1 -name "*.md" -type f -exec basename {} .md \\; 2>/dev/null | sort; end; end)'

# Alias: nls (note -l)
complete -c nls -f
complete -c nls -n '__fish_is_first_token' -a '(if test -f ~/.note; set notesdir (grep "^notesdir=" ~/.note | cut -d= -f2 | sed "s|~|$HOME|"); if test -d "$notesdir"; find "$notesdir" -maxdepth 1 -name "*.md" -type f -exec basename {} .md \\; 2>/dev/null | sort; end; end)'

# Alias: nrm (note -d)
complete -c nrm -f
complete -c nrm -n '__fish_is_first_token' -a '(if test -f ~/.note; set notesdir (grep "^notesdir=" ~/.note | cut -d= -f2 | sed "s|~|$HOME|"); if test -d "$notesdir"; find "$notesdir" -maxdepth 1 -name "*.md" -type f -exec basename {} .md \\; 2>/dev/null | sort; end; end)'
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
		fmt.Printf("    source ~/%s\n", BashCentralizedConfig)
	case "zsh":
		fmt.Printf("    source ~/.zshrc\n")
		fmt.Printf("    source ~/%s\n", ZshCentralizedConfig)
	case "fish":
		fmt.Println("    (restart your shell)")
		fmt.Printf("    source %s\n", filepath.Join(homeDir, ".config", "fish", "completions", "note.fish"))
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
		// Remove centralized config file
		centralizedFile := filepath.Join(homeDir, BashCentralizedConfig)
		os.Remove(centralizedFile)

		// Remove legacy .note.bash file
		legacyFile := filepath.Join(homeDir, ".note.bash")
		os.Remove(legacyFile)

		// Clean up shell config files
		cleanupShellConfig(filepath.Join(homeDir, ".bashrc"))
		cleanupShellConfig(filepath.Join(homeDir, ".bash_profile"))
		cleanupShellConfig(filepath.Join(homeDir, ".profile"))

	case "zsh":
		// Remove centralized config file
		centralizedFile := filepath.Join(homeDir, ZshCentralizedConfig)
		os.Remove(centralizedFile)

		// Remove legacy .note.zsh file
		legacyFile := filepath.Join(homeDir, ".note.zsh")
		os.Remove(legacyFile)

		// Clean up .zshrc
		cleanupShellConfig(filepath.Join(homeDir, ".zshrc"))

	case "fish":
		// Remove centralized config file
		centralizedFile := filepath.Join(homeDir, FishCentralizedConfig)
		os.Remove(centralizedFile)

		// Remove existing fish completion file
		fishCompletionDir := filepath.Join(homeDir, ".config", "fish", "completions")
		noteCompletionFile := filepath.Join(fishCompletionDir, "note.fish")
		os.Remove(noteCompletionFile)

		// Clean up fish config
		fishConfig := filepath.Join(homeDir, ".config", "fish", "config.fish")
		cleanupShellConfig(fishConfig)
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

		// Skip Note CLI integration header and following source line
		if strings.Contains(line, "# Note CLI integration") {
			skipNext = true
			continue
		}

		// Skip centralized config source lines
		if strings.Contains(line, BashCentralizedConfig) ||
			strings.Contains(line, ZshCentralizedConfig) ||
			strings.Contains(line, FishCentralizedConfig) {
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

// Centralized config file paths
const (
	BashCentralizedConfig = ".note_bash_rc"
	ZshCentralizedConfig  = ".note_zsh_rc"
	FishCentralizedConfig = ".note_fish_rc"
)

// generateBashConfig generates the complete bash config content
func generateBashConfig(aliasesEnabled, completionEnabled bool, notePath string) string {
	var content strings.Builder

	content.WriteString("# Note CLI Shell Integration\n")
	content.WriteString("# Generated by note CLI - Do not edit manually\n")
	content.WriteString("# Regenerate with: note --configure\n\n")

	if aliasesEnabled {
		content.WriteString("# ============= ALIASES =============\n")
		content.WriteString(fmt.Sprintf("alias n='%s'\n", notePath))
		content.WriteString(fmt.Sprintf("alias nls='%s -l'\n", notePath))
		content.WriteString(fmt.Sprintf("alias nrm='%s -d'\n", notePath))
		content.WriteString("\n")
	}

	if completionEnabled {
		content.WriteString("# ============= COMPLETION =============\n")
		content.WriteString(`_note_complete() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local prev="${COMP_WORDS[COMP_CWORD-1]}"

    # If we're on the first argument
    if [[ ${COMP_CWORD} -eq 1 ]]; then
        # If user starts typing a dash, offer flags
        if [[ "$cur" == -* ]]; then
            local flags="-l -s -a -d -v --config --configure --autocomplete --alias --help --version -h"
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
    # If previous was -l, -a, or -d, offer note names
    elif [[ "$prev" == "-l" || "$prev" == "-a" || "$prev" == "-d" ]]; then
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

# Register completion for note and its aliases
complete -F _note_complete note
complete -F _note_complete n
complete -F _note_complete nls
complete -F _note_complete nrm
`)
	}

	return content.String()
}

// generateZshConfig generates the complete zsh config content
func generateZshConfig(aliasesEnabled, completionEnabled bool, notePath string) string {
	var content strings.Builder

	content.WriteString("# Note CLI Shell Integration\n")
	content.WriteString("# Generated by note CLI - Do not edit manually\n")
	content.WriteString("# Regenerate with: note --configure\n\n")

	if aliasesEnabled {
		content.WriteString("# ============= ALIASES =============\n")
		content.WriteString(fmt.Sprintf("alias n='%s'\n", notePath))
		content.WriteString(fmt.Sprintf("alias nls='%s -l'\n", notePath))
		content.WriteString(fmt.Sprintf("alias nrm='%s -d'\n", notePath))
		content.WriteString("\n")
	}

	if completionEnabled {
		content.WriteString("# ============= COMPLETION =============\n")
		content.WriteString("autoload -U +X compinit && compinit\n\n")
		content.WriteString(`_note_complete() {
    local cur="${words[CURRENT]}"
    local prev="${words[CURRENT-1]}"

    # If we're on the first argument
    if [[ $CURRENT -eq 2 ]]; then
        # If user starts typing a dash, offer flags
        if [[ "$cur" == -* ]]; then
            local flags=("-l" "-s" "-a" "-d" "-v" "--config" "--configure" "--autocomplete" "--alias" "--help" "--version" "-h")
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

    # If previous was -l, -a, or -d, offer note names
    elif [[ "$prev" == "-l" || "$prev" == "-a" || "$prev" == "-d" ]]; then
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

# Register completion for note and its aliases
compdef _note_complete note
compdef _note_complete n
compdef _note_complete nls
compdef _note_complete nrm
`)
	}

	return content.String()
}

// generateFishConfig generates the fish config content (aliases only, completion stays in standard location)
func generateFishConfig(aliasesEnabled bool, notePath string) string {
	var content strings.Builder

	content.WriteString("# Note CLI Shell Integration\n")
	content.WriteString("# Generated by note CLI - Do not edit manually\n")
	content.WriteString("# Regenerate with: note --configure\n\n")

	if aliasesEnabled {
		content.WriteString("# ============= ALIASES =============\n")
		content.WriteString(fmt.Sprintf("alias n '%s'\n", notePath))
		content.WriteString(fmt.Sprintf("alias nls '%s -l'\n", notePath))
		content.WriteString(fmt.Sprintf("alias nrm '%s -d'\n", notePath))
	}

	return content.String()
}

// WriteCentralizedConfig writes the centralized config file for the specified shell
func WriteCentralizedConfig(shell string, aliasesEnabled, completionEnabled bool) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %w", err)
	}

	// Get the path to the note binary
	notePath, err := os.Executable()
	if err != nil {
		// Fallback to checking PATH
		notePath, err = exec.LookPath("note")
		if err != nil {
			return fmt.Errorf("could not determine note command path: %w", err)
		}
	}

	var configPath string
	var content string

	switch shell {
	case "bash":
		configPath = filepath.Join(homeDir, BashCentralizedConfig)
		content = generateBashConfig(aliasesEnabled, completionEnabled, notePath)
	case "zsh":
		configPath = filepath.Join(homeDir, ZshCentralizedConfig)
		content = generateZshConfig(aliasesEnabled, completionEnabled, notePath)
	case "fish":
		configPath = filepath.Join(homeDir, FishCentralizedConfig)
		content = generateFishConfig(aliasesEnabled, notePath)
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// EnsureSourceLine adds the source line to the shell's RC file if not already present
func EnsureSourceLine(shell string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %w", err)
	}

	var rcPath string
	var sourceLine string
	var configFile string

	switch shell {
	case "bash":
		rcPath = filepath.Join(homeDir, ".bashrc")
		configFile = BashCentralizedConfig
		sourceLine = fmt.Sprintf("\n# Note CLI integration\n[ -f ~/%s ] && source ~/%s\n", configFile, configFile)
	case "zsh":
		rcPath = filepath.Join(homeDir, ".zshrc")
		configFile = ZshCentralizedConfig
		sourceLine = fmt.Sprintf("\n# Note CLI integration\n[ -f ~/%s ] && source ~/%s\n", configFile, configFile)
	case "fish":
		// Create fish config directory if it doesn't exist
		fishConfigDir := filepath.Join(homeDir, ".config", "fish")
		if err := os.MkdirAll(fishConfigDir, 0755); err != nil {
			return fmt.Errorf("error creating fish config directory: %w", err)
		}
		rcPath = filepath.Join(fishConfigDir, "config.fish")
		configFile = FishCentralizedConfig
		sourceLine = fmt.Sprintf("\n# Note CLI integration\ntest -f ~/%s; and source ~/%s\n", configFile, configFile)
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}

	// Check if source line already exists
	if content, err := os.ReadFile(rcPath); err == nil {
		contentStr := string(content)
		// Check for either the config file name or the full source pattern
		if strings.Contains(contentStr, configFile) ||
			strings.Contains(contentStr, "# Note CLI integration") {
			// Source line already exists
			return nil
		}
	}

	// Append source line to RC file
	file, err := os.OpenFile(rcPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error opening %s: %w", rcPath, err)
	}
	defer file.Close()

	if _, err := file.WriteString(sourceLine); err != nil {
		return fmt.Errorf("error writing to %s: %w", rcPath, err)
	}

	return nil
}

// GetCentralizedConfigStatus checks what features are enabled in the centralized config
func GetCentralizedConfigStatus(shell string) (hasAliases, hasCompletion bool) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, false
	}

	var configPath string
	switch shell {
	case "bash":
		configPath = filepath.Join(homeDir, BashCentralizedConfig)
	case "zsh":
		configPath = filepath.Join(homeDir, ZshCentralizedConfig)
	case "fish":
		configPath = filepath.Join(homeDir, FishCentralizedConfig)
		// For fish, completion is stored separately in the standard location
		fishCompletionDir := filepath.Join(homeDir, ".config", "fish", "completions")
		fishCompletionFile := filepath.Join(fishCompletionDir, "note.fish")
		if _, err := os.Stat(fishCompletionFile); err == nil {
			hasCompletion = true
		}
	default:
		return false, false
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		// For fish, we may have completion but no config file (aliases)
		// Return what we've already detected
		return hasAliases, hasCompletion
	}

	contentStr := string(content)
	hasAliases = strings.Contains(contentStr, "# ============= ALIASES =============")
	if shell != "fish" {
		hasCompletion = strings.Contains(contentStr, "# ============= COMPLETION =============")
	}

	return hasAliases, hasCompletion
}

// CleanupLegacyConfig removes old-style configuration files and inline entries
func CleanupLegacyConfig(shell string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %w", err)
	}

	switch shell {
	case "bash":
		// Remove old .note.bash file
		legacyBashFile := filepath.Join(homeDir, ".note.bash")
		os.Remove(legacyBashFile)

		// Clean up legacy entries from .bashrc
		bashrc := filepath.Join(homeDir, ".bashrc")
		cleanupLegacyShellConfig(bashrc)
		cleanupLegacyShellConfig(filepath.Join(homeDir, ".bash_profile"))
		cleanupLegacyShellConfig(filepath.Join(homeDir, ".profile"))

	case "zsh":
		// Remove old .note.zsh file
		legacyZshFile := filepath.Join(homeDir, ".note.zsh")
		os.Remove(legacyZshFile)

		// Clean up legacy entries from .zshrc
		zshrc := filepath.Join(homeDir, ".zshrc")
		cleanupLegacyShellConfig(zshrc)

	case "fish":
		// Clean up legacy entries from config.fish
		fishConfig := filepath.Join(homeDir, ".config", "fish", "config.fish")
		cleanupLegacyFishConfig(fishConfig)
	}

	return nil
}

// cleanupLegacyShellConfig removes old note command aliases and completion source lines from shell config
func cleanupLegacyShellConfig(configFile string) {
	content, err := os.ReadFile(configFile)
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	inNoteSection := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Detect start of note sections
		if trimmedLine == "# note command aliases" || trimmedLine == "# note command completion" {
			inNoteSection = true
			continue
		}

		// Skip old source lines for .note.bash/.note.zsh
		if strings.Contains(line, ".note.bash") || strings.Contains(line, ".note.zsh") {
			continue
		}

		// Skip inline alias definitions that are part of note
		if inNoteSection {
			// Check if this line is a note-related alias
			if (strings.Contains(line, "alias n=") || strings.Contains(line, "alias nls=") || strings.Contains(line, "alias nrm=")) &&
				strings.Contains(line, "note") {
				continue
			}
			// Check if this line is a note-related source/completion line
			if strings.Contains(line, "source") && strings.Contains(line, "note") {
				continue
			}
			if strings.Contains(line, "autoload") && strings.Contains(line, "compinit") {
				continue
			}
			// If we hit a non-empty, non-note line, we're out of the section
			if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "#") {
				inNoteSection = false
			}
			// Skip empty lines within the note section
			if trimmedLine == "" {
				continue
			}
		}

		// Skip standalone note alias lines (not in a section)
		if (strings.Contains(line, "alias n='") || strings.Contains(line, "alias nls='") || strings.Contains(line, "alias nrm='")) &&
			strings.Contains(line, "note") {
			continue
		}

		newLines = append(newLines, line)
	}

	// Remove consecutive empty lines at the end
	for len(newLines) > 1 && strings.TrimSpace(newLines[len(newLines)-1]) == "" && strings.TrimSpace(newLines[len(newLines)-2]) == "" {
		newLines = newLines[:len(newLines)-1]
	}

	// Write cleaned content back
	newContent := strings.Join(newLines, "\n")
	if len(newContent) > 0 && !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}
	os.WriteFile(configFile, []byte(newContent), 0644)
}

// cleanupLegacyFishConfig removes old note command aliases from fish config
func cleanupLegacyFishConfig(configFile string) {
	content, err := os.ReadFile(configFile)
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	skipNext := false
	skipCount := 0

	for _, line := range lines {
		// Skip "# note command aliases" header and following alias lines
		if strings.Contains(line, "# note command aliases") {
			skipNext = true
			skipCount = 3 // Skip the next 3 lines (the alias lines)
			continue
		}

		// Skip inline fish alias definitions
		if (strings.Contains(line, "alias n ") || strings.Contains(line, "alias nls ") || strings.Contains(line, "alias nrm ")) &&
			strings.Contains(line, "note") {
			continue
		}

		if skipNext && skipCount > 0 {
			skipCount--
			if skipCount == 0 {
				skipNext = false
			}
			continue
		}

		newLines = append(newLines, line)
	}

	// Remove trailing empty lines
	for len(newLines) > 0 && strings.TrimSpace(newLines[len(newLines)-1]) == "" {
		newLines = newLines[:len(newLines)-1]
	}

	// Write cleaned content back
	newContent := strings.Join(newLines, "\n")
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}
	os.WriteFile(configFile, []byte(newContent), 0644)
}
