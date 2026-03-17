#!/bin/bash

# Bash completion for markdown-runner

# ============================================================================
# Configuration & Data Structures
# ============================================================================

# Flag equivalents (short to long mapping)
declare -gA FLAG_EQUIV=(
    [-d]="--dry-run"     [--dry-run]="-d"
    [-l]="--list"        [--list]="-l"
    [-i]="--interactive" [--interactive]="-i"
    [-s]="--start-from"  [--start-from]="-s"
    [-B]="--break-at"    [--break-at]="-B"
    [-t]="--timeout"     [--timeout]="-t"
    [-u]="--update-files" [--update-files]="-u"
    [-f]="--filter"      [--filter]="-f"
    [-r]="--recursive"   [--recursive]="-r"
    [-v]="--verbose"     [--verbose]="-v"
    [-q]="--quiet"       [--quiet]="-q"
    [-h]="--help"        [--help]="-h"
)

# Flags that require a value
declare -gA FLAGS_WITH_VALUES=(
    [-B]=1 [--break-at]=1
    [-s]=1 [--start-from]=1
    [-t]=1 [--timeout]=1
    [-f]=1 [--filter]=1
    [--view]=1
)

# Flag incompatibilities
# Format: "flag:incompatible1,incompatible2,..."
declare -ga FLAG_INCOMPATIBILITIES=(
    # Mutual exclusions
    "-v:-q,--quiet"
    "-q:-v,--verbose"
    "-d:-u,--update-files"
    "-u:-d,--dry-run"
    "--dry-run:-u,--update-files"
    "--update-files:-d,--dry-run"
    "-B:--ignore-breakpoints"
    "--break-at:--ignore-breakpoints"
    "--ignore-breakpoints:-B,--break-at"

    # Help excludes everything
    "-h:-d,--dry-run,-l,--list,-i,--interactive,-s,--start-from,-B,--break-at,-t,--timeout,-u,--update-files,--ignore-breakpoints,-f,--filter,-r,--recursive,--view,-v,--verbose,-q,--quiet,--no-styling"
    "--help:-d,--dry-run,-l,--list,-i,--interactive,-s,--start-from,-B,--break-at,-t,--timeout,-u,--update-files,--ignore-breakpoints,-f,--filter,-r,--recursive,--view,-v,--verbose,-q,--quiet,--no-styling"

    # List mode excludes execution flags
    "-l:-i,--interactive,-B,--break-at,-s,--start-from,-d,--dry-run,-t,--timeout,-u,--update-files"
    "--list:-i,--interactive,-B,--break-at,-s,--start-from,-d,--dry-run,-t,--timeout,-u,--update-files"

    # Interactive flags exclude list/help
    "-i:--view,-l,--list,-h,--help"
    "--interactive:--view,-l,--list,-h,--help"
    "-B:--view,-l,--list,-h,--help"
    "--break-at:--view,-l,--list,-h,--help"
    "-s:--view,-l,--list,-h,--help"
    "--start-from:--view,-l,--list,-h,--help"
)

# All available flags
ALL_FLAGS="-d --dry-run -l --list -i --interactive -s --start-from -B --break-at -t --timeout -u --update-files --ignore-breakpoints -f --filter -r --recursive --view -v --verbose -q --quiet --no-styling -h --help"

# ============================================================================
# Utility Functions
# ============================================================================

# Check if a filename has a markdown extension
# Args: $1 - filename to check
# Returns: 0 if markdown file, 1 otherwise
_is_markdown_file() {
    [[ "$1" =~ \.(md|MD|Markdown|markdown)$ ]]
}

# Check if a flag requires a value argument
# Args: $1 - flag to check (e.g., "-B", "--timeout")
# Returns: 0 if flag takes a value, 1 otherwise
_flag_takes_value() {
    [[ -n "${FLAGS_WITH_VALUES[$1]}" ]]
}

# Find markdown files in a path, respecting recursive flag
# Args: $1 - path to search (defaults to current directory if empty)
#       $2 - "true" for recursive search, anything else for non-recursive
# Output: newline-separated list of markdown file paths, sorted
_find_markdown_files() {
    local path="${1:-.}"
    local recursive="$2"

    [[ -f "$path" ]] && { echo "$path"; return; }
    [[ ! -d "$path" ]] && return

    local depth_flag="-maxdepth 1"
    [[ "$recursive" == "true" ]] && depth_flag=""

    find "$path" $depth_flag -type f \( -name "*.md" -o -name "*.MD" -o -name "*.Markdown" -o -name "*.markdown" \) 2>/dev/null | sort
}

