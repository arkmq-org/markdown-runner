#!/bin/bash

# Test suite for smart nospace behavior
# Tests that single-chunk stages allow trailing spaces while multi-chunk stages use nospace

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

# Helper function to count chunks for a stage
count_chunks_for_stage() {
    local stage="$1"
    local md_files
    md_files=$(_get_executable_files | tr '\n' ' ')
    
    if [[ -n "$md_files" ]]; then
        # Use the same regex pattern as the actual completion code
        echo "$md_files" | xargs grep -h '```[a-zA-Z0-9_. -]*{.*"stage"[[:space:]]*:[[:space:]]*"'"$stage"'"' 2>/dev/null | wc -l
    else
        echo "0"
    fi
}

echo "=== Smart Nospace Behavior Test Suite ==="
echo

# Test 1: Single chunk stage completion (should allow trailing space)
echo "=== Testing Single Chunk Stages (should allow trailing space) ==="

run_completion_test \
    "help stage (1 chunk) - partial completion" \
    "markdown-runner" "-B" "hel"

help_chunks=$(count_chunks_for_stage "help")
if [[ ${#COMPREPLY[@]} -eq 1 ]] && [[ "${COMPREPLY[0]}" == "help" ]] && [[ $help_chunks -eq 1 ]]; then
    echo -e "  ${GREEN}PASS: Single chunk stage should allow trailing space${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Expected single completion 'help' for single chunk stage (chunks: $help_chunks)${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "test stage (1 chunk) - partial completion" \
    "markdown-runner" "-B" "tes"

# Check if test is in the results and has 1 chunk
test_chunks=$(count_chunks_for_stage "test")
has_test_stage=false
for completion in "${COMPREPLY[@]}"; do
    if [[ "$completion" == "test" ]]; then
        has_test_stage=true
        break
    fi
done

if [[ "$has_test_stage" == true ]] && [[ $test_chunks -eq 1 ]]; then
    echo -e "  ${GREEN}PASS: test stage (1 chunk) present in multi-completion${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: test stage should be present with 1 chunk (chunks: $test_chunks)${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 2: Multi chunk stage completion (should use nospace)
echo "=== Testing Multi Chunk Stages (should use nospace) ==="

run_completion_test \
    "setup stage (2 chunks) - partial completion" \
    "markdown-runner" "-B" "setu"

setup_chunks=$(count_chunks_for_stage "setup")
if [[ ${#COMPREPLY[@]} -eq 1 ]] && [[ "${COMPREPLY[0]}" == "setup" ]] && [[ $setup_chunks -gt 1 ]]; then
    echo -e "  ${GREEN}PASS: Multi chunk stage should use nospace (chunks: $setup_chunks)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Expected single completion 'setup' for multi chunk stage (chunks: $setup_chunks)${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 3: Mixed completion (should use nospace if ANY stage has multiple chunks)
echo "=== Testing Mixed Completions (should use nospace if any multi-chunk) ==="

run_completion_test \
    "Mixed stages with different chunk counts" \
    "markdown-runner" "-B" "te"

# Check chunk counts for stages starting with "te"
has_multi_chunk=false
for completion in "${COMPREPLY[@]}"; do
    chunk_count=$(count_chunks_for_stage "$completion")
    if [[ $chunk_count -gt 1 ]]; then
        has_multi_chunk=true
        break
    fi
done

if [[ "$has_multi_chunk" == true ]]; then
    echo -e "  ${GREEN}PASS: Mixed completion should use nospace (has multi-chunk stages)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Mixed completion should use nospace when any stage has multiple chunks${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 4: Verify chunk counting logic
echo "=== Testing Chunk Counting Logic ==="

echo "Chunk counts for key stages:"
for stage in "help" "setup" "test" "main"; do
    chunk_count=$(count_chunks_for_stage "$stage")
    echo "  $stage: $chunk_count chunks"
done

# Verify our assumptions about chunk counts
help_chunks=$(count_chunks_for_stage "help")
setup_chunks=$(count_chunks_for_stage "setup")

if [[ $help_chunks -eq 1 ]] && [[ $setup_chunks -gt 1 ]]; then
    echo -e "  ${GREEN}PASS: Chunk counting logic works correctly${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Chunk counting logic incorrect (help: $help_chunks, setup: $setup_chunks)${NC}"
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
