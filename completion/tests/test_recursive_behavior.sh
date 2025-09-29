#!/bin/bash

# Test suite for recursive flag behavior in completion
# Tests the distinction between recursive and non-recursive completion modes

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
    local expected_behavior="$2"
    shift 2
    local comp_words=("$@")
    
    echo -e "${YELLOW}Testing: $test_name${NC}"
    echo "  COMP_WORDS: ${comp_words[*]}"
    echo "  Expected: $expected_behavior"
    
    TESTS_RUN=$((TESTS_RUN + 1))
    
    # Set up completion variables
    COMP_WORDS=("${comp_words[@]}")
    COMP_CWORD=$((${#COMP_WORDS[@]} - 1))
    
    # Clear previous results
    COMPREPLY=()
    
    # Run completion
    _markdown_runner_completion
    
    echo "  Completions found: ${#COMPREPLY[@]}"
    echo "  First 10 completions: ${COMPREPLY[@]:0:10}"
    
    return 0  # We'll do manual verification for now
}

# Helper function to check if all completions end with @
check_all_file_at_completions() {
    for completion in "${COMPREPLY[@]}"; do
        if [[ "$completion" != *@ ]]; then
            return 1  # Found a completion that doesn't end with @
        fi
    done
    return 0  # All completions end with @
}

# Helper function to check if completions contain stage names (not ending with @)
check_contains_stage_names() {
    for completion in "${COMPREPLY[@]}"; do
        if [[ "$completion" != *@ ]]; then
            return 0  # Found a stage name
        fi
    done
    return 1  # No stage names found
}

echo "=== Recursive Behavior Test Suite ==="
echo

# Test 1: Non-recursive should show stage names + file@ completions
run_completion_test \
    "Non-recursive -B should show stages + file@ completions" \
    "Mix of stage names and file@ completions" \
    "markdown-runner" "-B" ""

if check_contains_stage_names && [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Contains stage names (non-recursive behavior)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Should contain stage names in non-recursive mode${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 2: Recursive should show ONLY file@ completions
run_completion_test \
    "Recursive -B should show ONLY file@ completions" \
    "Only file@ completions, no stage names" \
    "markdown-runner" "-r" "-B" ""

if check_all_file_at_completions && [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: All completions are file@ format (recursive behavior)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Should only show file@ completions in recursive mode${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 3: Non-recursive with partial stage name should work
run_completion_test \
    "Non-recursive partial stage name" \
    "Matching stage names" \
    "markdown-runner" "-B" "hel"

if [[ ${#COMPREPLY[@]} -gt 0 ]] && ! check_all_file_at_completions; then
    echo -e "  ${GREEN}PASS: Shows matching stage names for partial input${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Should show matching stage names for partial input${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 4: Recursive with partial should show only matching file@ completions
run_completion_test \
    "Recursive partial should filter file@ completions" \
    "Only file@ completions matching the partial input" \
    "markdown-runner" "-r" "-B" "README"

# For this test, we expect either 0 completions (no files match) or file@ completions
if [[ ${#COMPREPLY[@]} -eq 0 ]] || check_all_file_at_completions; then
    echo -e "  ${GREEN}PASS: Recursive partial shows only file@ completions${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Recursive partial should only show file@ completions${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 5: --recursive long form should work the same as -r
run_completion_test \
    "Long form --recursive should work like -r" \
    "Only file@ completions, no stage names" \
    "markdown-runner" "--recursive" "-B" ""

if check_all_file_at_completions && [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: --recursive works the same as -r${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: --recursive should work the same as -r${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 6: Recursive with -s flag should also work
run_completion_test \
    "Recursive -s should show ONLY file@ completions" \
    "Only file@ completions, no stage names" \
    "markdown-runner" "-r" "-s" ""

if check_all_file_at_completions && [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Recursive -s shows only file@ completions${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Recursive -s should only show file@ completions${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 7: Directory + recursive should show only file@ completions from that directory
run_completion_test \
    "Directory + recursive should show ONLY file@ from directory" \
    "Only file@ completions from specified directory" \
    "markdown-runner" "test/cases/" "-r" "-B" ""

if check_all_file_at_completions && [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Directory + recursive shows only file@ completions${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Directory + recursive should only show file@ completions${NC}"
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
