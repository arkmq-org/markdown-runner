#!/bin/bash

# Test suite for positional argument completion behavior
# Tests that after providing a markdown file, completion switches to flags

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

# Helper function to check if completions are flags
check_all_flags() {
    for completion in "${COMPREPLY[@]}"; do
        if [[ "$completion" != -* ]]; then
            return 1  # Found a non-flag completion
        fi
    done
    return 0  # All completions are flags
}

# Helper function to check if completions are files/directories
check_files_or_dirs() {
    for completion in "${COMPREPLY[@]}"; do
        if [[ "$completion" == *.md ]] || [[ "$completion" == */ ]] || [[ -f "$completion" ]] || [[ -d "$completion" ]]; then
            return 0  # Found a file or directory
        fi
    done
    return 1  # No files or directories found
}

echo "=== Positional Arguments Test Suite ==="
echo

# Test 1: No markdown file provided (should show files)
echo "=== Testing No Markdown File Provided ==="

run_completion_test \
    "Empty completion (should show files)" \
    "markdown-runner" ""

if check_files_or_dirs && [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: No file provided shows files/directories${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: No file provided should show files/directories${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Partial file completion" \
    "markdown-runner" "README"

if check_files_or_dirs && [[ ${#COMPREPLY[@]} -gt 0 ]]; then
    echo -e "  ${GREEN}PASS: Partial file completion shows files${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Partial file completion should show files${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 2: Markdown file provided (should show flags)
echo "=== Testing Markdown File Provided ==="

run_completion_test \
    "README.md provided (should show flags)" \
    "markdown-runner" "README.md" ""

if check_all_flags && [[ ${#COMPREPLY[@]} -gt 10 ]]; then
    echo -e "  ${GREEN}PASS: Markdown file provided shows flags${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Markdown file provided should show flags${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "File with path provided (should show flags)" \
    "markdown-runner" "test/cases/parallel.md" ""

if check_all_flags && [[ ${#COMPREPLY[@]} -gt 10 ]]; then
    echo -e "  ${GREEN}PASS: File with path shows flags${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: File with path should show flags${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Different markdown extension (.MD)" \
    "markdown-runner" "file.MD" ""

if check_all_flags && [[ ${#COMPREPLY[@]} -gt 10 ]]; then
    echo -e "  ${GREEN}PASS: .MD extension shows flags${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: .MD extension should show flags${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Different markdown extension (.Markdown)" \
    "markdown-runner" "file.Markdown" ""

if check_all_flags && [[ ${#COMPREPLY[@]} -gt 10 ]]; then
    echo -e "  ${GREEN}PASS: .Markdown extension shows flags${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: .Markdown extension should show flags${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 3: Directory provided (should show flags, not more files)
echo "=== Testing Directory Provided ==="

run_completion_test \
    "Directory provided (should show flags)" \
    "markdown-runner" "test/" ""

if check_all_flags && [[ ${#COMPREPLY[@]} -gt 10 ]]; then
    echo -e "  ${GREEN}PASS: Directory provided shows flags${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Directory provided should show flags${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Specific directory (test/cases/) should show flags" \
    "markdown-runner" "test/cases/" ""

if check_all_flags && [[ ${#COMPREPLY[@]} -gt 10 ]]; then
    echo -e "  ${GREEN}PASS: Specific directory shows flags${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Specific directory should show flags${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Nested directory should show flags" \
    "markdown-runner" "test/cases/recursive/" ""

if check_all_flags && [[ ${#COMPREPLY[@]} -gt 10 ]]; then
    echo -e "  ${GREEN}PASS: Nested directory shows flags${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Nested directory should show flags${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 4: File + flags combination
echo "=== Testing File + Flags Combination ==="

run_completion_test \
    "File + flag completion" \
    "markdown-runner" "README.md" "-v" ""

if check_all_flags && [[ ${#COMPREPLY[@]} -gt 10 ]]; then
    echo -e "  ${GREEN}PASS: File + flag shows more flags${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: File + flag should show more flags${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "Multiple flags + file" \
    "markdown-runner" "-v" "-d" "README.md" ""

if check_all_flags && [[ ${#COMPREPLY[@]} -gt 10 ]]; then
    echo -e "  ${GREEN}PASS: Multiple flags + file shows flags${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Multiple flags + file should show flags${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 5: Flag value completion (should not be affected)
echo "=== Testing Flag Value Completion ==="

run_completion_test \
    "File + break-at flag value" \
    "markdown-runner" "README.md" "-B" ""

# Should show stage names, not flags
if [[ ${#COMPREPLY[@]} -gt 0 ]] && ! check_all_flags; then
    echo -e "  ${GREEN}PASS: File + flag value shows flag-specific completion${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: File + flag value should show flag-specific completion${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "File + timeout flag value" \
    "markdown-runner" "README.md" "-t" ""

# Should show timeout values
expected_timeouts=("1" "5" "10" "30" "60")
has_timeout_values=false
for timeout in "${expected_timeouts[@]}"; do
    for completion in "${COMPREPLY[@]}"; do
        if [[ "$completion" == "$timeout" ]]; then
            has_timeout_values=true
            break 2
        fi
    done
done

if [[ "$has_timeout_values" == true ]]; then
    echo -e "  ${GREEN}PASS: File + timeout flag shows timeout values${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: File + timeout flag should show timeout values${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 6: Partial flag completion
echo "=== Testing Partial Flag Completion ==="

run_completion_test \
    "File + partial flag" \
    "markdown-runner" "README.md" "-"

if check_all_flags && [[ ${#COMPREPLY[@]} -gt 10 ]]; then
    echo -e "  ${GREEN}PASS: File + partial flag shows all flags${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: File + partial flag should show all flags${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "File + specific flag prefix" \
    "markdown-runner" "README.md" "--dry"

if [[ ${#COMPREPLY[@]} -eq 1 ]] && [[ "${COMPREPLY[0]}" == "--dry-run" ]]; then
    echo -e "  ${GREEN}PASS: File + specific flag prefix completes correctly${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: File + specific flag prefix should complete to --dry-run${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

# Test 7: Edge cases
echo "=== Testing Edge Cases ==="

run_completion_test \
    "Non-existent markdown file" \
    "markdown-runner" "nonexistent.md" ""

if check_all_flags && [[ ${#COMPREPLY[@]} -gt 10 ]]; then
    echo -e "  ${GREEN}PASS: Non-existent .md file still triggers flag mode${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: Non-existent .md file should trigger flag mode${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo

run_completion_test \
    "File with no extension" \
    "markdown-runner" "somefile" ""

if check_files_or_dirs; then
    echo -e "  ${GREEN}PASS: File with no extension shows files (not flag mode)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "  ${RED}FAIL: File with no extension should show files${NC}"
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
