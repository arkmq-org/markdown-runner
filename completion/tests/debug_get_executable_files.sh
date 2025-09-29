#!/bin/bash

# Debug script for _get_executable_files function
# This script helps debug issues with file discovery in completion

source "$(dirname "$0")/../bash_completion.sh"

echo "=== Debug: _get_executable_files function ==="
echo

# Test case 1: Normal completion scenario
echo "Test 1: Normal completion with -B README.md@"
COMP_WORDS=("markdown-runner" "-B" "README.md@")
COMP_CWORD=2

echo "COMP_WORDS: ${COMP_WORDS[@]}"

echo "=== Checking arguments in COMP_WORDS (with @ filter) ==="
for arg in "${COMP_WORDS[@]}"; do
    echo "Arg: '$arg'"
    if [[ "$arg" != -* ]] && [[ "$arg" != "markdown-runner" ]] && [[ "$arg" != *.md ]] && [[ "$arg" != *@* ]]; then
        echo "  -> Would be selected as target_path"
    else
        echo "  -> Skipped (is flag, is markdown-runner, is *.md, or contains @)"
    fi
done

echo
echo "=== Calling _get_executable_files ==="
files=$(_get_executable_files)
echo "Files found: '$files'"

echo
echo "=== Manual test of _find_markdown_files ==="
manual_files=$(_find_markdown_files "." "false")
echo "Manual files: '$manual_files'"

echo
echo "Test 2: With directory argument"
COMP_WORDS=("markdown-runner" "-B" "test/" "README.md@")
echo "COMP_WORDS: ${COMP_WORDS[@]}"
files2=$(_get_executable_files)
echo "Files found: '$files2'"

echo
echo "Test 3: With recursive flag"
COMP_WORDS=("markdown-runner" "-r" "-B" "README.md@")
echo "COMP_WORDS: ${COMP_WORDS[@]}"
files3=$(_get_executable_files)
echo "Files found: '$files3'"
