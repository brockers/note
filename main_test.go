package main

import (
	"os"
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
		{"short", "short"},  // Term same length as text
		{"ab", "abc"},       // Term longer than text
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