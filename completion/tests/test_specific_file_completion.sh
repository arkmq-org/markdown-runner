#!/bin/bash

# Test suite for specific file completion functionality
# Tests that when a specific markdown file is provided, completion only shows stages from that file

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
    COMP_LINE="${COMP_WORDS[*]}"
    COMP_POINT=${#COMP_LINE}
    
    # Clear previous results
    COMPREPLY=()
    
    # Run completion
    _markdown_runner_completion
    
    echo "  Completions found: ${#COMPREPLY[@]}"
    echo "  Completions: ${COMPREPLY[*]}"
    
    return 0  # We'll do manual verification for now
}

# Helper function to check if completions contain a specific item
check_completion_contains() {
    local item="$1"
    for completion in "${COMPREPLY[@]}"; do
        if [[ "$completion" == "$item" ]]; then
            return 0
        fi
    done
    return 1
}

# Helper function to check if completions do NOT contain a specific item
check_completion_not_contains() {
    local item="$1"
    for completion in "${COMPREPLY[@]}"; do
        if [[ "$completion" == "$item" ]]; then
            return 1
        fi
    done
    return 0
}

echo "=== Specific File Completion Test Suite ==="
echo

# Test 1: Specific file should only show stages from that file (NOT README.md@)
run_completion_test \
    "Specific file should NOT include README.md@" \
    "markdown-runner" "test/cases/parallel.md" "-B" ""

if check_completion_not_contains "README.md@" && [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Does not include README.md@${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Should not include README.md@${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 2: No specific file should NOT include file@ completions (only stage names)
run_completion_test \
    "No specific file should NOT include README.md@" \
    "markdown-runner" "-B" ""

if check_completion_not_contains "README.md@" && [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Does not include README.md@ when no specific file${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Should not include README.md@ when no specific file${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 3: Directory context should show file@ completions
run_completion_test \
    "Directory context should show file@ completions" \
    "markdown-runner" "test/cases/" "-B" ""

if check_completion_contains "parallel.md@" && check_completion_contains "happy.md@"; then
    echo -e "  ${GREEN}PASS: Shows file@ completions for directory${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Should show file@ completions for directory${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 4: Specific file with different extensions
if [[ -f "test/cases/happy.md" ]]; then
    run_completion_test \
        "Different specific file should not include README.md@" \
        "markdown-runner" "test/cases/happy.md" "-B" ""
    
    if check_completion_not_contains "README.md@" && [[ ${#COMPREPLY[@]} -gt 0 ]]; then
        echo -e "  ${GREEN}PASS: Different file does not include README.md@${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "  ${RED}FAIL: Different file should not include README.md@${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    echo
fi

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
