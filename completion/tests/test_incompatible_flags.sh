#!/bin/bash

# Test suite for incompatible flag filtering
# Tests that logically incompatible flags are excluded from completion suggestions

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
        # Handle both standard format and enhanced format with descriptions
        if [[ "$completion" == "$flag" ]] || [[ "$completion" == "$flag:"* ]]; then
            return 1  # Flag found (not excluded)
        fi
    done
    return 0  # Flag not found (excluded)
}

# Helper function to check if a flag is included in completions
check_flag_included() {
    local flag="$1"
    for completion in "${COMPREPLY[@]}"; do
        # Handle both standard format and combined format with descriptions
        if [[ "$completion" == "$flag" ]] || 
           [[ "$completion" == "$flag:"* ]] ||
           [[ "$completion" == "$flag, "* ]] ||
           [[ "$completion" == *", $flag:"* ]]; then
            return 0  # Flag found (included)
        fi
    done
    return 1  # Flag not found (not included)
}

echo "=== Incompatible Flags Test Suite ==="
echo

# Test 1: CI Mode Incompatibilities
echo "=== Testing CI Mode Incompatibilities ==="

run_completion_test \
    "CI mode excludes interactive flags" \
    "markdown-runner" "--view" "ci" "README.md" "-"

if check_flag_excluded "-i" && check_flag_excluded "--interactive" && 
   check_flag_excluded "-B" && check_flag_excluded "--break-at" && 
   check_flag_excluded "-s" && check_flag_excluded "--start-from"; then
    echo -e "  ${GREEN}PASS: CI mode excludes interactive flags${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: CI mode should exclude interactive flags${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "CI mode allows non-interactive flags" \
    "markdown-runner" "--view" "ci" "README.md" "-"

if check_flag_included "-d" && check_flag_included "--dry-run" && 
   check_flag_included "-l" && check_flag_included "--list"; then
    echo -e "  ${GREEN}PASS: CI mode allows non-interactive flags${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: CI mode should allow non-interactive flags${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 2: Default View Mode (should not exclude interactive flags)
echo "=== Testing Default View Mode ==="

run_completion_test \
    "Default view allows interactive flags" \
    "markdown-runner" "--view" "default" "README.md" "-"

if check_flag_included "-i" && check_flag_included "--interactive" && 
   check_flag_included "-B" && check_flag_included "--break-at"; then
    echo -e "  ${GREEN}PASS: Default view allows interactive flags${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Default view should allow interactive flags${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 3: List Mode Incompatibilities
echo "=== Testing List Mode Incompatibilities ==="

run_completion_test \
    "List mode excludes execution flags" \
    "markdown-runner" "-l" "README.md" "-"

if check_flag_excluded "-i" && check_flag_excluded "--interactive" && 
   check_flag_excluded "-B" && check_flag_excluded "--break-at" && 
   check_flag_excluded "-s" && check_flag_excluded "--start-from" &&
   check_flag_excluded "-d" && check_flag_excluded "--dry-run" &&
   check_flag_excluded "-t" && check_flag_excluded "--timeout" &&
   check_flag_excluded "-u" && check_flag_excluded "--update-files"; then
    echo -e "  ${GREEN}PASS: List mode excludes execution flags${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: List mode should exclude execution flags${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "List mode allows compatible flags" \
    "markdown-runner" "-l" "README.md" "-"

if check_flag_included "-r" && check_flag_included "--recursive" && 
   check_flag_included "-f" && check_flag_included "--filter" &&
   check_flag_included "-v" && check_flag_included "--verbose"; then
    echo -e "  ${GREEN}PASS: List mode allows compatible flags${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: List mode should allow compatible flags${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 4: Help Mode Incompatibilities
echo "=== Testing Help Mode Incompatibilities ==="

run_completion_test \
    "Help mode excludes all other flags" \
    "markdown-runner" "-h" "-"

if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
    echo -e "  ${GREEN}PASS: Help mode excludes all other flags${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Help mode should exclude all other flags, got ${#COMPREPLY[@]}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Long help mode excludes all other flags" \
    "markdown-runner" "--help" "-"

if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
    echo -e "  ${GREEN}PASS: Long help mode excludes all other flags${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Long help mode should exclude all other flags, got ${#COMPREPLY[@]}${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 5: Dry Run Incompatibilities
echo "=== Testing Dry Run Incompatibilities ==="

run_completion_test \
    "Dry run excludes update-files" \
    "markdown-runner" "--dry-run" "README.md" "-"

if check_flag_excluded "-u" && check_flag_excluded "--update-files"; then
    echo -e "  ${GREEN}PASS: Dry run excludes update-files${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Dry run should exclude update-files${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Dry run allows other flags" \
    "markdown-runner" "--dry-run" "README.md" "-"

