#!/bin/bash

# Test suite for stage validation functionality
# Tests the _is_valid_stage_name function

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

# Helper function to test stage validation
test_stage_validation() {
    local test_name="$1"
    local stage_name="$2"
    local expected_result="$3"  # "valid" or "invalid"
    
    echo -e "${YELLOW}Testing: $test_name${NC}"
    echo "  Stage name: '$stage_name'"
    
    TESTS_RUN=$((TESTS_RUN + 1))
    
    # Set up minimal completion context
    COMP_WORDS=("markdown-runner" "-B" "$stage_name")
    COMP_CWORD=2
    
    # Call the function
    local result
    if _is_valid_stage_name "$stage_name"; then
        result="valid"
    else
        result="invalid"
    fi
    
    echo "  Result: $result"
    echo "  Expected: $expected_result"
    
    if [[ "$result" == "$expected_result" ]]; then
        echo -e "  ${GREEN}PASS${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "  ${RED}FAIL${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    echo
}

echo "=== Stage Validation Test Suite ==="
echo

# Test 1: Valid stage names
test_stage_validation \
    "help is valid stage" \
    "help" \
    "valid"

test_stage_validation \
    "setup is valid stage" \
    "setup" \
    "valid"

test_stage_validation \
    "main is valid stage" \
    "main" \
    "valid"

test_stage_validation \
    "test is valid stage" \
    "test" \
    "valid"

test_stage_validation \
    "test1 is valid stage" \
    "test1" \
    "valid"

test_stage_validation \
    "teardown is valid stage" \
    "teardown" \
    "valid"

test_stage_validation \
    "integration-test is valid stage" \
    "integration-test" \
    "valid"

# Test 2: Invalid stage names
test_stage_validation \
    "nonexistent is invalid stage" \
    "nonexistent" \
    "invalid"

test_stage_validation \
    "partial hel is invalid stage" \
    "hel" \
    "invalid"

test_stage_validation \
    "partial tes is invalid stage" \
    "tes" \
    "invalid"

test_stage_validation \
    "empty string is invalid stage" \
    "" \
    "invalid"

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
