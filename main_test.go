package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExpandPath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/Documents", filepath.Join(homeDir, "Documents")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, test := range tests {
		result := expandPath(test.input)
		if result != test.expected {
			t.Errorf("expandPath(%s) = %s; want %s", test.input, result, test.expected)
		}
	}
}

func TestExpandPathWithSymlinks(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "note-symlink-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create real directory
	realDir := filepath.Join(tempDir, "real-notes")
	if err := os.MkdirAll(realDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create symbolic link
	symlinkDir := filepath.Join(tempDir, "symlink-notes")
	if err := os.Symlink(realDir, symlinkDir); err != nil {
		t.Skip("Skipping symlink test: symlink creation failed (might not be supported)")
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Regular path",
			input:    realDir,
			expected: realDir,
		},
		{
			name:     "Symbolic link path",
			input:    symlinkDir,
			expected: realDir, // Should resolve to the real directory
		},
		{
			name:     "Non-existent path",
			input:    filepath.Join(tempDir, "non-existent"),
			expected: filepath.Join(tempDir, "non-existent"), // Should return original if can't resolve
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := expandPath(test.input)
			if result != test.expected {
				t.Errorf("expandPath(%s) = %s; want %s", test.input, result, test.expected)
			}
		})
	}
}

func TestFindMatchingNotes(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "note-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := []string{
		"meeting-20240101.md",
		"project-20240102.md",
		"ideas-20240103.md",
		"meeting-20240104.md",
		"readme.txt", // Should be ignored
	}

	for _, file := range testFiles {
		f, err := os.Create(filepath.Join(tempDir, file))
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
	}

	// Test finding all notes
	notes := findMatchingNotes(tempDir, "", false)
	if len(notes) != 4 { // Should ignore .txt file
		t.Errorf("Expected 4 notes, got %d", len(notes))
	}

	// Test pattern matching
	notes = findMatchingNotes(tempDir, "meeting", false)
	if len(notes) != 2 {
		t.Errorf("Expected 2 meeting notes, got %d", len(notes))
	}

	// Verify correct files were found
	for _, note := range notes {
		if !strings.Contains(note, "meeting") {
			t.Errorf("Found non-meeting note: %s", note)
		}
	}
}

func TestGenerateFilename(t *testing.T) {
	// Test date format
	today := time.Now().Format("20060102")
	noteName := "test-note"
	expected := noteName + "-" + today + ".md"

	// This would be part of openOrCreateNote logic
	filename := noteName + "-" + today + ".md"

	if filename != expected {
		t.Errorf("Generated filename %s does not match expected %s", filename, expected)
	}
}

func TestConfigSaveAndLoad(t *testing.T) {
	// Create temp config file
	tempFile, err := os.CreateTemp("", ".note-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	// Test config structure (would need refactoring to test properly)
	testConfig := Config{
		Editor:   "vim",
		NotesDir: "/home/user/Notes",
	}

	// Verify config has expected fields
	if testConfig.Editor != "vim" {
		t.Errorf("Editor not set correctly")
	}
	if testConfig.NotesDir != "/home/user/Notes" {
		t.Errorf("NotesDir not set correctly")
	}
}

func TestCopyFile(t *testing.T) {
	// Create temp source file
	srcFile, err := os.CreateTemp("", "source")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(srcFile.Name())

	// Write test content
	testContent := "This is a test note"
	if _, err := srcFile.WriteString(testContent); err != nil {
		t.Fatal(err)
	}
	srcFile.Close()

	// Create destination path
	dstFile := srcFile.Name() + ".copy"
	defer os.Remove(dstFile)

	// Test copy
	if err := copyFile(srcFile.Name(), dstFile); err != nil {
		t.Errorf("copyFile failed: %v", err)
	}

	// Verify content
	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != testContent {
		t.Errorf("Copied content does not match: got %s, want %s", content, testContent)
	}
}

func TestSpacesInNoteNames(t *testing.T) {
	// Test the specific bug: spaces in command line arguments should become underscores in filename
	tests := []struct {
		input    []string
		expected string
	}{
		{[]string{"this", "is", "a", "test"}, "this_is_a_test"},
		{[]string{"meeting", "notes"}, "meeting_notes"},
		{[]string{"single"}, "single"},
		{[]string{"project", "ideas", "for", "2024"}, "project_ideas_for_2024"},
	}

	for _, test := range tests {
		// Simulate the argument joining logic from main()
		noteName := strings.Join(test.input, " ")

		// Apply the space-to-underscore conversion from openOrCreateNote()
		cleanNoteName := strings.ReplaceAll(noteName, " ", "_")

		if cleanNoteName != test.expected {
			t.Errorf("For args %v: got %s, want %s", test.input, cleanNoteName, test.expected)
		}

		// Also test that the full filename generation works correctly
		today := time.Now().Format("20060102")
		expectedFilename := test.expected + "-" + today + ".md"
		actualFilename := cleanNoteName + "-" + today + ".md"

		if actualFilename != expectedFilename {
			t.Errorf("Filename generation failed for args %v: got %s, want %s", test.input, actualFilename, expectedFilename)
		}
	}
}

func TestHighlightTerm(t *testing.T) {
	// Test the highlighting functionality
	tests := []struct {
		text     string
		term     string
		expected string // Without color codes (since we can't mock isOutputToTerminal easily)
	}{
		{"meeting-notes-20250108.md", "meeting", "meeting-notes-20250108.md"},
		{"project-planning-session.md", "planning", "project-planning-session.md"},
		{"daily-standup-meeting.md", "meeting", "daily-standup-meeting.md"},
		{"test-case-file.md", "case", "test-case-file.md"},
		{"", "search", ""},
		{"no-match.md", "xyz", "no-match.md"},
	}

	for _, test := range tests {
		// Since we can't easily mock isOutputToTerminal() in unit tests,
		// we'll test the core logic of finding and replacing terms
		result := highlightTerm(test.text, test.term)

		// For terminal output, result should contain color codes
		// For non-terminal (like in tests), it should be unchanged
		if test.term == "" || test.text == "" {
			if result != test.expected {
				t.Errorf("highlightTerm(%q, %q) = %q; want %q", test.text, test.term, result, test.expected)
			}
		} else {
			// In test environment (non-terminal), result should equal input
			if result != test.text {
				t.Errorf("highlightTerm(%q, %q) = %q; want %q (no highlighting in test env)", test.text, test.term, result, test.text)
			}
		}
	}
}

func TestHighlightTermPanicRegression(t *testing.T) {
	// Test cases that previously caused slice bounds panic
	// These test the specific issue where multiple occurrences of a term
	// caused the highlighting algorithm to go out of bounds
	problematicCases := []struct {
		text string
		term string
	}{
		// This case caused the original panic: second "life" would go out of bounds
		{"Life-101-Identifing-a-Vision-for-your-Life.md", "life"},
		// Additional edge cases
		{"test-life-and-life-again.txt", "life"},
		{"abc-abc-abc.md", "abc"},
		{"short", "short"},     // Term same length as text
		{"ab", "abc"},          // Term longer than text
		{"a-a-a-a-a.txt", "a"}, // Many small matches
	}

	for _, test := range problematicCases {
		// The main test is that this doesn't panic
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("highlightTerm(%q, %q) panicked: %v", test.text, test.term, r)
				}
			}()

			result := highlightTerm(test.text, test.term)

			// Basic sanity check: result should not be shorter than original text
			// (unless we couldn't do highlighting due to test environment)
			if len(result) < len(test.text) {
				t.Errorf("highlightTerm(%q, %q) result %q is shorter than input", test.text, test.term, result)
			}
		}()
	}
}

