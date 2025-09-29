#!/bin/bash

# Test suite for file discovery functionality in completion
# Tests the _get_executable_files function in various contexts

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

# Helper function to test file discovery
test_file_discovery() {
    local test_name="$1"
    local comp_words_str="$2"
    local comp_cword="$3"
    local expected_result="$4"
    
    echo -e "${YELLOW}Testing: $test_name${NC}"
    echo "  COMP_WORDS: $comp_words_str"
    echo "  COMP_CWORD: $comp_cword"
    
    TESTS_RUN=$((TESTS_RUN + 1))
    
    # Set up completion variables
    IFS=' ' read -ra COMP_WORDS <<< "$comp_words_str"
    COMP_CWORD=$comp_cword
    
    # Call the function
    local result
    result=$(_get_executable_files)
    
    echo "  Result: '$result'"
    echo "  Expected: '$expected_result'"
    
    if [[ "$result" == "$expected_result" ]]; then
        echo -e "  ${GREEN}PASS${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "  ${RED}FAIL${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    echo
}

echo "=== File Discovery Test Suite ==="
echo

# Test 1: Basic flag completion contexts
test_file_discovery \
    "Basic -B flag completion" \
    "markdown-runner -B help" \
    2 \
    "./README.md"

test_file_discovery \
    "File@stage completion context" \
    "markdown-runner -B README.md@" \
    2 \
    "./README.md"

test_file_discovery \
    "Stage/chunk completion context" \
    "markdown-runner -B setup/" \
    2 \
    "./README.md"

# Test 2: Directory arguments should be respected
test_file_discovery \
    "Explicit directory argument" \
    "markdown-runner test/ -B help" \
    3 \
    ""

test_file_discovery \
    "Current directory default" \
    "markdown-runner -B" \
    2 \
    "./README.md"

# Test 3: Edge cases
test_file_discovery \
    "Multiple flags" \
    "markdown-runner -v -B help" \
    3 \
    "./README.md"

test_file_discovery \
    "Flag with value" \
    "markdown-runner -t 30 -B help" \
    4 \
    "./README.md"

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