# Extract target path and recursive flag from the command line
# Scans COMP_WORDS to find the directory/file argument and -r/--recursive flag.
# Skips over flags, flag values, and @-syntax arguments to find the target path.
# Output: pipe-delimited string "path|recursive" (e.g., ".|false" or "./docs|true")
_get_target_context() {
    local target_path=""
    local recursive=false
    local i

    # Check for recursive flag
    for ((i=1; i<${#COMP_WORDS[@]}; i++)); do
        [[ "${COMP_WORDS[i]}" == "-r" || "${COMP_WORDS[i]}" == "--recursive" ]] && recursive=true
    done

    # Find target path (skip flags and their values)
    for ((i=1; i<COMP_CWORD; i++)); do
        local word="${COMP_WORDS[i]}"

        # Skip flags
        if [[ "$word" == -* ]]; then
            _flag_takes_value "$word" && ((i++))
            continue
        fi

        # Skip @-containing arguments (flag values)
        [[ "$word" == *@* ]] && continue

        # Skip .md files (they're flag values for -B/-s)
        _is_markdown_file "$word" && continue

        # Found target path
        if [[ -d "$word" || -f "$word" ]]; then
            target_path="$word"
            break
        fi
    done

    echo "${target_path:-.}|$recursive"
}

# Get list of markdown files based on current command context
# Combines _get_target_context and _find_markdown_files to return the appropriate
# markdown files for completion based on what the user has typed so far.
# Output: newline-separated list of markdown file paths
_get_executable_files() {
    local context
    context=$(_get_target_context)
    IFS='|' read -r path recursive <<< "$context"
    _find_markdown_files "$path" "$recursive"
}

# Find the first markdown file mentioned in command line arguments
# Scans COMP_WORDS for any argument with a markdown extension.
# Useful for determining which file to extract stages/chunks from.
# Output: path to the markdown file, or empty if none found
_find_markdown_in_args() {
    local arg
    for arg in "${COMP_WORDS[@]}"; do
        _is_markdown_file "$arg" && { echo "$arg"; return; }
    done
}

# ============================================================================
# Flag Management
# ============================================================================

# Get all flags that should be excluded from completion
# Collects flags already used, their equivalents (short/long forms), and
# flags that are incompatible with currently used flags.
# Output: newline-separated list of flags to exclude, sorted and deduplicated
_get_excluded_flags() {
    local excluded=()
    local i

    # Collect used flags and their equivalents
    for ((i=1; i<COMP_CWORD; i++)); do
        local arg="${COMP_WORDS[i]}"
        [[ "$arg" != -* ]] && continue

        excluded+=("$arg")
        [[ -n "${FLAG_EQUIV[$arg]}" ]] && excluded+=("${FLAG_EQUIV[$arg]}")
    done

    # Add incompatible flags
    for ((i=1; i<COMP_CWORD; i++)); do
        local arg="${COMP_WORDS[i]}"
        [[ "$arg" != -* ]] && continue

        # Special case: --view ci
        if [[ "$arg" == "--view" && "${COMP_WORDS[i+1]}" == "ci" ]]; then
            excluded+=("-i" "--interactive" "-B" "--break-at" "-s" "--start-from")
            continue
        fi

        # Check incompatibility table
        local rule incomp_list
        for rule in "${FLAG_INCOMPATIBILITIES[@]}"; do
            IFS=':' read -r flag incomp_list <<< "$rule"
            if [[ "$arg" == "$flag" ]]; then
                IFS=',' read -ra incomp <<< "$incomp_list"
                excluded+=("${incomp[@]}")
            fi
        done
    done

    printf '%s\n' "${excluded[@]}" | sort -u
}

# Get list of flags available for completion
# Filters ALL_FLAGS to remove flags that are already used or incompatible.
# Output: space-separated list of available flags
_get_available_flags() {
    local excluded
    excluded=$(_get_excluded_flags)

    local available=()
    local flag ex is_excluded
    for flag in $ALL_FLAGS; do
        is_excluded=false
        while IFS= read -r ex; do
            [[ "$flag" == "$ex" ]] && { is_excluded=true; break; }
        done <<< "$excluded"
        [[ "$is_excluded" == false ]] && available+=("$flag")
    done

    echo "${available[*]}"
}

# ============================================================================
# Stage & Chunk Discovery
# ============================================================================

# Extract unique stage names from markdown files
# Searches for code blocks with JSON metadata containing "stage" property.
# Args: $1 - space-separated list of markdown file paths
# Output: newline-separated list of unique stage names, sorted
_get_stages() {
    local files="$1"
    [[ -z "$files" ]] && return

    echo "$files" | xargs grep -h '```[a-zA-Z0-9_. -]*{.*"stage"' 2>/dev/null | \
        sed -n 's/.*"stage"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | \
        sort -u
}

# Check if a stage name exists in the relevant markdown files
# Searches for the stage in either a specific file mentioned in args or all executable files.
# Args: $1 - stage name to validate
# Returns: 0 if stage exists, 1 otherwise
_is_valid_stage() {
    local stage="$1"
    local files target_file

    # Check for specific markdown file in args first
    target_file=$(_find_markdown_in_args)
    if [[ -n "$target_file" ]]; then
        files="$target_file"
    else
        files=$(_get_executable_files | tr '\n' ' ')
    fi

    [[ -z "$files" ]] && return 1

    echo "$files" | xargs grep -lq '```[a-zA-Z0-9_. -]*{.*"stage"[[:space:]]*:[[:space:]]*"'"$stage"'"' 2>/dev/null
}

# Backward compatibility alias for _is_valid_stage
_is_valid_stage_name() {
    _is_valid_stage "$@"
}

# Get chunk IDs or indices for a specific stage
# Extracts chunks belonging to a stage. If chunk has an "id" property, uses that;
# otherwise falls back to zero-based index.
# Args: $1 - stage name
#       $2 - space-separated list of markdown file paths
# Output: newline-separated list of chunk identifiers
_get_chunks() {
    local stage="$1"
    local files="$2"
    [[ -z "$files" ]] && return

    local chunks
    chunks=$(echo "$files" | xargs grep -h '```[a-zA-Z0-9_. -]*{.*"stage"[[:space:]]*:[[:space:]]*"'"$stage"'"' 2>/dev/null)

    local index=0 chunk_id
    while IFS= read -r line; do
        [[ -z "$line" ]] && continue
        chunk_id=$(echo "$line" | sed -n 's/.*"id"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
        echo "${chunk_id:-$index}"
        ((index++))
    done <<< "$chunks"
}

# Count the number of chunks for a specific stage
# Args: $1 - stage name
#       $2 - space-separated list of markdown file paths
# Output: total count of chunks (summed across all files)
_count_chunks() {
    local stage="$1"
    local files="$2"
    echo "$files" | xargs grep -ch "^\`\`\`[a-zA-Z0-9_. -]*{.*\"stage\"[[:space:]]*:[[:space:]]*\"$stage\"" 2>/dev/null | awk '{s+=$1} END {print s}'
}

# ============================================================================
# Completion Functions
# ============================================================================

# Complete stage names with smart nospace handling
# Provides stage name completions. For multi-chunk stages, disables automatic space
# after completion so user can type "/" to access chunk completion.
# Args: $1 - current word being completed
#       $2 - optional: space-separated list of markdown files (auto-detected if empty)
# Side effects: Sets COMPREPLY array; may disable nospace via compopt
_complete_stages() {
    local cur="$1"
    local files="$2"

    # If files not provided, get them intelligently
    if [[ -z "$files" ]]; then
        local target_file
        target_file=$(_find_markdown_in_args)
        if [[ -n "$target_file" ]]; then
            files="$target_file"
        else
            files=$(_get_executable_files | tr '\n' ' ')
        fi
    fi

    local stages
    stages=$(_get_stages "$files")
    COMPREPLY=( $(compgen -W "$stages" -- "$cur") )

    # Smart UX: If ANY of the offered stages has multiple chunks, disable auto-space
    # This allows user to press Tab again after completing a stage name to add "/" and see chunks
    # Example: User types "dep" -> Tab -> "deploy" (no space) -> Tab -> "deploy/" -> Tab -> "deploy/0"
    if [[ ${#COMPREPLY[@]} -gt 0 ]]; then
        local has_multi_chunk=false
        local stage
        for stage in "${COMPREPLY[@]}"; do
            local count
            count=$(_count_chunks "$stage" "$files")
            [[ $count -gt 1 ]] && { has_multi_chunk=true; break; }
        done
        [[ "$has_multi_chunk" == true ]] && compopt -o nospace 2>/dev/null
    fi
}

# Complete chunk identifiers for a specific stage
# Provides completions in "stage/chunk" format for all chunks in the stage.
# Args: $1 - stage name
#       $2 - partial chunk identifier being typed
#       $3 - optional: space-separated list of markdown files (auto-detected if empty)
# Side effects: Sets COMPREPLY array
_complete_chunks() {
    local stage="$1"
    local chunk_prefix="$2"
    local files="$3"

    # If files not provided, get them intelligently
    if [[ -z "$files" ]]; then
        local target_file
        target_file=$(_find_markdown_in_args)
        if [[ -n "$target_file" ]]; then
            files="$target_file"
        else
            files=$(_get_executable_files | tr '\n' ' ')
        fi
    fi

    local chunks
    chunks=$(_get_chunks "$stage" "$files")

    local completions=()
    local chunk
    while IFS= read -r chunk; do
        [[ -n "$chunk" ]] && completions+=("${stage}/${chunk}")
    done <<< "$chunks"

    COMPREPLY=( $(compgen -W "${completions[*]}" -- "${stage}/${chunk_prefix}") )
}

# Complete file@stage or file@stage/chunk syntax
# Handles completions for the "filename@stage" and "filename@stage/chunk" formats.
# When multiple files have the same basename, chooses the one with the shortest path.
# Args: $1 - current word being completed (contains @)
# Side effects: Sets COMPREPLY array; may disable nospace via compopt
_complete_file_at() {
    local cur="$1"
    local file_part="${cur%%@*}"   # Extract filename before @
    local after_at="${cur#*@}"      # Extract everything after @

    # Smart file resolution: when multiple files have same basename (e.g., tests/test.md and docs/test.md)
    # prefer the one with shortest path (closest to project root) as it's more likely to be the intended file
    local target_file files candidates=()
    files=$(_get_executable_files)
    while IFS= read -r f; do
        if [[ "$(basename "$f")" == "$file_part" ]]; then
            candidates+=("$f")
        fi
    done <<< "$files"

    # Choose file with shortest path (e.g., "test.md" over "subdir/test.md")
    if [[ ${#candidates[@]} -gt 0 ]]; then
        target_file="${candidates[0]}"
        for f in "${candidates[@]}"; do
            [[ ${#f} -lt ${#target_file} ]] && target_file="$f"
        done
    fi

    [[ -z "$target_file" ]] && return

    # Scenario 1: User already has stage and is typing chunk (e.g., "test.md@deploy/chu")
    if [[ "$after_at" == */* ]]; then
        local stage="${after_at%%/*}"      # Extract stage name
        local chunk_prefix="${after_at#*/}" # Extract partial chunk ID
        local chunks completions=()
        chunks=$(_get_chunks "$stage" "$target_file")
        while IFS= read -r chunk; do
            [[ -n "$chunk" ]] && completions+=("${file_part}@${stage}/${chunk}")
        done <<< "$chunks"
        COMPREPLY=( $(compgen -W "${completions[*]}" -- "$cur") )
    else
        # Scenario 2 & 3: User is typing stage name after @ (e.g., "test.md@dep" or "test.md@deploy")
        # Need to determine if what they've typed is a complete stage name or still partial
        local is_complete=false
        if [[ -n "$after_at" ]]; then
            grep -q '```[a-zA-Z0-9_. -]*{.*"stage"[[:space:]]*:[[:space:]]*"'"$after_at"'"' "$target_file" 2>/dev/null && is_complete=true
        fi

        if [[ "$is_complete" == true ]]; then
            # Scenario 2: Complete stage typed (e.g., "test.md@deploy")
            # Show chunks with file@stage/chunk format
            local chunks completions=()
            chunks=$(_get_chunks "$after_at" "$target_file")
            while IFS= read -r chunk; do
                [[ -n "$chunk" ]] && completions+=("${file_part}@${after_at}/${chunk}")
            done <<< "$chunks"
            COMPREPLY=( $(compgen -W "${completions[*]}" -- "$cur") )
        else
            # Scenario 3: Partial stage typed (e.g., "test.md@dep")
            # Complete stage names with file@stage format, no auto-space to allow typing more
            local stages
            stages=$(_get_stages "$target_file")
            local completions=()
            for stage in $stages; do
                completions+=("${file_part}@${stage}")
            done
            COMPREPLY=( $(compgen -W "${completions[*]}" -- "$cur") )
            compopt -o nospace 2>/dev/null  # No space so user can keep typing or add /
        fi
    fi
}

# Check if we're currently completing a value for a flag
# Returns: 0 if the previous word is a flag that takes a value, 1 otherwise
_is_completing_flag_value() {
    [[ $COMP_CWORD -gt 0 ]] && _flag_takes_value "${COMP_WORDS[COMP_CWORD-1]}"
}

# Skip past a flag and its value when iterating through COMP_WORDS
# Args: $1 - name of index variable (passed by reference)
#       $2 - flag to check
# Side effects: Increments the index variable if flag takes a value
_skip_flag_and_value() {
    local -n idx=$1
    local flag="$2"
    _flag_takes_value "$flag" && ((idx++))
}

# Determine if completion should offer file@stage syntax
# In directory context (when a directory is the target or -r is used), completions
# should offer "filename@" syntax to disambiguate files in subdirectories.
# Returns: 0 if in directory context, 1 otherwise
_is_directory_context() {
    # First check if we're completing a flag value
    if _is_completing_flag_value; then
        local prev_word="${COMP_WORDS[COMP_CWORD-1]}"
        # Special cases where flag value completion should be treated as directory context:
        if [[ "$prev_word" == "-B" || "$prev_word" == "--break-at" || "$prev_word" == "-s" || "$prev_word" == "--start-from" ]]; then
            # Case 1: Recursive mode
            if [[ " ${COMP_WORDS[@]} " =~ " -r " || " ${COMP_WORDS[@]} " =~ " --recursive " ]]; then
                return 0  # Recursive mode: treat as directory context
            fi

            # Case 2: Explicit directory argument present AND we're not completing a complete stage name
            local cur="${COMP_WORDS[COMP_CWORD]}"
            local is_complete_stage=false

            # Check if current word is a complete stage name
            if [[ -n "$cur" && "$cur" =~ ^[a-zA-Z0-9_-]+$ ]]; then
                _is_valid_stage_name "$cur" && is_complete_stage=true
            fi

            # Only apply directory context if we're not completing a complete stage name
            if [[ "$is_complete_stage" == false ]]; then
                for ((i=1; i<COMP_CWORD; i++)); do
                    local word="${COMP_WORDS[i]}"
                    # Skip flags and their values
                    if [[ "$word" == -* ]]; then
                        _skip_flag_and_value i "$word"
                        continue
                    fi
                    # Check if this is a directory argument
                    [[ -d "$word" ]] && return 0
                done
            fi
        fi
        return 1  # We're completing a flag value, not a directory
    fi

    local target_path=""

    # Find the target path from command line arguments (non-flag argument)
    for ((i=1; i<COMP_CWORD; i++)); do
        local word="${COMP_WORDS[i]}"
        # Skip flags and their values
        if [[ "$word" == -* ]]; then
            _skip_flag_and_value i "$word"
            continue
        fi
        # This should be the target path
        target_path="$word"
        break
    done

    # If no path specified, default to current directory
    [[ -z "$target_path" ]] && target_path="."

    # Check if it's a directory
    [[ -d "$target_path" ]]
}

# Complete values for -B/--break-at and -s/--start-from flags
# Handles multiple formats: stage, stage/chunk, file@stage, file@stage/chunk
# Args: $1 - current word being completed
#       $2 - true for -B/--break-at, false for -s/--start-from (currently unused)
# Side effects: Sets COMPREPLY array; may disable nospace via compopt
_complete_break_or_start() {
    local cur="$1"
    local is_break_at="$2"  # true for -B, false for -s

    # file@... format (e.g., test.md@stage or file.md@stage/chunk)
    if [[ "$cur" == *@* ]]; then
        # Delegate to _complete_file_at to handle filename@stage or filename@stage/chunk completions
        _complete_file_at "$cur"

        # Smart UX: if user just typed the @ (like "test.md@"), disable auto-space
        # This allows them to immediately see/type stage names without a trailing space breaking the syntax
        # Example: "test.md@" -> Tab -> "test.md@setup" (not "test.md@ setup")
        [[ "$cur" == *@ ]] && compopt -o nospace 2>/dev/null
        return
    fi

    # stage/chunk format (e.g., "setup/0" or "deploy/init")
    if [[ "$cur" == */* ]]; then
        # Extract stage and chunk parts using bash parameter expansion
        local stage="${cur%%/*}"      # Everything before the first /
        local chunk_prefix="${cur#*/}" # Everything after the first /
        _complete_chunks "$stage" "$chunk_prefix" ""
        return
    fi

    # Smart completion when user has typed a complete stage name
    if _is_valid_stage "$cur"; then
        # Determine which markdown file(s) to search
        local target_file files
        target_file=$(_find_markdown_in_args)
        if [[ -n "$target_file" ]]; then
            files="$target_file"  # Specific file mentioned in args
        else
            files=$(_get_executable_files | tr '\n' ' ')  # All applicable files
        fi

        local chunk_count
        chunk_count=$(_count_chunks "$cur" "$files")

        if [[ $chunk_count -gt 1 ]]; then
            # Smart UX: Multi-chunk stage gets "/" appended with nospace
            # This guides the user to continue typing the chunk ID
            # Example: "setup" -> Tab -> "setup/" (ready for chunk selection)
            COMPREPLY=("${cur}/")
            compopt -o nospace 2>/dev/null
        else
            # Single-chunk stage: show the chunk ID directly (usually "0" or named ID)
            _complete_chunks "$cur" "" ""
        fi
        return
    fi

    # Directory context: offer file@stage syntax for disambiguation
    # When -r is used or a directory is specified, multiple files may have same stages
    # so we offer "filename@" completions to let user specify which file
    if _is_directory_context; then
        local files completions=()
        files=$(_get_executable_files)
        while IFS= read -r f; do
            # Add @ suffix to each basename to indicate file@stage format is available
            [[ -n "$f" ]] && completions+=("$(basename "$f")@")
        done <<< "$files"
        COMPREPLY=( $(compgen -W "${completions[*]}" -- "$cur") )
        # Disable space after @ so user can immediately type stage name
        compopt -o nospace 2>/dev/null
        return
    fi

    # Default: show stages
    _complete_stages "$cur" ""
}

# Complete markdown files and directories
# Provides file/directory completions, with directories having trailing slashes.
# Args: $1 - current word being completed
# Side effects: Sets COMPREPLY array; disables nospace for directories
_complete_files() {
    local cur="$1"

    local files dirs md_files=() completions=()
    files=($(compgen -f -- "$cur"))
    dirs=($(compgen -d -- "$cur"))

    # Filter to .md files
    for file in "${files[@]}"; do
        _is_markdown_file "$file" && md_files+=("$file")
    done

    # Add directories with trailing slash
    for dir in "${dirs[@]}"; do
        completions+=("${dir}/")
    done

    completions+=("${md_files[@]}")
    COMPREPLY=("${completions[@]}")

    # Disable space for directories
    [[ ${#dirs[@]} -gt 0 ]] && compopt -o nospace 2>/dev/null
}

# ============================================================================
# Main Completion Handler
# ============================================================================

# Main entry point for markdown-runner bash completion
# Called by bash whenever user presses Tab after typing "markdown-runner".
# Orchestrates all completion logic: flag values, flags, help, and file/directory completions.
# Uses bash built-in variables:
#   - COMP_WORDS: array of all words on the command line
#   - COMP_CWORD: index of the word being completed
#   - COMPREPLY: array where completion results are stored
_markdown_runner_completion() {
    local cur prev
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # Handle flag value completion
    case "$prev" in
        -B|--break-at)
            _complete_break_or_start "$cur" true
            [[ "$cur" == *@ ]] && compopt -o nospace 2>/dev/null
            return
            ;;
        -s|--start-from)
            _complete_break_or_start "$cur" false
            [[ "$cur" == *@ ]] && compopt -o nospace 2>/dev/null
            return
            ;;
        -t|--timeout)
            COMPREPLY=( $(compgen -W "1 5 10 30 60" -- "$cur") )
            return
            ;;
        --view)
            COMPREPLY=( $(compgen -W "default ci" -- "$cur") )
            return
            ;;
        -f|--filter)
            return  # No completion for regex
            ;;
    esac

    # Handle flag completion
    if [[ "$cur" == -* ]]; then
        # Enhanced help for bare "-"
        if [[ "$cur" == "-" ]]; then
            local available
            available=$(_get_available_flags)
            # Create combined short/long format with descriptions
            local items=()
            while IFS='|' read -r short long desc; do
                [[ -z "$short" && -z "$long" ]] && continue
                # Check if both forms are available
                if [[ -n "$short" ]] && [[ -n "$long" ]]; then
                    if [[ " $available " == *" $short "* ]] && [[ " $available " == *" $long "* ]]; then
                        items+=("$short, $long:$desc")
                    elif [[ " $available " == *" $short "* ]]; then
                        items+=("$short:$desc")
                    elif [[ " $available " == *" $long "* ]]; then
                        items+=("$long:$desc")
                    fi
                elif [[ -n "$long" ]] && [[ " $available " == *" $long "* ]]; then
                    items+=("$long:$desc")
                fi
            done <<EOF
-d|--dry-run|(dry-run)
-l|--list|(list files)
-i|--interactive|(interactive)
-s|--start-from|(start from stage)
-B|--break-at|(break at stage)
-t|--timeout|(timeout)
-u|--update-files|(update files)
|--ignore-breakpoints|(ignore breakpoints)
-f|--filter|(filter)
-r|--recursive|(recursive)
|--view|(view mode)
-v|--verbose|(verbose)
-q|--quiet|(quiet)
|--no-styling|(no styling)
-h|--help|(help)
EOF
            COMPREPLY=("${items[@]}")
            compopt -o nospace 2>/dev/null
        else
            # Standard flag completion
            local available
            available=$(_get_available_flags)
            COMPREPLY=( $(compgen -W "$available" -- "$cur") )
        fi
        return
    fi

    # Handle help command - show full categorized help
    if [[ "$cur" == "help" || "$cur" == "hel" || "$cur" == "he" ]]; then
        COMPREPLY=(
            "Modes:"
            "  -d, --dry-run              Just list what would be executed without doing it"
            "  -l, --list                 Just list the files found"
            ""
            "Execution Control:"
            "  -i, --interactive          Prompt to press enter between each chunk"
            "  -s, --start-from string    Start from a specific stage (stage or file@stage)"
            "  -B, --break-at string      Start debugging from a specific stage or chunk"
            "  -t, --timeout int          The timeout in minutes for every executed command"
            "  -u, --update-files         Update the chunk output section in the markdown files"
            "      --ignore-breakpoints   Ignore the breakpoints"
            ""
            "File Selection:"
            "  -f, --filter string        Run only the files matching the regex"
            "  -r, --recursive            Search for markdown files recursively"
            ""
            "Output & Logging:"
            "      --view string          UI to be used, can be 'default' or 'ci'"
            "  -v, --verbose              Print more logs"
            "  -q, --quiet                Disable output"
            "      --no-styling           Disable spinners in CLI"
            ""
            "Help:"
            "  -h, --help                 Show this help message"
        )
        compopt -o nospace 2>/dev/null
        return
    fi

    # Check if file/dir already provided
    local has_file=false
    local i
    for ((i=1; i<COMP_CWORD; i++)); do
        local arg="${COMP_WORDS[i]}"
        [[ "$arg" == -* ]] && { _flag_takes_value "$arg" && ((i++)); continue; }
        _is_markdown_file "$arg" || [[ -d "$arg" ]] && { has_file=true; break; }
    done

    if [[ "$has_file" == true ]]; then
        # File provided, show flags
        local available
        available=$(_get_available_flags)
        COMPREPLY=( $(compgen -W "$available" -- "$cur") )
    else
        # No file, complete files/directories
        _complete_files "$cur"
    fi
}

# Register completion
complete -F _markdown_runner_completion markdown-runner