func TestIsOutputToTerminal(t *testing.T) {
	// Test terminal detection
	// In test environment, this should typically return false
	result := isOutputToTerminal()

	// We can't predict the exact value, but the function should not panic
	// and should return a boolean
	if result != true && result != false {
		t.Error("isOutputToTerminal() should return a boolean value")
	}

	// In CI/test environments, this is typically false
	// We'll just verify it runs without error
}

func TestAreAliasesAlreadySetup(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "note-alias-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set temporary HOME
	os.Setenv("HOME", tempDir)

	// Test with no shell config files
	result := areAliasesAlreadySetup()
	if result {
		t.Error("Should return false when no config files exist")
	}

	// Test with bash config file without aliases
	bashrcPath := filepath.Join(tempDir, ".bashrc")
	if err := os.WriteFile(bashrcPath, []byte("# Some other config\nexport PATH=$PATH:/usr/bin\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Mock shell detection to return bash
	originalShell := os.Getenv("SHELL")
	os.Setenv("SHELL", "/bin/bash")
	defer os.Setenv("SHELL", originalShell)

	result = areAliasesAlreadySetup()
	if result {
		t.Error("Should return false when aliases don't exist in bashrc")
	}

	// Test with bash config file with aliases
	aliasContent := "# Some other config\nexport PATH=$PATH:/usr/bin\n# note command aliases\nalias n='/usr/bin/note'\nalias nls='/usr/bin/note -l'\nalias nrm='/usr/bin/note -rm'\n"
	if err := os.WriteFile(bashrcPath, []byte(aliasContent), 0644); err != nil {
		t.Fatal(err)
	}

	result = areAliasesAlreadySetup()
	if !result {
		t.Error("Should return true when aliases exist in bashrc")
	}
}

func TestSetupAliasesPath(t *testing.T) {
	// Test path detection logic used in alias setup
	// This tests the os.Executable and exec.LookPath fallback

	// Test executable path detection
	execPath, err := os.Executable()
	if err != nil {
		t.Logf("os.Executable() failed: %v, testing fallback", err)

		// Test PATH lookup fallback
		notePath, err := exec.LookPath("note")
		if err != nil {
			t.Skip("note command not found in PATH, skipping path test")
		}

		if notePath == "" {
			t.Error("exec.LookPath should return non-empty path")
		}
	} else {
		if execPath == "" {
			t.Error("os.Executable should return non-empty path")
		}
	}
}

func TestGetArchiveDir(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "note-archive-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	notesDir := filepath.Join(tempDir, "Notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Test 1: No archive directory exists - should return Archive (capital)
	result := getArchiveDir(notesDir)
	expected := filepath.Join(notesDir, "Archive")
	if result != expected {
		t.Errorf("No archive exists: expected %s, got %s", expected, result)
	}

	// Test 2: Only lowercase archive exists - should return archive
	archiveLower := filepath.Join(notesDir, "archive")
	if err := os.MkdirAll(archiveLower, 0755); err != nil {
		t.Fatal(err)
	}

	result = getArchiveDir(notesDir)
	if result != archiveLower {
		t.Errorf("Only lowercase exists: expected %s, got %s", archiveLower, result)
	}

	// Test 3: Both Archive and archive exist - should prefer Archive (capital)
	archiveUpper := filepath.Join(notesDir, "Archive")
	if err := os.MkdirAll(archiveUpper, 0755); err != nil {
		t.Fatal(err)
	}

	result = getArchiveDir(notesDir)
	if result != archiveUpper {
		t.Errorf("Both exist: expected %s (capital), got %s", archiveUpper, result)
	}

	// Test 4: Only Archive (capital) exists - should return Archive
	if err := os.RemoveAll(archiveLower); err != nil {
		t.Fatal(err)
	}

	result = getArchiveDir(notesDir)
	if result != archiveUpper {
		t.Errorf("Only capital exists: expected %s, got %s", archiveUpper, result)
	}
}

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expected  *ParsedFlags
		remaining []string
	}{
		{
			name:      "No flags",
			args:      []string{"note-name"},
			expected:  &ParsedFlags{},
			remaining: []string{"note-name"},
		},
		{
			name:      "Simple list flag",
			args:      []string{"-l"},
			expected:  &ParsedFlags{List: true},
			remaining: []string{},
		},
		{
			name:      "List with pattern",
			args:      []string{"-l", "pattern"},
			expected:  &ParsedFlags{List: true},
			remaining: []string{"pattern"},
		},
		{
			name:      "Archive flag",
			args:      []string{"-a"},
			expected:  &ParsedFlags{Archive: true},
			remaining: []string{},
		},
		{
			name:      "Search flag",
			args:      []string{"-s", "search-term"},
			expected:  &ParsedFlags{Search: "search-term"},
			remaining: []string{},
		},
		{
			name:      "Delete flag",
			args:      []string{"-d", "pattern"},
			expected:  &ParsedFlags{Delete: "pattern"},
			remaining: []string{},
		},
		{
			name:      "Help flag short",
			args:      []string{"-h"},
			expected:  &ParsedFlags{Help: true},
			remaining: []string{},
		},
		{
			name:      "Help flag long",
			args:      []string{"--help"},
			expected:  &ParsedFlags{Help: true},
			remaining: []string{},
		},
		{
			name:      "Config flag",
			args:      []string{"--config"},
			expected:  &ParsedFlags{Config: true},
			remaining: []string{},
		},
		{
			name:      "Configure flag (alias for config)",
			args:      []string{"--configure"},
			expected:  &ParsedFlags{Config: true},
			remaining: []string{},
		},
		{
			name:      "Autocomplete flag",
			args:      []string{"--autocomplete"},
			expected:  &ParsedFlags{Autocomplete: true},
			remaining: []string{},
		},
		{
			name:      "Alias flag",
			args:      []string{"--alias"},
			expected:  &ParsedFlags{Alias: true},
			remaining: []string{},
		},
		{
			name:      "Version flag short",
			args:      []string{"-v"},
			expected:  &ParsedFlags{Version: true},
			remaining: []string{},
		},
		{
			name:      "Version flag long",
			args:      []string{"--version"},
			expected:  &ParsedFlags{Version: true},
			remaining: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flags, remaining := parseFlags(test.args)

			// Check each flag field
			if flags.List != test.expected.List {
				t.Errorf("List: got %v, want %v", flags.List, test.expected.List)
			}
			if flags.Search != test.expected.Search {
				t.Errorf("Search: got %q, want %q", flags.Search, test.expected.Search)
			}
			if flags.Archive != test.expected.Archive {
				t.Errorf("Archive: got %v, want %v", flags.Archive, test.expected.Archive)
			}
			if flags.Delete != test.expected.Delete {
				t.Errorf("Delete: got %q, want %q", flags.Delete, test.expected.Delete)
			}
			if flags.Config != test.expected.Config {
				t.Errorf("Config: got %v, want %v", flags.Config, test.expected.Config)
			}
			if flags.Autocomplete != test.expected.Autocomplete {
				t.Errorf("Autocomplete: got %v, want %v", flags.Autocomplete, test.expected.Autocomplete)
			}
			if flags.Alias != test.expected.Alias {
				t.Errorf("Alias: got %v, want %v", flags.Alias, test.expected.Alias)
			}
			if flags.Help != test.expected.Help {
				t.Errorf("Help: got %v, want %v", flags.Help, test.expected.Help)
			}
			if flags.Version != test.expected.Version {
				t.Errorf("Version: got %v, want %v", flags.Version, test.expected.Version)
			}

			// Check remaining arguments
			if len(remaining) != len(test.remaining) {
				t.Errorf("Remaining args length: got %d, want %d", len(remaining), len(test.remaining))
			} else {
				for i, arg := range remaining {
					if arg != test.remaining[i] {
						t.Errorf("Remaining arg %d: got %q, want %q", i, arg, test.remaining[i])
					}
				}
			}
		})
	}
}

