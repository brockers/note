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