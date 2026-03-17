#!/bin/bash

# Test Framework for Bash Completion Tests (JSON-based)
# Runs declarative test data from JSON

# Source the completion script
SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
source "$SCRIPT_DIR/../bash_completion.sh"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# ============================================================================
# Core Test Execution
# ============================================================================

# Run a single test case from JSON data
# Args: $1 = test JSON object (as string)
run_test_json() {
    local test_json="$1"

    TESTS_RUN=$((TESTS_RUN + 1))

    # Parse test data using jq
    local test_name
    test_name=$(echo "$test_json" | jq -r '.name')

    local assertion_type
    assertion_type=$(echo "$test_json" | jq -r '.assertion')

    # Parse COMP_WORDS array
    local comp_words_json
    comp_words_json=$(echo "$test_json" | jq -c '.comp_words')

    # Convert JSON array to bash array
    local i=0
    COMP_WORDS=()
    while IFS= read -r word; do
        COMP_WORDS+=("$word")
        ((i++))
    done < <(echo "$comp_words_json" | jq -r '.[]')

    COMP_CWORD=$((${#COMP_WORDS[@]} - 1))

    # Clear previous results
    COMPREPLY=()

    # Run completion
    _markdown_runner_completion 2>/dev/null

    # Parse expected value(s)
    local expected_type
    expected_type=$(echo "$test_json" | jq -r '.expected | type')

    local result=0

    if [[ "$expected_type" == "array" ]]; then
        # Expected is an array - parse it
        local expected=()
        while IFS= read -r val; do
            expected+=("$val")
        done < <(echo "$test_json" | jq -r '.expected[]')

        case "$assertion_type" in
            exact)
                assert_exact "${expected[@]}"
                result=$?
                ;;
            contains)
                assert_contains "${expected[@]}"
                result=$?
                ;;
            excludes)
                assert_excludes "${expected[@]}"
                result=$?
                ;;
            all_match)
                # For all_match, expected should be single pattern
                assert_all_match "${expected[0]}"
                result=$?
                ;;
            any_match)
                # For any_match, expected should be single pattern
                assert_any_match "${expected[0]}"
                result=$?
                ;;
            *)
                echo -e "${RED}Unknown assertion type: $assertion_type${NC}"
                return 1
                ;;
        esac
    else
        # Expected is a number
        local expected_num
        expected_num=$(echo "$test_json" | jq -r '.expected')

        case "$assertion_type" in
            count_gt)
                assert_count_gt "$expected_num"
                result=$?
                ;;
            count_eq)
                assert_count_eq "$expected_num"
                result=$?
                ;;
            *)
                echo -e "${RED}Unknown assertion type: $assertion_type${NC}"
                return 1
                ;;
        esac
    fi

    # Report result
    if [[ $result -eq 0 ]]; then
        echo -e "${GREEN}✓${NC} $test_name"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} $test_name"
        echo -e "  ${YELLOW}COMP_WORDS:${NC} $(printf '%s|' "${COMP_WORDS[@]}" | sed 's/|$//')"

        # Format expected for display
        local expected_display
        if [[ "$expected_type" == "array" ]]; then
            expected_display=$(echo "$test_json" | jq -r '.expected | join("|")')
        else
            expected_display=$(echo "$test_json" | jq -r '.expected')
        fi
        echo -e "  ${YELLOW}Expected ($assertion_type):${NC} $expected_display"

        echo -e "  ${YELLOW}Got:${NC} $(printf '%s|' "${COMPREPLY[@]}" | sed 's/|$//')"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    return 0  # Don't fail the whole suite on single test failure
}

# ============================================================================
# Assertion Functions
# ============================================================================

