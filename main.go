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

func main() {
	config := loadOrCreateConfig()

	// Parse flags
	var (
		listFlag    = flag.Bool("ls", false, "List all current notes")
		listFlagAlt = flag.Bool("l", false, "List all current notes (short form)")
		searchFlag  = flag.String("s", "", "Full-text search in notes")
		archiveFlag = flag.Bool("a", false, "List/search all notes including archived")
		removeFlag  = flag.String("rm", "", "Archive matching notes")
		configFlag  = flag.Bool("config", false, "Run setup/reconfigure")
		helpFlag    = flag.Bool("help", false, "Show help")
		helpFlagAlt = flag.Bool("h", false, "Show help (short form)")
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

	// Handle listing
	if *listFlag || *listFlagAlt {
		pattern := ""
		if flag.NArg() > 0 {
			pattern = flag.Arg(0)
		}
		listNotes(config, pattern, false)
		return
	}

	// Handle archive listing
	if *archiveFlag {
		pattern := ""
		if flag.NArg() > 0 {
			pattern = flag.Arg(0)
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

	noteName := flag.Arg(0)
	openOrCreateNote(config, noteName)
}

func loadOrCreateConfig() Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		os.Exit(1)
	}

	configPath := filepath.Join(homeDir, ".note")

	// Check if config exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// First run, create config
		return runSetup()
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
		return runSetup()
	}

	return config
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

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, path[2:])
	}
	return path
}

func openOrCreateNote(config Config, noteName string) {
	// Check if it's a specific file with date already
	if strings.HasSuffix(noteName, ".md") {
		// Open specific file
		notePath := filepath.Join(config.NotesDir, noteName)
		openInEditor(config.Editor, notePath)
		return
	}

	// Generate today's date
	today := time.Now().Format("20060102")
	filename := fmt.Sprintf("%s-%s.md", noteName, today)
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

	// Create new note
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
		fmt.Println(note)
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
		if pattern == "" || strings.Contains(strings.ToLower(info.Name()), strings.ToLower(pattern)) {
			notes = append(notes, info.Name())
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
  --config                 Run setup/reconfigure
  -ls, -l [pattern]        List notes (optionally matching pattern)
  -s [term]                Full-text search in notes
  -rm [pattern]            Archive matching notes
  -a [pattern]             List/search all notes including archived
  --help, -h               Show this help message

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

For more information, see: https://github.com/bobby/note`)
}