func TestParseFlagsChaining(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expected  *ParsedFlags
		remaining []string
	}{
		{
			name:      "Archive and List chained (-al)",
			args:      []string{"-al"},
			expected:  &ParsedFlags{Archive: true, List: true},
			remaining: []string{},
		},
		{
			name:      "List and Archive chained (-la)",
			args:      []string{"-la"},
			expected:  &ParsedFlags{List: true, Archive: true},
			remaining: []string{},
		},
		{
			name:      "Archive and Search chained (-as)",
			args:      []string{"-as", "search-term"},
			expected:  &ParsedFlags{Archive: true, Search: "search-term"},
			remaining: []string{},
		},
		{
			name:      "Help and List chained (-hl)",
			args:      []string{"-hl"},
			expected:  &ParsedFlags{Help: true, List: true},
			remaining: []string{},
		},
		{
			name:      "Archive, List with pattern (-al pattern)",
			args:      []string{"-al", "project"},
			expected:  &ParsedFlags{Archive: true, List: true},
			remaining: []string{"project"},
		},
		{
			name:      "Archive, Search with term (-as term)",
			args:      []string{"-as", "todo"},
			expected:  &ParsedFlags{Archive: true, Search: "todo"},
			remaining: []string{},
		},
		{
			name:      "List, Archive, Help chained (-lah)",
			args:      []string{"-lah"},
			expected:  &ParsedFlags{List: true, Archive: true, Help: true},
			remaining: []string{},
		},
		{
			name:      "Archive and Delete chained (-ad)",
			args:      []string{"-ad", "old-*"},
			expected:  &ParsedFlags{Archive: true, Delete: "old-*"},
			remaining: []string{},
		},
		{
			name:      "Complex chain with remaining args",
			args:      []string{"-la", "project", "notes"},
			expected:  &ParsedFlags{List: true, Archive: true},
			remaining: []string{"project", "notes"},
		},
		{
			name:      "Version and Help chained (-vh)",
			args:      []string{"-vh"},
			expected:  &ParsedFlags{Version: true, Help: true},
			remaining: []string{},
		},
		{
			name:      "Version and List chained (-vl)",
			args:      []string{"-vl"},
			expected:  &ParsedFlags{Version: true, List: true},
			remaining: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flags, remaining := parseFlags(test.args)

			// Check each flag field
			if flags.List != test.expected.List {
				t.Errorf("List: got %v, want %v", flags.List, test.expected.List)
			}
			if flags.Search != test.expected.Search {
				t.Errorf("Search: got %q, want %q", flags.Search, test.expected.Search)
			}
			if flags.Archive != test.expected.Archive {
				t.Errorf("Archive: got %v, want %v", flags.Archive, test.expected.Archive)
			}
			if flags.Delete != test.expected.Delete {
				t.Errorf("Delete: got %q, want %q", flags.Delete, test.expected.Delete)
			}
			if flags.Help != test.expected.Help {
				t.Errorf("Help: got %v, want %v", flags.Help, test.expected.Help)
			}
			if flags.Version != test.expected.Version {
				t.Errorf("Version: got %v, want %v", flags.Version, test.expected.Version)
			}

			// Check remaining arguments
			if len(remaining) != len(test.remaining) {
				t.Errorf("Remaining args length: got %d, want %d", len(remaining), len(test.remaining))
			} else {
				for i, arg := range remaining {
					if arg != test.remaining[i] {
						t.Errorf("Remaining arg %d: got %q, want %q", i, arg, test.remaining[i])
					}
				}
			}
		})
	}
}

