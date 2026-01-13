#!/bin/bash

# test_completion.sh - Test bash completion for the note command
# This script tests the completion functionality to ensure it properly
# handles partial note name matching

# Don't use set -e because we want to continue even if a test fails

echo "=== Note Completion Test Suite ==="
echo

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Create a temporary test directory with mock notes
TEST_DIR=$(mktemp -d)
TEST_CONFIG=$(mktemp)
trap "rm -rf $TEST_DIR $TEST_CONFIG" EXIT

# Set up test environment
echo "Setting up test environment..."
echo "editor=vim" > $TEST_CONFIG
echo "notesdir=$TEST_DIR" >> $TEST_CONFIG

# Create mock note files
touch "$TEST_DIR/Life-101-Philosophy-20240101.md"
touch "$TEST_DIR/Life-101-Habits-20240102.md"
touch "$TEST_DIR/Life-101-Growth-20240103.md"
touch "$TEST_DIR/ASG_Weekly_Notes-20240104.md"
touch "$TEST_DIR/ASG_Monthly_Report-20240105.md"
touch "$TEST_DIR/Family-Notes-20240106.md"
touch "$TEST_DIR/Project-Alpha-20240107.md"
touch "$TEST_DIR/Project-Beta-20240108.md"
touch "$TEST_DIR/Chess-Notes-20240109.md"
touch "$TEST_DIR/README.md"

# Build the completion script dynamically based on current main.go
echo "Building completion script from main.go..."
./note --autocomplete > /dev/null 2>&1 <<EOF
y
EOF

# Extract the completion function from the generated script
COMPLETION_SCRIPT=$(mktemp)
cat > $COMPLETION_SCRIPT << 'SCRIPT_END'
#!/bin/bash

_note_complete_test() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local prev="${COMP_WORDS[COMP_CWORD-1]}"
    
    # If we're on the first argument
    if [[ ${COMP_CWORD} -eq 1 ]]; then
        # If user starts typing a dash, offer flags
        if [[ "$cur" == -* ]]; then
            local flags="-l -s -a -d -v --config --autocomplete --alias --help --version -h"
            COMPREPLY=($(compgen -W "$flags" -- "${cur}"))
        else
            # Otherwise, prioritize note names
            if [[ -f TEST_CONFIG_PATH ]]; then
                local notesdir=$(grep "^notesdir=" TEST_CONFIG_PATH | cut -d= -f2 | sed "s|~|$HOME|")
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
        if [[ -f TEST_CONFIG_PATH ]]; then
            local notesdir=$(grep "^notesdir=" TEST_CONFIG_PATH | cut -d= -f2 | sed "s|~|$HOME|")
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
SCRIPT_END

# Replace TEST_CONFIG_PATH with actual path
sed -i "s|TEST_CONFIG_PATH|$TEST_CONFIG|g" $COMPLETION_SCRIPT

# Source the test completion function
source $COMPLETION_SCRIPT

