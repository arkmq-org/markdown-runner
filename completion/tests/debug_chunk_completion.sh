#!/bin/bash

# Debug script for chunk completion
# This script helps debug issues with chunk discovery in completion

source "$(dirname "$0")/../bash_completion.sh"

echo "=== Debug: Chunk completion ==="
echo

# Test chunk completion for setup stage
echo "=== Testing chunk completion for setup stage ==="
stage="setup"
target_file="README.md"

if [[ -f "$target_file" ]]; then
    echo "File exists: $target_file"
    
    echo
    echo "=== Raw grep output for stage '$stage' ==="
    grep -h '```[a-zA-Z0-9_. -]*{.*"stage"[[:space:]]*:[[:space:]]*"'"$stage"'"' "$target_file" 2>/dev/null
    
    echo
    echo "=== Processing chunks ==="
    all_chunks=$(grep -h '```[a-zA-Z0-9_. -]*{.*"stage"[[:space:]]*:[[:space:]]*"'"$stage"'"' "$target_file" 2>/dev/null)
    
    index=0
    chunk_list=""
    while IFS= read -r line; do
        if [[ -n "$line" ]]; then
            echo "Processing line: $line"
            
            # Extract ID if present
            chunk_id=$(echo "$line" | sed -n 's/.*"id"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
            if [[ -n "$chunk_id" ]]; then
                chunk_list="$chunk_list $chunk_id"
                echo "  Found chunk ID: $chunk_id"
            else
                chunk_list="$chunk_list $index"
                echo "  Using index: $index"
            fi
            ((index++))
        fi
    done <<< "$all_chunks"
    
    echo
    echo "=== Final chunk list ==="
    echo "Chunks: $chunk_list"
    
    echo
    echo "=== Building completions ==="
    completions=()
    for chunk in $chunk_list; do
        completions+=("${stage}/${chunk}")
        echo "Added completion: ${stage}/${chunk}"
    done
    
    echo
    echo "=== Final completions ==="
    printf '%s\n' "${completions[@]}"
    
else
    echo "$target_file not found!"
fi

echo
echo "=== Testing file@stage/chunk completion ==="
file_part="README.md"
stage="setup"
chunk_prefix=""

echo "Testing: ${file_part}@${stage}/"

# Find the actual file path
target_file=""
executable_files=$(_get_executable_files)
while IFS= read -r file; do
    if [[ -n "$file" ]] && [[ "$(basename "$file")" == "$file_part" ]]; then
        target_file="$file"
        break
    fi
done <<< "$executable_files"

if [[ -n "$target_file" ]]; then
    echo "Found target file: $target_file"
    
    # Extract chunks for the specific stage from the specific file
    all_chunks=$(grep -h '```[a-zA-Z0-9_. -]*{.*"stage"[[:space:]]*:[[:space:]]*"'"$stage"'"' "$target_file" 2>/dev/null)
    
    # Process chunks and collect IDs and indices
    index=0
    chunk_list=""
    while IFS= read -r line; do
        if [[ -n "$line" ]]; then
            # Extract ID if present
            chunk_id=$(echo "$line" | sed -n 's/.*"id"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
            if [[ -n "$chunk_id" ]]; then
                chunk_list="$chunk_list $chunk_id"
            else
                chunk_list="$chunk_list $index"
            fi
            ((index++))
        fi
    done <<< "$all_chunks"
    
    # Format completions as file@stage/chunk
    completions=()
    for chunk in $chunk_list; do
        completions+=("${file_part}@${stage}/${chunk}")
    done
    
    echo "File@stage/chunk completions:"
    printf '%s\n' "${completions[@]}"
fi
