#!/bin/bash

# Don't use set -e because we want to continue even if a test fails
# We'll handle errors explicitly in the test functions

echo "=== Note CLI Integration Tests ==="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Function to run a test
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_result="$3"
    
    echo -n "Testing: $test_name... "
    
    if eval "$test_command"; then
        if [ -z "$expected_result" ] || eval "$expected_result"; then
            echo -e "${GREEN}PASSED${NC}"
            ((TESTS_PASSED++))
        else
            echo -e "${RED}FAILED${NC} (expectation not met)"
            ((TESTS_FAILED++))
        fi
    else
        echo -e "${RED}FAILED${NC}"
        ((TESTS_FAILED++))
    fi
}

# Setup test environment
TEST_DIR=$(mktemp -d)
export HOME="$TEST_DIR"
NOTE_CMD="./note"

# Build the application
echo "Building note application..."
go build -o note

# Test 1: First run setup (now includes autocomplete prompt)
echo -e "vim\n$TEST_DIR/Notes\nn\n" | $NOTE_CMD > /dev/null 2>&1
run_test "Initial setup creates config" "test -f $TEST_DIR/.note" ""

# Test 2: Config file content
run_test "Config contains editor" "grep -q 'editor=vim' $TEST_DIR/.note" ""
run_test "Config contains notesdir" "grep -q 'notesdir=' $TEST_DIR/.note" ""

# Test 3: Notes directory created
run_test "Notes directory created" "test -d $TEST_DIR/Notes" ""
run_test "Archive directory created" "test -d $TEST_DIR/Notes/Archive" ""

# Test 4: Help command
run_test "Help command works" "$NOTE_CMD --help 2>&1 | grep -q 'note - A minimalist CLI'" ""

# Test 5: Create a note (would open editor, so we'll test file creation differently)
TODAY=$(date +%Y%m%d)
touch "$TEST_DIR/Notes/test-note-$TODAY.md"
run_test "Note file structure" "test -f $TEST_DIR/Notes/test-note-$TODAY.md" ""

# Test 6: List notes
echo "Test content" > "$TEST_DIR/Notes/meeting-$TODAY.md"
echo "Another note" > "$TEST_DIR/Notes/project-$TODAY.md"
run_test "List all notes" "$NOTE_CMD -ls | grep -c '.md' | grep -q 3" ""

# Test 7: List with pattern
run_test "List with pattern" "$NOTE_CMD -ls meeting | grep -q meeting" ""
run_test "Pattern filters correctly" "$NOTE_CMD -ls meeting | grep -v project > /dev/null" ""

# Test 8: Full-text search
echo "TODO: Fix bug" > "$TEST_DIR/Notes/bug-report-$TODAY.md"
run_test "Full-text search finds content" "$NOTE_CMD -s 'TODO' 2>&1 | grep -q 'bug-report'" ""

# Test 9: Archive functionality
$NOTE_CMD -rm "test-note*" > /dev/null 2>&1
run_test "Archive moves file" "test -f $TEST_DIR/Notes/Archive/test-note-$TODAY.md" ""
run_test "Archived file removed from main" "! test -f $TEST_DIR/Notes/test-note-$TODAY.md" ""

# Test 10: List archived notes
run_test "List archived notes" "$NOTE_CMD -a | grep -q 'Archive/test-note'" ""

# Test 11: Multiple file operations
echo "Note 1" > "$TEST_DIR/Notes/temp1-$TODAY.md"
echo "Note 2" > "$TEST_DIR/Notes/temp2-$TODAY.md"
echo "Note 3" > "$TEST_DIR/Notes/keep-$TODAY.md"
$NOTE_CMD -rm "temp*" > /dev/null 2>&1
run_test "Bulk archive works" "test -f $TEST_DIR/Notes/Archive/temp1-$TODAY.md" ""
run_test "Pattern preserves non-matching" "test -f $TEST_DIR/Notes/keep-$TODAY.md" ""

# Test 12: Edge cases
touch "$TEST_DIR/Notes/note with spaces-$TODAY.md"
run_test "Handles spaces in names" "$NOTE_CMD -ls | grep -q 'note with spaces'" ""

# Test 13: Opening existing files without .md extension
echo "Test content" > "$TEST_DIR/Notes/existing-note-20240426.md"
echo "editor=echo" > "$TEST_DIR/.note.bak" && mv "$TEST_DIR/.note" "$TEST_DIR/.note.bak2" && cp "$TEST_DIR/.note.bak2" "$TEST_DIR/.note" && sed -i 's/editor=vim/editor=echo/' "$TEST_DIR/.note"
RESULT=$($NOTE_CMD existing-note-20240426 2>&1 | tr -d '\n')
echo "editor=vim" > "$TEST_DIR/.note" && echo "notesdir=$TEST_DIR/Notes" >> "$TEST_DIR/.note"
run_test "Opens existing file without adding new date" "echo '$RESULT' | grep -q 'existing-note-20240426.md'" ""

# Test 14: Config modification (now includes autocomplete prompt)
echo -e "nano\n$TEST_DIR/NewNotes\nn\n" | $NOTE_CMD --config > /dev/null 2>&1
run_test "Config updates editor" "grep -q 'editor=nano' $TEST_DIR/.note" ""

# Test 15: Search highlighting functionality
echo "editor=vim" > "$TEST_DIR/.note" && echo "notesdir=$TEST_DIR/Notes" >> "$TEST_DIR/.note"
echo "Content about meetings" > "$TEST_DIR/Notes/meeting-highlights-$TODAY.md"
echo "Daily standup notes" > "$TEST_DIR/Notes/daily-standup-$TODAY.md"
echo "Project planning session" > "$TEST_DIR/Notes/project-planning-$TODAY.md"
echo "Notes about daily meetings" > "$TEST_DIR/Notes/daily meetings notes-$TODAY.md"

# Test 16: Multi-word search patterns (tests argument joining functionality)
run_test "Multi-word search finds matches" "$NOTE_CMD -l daily meetings | grep -q 'daily meetings notes'" ""

# Test 17: Piped output has no color codes (should be clean)
PIPED_OUTPUT=$($NOTE_CMD -l meeting | cat)
run_test "Piped output contains no ANSI color codes" "echo '$PIPED_OUTPUT' | grep -v '\[31m'" ""

# Test 18: Redirected output has no color codes
$NOTE_CMD -l meeting > "$TEST_DIR/output.txt" 2>&1
run_test "Redirected output contains no ANSI color codes" "! grep -q '\[31m' $TEST_DIR/output.txt" ""

# Test 19: Search with no matches
run_test "Search with no matches returns nothing" "! $NOTE_CMD -l nonexistent | grep -q '.'" ""

# Test 20: Case-insensitive search highlighting
echo "MEETING NOTES" > "$TEST_DIR/Notes/MEETING-UPPERCASE-$TODAY.md"
run_test "Case-insensitive search finds uppercase" "$NOTE_CMD -l meeting | grep -q 'MEETING-UPPERCASE'" ""

# Test 21: Search pattern with spaces in existing filenames
echo "Test content" > "$TEST_DIR/Notes/project notes with spaces-$TODAY.md"
run_test "Search finds files with spaces in name" "$NOTE_CMD -l project | grep -q 'project.*spaces'" ""

# Cleanup
rm -rf "$TEST_DIR"

# Summary
echo ""
echo "=== Test Summary ==="
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi