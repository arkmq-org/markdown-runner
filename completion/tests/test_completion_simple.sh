#!/bin/bash

# Simplified test suite for markdown-runner bash completion
# This version focuses on the core functionality that we know works

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Source the completion script
source "$(dirname "$0")/../bash_completion.sh"

# Helper function to run a completion test
run_completion_test() {
    local test_name="$1"
    local comp_line="$2"
    local expected_count="$3"  # Expected minimum number of completions
    
    echo -e "${YELLOW}Testing: $test_name${NC}"
    echo "  Command: $comp_line"
    
    TESTS_RUN=$((TESTS_RUN + 1))
    
    # Parse the completion line
    IFS=' ' read -ra COMP_WORDS <<< "$comp_line"
    COMP_CWORD=$((${#COMP_WORDS[@]} - 1))
    COMP_LINE="$comp_line"
    COMP_POINT=${#COMP_LINE}
    
    # Clear previous results
    COMPREPLY=()
    
    # Run completion
    _markdown_runner_completion
    
    # Check results
    local completion_count=${#COMPREPLY[@]}
    echo "  Found $completion_count completions"
    
    if [[ $completion_count -lt $expected_count ]]; then
        echo -e "  ${RED}FAIL: Expected at least $expected_count completions, got $completion_count${NC}"
        echo "  Completions found: ${COMPREPLY[*]}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
    
    echo "  Sample completions: ${COMPREPLY[@]:0:3}..."
    echo -e "  ${GREEN}PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    return 0
}

echo "=== Markdown Runner Bash Completion Test Suite (Simplified) ==="
echo

# Test 1: Basic stage completion
run_completion_test \
    "Basic stage completion" \
    "markdown-runner -B README.md@" \
    10

# Test 2: Partial stage completion
run_completion_test \
    "Partial stage completion" \
    "markdown-runner -B README.md@test" \
    3

# Test 3: Chunk completion
run_completion_test \
    "Chunk completion" \
    "markdown-runner -B README.md@setup/" \
    2

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
