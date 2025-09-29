#!/bin/bash

# Test suite for enhanced completion features
# Tests the new categorized help and description functionality

source "$(dirname "$0")/../bash_completion.sh"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to run a completion test
run_completion_test() {
    local test_name="$1"
    shift
    local comp_words=("$@")
    
    echo -e "${YELLOW}Testing: $test_name${NC}"
    echo "  COMP_WORDS: ${comp_words[*]}"
    
    TESTS_RUN=$((TESTS_RUN + 1))
    
    # Set up completion variables
    COMP_WORDS=("${comp_words[@]}")
    COMP_CWORD=$((${#COMP_WORDS[@]} - 1))
    
    # Clear previous results
    COMPREPLY=()
    
    # Run completion
    _markdown_runner_completion
    
    echo "  Completions found: ${#COMPREPLY[@]}"
    echo "  First few: ${COMPREPLY[*]:0:3}..."
    
    return 0
}

# Helper function to check if completions contain descriptions
check_has_descriptions() {
    for completion in "${COMPREPLY[@]}"; do
        if [[ "$completion" == *":"* ]]; then
            return 0  # Found description format
        fi
    done
    return 1  # No descriptions found
}

# Helper function to check if completions are categorized help
check_is_categorized_help() {
    for completion in "${COMPREPLY[@]}"; do
        if [[ "$completion" == "Modes:" ]] || [[ "$completion" == "Execution Control:" ]]; then
            return 0  # Found categorized help
        fi
    done
    return 1  # Not categorized help
}

# Helper function to check if a specific flag with description exists
check_flag_with_description() {
    local flag="$1"
    local expected_desc="$2"
    for completion in "${COMPREPLY[@]}"; do
        # Handle both single format and combined format
        if [[ "$completion" == "$flag:($expected_desc)" ]] || 
           [[ "$completion" == "$flag, "* ]] && [[ "$completion" == *":($expected_desc)" ]]; then
            return 0  # Found exact match
        fi
    done
    return 1  # Not found
}

echo "=== Enhanced Completion Test Suite ==="
echo

# Test 1: Enhanced Help with Descriptions
echo "=== Testing Enhanced Help with Descriptions ==="

run_completion_test \
    "Bare dash shows enhanced help with descriptions" \
    "markdown-runner" "-"

if check_has_descriptions && [[ ${#COMPREPLY[@]} -gt 10 ]] && [[ ${#COMPREPLY[@]} -lt 20 ]]; then
    echo -e "  ${GREEN}PASS: Enhanced help shows descriptions (${#COMPREPLY[@]} entries, no duplicates)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Enhanced help should show descriptions (expected 10-20 entries, got ${#COMPREPLY[@]})${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Enhanced help includes specific flags with descriptions" \
    "markdown-runner" "-"

if check_flag_with_description "-d" "dry-run" && check_flag_with_description "-i" "interactive"; then
    echo -e "  ${GREEN}PASS: Enhanced help includes expected flag descriptions${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Enhanced help should include expected flag descriptions${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 2: Full Categorized Help
echo "=== Testing Full Categorized Help ==="

run_completion_test \
    "Help command shows full categorized help" \
    "markdown-runner" "help"

if check_is_categorized_help && [[ ${#COMPREPLY[@]} -gt 15 ]]; then
    echo -e "  ${GREEN}PASS: Help command shows categorized help${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Help command should show categorized help${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Partial help command works" \
    "markdown-runner" "hel"

if check_is_categorized_help; then
    echo -e "  ${GREEN}PASS: Partial help command works${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Partial help command should work${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 3: Standard Completion Still Works
echo "=== Testing Standard Completion Compatibility ==="

run_completion_test \
    "Partial flag completion works normally" \
    "markdown-runner" "-d"

if ! check_has_descriptions && [[ ${#COMPREPLY[@]} -eq 1 ]] && [[ "${COMPREPLY[0]}" == "-d" ]]; then
    echo -e "  ${GREEN}PASS: Partial flag completion works normally${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Partial flag completion should work normally${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Long flag completion works normally" \
    "markdown-runner" "--dry"

if ! check_has_descriptions && [[ ${#COMPREPLY[@]} -eq 1 ]] && [[ "${COMPREPLY[0]}" == "--dry-run" ]]; then
    echo -e "  ${GREEN}PASS: Long flag completion works normally${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Long flag completion should work normally${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 4: Enhanced Help with Incompatible Filtering
echo "=== Testing Enhanced Help with Incompatible Filtering ==="

run_completion_test \
    "Enhanced help respects incompatible flags (CI mode)" \
    "markdown-runner" "--view" "ci" "-"

if check_has_descriptions && [[ ${#COMPREPLY[@]} -lt 15 ]]; then
    # Should have fewer flags due to incompatible filtering
    echo -e "  ${GREEN}PASS: Enhanced help respects incompatible flags (${#COMPREPLY[@]} entries)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Enhanced help should respect incompatible flags (expected <15, got ${#COMPREPLY[@]})${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Enhanced help respects incompatible flags (verbose)" \
    "markdown-runner" "-v" "-"

if check_has_descriptions && ! check_flag_with_description "-q" "quiet"; then
    # Should not include quiet flag when verbose is used
    echo -e "  ${GREEN}PASS: Enhanced help excludes incompatible quiet flag${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Enhanced help should exclude incompatible quiet flag${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 5: Context Awareness
echo "=== Testing Context Awareness ==="

run_completion_test \
    "File completion still works when no flags" \
    "markdown-runner" "README"

if ! check_has_descriptions && ! check_is_categorized_help; then
    echo -e "  ${GREEN}PASS: File completion works when no flags${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: File completion should work when no flags${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Flag value completion still works" \
    "markdown-runner" "--view" "c"

if ! check_has_descriptions && ! check_is_categorized_help; then
    echo -e "  ${GREEN}PASS: Flag value completion works${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Flag value completion should work${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 6: Edge Cases
echo "=== Testing Edge Cases ==="

run_completion_test \
    "Enhanced help works with existing file argument" \
    "markdown-runner" "README.md" "-"

if check_has_descriptions && [[ ${#COMPREPLY[@]} -eq 15 ]]; then
    echo -e "  ${GREEN}PASS: Enhanced help works with existing file argument (${#COMPREPLY[@]} entries)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Enhanced help should work with existing file argument (expected 15, got ${#COMPREPLY[@]})${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Help command works with existing flags" \
    "markdown-runner" "-v" "help"

if check_is_categorized_help; then
    echo -e "  ${GREEN}PASS: Help command works with existing flags${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Help command should work with existing flags${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

echo "=== Test Results ==="
echo "Tests run: $TESTS_RUN"
echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
if [[ $TESTS_FAILED -gt 0 ]]; then
    echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
    exit 1
else
    echo -e "Tests failed: ${GREEN}0${NC}"
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
fi
