#!/bin/bash

# Debug script for stage extraction from markdown files
# This script helps debug issues with stage discovery in completion

source "$(dirname "$0")/../bash_completion.sh"

echo "=== Debug: Stage extraction from markdown files ==="
echo

# Test stage extraction from README.md
echo "=== Testing stage extraction from README.md ==="
if [[ -f "README.md" ]]; then
    echo "File exists: README.md"
    
    echo
    echo "=== Raw grep output ==="
    grep -h '```[a-zA-Z0-9_. -]*{.*"stage"' "README.md" 2>/dev/null | head -10
    
    echo
    echo "=== Extracted stages ==="
    stages=$(grep -h '```[a-zA-Z0-9_. -]*{.*"stage"' "README.md" 2>/dev/null | \
             sed -n 's/.*"stage"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | \
             sort -u)
    echo "$stages"
    
    echo
    echo "=== Stage count ==="
    stage_count=$(echo "$stages" | wc -l)
    echo "Found $stage_count stages"
    
else
    echo "README.md not found!"
fi

echo
echo "=== Testing file@stage completion logic ==="
file_part="README.md"
stage_prefix=""

# Find the actual file path
target_file=""
executable_files=$(_get_executable_files)
echo "Executable files: $executable_files"

while IFS= read -r file; do
    echo "Checking file: '$file' vs basename: '$(basename "$file")'"
    if [[ -n "$file" ]] && [[ "$(basename "$file")" == "$file_part" ]]; then
        target_file="$file"
        echo "Found target file: '$target_file'"
        break
    fi
done <<< "$executable_files"

if [[ -n "$target_file" ]]; then
    echo
    echo "=== Building completions for $target_file ==="
    stages=$(grep -h '```[a-zA-Z0-9_. -]*{.*"stage"' "$target_file" 2>/dev/null | \
             sed -n 's/.*"stage"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | \
             sort -u)
    
    completions=()
    for stage in $stages; do
        completions+=("${file_part}@${stage}")
        echo "Added completion: ${file_part}@${stage}"
    done
    
    echo
    echo "=== Final completions ==="
    printf '%s\n' "${completions[@]}"
else
    echo "No target file found!"
fi
