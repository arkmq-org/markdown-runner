#!/bin/bash

# Test suite for markdown-runner bash completion
# This script tests various completion scenarios to ensure they work correctly

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
    local expected_pattern="$3"
    local should_contain="$4"  # Optional: specific completion that should be present
    
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
    
    if [[ $completion_count -eq 0 ]]; then
        echo -e "  ${RED}FAIL: No completions found${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
    
    # Check if expected pattern matches (if provided)
    if [[ -n "$expected_pattern" ]]; then
        local pattern_match=false
        for completion in "${COMPREPLY[@]}"; do
            if [[ "$completion" =~ $expected_pattern ]]; then
                pattern_match=true
                break
            fi
        done
        
        if [[ "$pattern_match" == false ]]; then
            echo -e "  ${RED}FAIL: No completion matches pattern '$expected_pattern'${NC}"
            echo "  Completions found: ${COMPREPLY[*]}"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            return 1
        fi
    fi
    
    # Check if specific completion is present (if provided)
    if [[ -n "$should_contain" ]]; then
        local contains=false
        for completion in "${COMPREPLY[@]}"; do
            if [[ "$completion" == "$should_contain" ]]; then
                contains=true
                break
            fi
        done
        
        if [[ "$contains" == false ]]; then
            echo -e "  ${RED}FAIL: Expected completion '$should_contain' not found${NC}"
            echo "  Completions found: ${COMPREPLY[*]}"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            return 1
        fi
    fi
    
    echo -e "  ${GREEN}PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    return 0
}

# Helper function to test that _get_executable_files works correctly
test_get_executable_files() {
    echo -e "${YELLOW}Testing: _get_executable_files function${NC}"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    # Test with normal command line
    COMP_WORDS=("markdown-runner" "-B" "README.md@")
    local files
    files=$(_get_executable_files)
    
    if [[ -z "$files" ]]; then
        echo -e "  ${RED}FAIL: _get_executable_files returned no files${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
    
    if [[ "$files" != *"README.md"* ]]; then
        echo -e "  ${RED}FAIL: _get_executable_files did not find README.md${NC}"
        echo "  Files found: $files"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
    
    echo -e "  ${GREEN}PASS: Found files: $files${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    return 0
}

echo "=== Markdown Runner Bash Completion Test Suite ==="
echo

# Test 1: Basic stage completion
run_completion_test \
    "Basic stage completion" \
    "markdown-runner -B README.md@" \
    "README\.md@.*" \
    "README.md@setup"

# Test 2: Partial stage completion
run_completion_test \
    "Partial stage completion" \
    "markdown-runner -B README.md@test" \
    "README\.md@test.*" \
    "README.md@test1"

# Test 3: Chunk completion
# NOTE: This test may hang due to an issue in the original test framework
# Use test_stage_chunk_completion.sh for reliable chunk completion testing
run_completion_test \
    "Chunk completion" \
    "markdown-runner -B README.md@setup/" \
    "README\.md@setup/.*" \
    "README.md@setup/init"

# Test 4: Stage completion without file prefix
run_completion_test \
    "Stage completion without file prefix" \
    "markdown-runner -B setup" \
    "setup"

# Test 5: Chunk completion without file prefix
run_completion_test \
    "Chunk completion without file prefix" \
    "markdown-runner -B setup/" \
    "setup/.*"

# Test 6: Test _get_executable_files function
# test_get_executable_files  # Commented out for now

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
