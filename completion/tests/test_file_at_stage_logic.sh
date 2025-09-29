#!/bin/bash

# Test suite for file@stage complete vs partial logic
# Tests the regression fix where complete stage names should show chunks, not stage names

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

# Helper function to check if all completions are chunks (contain /)
check_all_chunks() {
    for completion in "${COMPREPLY[@]}"; do
        if [[ "$completion" != */* ]]; then
            return 1  # Found a non-chunk completion
        fi
    done
    return 0  # All completions are chunks
}

# Helper function to check if all completions are stage names (don't contain /)
check_all_stages() {
    for completion in "${COMPREPLY[@]}"; do
        if [[ "$completion" == */* ]]; then
            return 1  # Found a chunk completion
        fi
    done
    return 0  # All completions are stage names
}

echo "=== File@Stage Complete vs Partial Logic Test Suite ==="
echo

# Test 1: Complete stage names should show chunks
echo "=== Testing Complete Stage Names (should show chunks) ==="

run_completion_test \
    "README.md@inner_test1 (complete stage) should show chunks" \
    "markdown-runner" "-B" "README.md@inner_test1"

if check_all_chunks && [[ ${#COMPREPLY[@]} -gt 1 ]]; then
    echo -e "  ${GREEN}PASS: Complete stage shows multiple chunks${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Complete stage should show chunks, not stage names${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "README.md@help (complete stage) should show chunks" \
    "markdown-runner" "-B" "README.md@help"

if check_all_chunks && [[ ${#COMPREPLY[@]} -ge 1 ]]; then
    echo -e "  ${GREEN}PASS: Complete stage shows chunks${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Complete stage should show chunks${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "README.md@setup (complete stage) should show chunks" \
    "markdown-runner" "-B" "README.md@setup"

if check_all_chunks && [[ ${#COMPREPLY[@]} -ge 1 ]]; then
    echo -e "  ${GREEN}PASS: Complete stage shows chunks${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Complete stage should show chunks${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 2: Partial stage names should complete to stage names
echo "=== Testing Partial Stage Names (should complete to stages) ==="

run_completion_test \
    "README.md@inner (partial stage) should complete to stage" \
    "markdown-runner" "-B" "README.md@inner"

if check_all_stages && [[ ${#COMPREPLY[@]} -ge 1 ]]; then
    echo -e "  ${GREEN}PASS: Partial stage completes to stage names${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Partial stage should complete to stage names, not chunks${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "README.md@hel (partial stage) should complete to stage" \
    "markdown-runner" "-B" "README.md@hel"

if check_all_stages && [[ ${#COMPREPLY[@]} -ge 1 ]]; then
    echo -e "  ${GREEN}PASS: Partial stage completes to stage names${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Partial stage should complete to stage names${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "README.md@se (partial stage) should complete to stage" \
    "markdown-runner" "-B" "README.md@se"

if check_all_stages && [[ ${#COMPREPLY[@]} -ge 1 ]]; then
    echo -e "  ${GREEN}PASS: Partial stage completes to stage names${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Partial stage should complete to stage names${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 3: Non-existent stages should show no completions
echo "=== Testing Non-existent Stages ==="

run_completion_test \
    "README.md@nonexistent (invalid stage) should show nothing" \
    "markdown-runner" "-B" "README.md@nonexistent"

if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
    echo -e "  ${GREEN}PASS: Non-existent stage shows no completions${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Non-existent stage should show no completions${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 4: File@stage/chunk format should still work
echo "=== Testing File@Stage/Chunk Format ==="

run_completion_test \
    "README.md@inner_test1/some (chunk prefix) should show matching chunks" \
    "markdown-runner" "-B" "README.md@inner_test1/some"

if check_all_chunks && [[ ${#COMPREPLY[@]} -ge 1 ]]; then
    # Check that all completions start with the expected prefix
    all_match_prefix=true
    for completion in "${COMPREPLY[@]}"; do
        if [[ "$completion" != "README.md@inner_test1/some"* ]]; then
            all_match_prefix=false
            break
        fi
    done
    
    if [[ "$all_match_prefix" == true ]]; then
        echo -e "  ${GREEN}PASS: Chunk prefix shows matching chunks${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "  ${RED}FAIL: Chunk prefix should show matching chunks only${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
else
    echo -e "  ${RED}FAIL: Chunk prefix should show matching chunks${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 5: Recursive vs non-recursive should work the same
echo "=== Testing Recursive vs Non-recursive ==="

run_completion_test \
    "Recursive: README.md@inner_test1 should show chunks" \
    "markdown-runner" "-r" "-B" "README.md@inner_test1"

recursive_count=${#COMPREPLY[@]}

run_completion_test \
    "Non-recursive: README.md@inner_test1 should show chunks" \
    "markdown-runner" "-B" "README.md@inner_test1"

non_recursive_count=${#COMPREPLY[@]}

if [[ $recursive_count -eq $non_recursive_count ]] && [[ $recursive_count -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Recursive and non-recursive show same results${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Recursive ($recursive_count) and non-recursive ($non_recursive_count) should show same results${NC}"
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
