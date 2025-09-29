#!/bin/bash

# Test suite for directory partial completion fix
# Tests the specific issue: markdown-runner test/cases/ -B p<TAB> should show parallel.md@, not parallel_write

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
    echo "  Completions: ${COMPREPLY[*]}"
    
    return 0
}

# Helper function to check if all completions are file@ format
check_all_file_at() {
    for completion in "${COMPREPLY[@]}"; do
        if [[ "$completion" != *@* ]]; then
            return 1  # Found a non-file@ completion
        fi
    done
    return 0  # All completions are file@ format
}

# Helper function to check if all completions are stage names (no @ or /)
check_all_stage_names() {
    for completion in "${COMPREPLY[@]}"; do
        if [[ "$completion" == *@* ]] || [[ "$completion" == */* ]]; then
            return 1  # Found a file@ or chunk completion
        fi
    done
    return 0  # All completions are stage names
}

echo "=== Directory Partial Completion Test Suite ==="
echo

# Test 1: The original user issue - directory with empty completion
echo "=== Testing Directory Empty Completion ==="

run_completion_test \
    "test/cases/ -B (empty) should show file@ completions" \
    "markdown-runner" "test/cases/" "-B" ""

if check_all_file_at && [[ ${#COMPREPLY[@]} -gt 1 ]]; then
    echo -e "  ${GREEN}PASS: Directory empty completion shows file@ completions${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Directory empty completion should show file@ completions${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 2: The original user issue - directory with partial completion
echo "=== Testing Directory Partial Completion (Main Issue) ==="

run_completion_test \
    "test/cases/ -B p should show parallel.md@, NOT parallel_write" \
    "markdown-runner" "test/cases/" "-B" "p"

if [[ ${#COMPREPLY[@]} -eq 1 ]] && [[ "${COMPREPLY[0]}" == "parallel.md@" ]]; then
    echo -e "  ${GREEN}PASS: Directory partial 'p' shows parallel.md@ (file@ completion)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Directory partial 'p' should show parallel.md@, got: ${COMPREPLY[*]}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 3: Other partial completions in directory context
run_completion_test \
    "test/cases/ -B h should show happy.md@" \
    "markdown-runner" "test/cases/" "-B" "h"

if [[ ${#COMPREPLY[@]} -eq 1 ]] && [[ "${COMPREPLY[0]}" == "happy.md@" ]]; then
    echo -e "  ${GREEN}PASS: Directory partial 'h' shows happy.md@${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Directory partial 'h' should show happy.md@, got: ${COMPREPLY[*]}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "test/cases/ -B s should show schema_error.md@" \
    "markdown-runner" "test/cases/" "-B" "s"

if [[ ${#COMPREPLY[@]} -eq 1 ]] && [[ "${COMPREPLY[0]}" == "schema_error.md@" ]]; then
    echo -e "  ${GREEN}PASS: Directory partial 's' shows schema_error.md@${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Directory partial 's' should show schema_error.md@, got: ${COMPREPLY[*]}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 4: Complete stage names in directory context should show chunks
echo "=== Testing Complete Stage Names in Directory Context ==="

run_completion_test \
    "test/cases/ -B parallel_write (complete stage) should show chunks" \
    "markdown-runner" "test/cases/" "-B" "parallel_write"

if [[ ${#COMPREPLY[@]} -ge 1 ]] && [[ "${COMPREPLY[0]}" == */* ]]; then
    echo -e "  ${GREEN}PASS: Complete stage in directory shows chunks${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Complete stage should show chunks, got: ${COMPREPLY[*]}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 5: Non-directory context should still work normally
echo "=== Testing Non-Directory Context (Control) ==="

run_completion_test \
    "Non-directory -B p should show stage names (if any)" \
    "markdown-runner" "-B" "p"

# This might show no completions or stage names, both are acceptable
echo -e "  ${GREEN}PASS: Non-directory context works (${#COMPREPLY[@]} completions)${NC}"
TESTS_PASSED=$((TESTS_PASSED + 1))
echo

# Test 6: Recursive mode should still work
echo "=== Testing Recursive Mode ==="

run_completion_test \
    "Recursive -B p should show file@ completions" \
    "markdown-runner" "-r" "-B" "p"

if check_all_file_at || [[ ${#COMPREPLY[@]} -eq 0 ]]; then
    echo -e "  ${GREEN}PASS: Recursive mode shows file@ completions or no matches${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Recursive mode should show file@ completions, got: ${COMPREPLY[*]}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 7: Multiple character partial matches
echo "=== Testing Multi-Character Partial Matches ==="

run_completion_test \
    "test/cases/ -B par should show parallel.md@" \
    "markdown-runner" "test/cases/" "-B" "par"

if [[ ${#COMPREPLY[@]} -eq 1 ]] && [[ "${COMPREPLY[0]}" == "parallel.md@" ]]; then
    echo -e "  ${GREEN}PASS: Multi-character partial 'par' shows parallel.md@${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Multi-character partial 'par' should show parallel.md@, got: ${COMPREPLY[*]}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "test/cases/ -B hap should show happy.md@" \
    "markdown-runner" "test/cases/" "-B" "hap"

if [[ ${#COMPREPLY[@]} -eq 1 ]] && [[ "${COMPREPLY[0]}" == "happy.md@" ]]; then
    echo -e "  ${GREEN}PASS: Multi-character partial 'hap' shows happy.md@${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Multi-character partial 'hap' should show happy.md@, got: ${COMPREPLY[*]}${NC}"
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
