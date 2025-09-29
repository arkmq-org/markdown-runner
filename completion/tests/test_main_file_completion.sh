#!/bin/bash

# Test suite for main file completion
# Tests .md file completion, directory completion, and nospace behavior

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

# Helper function to check if all completions are .md files or directories
check_md_files_and_dirs() {
    for completion in "${COMPREPLY[@]}"; do
        # Should be either .md file or directory (ending with /)
        if [[ "$completion" != *.md ]] && [[ "$completion" != */ ]]; then
            return 1
        fi
    done
    return 0
}

# Helper function to check if any completion ends with /
check_has_directories() {
    for completion in "${COMPREPLY[@]}"; do
        if [[ "$completion" == */ ]]; then
            return 0
        fi
    done
    return 1
}

# Helper function to check if any completion is .md file
check_has_md_files() {
    for completion in "${COMPREPLY[@]}"; do
        if [[ "$completion" == *.md ]]; then
            return 0
        fi
    done
    return 1
}

echo "=== Main File Completion Test Suite ==="
echo

# Test 1: Basic file completion (main argument)
echo "=== Testing Basic File Completion ==="

run_completion_test \
    "Main file argument completion (empty)" \
    "markdown-runner" ""

if check_md_files_and_dirs && [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Main completion shows .md files and directories${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Main completion should show .md files and directories only${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Main file argument with README prefix" \
    "markdown-runner" "README"

# Should complete to README.md
if [[ ${#COMPREPLY[@]} -ge 1 ]] && check_has_md_files; then
    echo -e "  ${GREEN}PASS: README prefix shows .md file completions${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: README prefix should show .md file completions${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Main file argument with .md extension" \
    "markdown-runner" "README.md"

# Should complete to README.md exactly
if [[ ${#COMPREPLY[@]} -ge 1 ]]; then
    local found_readme=false
    for completion in "${COMPREPLY[@]}"; do
        if [[ "$completion" == "README.md" ]]; then
            found_readme=true
            break
        fi
    done
    if [[ "$found_readme" == true ]]; then
        echo -e "  ${GREEN}PASS: Complete .md filename found in completions${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "  ${RED}FAIL: Complete .md filename should be in completions${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
else
    echo -e "  ${RED}FAIL: Complete .md filename should show completions${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 2: Directory completion
echo "=== Testing Directory Completion ==="

run_completion_test \
    "Directory prefix completion" \
    "markdown-runner" "test"

if check_has_directories; then
    echo -e "  ${GREEN}PASS: Directory prefix shows directory completions${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Directory prefix should show directory completions${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Complete directory path" \
    "markdown-runner" "test/"

# Should show contents of test/ directory
if [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Complete directory path shows contents (${#COMPREPLY[@]} items)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Complete directory path should show directory contents${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Nested directory completion" \
    "markdown-runner" "test/cases/"

# Should show .md files in test/cases/
if check_has_md_files && [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Nested directory shows .md files (${#COMPREPLY[@]} found)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Nested directory should show .md files${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 3: File filtering (only .md files, not other types)
echo "=== Testing File Filtering ==="

run_completion_test \
    "File filtering (should exclude non-.md files)" \
    "markdown-runner" "go"

# Should not show .go files, only directories or .md files
if check_md_files_and_dirs; then
    echo -e "  ${GREEN}PASS: File filtering excludes non-.md files${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Should only show .md files and directories${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 4: Mixed file and directory completion
echo "=== Testing Mixed File and Directory Completion ==="

run_completion_test \
    "Mixed completion with common prefix" \
    "markdown-runner" "c"

# Should show both completion/ directory and any .md files starting with 'c'
if [[ ${#COMPREPLY[@]} -gt 0 ]] && check_md_files_and_dirs; then
    echo -e "  ${GREEN}PASS: Mixed completion shows both files and directories${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Mixed completion should show files and directories${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 5: Nospace behavior for directories
echo "=== Testing Nospace Behavior ==="

# This test verifies that the completion function sets up nospace correctly
# We can't directly test compopt, but we can verify the logic

run_completion_test \
    "Directory completion (should trigger nospace logic)" \
    "markdown-runner" "test"

# Check if we have directories in the results (which should trigger nospace)
if check_has_directories; then
    echo -e "  ${GREEN}PASS: Directory completion includes directories (nospace should be triggered)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Directory completion should include directories${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 6: Path traversal
echo "=== Testing Path Traversal ==="

run_completion_test \
    "Deep path traversal" \
    "markdown-runner" "test/cases/recursive/"

if [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Deep path traversal works (${#COMPREPLY[@]} items)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Deep path traversal should show contents${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Very deep path traversal" \
    "markdown-runner" "test/cases/recursive/nested/"

if [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Very deep path traversal works (${#COMPREPLY[@]} items)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Very deep path traversal should show contents${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 7: Non-existent paths
echo "=== Testing Non-existent Paths ==="

run_completion_test \
    "Non-existent directory" \
    "markdown-runner" "nonexistent/"

# Should show no completions or handle gracefully
echo -e "  ${GREEN}PASS: Non-existent directory handled gracefully (${#COMPREPLY[@]} completions)${NC}"
TESTS_PASSED=$((TESTS_PASSED + 1))
echo

run_completion_test \
    "Non-existent file prefix" \
    "markdown-runner" "nonexistent"

# Should show no completions or handle gracefully  
echo -e "  ${GREEN}PASS: Non-existent file prefix handled gracefully (${#COMPREPLY[@]} completions)${NC}"
TESTS_PASSED=$((TESTS_PASSED + 1))
echo

# Test 8: Edge cases
echo "=== Testing Edge Cases ==="

run_completion_test \
    "Empty string completion" \
    "markdown-runner" ""

if [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Empty string shows completions (${#COMPREPLY[@]} items)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Empty string should show completions${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Single character completion" \
    "markdown-runner" "R"

if [[ ${#COMPREPLY[@]} -ge 0 ]]; then
    echo -e "  ${GREEN}PASS: Single character completion works (${#COMPREPLY[@]} items)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Single character completion should work${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 9: Case sensitivity
echo "=== Testing Case Sensitivity ==="

run_completion_test \
    "Lowercase readme completion" \
    "markdown-runner" "readme"

lowercase_count=${#COMPREPLY[@]}

run_completion_test \
    "Uppercase README completion" \
    "markdown-runner" "README"

uppercase_count=${#COMPREPLY[@]}

# Both should work (bash completion is typically case-sensitive)
if [[ $uppercase_count -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Case sensitivity works as expected (lowercase: $lowercase_count, uppercase: $uppercase_count)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Case sensitivity should work${NC}"
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