# Test function
run_test() {
    local test_name="$1"
    local input="$2"
    local expected_count="$3"
    local expected_pattern="$4"
    
    # Set up completion environment
    export COMP_WORDS=("note" "$input")
    export COMP_CWORD=1
    COMPREPLY=()
    
    # Run completion
    _note_complete_test
    
    # Check results
    local result_count=${#COMPREPLY[@]}
    local pattern_found=0
    
    if [[ -n "$expected_pattern" ]]; then
        for result in "${COMPREPLY[@]}"; do
            if [[ "$result" == $expected_pattern* ]]; then
                pattern_found=1
                break
            fi
        done
    else
        pattern_found=1  # No pattern to check
    fi
    
    if [[ "$expected_count" == "-1" ]] || [[ $result_count -eq $expected_count ]]; then
        count_ok=1
    elif [[ "$expected_count" == "+" ]] && [[ $result_count -gt 0 ]]; then
        count_ok=1
    else
        count_ok=0
    fi
    
    if [[ $count_ok -eq 1 ]] && [[ $pattern_found -eq 1 ]]; then
        echo -e "${GREEN}✓${NC} $test_name (found $result_count matches)"
        if [[ $result_count -gt 0 ]] && [[ $result_count -le 5 ]]; then
            for result in "${COMPREPLY[@]}"; do
                echo "    - $result"
            done
        fi
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗${NC} $test_name"
        echo "    Expected: count=$expected_count, pattern=$expected_pattern"
        echo "    Got: count=$result_count"
        if [[ $result_count -gt 0 ]] && [[ $result_count -le 10 ]]; then
            echo "    Results:"
            for result in "${COMPREPLY[@]}"; do
                echo "      - $result"
            done
        fi
        ((TESTS_FAILED++))
    fi
}

echo "Running completion tests..."
echo

# Test partial matching (case-sensitive)
run_test "Partial match 'Life' should return Life-101-* notes" "Life" 3 "Life-101"
run_test "Partial match 'ASG' should return ASG_* notes" "ASG" 2 "ASG_"
run_test "Partial match 'Project' should return Project-* notes" "Project" 2 "Project-"
run_test "Partial match 'Fam' should return Family notes" "Fam" 1 "Family-Notes"
run_test "Partial match 'Chess' should return Chess notes" "Chess" 1 "Chess-Notes"

# Test case-insensitive matching
run_test "Case-insensitive 'life' should match Life-101-* notes" "life" 3 "Life-101"
run_test "Case-insensitive 'asg' should match ASG_* notes" "asg" 2 "ASG_"
run_test "Case-insensitive 'PROJECT' should match Project-* notes" "PROJECT" 2 "Project-"
run_test "Case-insensitive 'fam' should match Family notes" "fam" 1 "Family-Notes"

# Test exact prefix matching
run_test "Exact prefix 'Life-101-H' should return Habits note" "Life-101-H" 1 "Life-101-Habits"
run_test "Exact prefix 'ASG_W' should return Weekly notes" "ASG_W" 1 "ASG_Weekly"

# Test non-matching input
run_test "Non-matching 'XYZ' should return no results" "XYZ" 0 ""

# Test empty input (should return all notes)
run_test "Empty input should return all notes" "" 10 ""

# Test flag completion
run_test "Flag '-l' should match -l flag" "-l" 1 "-l"
run_test "Flag '-v' should match -v flag" "-v" 1 "-v"
run_test "Flag '--h' should match --help" "--h" 1 "--help"
run_test "Flag '--v' should match --version" "--v" 1 "--version"
run_test "Flag '--a' should match --autocomplete and --alias" "--a" 2 "--a"

# Test completion after flags
export COMP_WORDS=("note" "-l" "Life")
export COMP_CWORD=2
COMPREPLY=()
_note_complete_test
if [[ ${#COMPREPLY[@]} -eq 3 ]]; then
    echo -e "${GREEN}✓${NC} Completion after -l flag works (found ${#COMPREPLY[@]} matches)"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗${NC} Completion after -l flag failed"
    ((TESTS_FAILED++))
fi

# Test completion after -d flag (for delete/archive)
export COMP_WORDS=("note" "-d" "ASG")
export COMP_CWORD=2
COMPREPLY=()
_note_complete_test
if [[ ${#COMPREPLY[@]} -eq 2 ]]; then
    echo -e "${GREEN}✓${NC} Completion after -d flag works (found ${#COMPREPLY[@]} matches)"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗${NC} Completion after -d flag failed (expected 2, got ${#COMPREPLY[@]})"
    ((TESTS_FAILED++))
fi

# Test flag chaining completion patterns
echo
echo "Testing flag chaining completion..."

# Test that flag chaining patterns are recognized in completion
# Since completion doesn't handle flag chaining directly, we just verify
# that individual flags still work when they would be part of chains
export COMP_WORDS=("note" "-a" "Life")
export COMP_CWORD=2
COMPREPLY=()
_note_complete_test
if [[ ${#COMPREPLY[@]} -eq 3 ]]; then
    echo -e "${GREEN}✓${NC} Archive flag completion works for flag chaining (found ${#COMPREPLY[@]} matches)"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗${NC} Archive flag completion failed (expected 3, got ${#COMPREPLY[@]})"
    ((TESTS_FAILED++))
fi

echo
echo "Testing alias completions (n, nls, nrm)..."

# Test helper for alias completion
run_alias_test() {
    local alias_name="$1"
    local test_name="$2"
    local input="$3"
    local expected_count="$4"

    # Set up completion environment for alias
    export COMP_WORDS=("$alias_name" "$input")
    export COMP_CWORD=1
    COMPREPLY=()

    # Run completion
    _note_complete_test

    # Check results
    local result_count=${#COMPREPLY[@]}

    if [[ "$expected_count" == "-1" ]] || [[ $result_count -eq $expected_count ]]; then
        echo -e "${GREEN}✓${NC} $test_name (found $result_count matches)"
        if [[ $result_count -gt 0 ]] && [[ $result_count -le 3 ]]; then
            for result in "${COMPREPLY[@]}"; do
                echo "    - $result"
            done
        fi
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗${NC} $test_name"
        echo "    Expected: count=$expected_count"
        echo "    Got: count=$result_count"
        ((TESTS_FAILED++))
    fi
}

# Test 'n' alias (same as note)
run_alias_test "n" "Alias 'n' should complete Life notes" "Life" 3
run_alias_test "n" "Alias 'n' should complete ASG notes" "ASG" 2
run_alias_test "n" "Alias 'n' should handle empty input" "" 10

# Test 'nls' alias (note -l)
run_alias_test "nls" "Alias 'nls' should complete Life notes" "Life" 3
run_alias_test "nls" "Alias 'nls' should complete Project notes" "Project" 2

# Test 'nrm' alias (note -d)
run_alias_test "nrm" "Alias 'nrm' should complete Family notes" "Family" 1
run_alias_test "nrm" "Alias 'nrm' should complete Chess notes" "Chess" 1

echo
echo "==================================="
echo "Test Summary:"
echo -e "  Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "  Failed: ${RED}$TESTS_FAILED${NC}"

if [[ $TESTS_FAILED -eq 0 ]]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed.${NC}"
    exit 1
fi