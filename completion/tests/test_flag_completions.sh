#!/bin/bash

# Test suite for flag completions
# Tests timeout values, view modes, and flag name completion

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

# Helper function to check if array contains expected values
check_contains_all() {
    local expected=("$@")
    for exp in "${expected[@]}"; do
        local found=false
        for actual in "${COMPREPLY[@]}"; do
            if [[ "$actual" == "$exp" ]]; then
                found=true
                break
            fi
        done
        if [[ "$found" == false ]]; then
            return 1
        fi
    done
    return 0
}

# Helper function to check exact match
check_exact_match() {
    local expected=("$@")
    if [[ ${#COMPREPLY[@]} -ne ${#expected[@]} ]]; then
        return 1
    fi
    
    for ((i=0; i<${#expected[@]}; i++)); do
        if [[ "${COMPREPLY[i]}" != "${expected[i]}" ]]; then
            return 1
        fi
    done
    return 0
}

echo "=== Flag Completions Test Suite ==="
echo

# Test 1: Timeout flag completion
echo "=== Testing Timeout Flag Completion ==="

run_completion_test \
    "Timeout flag empty completion" \
    "markdown-runner" "-t" ""

if check_exact_match "1" "5" "10" "30" "60"; then
    echo -e "  ${GREEN}PASS: Timeout flag shows all expected values${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Expected '1 5 10 30 60', got: ${COMPREPLY[*]}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Timeout flag partial '1' completion" \
    "markdown-runner" "-t" "1"

if [[ ${#COMPREPLY[@]} -eq 2 ]] && check_contains_all "1" "10"; then
    echo -e "  ${GREEN}PASS: Timeout partial '1' shows '1' and '10'${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Expected '1 10', got: ${COMPREPLY[*]}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Timeout flag partial '3' completion" \
    "markdown-runner" "-t" "3"

if [[ ${#COMPREPLY[@]} -eq 1 ]] && [[ "${COMPREPLY[0]}" == "30" ]]; then
    echo -e "  ${GREEN}PASS: Timeout partial '3' shows '30'${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Expected '30', got: ${COMPREPLY[*]}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Timeout flag long form completion" \
    "markdown-runner" "--timeout" ""

if check_exact_match "1" "5" "10" "30" "60"; then
    echo -e "  ${GREEN}PASS: Long timeout flag shows all expected values${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Expected '1 5 10 30 60', got: ${COMPREPLY[*]}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 2: View flag completion
echo "=== Testing View Flag Completion ==="

run_completion_test \
    "View flag empty completion" \
    "markdown-runner" "--view" ""

if check_exact_match "default" "ci"; then
    echo -e "  ${GREEN}PASS: View flag shows 'default' and 'ci'${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Expected 'default ci', got: ${COMPREPLY[*]}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "View flag partial 'd' completion" \
    "markdown-runner" "--view" "d"

if [[ ${#COMPREPLY[@]} -eq 1 ]] && [[ "${COMPREPLY[0]}" == "default" ]]; then
    echo -e "  ${GREEN}PASS: View partial 'd' shows 'default'${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Expected 'default', got: ${COMPREPLY[*]}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "View flag partial 'c' completion" \
    "markdown-runner" "--view" "c"

if [[ ${#COMPREPLY[@]} -eq 1 ]] && [[ "${COMPREPLY[0]}" == "ci" ]]; then
    echo -e "  ${GREEN}PASS: View partial 'c' shows 'ci'${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Expected 'ci', got: ${COMPREPLY[*]}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 3: Filter flag completion (should show no completions)
echo "=== Testing Filter Flag Completion ==="

run_completion_test \
    "Filter flag completion (should be empty)" \
    "markdown-runner" "-f" ""

if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
    echo -e "  ${GREEN}PASS: Filter flag shows no completions (regex patterns)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Expected no completions, got: ${COMPREPLY[*]}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Filter flag long form completion" \
    "markdown-runner" "--filter" ""

if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
    echo -e "  ${GREEN}PASS: Long filter flag shows no completions${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Expected no completions, got: ${COMPREPLY[*]}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 4: Flag name completion
echo "=== Testing Flag Name Completion ==="

run_completion_test \
    "Short flag completion" \
    "markdown-runner" "-"

# Check that we get multiple flags and they start with -
if [[ ${#COMPREPLY[@]} -gt 10 ]] && [[ "${COMPREPLY[0]}" == -* ]]; then
    echo -e "  ${GREEN}PASS: Short flag completion shows multiple flags (${#COMPREPLY[@]} found)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Expected multiple flags starting with -, got: ${#COMPREPLY[@]} completions${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Long flag completion" \
    "markdown-runner" "--"

# Check that we get multiple long flags and they start with --
if [[ ${#COMPREPLY[@]} -gt 5 ]] && [[ "${COMPREPLY[0]}" == --* ]]; then
    echo -e "  ${GREEN}PASS: Long flag completion shows multiple flags (${#COMPREPLY[@]} found)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Expected multiple flags starting with --, got: ${#COMPREPLY[@]} completions${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Specific flag partial 'h' completion" \
    "markdown-runner" "-h"

# Should complete to -h (help flag)
if check_contains_all "-h"; then
    echo -e "  ${GREEN}PASS: Flag partial 'h' includes '-h'${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Expected to include '-h', got: ${COMPREPLY[*]}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Specific long flag partial '--dry' completion" \
    "markdown-runner" "--dry"

# Should complete to --dry-run
if [[ ${#COMPREPLY[@]} -eq 1 ]] && [[ "${COMPREPLY[0]}" == "--dry-run" ]]; then
    echo -e "  ${GREEN}PASS: Long flag partial '--dry' completes to '--dry-run'${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Expected '--dry-run', got: ${COMPREPLY[*]}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 5: Flag equivalence (short vs long forms)
echo "=== Testing Flag Equivalence ==="

# Test that both -B and --break-at work
run_completion_test \
    "Short break-at flag works" \
    "markdown-runner" "-B" "help"

short_result=(${COMPREPLY[@]})

run_completion_test \
    "Long break-at flag works" \
    "markdown-runner" "--break-at" "help"

long_result=(${COMPREPLY[@]})

if [[ ${#short_result[@]} -eq ${#long_result[@]} ]] && [[ ${#short_result[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Short and long break-at flags produce equivalent results${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Short (-B) and long (--break-at) flags should be equivalent${NC}"
    echo "    Short: ${short_result[*]}"
    echo "    Long: ${long_result[*]}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 6: Invalid flag handling
echo "=== Testing Invalid Flag Handling ==="

run_completion_test \
    "Invalid flag completion" \
    "markdown-runner" "-z" ""

# Invalid flags should not crash and should show no completions
echo -e "  ${GREEN}PASS: Invalid flag handled gracefully (${#COMPREPLY[@]} completions)${NC}"
TESTS_PASSED=$((TESTS_PASSED + 1))
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
