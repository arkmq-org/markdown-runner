#!/bin/bash

# Test suite for directory context detection
# Tests the _is_directory_context function

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

# Helper function to test directory context detection
test_directory_context() {
    local test_name="$1"
    local comp_words_str="$2"
    local comp_cword="$3"
    local expected_result="$4"  # "true" or "false"
    
    echo -e "${YELLOW}Testing: $test_name${NC}"
    echo "  COMP_WORDS: $comp_words_str"
    echo "  COMP_CWORD: $comp_cword"
    
    TESTS_RUN=$((TESTS_RUN + 1))
    
    # Set up completion variables
    IFS=' ' read -ra COMP_WORDS <<< "$comp_words_str"
    COMP_CWORD=$comp_cword
    
    # Call the function
    local result
    if _is_directory_context; then
        result="true"
    else
        result="false"
    fi
    
    echo "  Result: $result"
    echo "  Expected: $expected_result"
    
    if [[ "$result" == "$expected_result" ]]; then
        echo -e "  ${GREEN}PASS${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "  ${RED}FAIL${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    echo
}

echo "=== Directory Context Detection Test Suite ==="
echo

# Test 1: Flag value contexts should NOT be directory contexts
test_directory_context \
    "Completing -B flag value" \
    "markdown-runner -B help" \
    2 \
    "false"

test_directory_context \
    "Completing -s flag value" \
    "markdown-runner -s setup" \
    2 \
    "false"

test_directory_context \
    "Completing -f flag value" \
    "markdown-runner -f pattern" \
    2 \
    "false"

# Test 2: Directory arguments should be directory contexts
test_directory_context \
    "Explicit directory argument" \
    "markdown-runner test/" \
    1 \
    "true"

test_directory_context \
    "Directory with flags" \
    "markdown-runner test/cases/ -B parallel_write" \
    3 \
    "false"  # We're completing a complete stage name, not directory context

# Test 3: Main path completion should be directory context
test_directory_context \
    "Completing main path argument" \
    "markdown-runner" \
    1 \
    "true"

test_directory_context \
    "Completing main path with flags before" \
    "markdown-runner -v" \
    2 \
    "true"

test_directory_context \
    "Recursive -B flag should be directory context" \
    "markdown-runner -r -B" \
    3 \
    "true"

test_directory_context \
    "Directory argument with -B flag should be directory context" \
    "markdown-runner test/cases/ -B" \
    3 \
    "true"

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