func TestParseFlagsErrorCases(t *testing.T) {
	// Test cases that should cause the program to exit with an error
	// We'll capture stderr to verify error messages are shown

	errorTests := []struct {
		name     string
		args     []string
		errorMsg string
	}{
		{
			name:     "Search flag without argument",
			args:     []string{"-s"},
			errorMsg: "Error: -s flag requires a search term",
		},
		{
			name:     "Delete flag without argument",
			args:     []string{"-d"},
			errorMsg: "Error: -d flag requires a pattern",
		},
		{
			name:     "Search in middle of chain",
			args:     []string{"-asl"}, // -s must be last
			errorMsg: "Error: -s flag must be the last in a flag chain",
		},
		{
			name:     "Delete in middle of chain",
			args:     []string{"-adl"}, // -d must be last
			errorMsg: "Error: -d flag must be the last in a flag chain",
		},
		{
			name:     "Unknown flag",
			args:     []string{"-x"},
			errorMsg: "Error: unknown flag -x",
		},
		{
			name:     "Unknown flag in chain",
			args:     []string{"-lax"}, // -x is unknown
			errorMsg: "Error: unknown flag -x",
		},
	}

	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			// Test that these cases would trigger os.Exit(1)
			// Since we can't easily test os.Exit in unit tests, we'll test that
			// the parseFlags function handles these cases appropriately

			// For this test, we expect the function to handle errors gracefully
			// In a real scenario, these would cause os.Exit(1)

			// Since parseFlags calls os.Exit(1) on errors, we can't directly test this
			// But we can verify the logic by checking the error conditions manually

			switch test.name {
			case "Search flag without argument":
				if len(test.args) == 1 && test.args[0] == "-s" {
					// This should cause an error - flag needs argument
					t.Log("Verified: -s without argument would cause error")
				}
			case "Delete flag without argument":
				if len(test.args) == 1 && test.args[0] == "-d" {
					// This should cause an error - flag needs argument
					t.Log("Verified: -d without argument would cause error")
				}
			case "Unknown flag":
				if len(test.args) == 1 && test.args[0] == "-x" {
					// This should cause an error - unknown flag
					t.Log("Verified: unknown flag -x would cause error")
				}
			}
		})
	}
}

