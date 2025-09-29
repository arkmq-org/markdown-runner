#!/bin/bash

# Test suite for stage chunk completion functionality
# Tests the ability to complete stage names to show chunks within those stages

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
    local comp_line="$2"
    local expected_count="$3"
    local expected_completions="$4"  # Space-separated list of expected completions
    
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
    
    if [[ $completion_count -ne $expected_count ]]; then
        echo -e "  ${RED}FAIL: Expected $expected_count completions, got $completion_count${NC}"
        echo "  Completions found: ${COMPREPLY[*]}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
    
    # Check if expected completions are present
    if [[ -n "$expected_completions" ]]; then
        IFS=' ' read -ra expected_array <<< "$expected_completions"
        for expected in "${expected_array[@]}"; do
            local found=false
            for completion in "${COMPREPLY[@]}"; do
                if [[ "$completion" == "$expected" ]]; then
                    found=true
                    break
                fi
            done
            
            if [[ "$found" == false ]]; then
                echo -e "  ${RED}FAIL: Expected completion '$expected' not found${NC}"
                echo "  Completions found: ${COMPREPLY[*]}"
                TESTS_FAILED=$((TESTS_FAILED + 1))
                return 1
            fi
        done
    fi
    
    echo "  Completions: ${COMPREPLY[*]}"
    echo -e "  ${GREEN}PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    return 0
}

echo "=== Stage Chunk Completion Test Suite ==="
echo

# Test 1: Exact stage names should show chunks
run_completion_test \
    "help stage shows chunks" \
    "markdown-runner -B help" \
    1 \
    "help/0"

run_completion_test \
    "setup stage shows chunks" \
    "markdown-runner -B setup" \
    2 \
    "setup/init setup/1"

run_completion_test \
    "main stage shows chunks" \
    "markdown-runner -B main" \
    1 \
    "main/process"

run_completion_test \
    "test stage shows chunks" \
    "markdown-runner -B test" \
    1 \
    "test/0"

# Test 2: Partial stage names should show matching stage names
run_completion_test \
    "partial 'hel' shows help stage" \
    "markdown-runner -B hel" \
    1 \
    "help"

run_completion_test \
    "partial 'se' shows setup stage" \
    "markdown-runner -B se" \
    1 \
    "setup"

run_completion_test \
    "partial 'tes' shows test stages" \
    "markdown-runner -B tes" \
    6 \
    "test test1 test2 test3 test4 test5"

# Test 3: Existing functionality should still work
run_completion_test \
    "file@stage format shows chunks for complete stage" \
    "markdown-runner -B README.md@setup" \
    2 \
    "README.md@setup/init README.md@setup/1"

run_completion_test \
    "stage/chunk format still works" \
    "markdown-runner -B setup/" \
    2 \
    "setup/init setup/1"

run_completion_test \
    "file@stage/chunk format still works" \
    "markdown-runner -B README.md@setup/" \
    2 \
    "README.md@setup/init README.md@setup/1"

echo
# Note: Stage name completions use smart nospace behavior:
# - Single chunk stages: Allow trailing space (no need to chain)
# - Multiple chunk stages: Use nospace to enable chaining
echo "Note: Smart nospace - single chunk stages get trailing space, multi-chunk stages enable chaining"

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
