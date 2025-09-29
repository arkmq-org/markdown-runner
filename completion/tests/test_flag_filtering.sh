#!/bin/bash

# Test suite for flag filtering functionality
# Tests that already-used flags are excluded from completion suggestions

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
    echo "  Completions: ${COMPREPLY[*]:0:5}..."
    
    return 0
}

# Helper function to check if a flag is excluded from completions
check_flag_excluded() {
    local flag="$1"
    for completion in "${COMPREPLY[@]}"; do
        # Handle both standard format and combined format with descriptions
        if [[ "$completion" == "$flag" ]] || 
           [[ "$completion" == "$flag:"* ]] ||
           [[ "$completion" == "$flag, "* ]] ||
           [[ "$completion" == *", $flag:"* ]]; then
            return 1  # Flag found (not excluded)
        fi
    done
    return 0  # Flag not found (excluded)
}

# Helper function to check if all flags are present (baseline)
check_all_flags_present() {
    # Should have all 15 deduplicated flags when none are used
    [[ ${#COMPREPLY[@]} -eq 15 ]]
}

echo "=== Flag Filtering Test Suite ==="
echo

# Test 1: Baseline - no flags used
echo "=== Testing Baseline (No Flags Used) ==="

run_completion_test \
    "No flags used - all should be available" \
    "markdown-runner" "README.md" "-"

if check_all_flags_present; then
    echo -e "  ${GREEN}PASS: All flags available when none used (${#COMPREPLY[@]} flags)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Should have all 15 deduplicated flags when none used, got ${#COMPREPLY[@]}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 2: Single flag exclusion
echo "=== Testing Single Flag Exclusion ==="

run_completion_test \
    "Recursive flag (-r) should be excluded" \
    "markdown-runner" "-r" "test/cases/parallel.md" "-"

if check_flag_excluded "-r" && check_flag_excluded "--recursive" && [[ ${#COMPREPLY[@]} -eq 14 ]]; then
    echo -e "  ${GREEN}PASS: Recursive flag and equivalent excluded (${#COMPREPLY[@]} flags remaining)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Recursive flag should be excluded${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Verbose flag (-v) should be excluded" \
    "markdown-runner" "-v" "README.md" "-"

if check_flag_excluded "-v" && check_flag_excluded "--verbose" && 
   check_flag_excluded "-q" && check_flag_excluded "--quiet" && [[ ${#COMPREPLY[@]} -eq 13 ]]; then
    echo -e "  ${GREEN}PASS: Verbose flag and equivalent excluded (plus incompatible quiet)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Verbose flag should be excluded (expected 13, got ${#COMPREPLY[@]})${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Break-at flag (-B) should be excluded" \
    "markdown-runner" "-B" "help" "README.md" "-"

if check_flag_excluded "-B" && check_flag_excluded "--break-at" && 
   check_flag_excluded "--view" && check_flag_excluded "-l" && check_flag_excluded "--list" &&
   check_flag_excluded "-h" && check_flag_excluded "--help" && 
   check_flag_excluded "--ignore-breakpoints" && [[ ${#COMPREPLY[@]} -eq 10 ]]; then
    echo -e "  ${GREEN}PASS: Break-at flag and equivalent excluded (plus incompatible)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Break-at flag should be excluded (expected 11, got ${#COMPREPLY[@]})${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 3: Long flag exclusion
echo "=== Testing Long Flag Exclusion ==="

run_completion_test \
    "Long dry-run flag should be excluded" \
    "markdown-runner" "--dry-run" "README.md" "-"

if check_flag_excluded "--dry-run" && check_flag_excluded "-d" && 
   check_flag_excluded "-u" && check_flag_excluded "--update-files" && [[ ${#COMPREPLY[@]} -eq 13 ]]; then
    echo -e "  ${GREEN}PASS: Long dry-run flag and equivalent excluded (plus incompatible update-files)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Long dry-run flag should be excluded (expected 13, got ${#COMPREPLY[@]})${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Long recursive flag should be excluded" \
    "markdown-runner" "--recursive" "test/" "-"

if check_flag_excluded "--recursive" && check_flag_excluded "-r" && [[ ${#COMPREPLY[@]} -eq 14 ]]; then
    echo -e "  ${GREEN}PASS: Long recursive flag and equivalent excluded${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Long recursive flag should be excluded${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 4: Multiple flags exclusion
echo "=== Testing Multiple Flags Exclusion ==="

run_completion_test \
    "Multiple flags should be excluded" \
    "markdown-runner" "-v" "-d" "README.md" "-"

if check_flag_excluded "-v" && check_flag_excluded "--verbose" && 
   check_flag_excluded "-d" && check_flag_excluded "--dry-run" && 
   check_flag_excluded "-q" && check_flag_excluded "--quiet" &&
   check_flag_excluded "-u" && check_flag_excluded "--update-files" && 
   [[ ${#COMPREPLY[@]} -eq 11 ]]; then
    echo -e "  ${GREEN}PASS: Multiple flags and equivalents excluded (${#COMPREPLY[@]} flags remaining)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Multiple flags should be excluded (expected 11, got ${#COMPREPLY[@]})${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Three flags should be excluded" \
    "markdown-runner" "-r" "-v" "-i" "README.md" "-"

if check_flag_excluded "-r" && check_flag_excluded "-v" && check_flag_excluded "-i" && 
   check_flag_excluded "-q" && check_flag_excluded "--quiet" &&
   check_flag_excluded "--view" && check_flag_excluded "-l" && check_flag_excluded "--list" &&
   check_flag_excluded "-h" && check_flag_excluded "--help" && [[ ${#COMPREPLY[@]} -eq 8 ]]; then
    echo -e "  ${GREEN}PASS: Three flags and equivalents excluded (${#COMPREPLY[@]} flags remaining)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Three flags should be excluded (expected 8, got ${#COMPREPLY[@]})${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 5: Mixed short and long flags
echo "=== Testing Mixed Short and Long Flags ==="

run_completion_test \
    "Mixed short and long flags" \
    "markdown-runner" "-v" "--dry-run" "README.md" "-"

if check_flag_excluded "-v" && check_flag_excluded "--verbose" && 
   check_flag_excluded "-d" && check_flag_excluded "--dry-run" && 
   check_flag_excluded "-q" && check_flag_excluded "--quiet" &&
   check_flag_excluded "-u" && check_flag_excluded "--update-files" && 
   [[ ${#COMPREPLY[@]} -eq 11 ]]; then
    echo -e "  ${GREEN}PASS: Mixed short/long flags excluded (plus incompatible)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Mixed short/long flags should be excluded (expected 11, got ${#COMPREPLY[@]})${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 6: Flags with values
echo "=== Testing Flags with Values ==="

run_completion_test \
    "Flag with value should be excluded" \
    "markdown-runner" "-t" "30" "README.md" "-"

if check_flag_excluded "-t" && check_flag_excluded "--timeout" && [[ ${#COMPREPLY[@]} -eq 14 ]]; then
    echo -e "  ${GREEN}PASS: Flag with value excluded${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Flag with value should be excluded${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 7: Edge cases
echo "=== Testing Edge Cases ==="

run_completion_test \
    "Flag at different positions" \
    "markdown-runner" "README.md" "-v" "-"

if check_flag_excluded "-v" && check_flag_excluded "--verbose" && 
   check_flag_excluded "-q" && check_flag_excluded "--quiet" && [[ ${#COMPREPLY[@]} -eq 13 ]]; then
    echo -e "  ${GREEN}PASS: Flag at different position excluded (plus incompatible)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Flag at different position should be excluded (expected 13, got ${#COMPREPLY[@]})${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Partial flag completion with exclusion" \
    "markdown-runner" "-v" "README.md" "--d"

# Should complete to --dry-run but exclude --debug-* if any existed
if [[ ${#COMPREPLY[@]} -eq 1 ]] && [[ "${COMPREPLY[0]}" == "--dry-run" ]]; then
    echo -e "  ${GREEN}PASS: Partial flag completion works with exclusion${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Partial flag completion should work with exclusion${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 8: All flag pairs
echo "=== Testing All Flag Pair Exclusions ==="

flag_pairs=(
    "-d:--dry-run"
    "-l:--list" 
    "-i:--interactive"
    "-s:--start-from"
    "-B:--break-at"
    "-t:--timeout"
    "-u:--update-files"
    "-f:--filter"
    "-r:--recursive"
    "-v:--verbose"
    "-q:--quiet"
    "-h:--help"
)

local pair_tests_passed=0
for pair in "${flag_pairs[@]}"; do
    IFS=':' read -r short long <<< "$pair"
    
    # Test short excludes long
    COMP_WORDS=("markdown-runner" "$short" "README.md" "-")
    COMP_CWORD=3
    COMPREPLY=()
    _markdown_runner_completion
    
    if check_flag_excluded "$short" && check_flag_excluded "$long"; then
        ((pair_tests_passed++))
    fi
done

if [[ $pair_tests_passed -eq ${#flag_pairs[@]} ]]; then
    echo -e "  ${GREEN}PASS: All flag pairs properly exclude equivalents (${#flag_pairs[@]}/${#flag_pairs[@]})${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Some flag pairs don't exclude equivalents ($pair_tests_passed/${#flag_pairs[@]})${NC}"
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
