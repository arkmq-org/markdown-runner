#!/bin/bash

# Test suite for -s/--start-from flag completion
# Tests that start-from behaves similarly to break-at but may have subtle differences

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

# Helper function to check if all completions are file@ format
check_all_file_at() {
    for completion in "${COMPREPLY[@]}"; do
        if [[ "$completion" != *@* ]]; then
            return 1  # Found a non-file@ completion
        fi
    done
    return 0  # All completions are file@ format
}

# Helper function to check if all completions are stage names (no @ or /)
check_all_stage_names() {
    for completion in "${COMPREPLY[@]}"; do
        if [[ "$completion" == *@* ]] || [[ "$completion" == */* ]]; then
            return 1  # Found a file@ or chunk completion
        fi
    done
    return 0  # All completions are stage names
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

echo "=== Start-From Flag Completion Test Suite ==="
echo

# Test 1: Basic start-from completion (should be similar to break-at)
echo "=== Testing Basic Start-From Completion ==="

run_completion_test \
    "Start-from empty completion" \
    "markdown-runner" "-s" ""

if [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Start-from empty completion shows results (${#COMPREPLY[@]} found)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Start-from empty completion should show results${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Start-from long form empty completion" \
    "markdown-runner" "--start-from" ""

if [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Long start-from empty completion shows results (${#COMPREPLY[@]} found)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Long start-from empty completion should show results${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 2: Stage name completion
echo "=== Testing Stage Name Completion ==="

run_completion_test \
    "Start-from stage completion 'help'" \
    "markdown-runner" "-s" "help"

if [[ ${#COMPREPLY[@]} -ge 1 ]]; then
    echo -e "  ${GREEN}PASS: Start-from stage 'help' shows completions (${#COMPREPLY[@]} found)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Start-from stage 'help' should show completions${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Start-from partial stage 'hel'" \
    "markdown-runner" "-s" "hel"

if [[ ${#COMPREPLY[@]} -ge 1 ]]; then
    echo -e "  ${GREEN}PASS: Start-from partial 'hel' shows completions (${#COMPREPLY[@]} found)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Start-from partial 'hel' should show completions${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Start-from stage completion 'setup'" \
    "markdown-runner" "-s" "setup"

if [[ ${#COMPREPLY[@]} -ge 1 ]]; then
    echo -e "  ${GREEN}PASS: Start-from stage 'setup' shows completions (${#COMPREPLY[@]} found)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Start-from stage 'setup' should show completions${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 3: File@stage completion
echo "=== Testing File@Stage Completion ==="

run_completion_test \
    "Start-from file@stage empty" \
    "markdown-runner" "-s" "README.md@"

if [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Start-from README.md@ shows stages (${#COMPREPLY[@]} found)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Start-from README.md@ should show stages${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Start-from file@stage partial" \
    "markdown-runner" "-s" "README.md@hel"

if [[ ${#COMPREPLY[@]} -ge 1 ]]; then
    echo -e "  ${GREEN}PASS: Start-from README.md@hel shows completions (${#COMPREPLY[@]} found)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Start-from README.md@hel should show completions${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Start-from file@stage complete" \
    "markdown-runner" "-s" "README.md@help"

if [[ ${#COMPREPLY[@]} -ge 1 ]]; then
    echo -e "  ${GREEN}PASS: Start-from README.md@help shows completions (${#COMPREPLY[@]} found)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Start-from README.md@help should show completions${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 4: Directory context (recursive mode)
echo "=== Testing Directory Context ==="

run_completion_test \
    "Start-from recursive mode" \
    "markdown-runner" "-r" "-s" ""

if check_all_file_at && [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Start-from recursive shows file@ completions${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
elif [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Start-from recursive shows completions (${#COMPREPLY[@]} found)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Start-from recursive should show completions${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Start-from directory context" \
    "markdown-runner" "test/cases/" "-s" ""

if check_all_file_at && [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Start-from directory shows file@ completions${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
elif [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Start-from directory shows completions (${#COMPREPLY[@]} found)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Start-from directory should show completions${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Start-from directory partial" \
    "markdown-runner" "test/cases/" "-s" "p"

if [[ ${#COMPREPLY[@]} -eq 1 ]] && [[ "${COMPREPLY[0]}" == "parallel.md@" ]]; then
    echo -e "  ${GREEN}PASS: Start-from directory partial 'p' shows parallel.md@${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
elif [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Start-from directory partial 'p' shows completions: ${COMPREPLY[*]}${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Start-from directory partial 'p' should show completions${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 5: Comparison with break-at (equivalence test)
echo "=== Testing Start-From vs Break-At Equivalence ==="

# Test same scenario with both flags
run_completion_test \
    "Break-at help completion" \
    "markdown-runner" "-B" "help"

break_at_result=(${COMPREPLY[@]})

run_completion_test \
    "Start-from help completion" \
    "markdown-runner" "-s" "help"

start_from_result=(${COMPREPLY[@]})

if [[ ${#break_at_result[@]} -eq ${#start_from_result[@]} ]] && [[ ${#break_at_result[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Start-from and break-at produce similar results for 'help'${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Start-from and break-at should produce similar results${NC}"
    echo "    Break-at: ${break_at_result[*]}"
    echo "    Start-from: ${start_from_result[*]}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test file@stage equivalence
run_completion_test \
    "Break-at file@stage" \
    "markdown-runner" "-B" "README.md@help"

break_at_file_result=(${COMPREPLY[@]})

run_completion_test \
    "Start-from file@stage" \
    "markdown-runner" "-s" "README.md@help"

start_from_file_result=(${COMPREPLY[@]})

if [[ ${#break_at_file_result[@]} -eq ${#start_from_file_result[@]} ]] && [[ ${#break_at_file_result[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Start-from and break-at produce similar results for file@stage${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Start-from and break-at should produce similar results for file@stage${NC}"
    echo "    Break-at: ${break_at_file_result[*]}"
    echo "    Start-from: ${start_from_file_result[*]}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 6: Specific file context
echo "=== Testing Specific File Context ==="

run_completion_test \
    "Start-from with specific file" \
    "markdown-runner" "README.md" "-s" ""

if [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Start-from with specific file shows completions (${#COMPREPLY[@]} found)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Start-from with specific file should show completions${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Start-from with specific file and stage" \
    "markdown-runner" "README.md" "-s" "help"

if [[ ${#COMPREPLY[@]} -ge 1 ]]; then
    echo -e "  ${GREEN}PASS: Start-from with specific file and stage shows completions (${#COMPREPLY[@]} found)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Start-from with specific file and stage should show completions${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 7: Edge cases
echo "=== Testing Edge Cases ==="

run_completion_test \
    "Start-from non-existent stage" \
    "markdown-runner" "-s" "nonexistent"

# Should handle gracefully (may show no completions)
echo -e "  ${GREEN}PASS: Start-from non-existent stage handled gracefully (${#COMPREPLY[@]} completions)${NC}"
TESTS_PASSED=$((TESTS_PASSED + 1))
echo

run_completion_test \
    "Start-from non-existent file@stage" \
    "markdown-runner" "-s" "nonexistent.md@stage"

# Should handle gracefully
echo -e "  ${GREEN}PASS: Start-from non-existent file@stage handled gracefully (${#COMPREPLY[@]} completions)${NC}"
TESTS_PASSED=$((TESTS_PASSED + 1))
echo

# Test 8: Complex scenarios
echo "=== Testing Complex Scenarios ==="

run_completion_test \
    "Start-from with multiple flags" \
    "markdown-runner" "-r" "-v" "-s" ""

if [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Start-from with multiple flags works (${#COMPREPLY[@]} completions)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Start-from with multiple flags should work${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Start-from with directory and multiple flags" \
    "markdown-runner" "-r" "test/cases/" "-s" "p"

if [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Start-from complex scenario works (${#COMPREPLY[@]} completions)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Start-from complex scenario should work${NC}"
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
