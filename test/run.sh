#!/bin/bash

# A simple integration test suite for the markdown-runner.

VERBOSE=false
if [ "$1" == "--verbose" ]; then
    VERBOSE=true
fi

# Function to print a colored message
print_msg() {
    COLOR=$1
    MSG=$2
    printf "\e[${COLOR}m%s\e[0m\n" "${MSG}"
}


# --- Test Setup ---
print_msg "33" "--- Building markdown-runner binary ---"
cd "$(dirname "$0")" # Run from the script's directory
(cd .. && go build -o test/markdown-runner)

# --- Test Cases ---
TEST_COUNTER=0
FAILED_TESTS=0

# Cleanup function to be called on exit
cleanup() {
    print_msg "33" "--- Cleaning up ---"
    rm -f markdown-runner
    rm -rf test_ws
    print_msg "32" "Cleanup complete."
}

trap cleanup EXIT

# A helper function to run a test
run_test() {
    ((TEST_COUNTER++))
    DESC=$1
    CMD=$2
    printf "\e[36mRunning test #${TEST_COUNTER}: ${DESC}...\e[0m"

    local runner_cmd="./markdown-runner"
    if [ "$VERBOSE" = "true" ]; then
        runner_cmd="./markdown-runner -v"
        printf "\n"
    fi

    local final_cmd="${CMD/.\/markdown-runner/$runner_cmd}"

    local cmd_to_run_quietly="${final_cmd}"
    if [ "$VERBOSE" = "false" ]; then
        if echo "${final_cmd}" | grep -q "grep"; then
            cmd_to_run_quietly="${final_cmd} > /dev/null"
        else
            cmd_to_run_quietly="${final_cmd} >/dev/null 2>&1"
        fi
    fi

    if eval "${cmd_to_run_quietly}"; then
        printf "\e[32m ✅\e[0m\n"
        if [ "$VERBOSE" = "true" ]; then
            eval "${final_cmd}"
        fi
        return 0
    else
        printf "\e[31m ❌\e[0m\n"
        echo -e "\e[31m--- Failing command output ---\e[0m"
        eval "${final_cmd}"
        echo -e "\e[31m------------------------------\e[0m"
        return 1
    fi
}

# --- Automatically Find and Run Tests ---
# Find all markdown files in the cases directory and run them as tests.
for test_file in cases/*.md; do
    test_name=$(basename "${test_file}" .md)
    # We expect schema_error.md to fail, so we invert the result with !
    if [ "${test_name}" == "schema_error" ]; then
        run_test "Schema error test (${test_name}) should fail as expected" \
            "! ./markdown-runner cases -f '.*'${test_file}" || ((FAILED_TESTS++))
    # The teardown test has a specific success condition
    elif [ "${test_name}" == "teardown" ]; then
        run_test "Teardown test (${test_name}) should execute teardown" \
            "./markdown-runner cases -f '.*'${test_file} 2>&1 | grep 'SUCCESS.*echo teardown should execute'" || ((FAILED_TESTS++))
    # The dry-run test has a specific success condition
    elif [ "${test_name}" == "happy" ]; then
        run_test "Dry run test for (${test_name}) should show DRY-RUN" \
            "./markdown-runner -d cases -f '.*'${test_file} 2>&1 | grep -E 'DRY-RUN.*|.*echo happy path'" || ((FAILED_TESTS++))
        run_test "Happy path test (${test_name}) should succeed" \
            "./markdown-runner cases -f '.*'${test_file}" || ((FAILED_TESTS++))
    else
        run_test "Test case ${test_name}" \
            "./markdown-runner cases -f '.*'${test_file}" || ((FAILED_TESTS++))
    fi
done

run_test "Recursive test" \
    "./markdown-runner -r cases/recursive 2>&1 | grep -c -E 'echo.*nested' | grep -q 2" || ((FAILED_TESTS++))

run_test "Recursive test with file filter" \
    "./markdown-runner -r cases/recursive -f '.*other.md' 2>&1 | grep -c 'nested script' | grep -q 1" || ((FAILED_TESTS++))

# --- Test Summary ---
print_msg "33" "\n--- Test Summary ---"
if [ "${FAILED_TESTS}" -eq 0 ]; then
    print_msg "32" "All ${TEST_COUNTER} tests passed!"
    exit 0
else
    print_msg "31" "${FAILED_TESTS} of ${TEST_COUNTER} tests failed."
    exit 1
fi