# Assert exact match (order and content)
assert_exact() {
    local expected=("$@")

    [[ ${#COMPREPLY[@]} -ne ${#expected[@]} ]] && return 1

    for ((i=0; i<${#expected[@]}; i++)); do
        [[ "${COMPREPLY[i]}" != "${expected[i]}" ]] && return 1
    done
    return 0
}

# Assert contains all specified values (order doesn't matter)
assert_contains() {
    local expected=("$@")

    for exp in "${expected[@]}"; do
        local found=false
        for actual in "${COMPREPLY[@]}"; do
            # Handle both plain flags and flags with descriptions (e.g., "-d" or "-d, --dry-run:(dry-run)")
            if [[ "$actual" == "$exp" ]] ||
               [[ "$actual" == "$exp:"* ]] ||
               [[ "$actual" == "$exp, "* ]] ||
               [[ "$actual" == *", $exp:"* ]]; then
                found=true
                break
            fi
        done
        [[ "$found" == false ]] && return 1
    done
    return 0
}

# Assert excludes all specified values
assert_excludes() {
    local excluded=("$@")

    for excl in "${excluded[@]}"; do
        for actual in "${COMPREPLY[@]}"; do
            # Handle both plain flags and flags with descriptions
            if [[ "$actual" == "$excl" ]] ||
               [[ "$actual" == "$excl:"* ]] ||
               [[ "$actual" == "$excl, "* ]] ||
               [[ "$actual" == *", $excl:"* ]]; then
                return 1  # Found excluded value
            fi
        done
    done
    return 0
}

# Assert count greater than
assert_count_gt() {
    local threshold="$1"
    [[ ${#COMPREPLY[@]} -gt $threshold ]]
}

# Assert count equals
assert_count_eq() {
    local expected="$1"
    [[ ${#COMPREPLY[@]} -eq $expected ]]
}

# Assert all completions match pattern
assert_all_match() {
    local pattern="$1"

    [[ ${#COMPREPLY[@]} -eq 0 ]] && return 1

    for completion in "${COMPREPLY[@]}"; do
        case "$pattern" in
            "*@*")
                [[ "$completion" != *@* ]] && return 1
                ;;
            "*/*")
                [[ "$completion" != */* ]] && return 1
                ;;
            "no_@_or_/")
                [[ "$completion" == *@* ]] || [[ "$completion" == */* ]] && return 1
                ;;
            *)
                [[ ! "$completion" =~ $pattern ]] && return 1
                ;;
        esac
    done
    return 0
}

# Assert any completion matches pattern
assert_any_match() {
    local pattern="$1"

    for completion in "${COMPREPLY[@]}"; do
        case "$pattern" in
            "*@*")
                [[ "$completion" == *@* ]] && return 0
                ;;
            "*/*")
                [[ "$completion" == */* ]] && return 0
                ;;
            *)
                [[ "$completion" =~ $pattern ]] && return 0
                ;;
        esac
    done
    return 1
}

# ============================================================================
# Test Suite Runner
# ============================================================================

# Run all tests from a test suite in JSON
# Args: $1 = suite name (e.g., "timeout_tests")
#       $2 = path to JSON file
run_test_suite_json() {
    local suite_name="$1"
    local json_file="$2"

    # Check if jq is available
    if ! command -v jq &> /dev/null; then
        echo -e "${RED}Error: jq is required but not installed${NC}"
        echo "Install with: sudo dnf install jq  (or apt-get, brew, etc.)"
        return 1
    fi

    # Read and parse test suite
    local suite_tests
    suite_tests=$(jq -c ".${suite_name}[]" "$json_file" 2>/dev/null)

    if [[ -z "$suite_tests" ]]; then
        echo -e "${YELLOW}Warning: No tests found for suite '${suite_name}'${NC}"
        return 0
    fi

    echo -e "\n${BLUE}=== $(echo "$suite_name" | sed 's/_/ /g' | sed 's/\b\(.\)/\u\1/g') ===${NC}"

    # Run each test
    while IFS= read -r test_json; do
        run_test_json "$test_json"
    done <<< "$suite_tests"
}

# Print final results
print_results() {
    echo
    echo "============================================"
    echo "Test Results:"
    echo "  Total:  $TESTS_RUN"
    echo -e "  ${GREEN}Passed: $TESTS_PASSED${NC}"
    [[ $TESTS_FAILED -gt 0 ]] && echo -e "  ${RED}Failed: $TESTS_FAILED${NC}" || echo "  Failed: 0"
    echo "============================================"

    [[ $TESTS_FAILED -eq 0 ]]
}
