#!/bin/bash

# Main Test Runner for Bash Completion Tests
# Runs all declarative tests from test_data.json

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Change to repository root so completion can find markdown files
# The completion script looks for .md files in the current directory
cd "$SCRIPT_DIR/../.."

# Check if jq is available
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed"
    echo "Install with:"
    echo "  Fedora/RHEL: sudo dnf install jq"
    echo "  Debian/Ubuntu: sudo apt-get install jq"
    echo "  macOS: brew install jq"
    exit 1
fi

# Source the test framework
source "$SCRIPT_DIR/test_framework.sh"

# Path to test data
TEST_DATA="$SCRIPT_DIR/test_data.json"

# Check if test data exists
if [[ ! -f "$TEST_DATA" ]]; then
    echo "Error: Test data file not found: $TEST_DATA"
    exit 1
fi

# Banner
echo "╔════════════════════════════════════════════════════════════════╗"
echo "║       Bash Completion Test Suite                               ║"
echo "╚════════════════════════════════════════════════════════════════╝"

# Run all test suites (automatically discover from JSON keys)
test_suites=$(jq -r 'keys[] | select(endswith("_tests"))' "$TEST_DATA" 2>/dev/null)

while IFS= read -r suite; do
    run_test_suite_json "$suite" "$TEST_DATA"
done <<< "$test_suites"

# Print final results
print_results

# Exit with failure if any tests failed
exit $?