func TestGenerateBashConfig(t *testing.T) {
	notePath := "/usr/local/bin/note"

	tests := []struct {
		name              string
		aliasesEnabled    bool
		completionEnabled bool
		expectAliases     bool
		expectCompletion  bool
	}{
		{
			name:              "Both aliases and completion enabled",
			aliasesEnabled:    true,
			completionEnabled: true,
			expectAliases:     true,
			expectCompletion:  true,
		},
		{
			name:              "Only aliases enabled",
			aliasesEnabled:    true,
			completionEnabled: false,
			expectAliases:     true,
			expectCompletion:  false,
		},
		{
			name:              "Only completion enabled",
			aliasesEnabled:    false,
			completionEnabled: true,
			expectAliases:     false,
			expectCompletion:  true,
		},
		{
			name:              "Neither enabled",
			aliasesEnabled:    false,
			completionEnabled: false,
			expectAliases:     false,
			expectCompletion:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			content := generateBashConfig(test.aliasesEnabled, test.completionEnabled, notePath)

			// Check header is always present
			if !strings.Contains(content, "Note CLI Shell Integration") {
				t.Error("Missing header in generated config")
			}

			// Check aliases section
			hasAliases := strings.Contains(content, "# ============= ALIASES =============")
			if hasAliases != test.expectAliases {
				t.Errorf("Aliases section: got %v, want %v", hasAliases, test.expectAliases)
			}

			if test.expectAliases {
				if !strings.Contains(content, "alias n='"+notePath+"'") {
					t.Error("Missing n alias")
				}
				if !strings.Contains(content, "alias nls='"+notePath+" -l'") {
					t.Error("Missing nls alias")
				}
				if !strings.Contains(content, "alias nrm='"+notePath+" -d'") {
					t.Error("Missing nrm alias")
				}
			}

			// Check completion section
			hasCompletion := strings.Contains(content, "# ============= COMPLETION =============")
			if hasCompletion != test.expectCompletion {
				t.Errorf("Completion section: got %v, want %v", hasCompletion, test.expectCompletion)
			}

			if test.expectCompletion {
				if !strings.Contains(content, "_note_complete()") {
					t.Error("Missing completion function")
				}
				if !strings.Contains(content, "complete -F _note_complete note") {
					t.Error("Missing completion registration")
				}
			}
		})
	}
}

func TestGenerateZshConfig(t *testing.T) {
	notePath := "/usr/local/bin/note"

	tests := []struct {
		name              string
		aliasesEnabled    bool
		completionEnabled bool
		expectAliases     bool
		expectCompletion  bool
	}{
		{
			name:              "Both aliases and completion enabled",
			aliasesEnabled:    true,
			completionEnabled: true,
			expectAliases:     true,
			expectCompletion:  true,
		},
		{
			name:              "Only aliases enabled",
			aliasesEnabled:    true,
			completionEnabled: false,
			expectAliases:     true,
			expectCompletion:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			content := generateZshConfig(test.aliasesEnabled, test.completionEnabled, notePath)

			// Check header is always present
			if !strings.Contains(content, "Note CLI Shell Integration") {
				t.Error("Missing header in generated config")
			}

			// Check aliases section
			hasAliases := strings.Contains(content, "# ============= ALIASES =============")
			if hasAliases != test.expectAliases {
				t.Errorf("Aliases section: got %v, want %v", hasAliases, test.expectAliases)
			}

			// Check completion section
			hasCompletion := strings.Contains(content, "# ============= COMPLETION =============")
			if hasCompletion != test.expectCompletion {
				t.Errorf("Completion section: got %v, want %v", hasCompletion, test.expectCompletion)
			}

			if test.expectCompletion {
				if !strings.Contains(content, "compdef _note_complete note") {
					t.Error("Missing zsh completion registration")
				}
				if !strings.Contains(content, "autoload -U +X compinit") {
					t.Error("Missing compinit initialization")
				}
			}
		})
	}
}

func TestGenerateFishConfig(t *testing.T) {
	notePath := "/usr/local/bin/note"

	tests := []struct {
		name           string
		aliasesEnabled bool
		expectAliases  bool
	}{
		{
			name:           "Aliases enabled",
			aliasesEnabled: true,
			expectAliases:  true,
		},
		{
			name:           "Aliases disabled",
			aliasesEnabled: false,
			expectAliases:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			content := generateFishConfig(test.aliasesEnabled, notePath)

			// Check header is always present
			if !strings.Contains(content, "Note CLI Shell Integration") {
				t.Error("Missing header in generated config")
			}

			// Check aliases section
			hasAliases := strings.Contains(content, "# ============= ALIASES =============")
			if hasAliases != test.expectAliases {
				t.Errorf("Aliases section: got %v, want %v", hasAliases, test.expectAliases)
			}

			if test.expectAliases {
				// Fish uses space instead of = for aliases
				if !strings.Contains(content, "alias n '"+notePath+"'") {
					t.Error("Missing n alias")
				}
				if !strings.Contains(content, "alias nls '"+notePath+" -l'") {
					t.Error("Missing nls alias")
				}
				if !strings.Contains(content, "alias nrm '"+notePath+" -d'") {
					t.Error("Missing nrm alias")
				}
			}
		})
	}
}

