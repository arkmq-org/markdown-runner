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
go build -o markdown-runner ../main.go

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

# Test 1: Happy Path
run_test "Happy path should succeed" \
    "./markdown-runner -m cases -f happy.md" || ((FAILED_TESTS++))

# Test 2: Schema Error
# We expect this to fail, so we invert the result with !
run_test "Schema error should cause a failure" \
    "! ./markdown-runner -m cases -f schema_error.md" || ((FAILED_TESTS++))

# Test 3: Teardown
# The command should fail, but the teardown output should be present.
run_test "Failing command with a valid teardown" \
    "./markdown-runner -m cases -f teardown.md 2>&1 | grep 'SUCCESS.*echo teardown should execute'" || ((FAILED_TESTS++))

# Test 4: Parallelism
run_test "Parallel execution should result in interleaved output" \
    "./markdown-runner -m cases -f parallel.md" || ((FAILED_TESTS++))

# Test 5: File Writer
run_test "Writer runtime should create a file with the correct content" \
    "./markdown-runner -m cases -f writer.md" || ((FAILED_TESTS++))

# Test 6: Dry Run
run_test "Dry run should display DRY-RUN message and not execute" \
    "./markdown-runner -d -m cases -f happy.md 2>&1 | grep -E 'DRY-RUN.*|.*echo happy path'" || ((FAILED_TESTS++))


# --- Test Summary ---
print_msg "33" "\n--- Test Summary ---"
if [ "${FAILED_TESTS}" -eq 0 ]; then
    print_msg "32" "All ${TEST_COUNTER} tests passed!"
    exit 0
else
    print_msg "31" "${FAILED_TESTS} of ${TEST_COUNTER} tests failed."
    exit 1
fi