if check_flag_included "-i" && check_flag_included "--interactive" && 
   check_flag_included "-B" && check_flag_included "--break-at"; then
    echo -e "  ${GREEN}PASS: Dry run allows other compatible flags${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Dry run should allow other compatible flags${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 6: Verbose vs Quiet Mutual Exclusion
echo "=== Testing Verbose vs Quiet Mutual Exclusion ==="

run_completion_test \
    "Verbose excludes quiet" \
    "markdown-runner" "-v" "README.md" "-"

if check_flag_excluded "-q" && check_flag_excluded "--quiet"; then
    echo -e "  ${GREEN}PASS: Verbose excludes quiet${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Verbose should exclude quiet${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Quiet excludes verbose" \
    "markdown-runner" "-q" "README.md" "-"

if check_flag_excluded "-v" && check_flag_excluded "--verbose"; then
    echo -e "  ${GREEN}PASS: Quiet excludes verbose${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Quiet should exclude verbose${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 7: Complex Combinations
echo "=== Testing Complex Combinations ==="

run_completion_test \
    "Multiple incompatible contexts" \
    "markdown-runner" "-v" "--dry-run" "-l" "README.md" "-"

# Should exclude: verbose+quiet, dry-run+update-files, list+execution flags
if check_flag_excluded "-q" && check_flag_excluded "--quiet" &&
   check_flag_excluded "-u" && check_flag_excluded "--update-files" &&
   check_flag_excluded "-i" && check_flag_excluded "-B" && check_flag_excluded "-s"; then
    echo -e "  ${GREEN}PASS: Multiple incompatible contexts handled correctly${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Multiple incompatible contexts should be handled${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 8: Bidirectional Incompatibilities
echo "=== Testing Bidirectional Incompatibilities ==="

run_completion_test \
    "Interactive flag excludes CI and list modes" \
    "markdown-runner" "-i" "README.md" "-"

if check_flag_excluded "--view" && check_flag_excluded "-l" && check_flag_excluded "--list" &&
   check_flag_excluded "-h" && check_flag_excluded "--help"; then
    echo -e "  ${GREEN}PASS: Interactive flag excludes CI and list modes${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Interactive flag should exclude CI and list modes${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Break-at flag excludes CI and list modes" \
    "markdown-runner" "-B" "help" "README.md" "-"

if check_flag_excluded "--view" && check_flag_excluded "-l" && check_flag_excluded "--list" &&
   check_flag_excluded "-h" && check_flag_excluded "--help"; then
    echo -e "  ${GREEN}PASS: Break-at flag excludes CI and list modes${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Break-at flag should exclude CI and list modes${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Start-from flag excludes CI and list modes" \
    "markdown-runner" "-s" "help" "README.md" "-"

if check_flag_excluded "--view" && check_flag_excluded "-l" && check_flag_excluded "--list" &&
   check_flag_excluded "-h" && check_flag_excluded "--help"; then
    echo -e "  ${GREEN}PASS: Start-from flag excludes CI and list modes${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Start-from flag should exclude CI and list modes${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Update-files flag excludes dry-run" \
    "markdown-runner" "-u" "README.md" "-"

if check_flag_excluded "-d" && check_flag_excluded "--dry-run"; then
    echo -e "  ${GREEN}PASS: Update-files flag excludes dry-run${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Update-files flag should exclude dry-run${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Break-at flag excludes ignore-breakpoints" \
    "markdown-runner" "-B" "help" "README.md" "-"

if check_flag_excluded "--ignore-breakpoints"; then
    echo -e "  ${GREEN}PASS: Break-at flag excludes ignore-breakpoints${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Break-at flag should exclude ignore-breakpoints${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Ignore-breakpoints flag excludes break-at" \
    "markdown-runner" "--ignore-breakpoints" "README.md" "-"

if check_flag_excluded "-B" && check_flag_excluded "--break-at"; then
    echo -e "  ${GREEN}PASS: Ignore-breakpoints flag excludes break-at${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Ignore-breakpoints flag should exclude break-at${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 9: Edge Cases
echo "=== Testing Edge Cases ==="

run_completion_test \
    "View flag without value (should not trigger CI exclusions)" \
    "markdown-runner" "--view" "README.md" "-"

# Should not exclude interactive flags since no "ci" value specified
if check_flag_included "-i" && check_flag_included "-B"; then
    echo -e "  ${GREEN}PASS: View flag without CI value allows interactive flags${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: View flag without CI value should allow interactive flags${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "No incompatible flags used (baseline)" \
    "markdown-runner" "-r" "README.md" "-"

# Should have normal exclusions (just -r and --recursive)
if [[ ${#COMPREPLY[@]} -eq 14 ]] && check_flag_excluded "-r" && check_flag_excluded "--recursive"; then
    echo -e "  ${GREEN}PASS: No incompatible flags shows normal exclusions${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: No incompatible flags should show normal exclusions (expected 14, got ${#COMPREPLY[@]})${NC}"
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