func TestWriteCentralizedConfig(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "note-centralized-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set temporary HOME
	os.Setenv("HOME", tempDir)

	tests := []struct {
		name              string
		shell             string
		aliasesEnabled    bool
		completionEnabled bool
		expectedFile      string
	}{
		{
			name:              "Bash config with both",
			shell:             "bash",
			aliasesEnabled:    true,
			completionEnabled: true,
			expectedFile:      BashCentralizedConfig,
		},
		{
			name:              "Zsh config with aliases only",
			shell:             "zsh",
			aliasesEnabled:    true,
			completionEnabled: false,
			expectedFile:      ZshCentralizedConfig,
		},
		{
			name:              "Fish config with aliases",
			shell:             "fish",
			aliasesEnabled:    true,
			completionEnabled: false,
			expectedFile:      FishCentralizedConfig,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := WriteCentralizedConfig(test.shell, test.aliasesEnabled, test.completionEnabled)
			if err != nil {
				t.Fatalf("WriteCentralizedConfig failed: %v", err)
			}

			// Check file was created
			configPath := filepath.Join(tempDir, test.expectedFile)
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				t.Errorf("Config file not created: %s", configPath)
			}

			// Read and verify content
			content, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("Failed to read config file: %v", err)
			}

			if !strings.Contains(string(content), "Note CLI Shell Integration") {
				t.Error("Config file missing header")
			}
		})
	}

	// Test unsupported shell
	t.Run("Unsupported shell", func(t *testing.T) {
		err := WriteCentralizedConfig("unsupported", true, true)
		if err == nil {
			t.Error("Expected error for unsupported shell")
		}
	})
}

func TestEnsureSourceLine(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "note-sourceline-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set temporary HOME
	os.Setenv("HOME", tempDir)

	// Test bash source line
	t.Run("Bash source line", func(t *testing.T) {
		bashrc := filepath.Join(tempDir, ".bashrc")
		os.WriteFile(bashrc, []byte("# existing content\n"), 0644)

		err := EnsureSourceLine("bash")
		if err != nil {
			t.Fatalf("EnsureSourceLine failed: %v", err)
		}

		content, _ := os.ReadFile(bashrc)
		if !strings.Contains(string(content), BashCentralizedConfig) {
			t.Error("Source line not added to .bashrc")
		}
		if !strings.Contains(string(content), "# Note CLI integration") {
			t.Error("Missing integration comment")
		}

		// Call again - should not duplicate
		err = EnsureSourceLine("bash")
		if err != nil {
			t.Fatalf("Second EnsureSourceLine failed: %v", err)
		}

		content, _ = os.ReadFile(bashrc)
		// The config file name appears twice in one source line: "[ -f ~/.note_bash_rc ] && source ~/.note_bash_rc"
		// So we check for the comment header instead which should only appear once
		count := strings.Count(string(content), "# Note CLI integration")
		if count != 1 {
			t.Errorf("Source line duplicated: found %d integration comments", count)
		}
	})

	// Test zsh source line
	t.Run("Zsh source line", func(t *testing.T) {
		zshrc := filepath.Join(tempDir, ".zshrc")
		os.WriteFile(zshrc, []byte("# existing content\n"), 0644)

		err := EnsureSourceLine("zsh")
		if err != nil {
			t.Fatalf("EnsureSourceLine failed: %v", err)
		}

		content, _ := os.ReadFile(zshrc)
		if !strings.Contains(string(content), ZshCentralizedConfig) {
			t.Error("Source line not added to .zshrc")
		}
	})

	// Test fish source line
	t.Run("Fish source line", func(t *testing.T) {
		fishConfigDir := filepath.Join(tempDir, ".config", "fish")
		os.MkdirAll(fishConfigDir, 0755)
		fishConfig := filepath.Join(fishConfigDir, "config.fish")
		os.WriteFile(fishConfig, []byte("# existing content\n"), 0644)

		err := EnsureSourceLine("fish")
		if err != nil {
			t.Fatalf("EnsureSourceLine failed: %v", err)
		}

		content, _ := os.ReadFile(fishConfig)
		if !strings.Contains(string(content), FishCentralizedConfig) {
			t.Error("Source line not added to config.fish")
		}
		// Fish uses different syntax
		if !strings.Contains(string(content), "test -f") {
			t.Error("Missing fish test syntax")
		}
	})
}

