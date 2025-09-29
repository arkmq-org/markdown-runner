#!/bin/bash

# Test suite for file@stage completion functionality
# Tests that file@stage completion shows all available stages from the correct file

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
    echo "  First 5 completions: ${COMPREPLY[@]:0:5}"
    
    return 0
}

# Helper function to check if all completions start with a prefix
check_all_start_with() {
    local prefix="$1"
    for completion in "${COMPREPLY[@]}"; do
        if [[ "$completion" != "$prefix"* ]]; then
            return 1
        fi
    done
    return 0
}

echo "=== File@Stage Completion Test Suite ==="
echo

# Test 1: README.md@ should show all stages from main README.md (not subdirectory ones)
run_completion_test \
    "README.md@ should show all stages from main README.md" \
    "markdown-runner" "-r" "-B" "README.md@"

if check_all_start_with "README.md@" && [[ ${#COMPREPLY[@]} -gt 10 ]]; then
    echo -e "  ${GREEN}PASS: Shows multiple stages from main README.md${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Should show multiple stages from main README.md (got ${#COMPREPLY[@]} completions)${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 2: Non-recursive should also work correctly
run_completion_test \
    "README.md@ should work in non-recursive mode" \
    "markdown-runner" "-B" "README.md@"

if check_all_start_with "README.md@" && [[ ${#COMPREPLY[@]} -gt 10 ]]; then
    echo -e "  ${GREEN}PASS: Shows multiple stages in non-recursive mode${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Should show multiple stages in non-recursive mode (got ${#COMPREPLY[@]} completions)${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 3: File not in current directory should return no completions (correct behavior)
run_completion_test \
    "parallel.md@ should return no completions (file not in current dir)" \
    "markdown-runner" "-B" "parallel.md@"

if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
    echo -e "  ${GREEN}PASS: Correctly returns no completions for file not in current directory${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Should return no completions for file not in current directory (got ${#COMPREPLY[@]} completions)${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 4: File@stage/chunk format should work
run_completion_test \
    "README.md@help/ should show chunks for help stage" \
    "markdown-runner" "-B" "README.md@help/"

if [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Shows chunks for specific stage${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Should show chunks for specific stage${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 5: Partial file@stage completion should work
run_completion_test \
    "README.md@hel should complete to help-related stages" \
    "markdown-runner" "-B" "README.md@hel"

if [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Shows matching stages for partial input${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Should show matching stages for partial input${NC}"
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