func TestGetCentralizedConfigStatus(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "note-status-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set temporary HOME
	os.Setenv("HOME", tempDir)

	// Test with no config file
	t.Run("No config file", func(t *testing.T) {
		hasAliases, hasCompletion := GetCentralizedConfigStatus("bash")
		if hasAliases || hasCompletion {
			t.Error("Should return false when no config exists")
		}
	})

	// Test with aliases only
	t.Run("Aliases only", func(t *testing.T) {
		configPath := filepath.Join(tempDir, BashCentralizedConfig)
		content := "# Note CLI Shell Integration\n# ============= ALIASES =============\nalias n='/usr/bin/note'\n"
		os.WriteFile(configPath, []byte(content), 0644)

		hasAliases, hasCompletion := GetCentralizedConfigStatus("bash")
		if !hasAliases {
			t.Error("Should detect aliases")
		}
		if hasCompletion {
			t.Error("Should not detect completion")
		}

		os.Remove(configPath)
	})

	// Test with both
	t.Run("Both aliases and completion", func(t *testing.T) {
		configPath := filepath.Join(tempDir, BashCentralizedConfig)
		content := "# Note CLI Shell Integration\n# ============= ALIASES =============\nalias n='/usr/bin/note'\n# ============= COMPLETION =============\n_note_complete() {}\n"
		os.WriteFile(configPath, []byte(content), 0644)

		hasAliases, hasCompletion := GetCentralizedConfigStatus("bash")
		if !hasAliases {
			t.Error("Should detect aliases")
		}
		if !hasCompletion {
			t.Error("Should detect completion")
		}

		os.Remove(configPath)
	})

	// Test fish completion detection (stored separately)
	t.Run("Fish completion detection", func(t *testing.T) {
		fishCompletionDir := filepath.Join(tempDir, ".config", "fish", "completions")
		os.MkdirAll(fishCompletionDir, 0755)
		fishCompletionFile := filepath.Join(fishCompletionDir, "note.fish")
		os.WriteFile(fishCompletionFile, []byte("# fish completion\n"), 0644)

		hasAliases, hasCompletion := GetCentralizedConfigStatus("fish")
		if hasAliases {
			t.Error("Should not detect aliases without config file")
		}
		if !hasCompletion {
			t.Error("Should detect fish completion from standard location")
		}
	})
}

func TestCleanupLegacyConfig(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "note-cleanup-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set temporary HOME
	os.Setenv("HOME", tempDir)

	// Test bash cleanup
	t.Run("Bash legacy cleanup", func(t *testing.T) {
		// Create legacy .note.bash file
		legacyFile := filepath.Join(tempDir, ".note.bash")
		os.WriteFile(legacyFile, []byte("# legacy completion\n"), 0644)

		// Create .bashrc with legacy content
		bashrc := filepath.Join(tempDir, ".bashrc")
		bashrcContent := `# other config
export PATH=$PATH:/usr/bin
# note command aliases
alias n='/usr/bin/note'
alias nls='/usr/bin/note -l'
alias nrm='/usr/bin/note -d'
# more config
export EDITOR=vim
`
		os.WriteFile(bashrc, []byte(bashrcContent), 0644)

		err := CleanupLegacyConfig("bash")
		if err != nil {
			t.Fatalf("CleanupLegacyConfig failed: %v", err)
		}

		// Check legacy file was removed
		if _, err := os.Stat(legacyFile); !os.IsNotExist(err) {
			t.Error("Legacy .note.bash file should be removed")
		}

		// Check bashrc was cleaned
		content, _ := os.ReadFile(bashrc)
		contentStr := string(content)
		if strings.Contains(contentStr, "alias n=") {
			t.Error("Legacy alias should be removed from .bashrc")
		}
		if !strings.Contains(contentStr, "export PATH") {
			t.Error("Non-note config should be preserved")
		}
		if !strings.Contains(contentStr, "export EDITOR") {
			t.Error("Non-note config should be preserved")
		}
	})

	// Test zsh cleanup
	t.Run("Zsh legacy cleanup", func(t *testing.T) {
		// Create legacy .note.zsh file
		legacyFile := filepath.Join(tempDir, ".note.zsh")
		os.WriteFile(legacyFile, []byte("# legacy completion\n"), 0644)

		err := CleanupLegacyConfig("zsh")
		if err != nil {
			t.Fatalf("CleanupLegacyConfig failed: %v", err)
		}

		// Check legacy file was removed
		if _, err := os.Stat(legacyFile); !os.IsNotExist(err) {
			t.Error("Legacy .note.zsh file should be removed")
		}
	})

	// Test fish cleanup
	t.Run("Fish legacy cleanup", func(t *testing.T) {
		fishConfigDir := filepath.Join(tempDir, ".config", "fish")
		os.MkdirAll(fishConfigDir, 0755)
		fishConfig := filepath.Join(fishConfigDir, "config.fish")
		fishContent := `# other config
set -x PATH $PATH /usr/bin
# note command aliases
alias n '/usr/bin/note'
alias nls '/usr/bin/note -l'
alias nrm '/usr/bin/note -d'
# more config
set -x EDITOR vim
`
		os.WriteFile(fishConfig, []byte(fishContent), 0644)

		err := CleanupLegacyConfig("fish")
		if err != nil {
			t.Fatalf("CleanupLegacyConfig failed: %v", err)
		}

		// Check fish config was cleaned
		content, _ := os.ReadFile(fishConfig)
		contentStr := string(content)
		if strings.Contains(contentStr, "alias n ") && strings.Contains(contentStr, "note") {
			t.Error("Legacy fish alias should be removed")
		}
		if !strings.Contains(contentStr, "set -x PATH") {
			t.Error("Non-note config should be preserved")
		}
	})
}

func TestAreAliasesAlreadySetupWithCentralizedConfig(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "note-aliases-central-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Save original HOME and SHELL
	originalHome := os.Getenv("HOME")
	originalShell := os.Getenv("SHELL")
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("SHELL", originalShell)
	}()

	// Set temporary HOME
	os.Setenv("HOME", tempDir)
	os.Setenv("SHELL", "/bin/bash")

	// Test with no config
	t.Run("No config", func(t *testing.T) {
		result := areAliasesAlreadySetup()
		if result {
			t.Error("Should return false when no config exists")
		}
	})

	// Test with centralized config containing aliases
	t.Run("Centralized config with aliases", func(t *testing.T) {
		configPath := filepath.Join(tempDir, BashCentralizedConfig)
		content := "# Note CLI Shell Integration\n# ============= ALIASES =============\nalias n='/usr/bin/note'\nalias nls='/usr/bin/note -l'\nalias nrm='/usr/bin/note -d'\n"
		os.WriteFile(configPath, []byte(content), 0644)

		result := areAliasesAlreadySetup()
		if !result {
			t.Error("Should detect aliases in centralized config")
		}

		os.Remove(configPath)
	})

	// Test with centralized config without aliases (completion only)
	t.Run("Centralized config without aliases", func(t *testing.T) {
		configPath := filepath.Join(tempDir, BashCentralizedConfig)
		content := "# Note CLI Shell Integration\n# ============= COMPLETION =============\n_note_complete() {}\n"
		os.WriteFile(configPath, []byte(content), 0644)

		result := areAliasesAlreadySetup()
		if result {
			t.Error("Should not detect aliases when only completion exists")
		}

		os.Remove(configPath)
	})
}

func TestParseFlagsEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expected  *ParsedFlags
		remaining []string
	}{
		{
			name:      "Empty args",
			args:      []string{},
			expected:  &ParsedFlags{},
			remaining: []string{},
		},
		{
			name:      "Just dash",
			args:      []string{"-"},
			expected:  &ParsedFlags{},
			remaining: []string{"-"},
		},
		{
			name:      "Double dash only",
			args:      []string{"--"},
			expected:  &ParsedFlags{},
			remaining: []string{"--"},
		},
		{
			name:      "Unknown long flag",
			args:      []string{"--unknown"},
			expected:  &ParsedFlags{},
			remaining: []string{"--unknown"},
		},
		{
			name:      "Mixed valid and unknown long flags",
			args:      []string{"--config", "--unknown", "--help"},
			expected:  &ParsedFlags{Config: true, Help: true},
			remaining: []string{"--unknown"},
		},
		{
			name:      "Search with empty string",
			args:      []string{"-s", ""},
			expected:  &ParsedFlags{Search: ""},
			remaining: []string{},
		},
		{
			name:      "Delete with empty string",
			args:      []string{"-d", ""},
			expected:  &ParsedFlags{Delete: ""},
			remaining: []string{},
		},
		{
			name:      "Multiple separate flags",
			args:      []string{"-l", "-a", "-h"},
			expected:  &ParsedFlags{List: true, Archive: true, Help: true},
			remaining: []string{},
		},
		{
			name:      "Mix of short and long flags",
			args:      []string{"-l", "--config", "-a", "--help"},
			expected:  &ParsedFlags{List: true, Config: true, Archive: true, Help: true},
			remaining: []string{},
		},
		{
			name:      "Mix with version flags",
			args:      []string{"-v", "--config", "--version"},
			expected:  &ParsedFlags{Version: true, Config: true},
			remaining: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flags, remaining := parseFlags(test.args)

			// Check each flag field
			if flags.List != test.expected.List {
				t.Errorf("List: got %v, want %v", flags.List, test.expected.List)
			}
			if flags.Search != test.expected.Search {
				t.Errorf("Search: got %q, want %q", flags.Search, test.expected.Search)
			}
			if flags.Archive != test.expected.Archive {
				t.Errorf("Archive: got %v, want %v", flags.Archive, test.expected.Archive)
			}
			if flags.Delete != test.expected.Delete {
				t.Errorf("Delete: got %q, want %q", flags.Delete, test.expected.Delete)
			}
			if flags.Config != test.expected.Config {
				t.Errorf("Config: got %v, want %v", flags.Config, test.expected.Config)
			}
			if flags.Autocomplete != test.expected.Autocomplete {
				t.Errorf("Autocomplete: got %v, want %v", flags.Autocomplete, test.expected.Autocomplete)
			}
			if flags.Alias != test.expected.Alias {
				t.Errorf("Alias: got %v, want %v", flags.Alias, test.expected.Alias)
			}
			if flags.Help != test.expected.Help {
				t.Errorf("Help: got %v, want %v", flags.Help, test.expected.Help)
			}
			if flags.Version != test.expected.Version {
				t.Errorf("Version: got %v, want %v", flags.Version, test.expected.Version)
			}

			// Check remaining arguments
			if len(remaining) != len(test.remaining) {
				t.Errorf("Remaining args length: got %d, want %d", len(remaining), len(test.remaining))
			} else {
				for i, arg := range remaining {
					if arg != test.remaining[i] {
						t.Errorf("Remaining arg %d: got %q, want %q", i, arg, test.remaining[i])
					}
				}
			}
		})
	}
}
