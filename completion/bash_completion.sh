#!/bin/bash

# Bash completion for markdown-runner
# Source this file or add it to your bash completion directory

# Flag mapping configuration - central definition of short/long flag pairs
_get_flag_equivalent() {
    local flag="$1"
    case "$flag" in
        -d) echo "--dry-run" ;;
        --dry-run) echo "-d" ;;
        -l) echo "--list" ;;
        --list) echo "-l" ;;
        -i) echo "--interactive" ;;
        --interactive) echo "-i" ;;
        -s) echo "--start-from" ;;
        --start-from) echo "-s" ;;
        -B) echo "--break-at" ;;
        --break-at) echo "-B" ;;
        -t) echo "--timeout" ;;
        --timeout) echo "-t" ;;
        -u) echo "--update-files" ;;
        --update-files) echo "-u" ;;
        -f) echo "--filter" ;;
        --filter) echo "-f" ;;
        -r) echo "--recursive" ;;
        --recursive) echo "-r" ;;
        -v) echo "--verbose" ;;
        --verbose) echo "-v" ;;
        -q) echo "--quiet" ;;
        --quiet) echo "-q" ;;
        -h) echo "--help" ;;
        --help) echo "-h" ;;
        *) echo "" ;; # No equivalent
    esac
}

# Collect flags that are already used on the command line
# Returns array of used flags including both short and long forms
_collect_used_flags() {
    local used_flags=()

    for ((i=1; i<COMP_CWORD; i++)); do
        local arg="${COMP_WORDS[i]}"
        if [[ "$arg" == -* ]]; then
            used_flags+=("$arg")
            # Also add the equivalent short/long form
            local equivalent
            equivalent=$(_get_flag_equivalent "$arg")
            if [[ -n "$equivalent" ]]; then
                used_flags+=("$equivalent")
            fi
        fi
    done

    printf '%s\n' "${used_flags[@]}"
}

# Collect flags that are incompatible with already used flags
# Returns array of flags that should be excluded due to incompatibility
_collect_incompatible_flags() {
    local incompatible_flags=()

    # Check for specific incompatibility patterns
    for ((i=1; i<COMP_CWORD; i++)); do
        local arg="${COMP_WORDS[i]}"
        local next_arg=""
        if [[ $((i+1)) -lt COMP_CWORD ]]; then
            next_arg="${COMP_WORDS[$((i+1))]}"
        fi

        case "$arg" in
            --view)
                if [[ "$next_arg" == "ci" ]]; then
                    # CI mode is incompatible with interactive features
                    incompatible_flags+=("-i" "--interactive" "-B" "--break-at" "-s" "--start-from")
                fi
                ;;
            -l|--list)
                # List mode is incompatible with execution-related flags
                incompatible_flags+=("-i" "--interactive" "-B" "--break-at" "-s" "--start-from" 
                                   "-d" "--dry-run" "-t" "--timeout" "-u" "--update-files")
                ;;
            -h|--help)
                # Help mode is incompatible with all other flags
                incompatible_flags+=("-d" "--dry-run" "-l" "--list" "-i" "--interactive" 
                                   "-s" "--start-from" "-B" "--break-at" "-t" "--timeout" 
                                   "-u" "--update-files" "--ignore-breakpoints" "-f" "--filter" 
                                   "-r" "--recursive" "--view" "-v" "--verbose" "-q" "--quiet" "--no-styling")
                ;;
            -d|--dry-run)
                # Dry run is incompatible with file modification
                incompatible_flags+=("-u" "--update-files")
                ;;
            -v|--verbose)
                # Verbose is incompatible with quiet
                incompatible_flags+=("-q" "--quiet")
                ;;
            -q|--quiet)
                # Quiet is incompatible with verbose
                incompatible_flags+=("-v" "--verbose")
                ;;
            -i|--interactive)
                # Interactive mode is incompatible with CI mode and list mode
                incompatible_flags+=("--view" "-l" "--list" "-h" "--help")
                ;;
            -B|--break-at)
                # Break-at (interactive debugging) is incompatible with CI mode, list mode, and ignore-breakpoints
                incompatible_flags+=("--view" "-l" "--list" "-h" "--help" "--ignore-breakpoints")
                ;;
            -s|--start-from)
                # Start-from (interactive debugging) is incompatible with CI mode and list mode
                incompatible_flags+=("--view" "-l" "--list" "-h" "--help")
                ;;
            -u|--update-files)
                # Update-files is incompatible with dry-run
                incompatible_flags+=("-d" "--dry-run")
                ;;
            --ignore-breakpoints)
                # Ignore-breakpoints is incompatible with break-at
                incompatible_flags+=("-B" "--break-at")
                ;;
        esac
    done

    printf '%s\n' "${incompatible_flags[@]}"
}

# Filter available flags by excluding used and incompatible flags
# Args: $1 - space-separated list of all possible flags
#       $2 - newline-separated list of used flags  
#       $3 - newline-separated list of incompatible flags
# Returns: space-separated list of available flags
_filter_available_flags() {
    local opts="$1"
    local used_flags="$2"
    local incompatible_flags="$3"
    local available_opts=""

    # Convert newline-separated lists to arrays for easier searching
    local used_array=()
    local incompatible_array=()

    while IFS= read -r flag; do
        [[ -n "$flag" ]] && used_array+=("$flag")
    done <<< "$used_flags"

    while IFS= read -r flag; do
        [[ -n "$flag" ]] && incompatible_array+=("$flag")
    done <<< "$incompatible_flags"

    # Check each option for exclusion
    for opt in $opts; do
        local is_excluded=false

        # Check if flag is already used
        for used in "${used_array[@]}"; do
            if [[ "$opt" == "$used" ]]; then
                is_excluded=true
                break
            fi
        done

        # Check if flag is incompatible
        if [[ "$is_excluded" == false ]]; then
            for incompatible in "${incompatible_array[@]}"; do
                if [[ "$opt" == "$incompatible" ]]; then
                    is_excluded=true
                    break
                fi
            done
        fi

        if [[ "$is_excluded" == false ]]; then
            available_opts="$available_opts $opt"
        fi
    done

    echo "$available_opts"
}

# Main function to get available flags (excluding already used and incompatible ones)
# Args: $1 - space-separated list of all possible flags
# Returns: space-separated list of available flags
_get_available_flags() {
    local opts="$1"

    # Collect exclusions
    local used_flags
    local incompatible_flags
    used_flags=$(_collect_used_flags)
    incompatible_flags=$(_collect_incompatible_flags)

    # Filter and return available flags
    _filter_available_flags "$opts" "$used_flags" "$incompatible_flags"
}

# Utility function to check if a flag takes a value
# Args: $1 - flag to check
# Returns: 0 if flag takes a value, 1 otherwise
_flag_takes_value() {
    local flag="$1"
    case "$flag" in
        -B|--break-at|-s|--start-from|-f|--filter|-t|--timeout|--view)
            return 0 ;;
        *)
            return 1 ;;
    esac
}

# Utility function to skip a flag and its value in command line parsing
# Args: $1 - current index (by reference)
#       $2 - flag at current index
# Modifies: increments index if flag takes a value
_skip_flag_and_value() {
    local -n index_ref=$1
    local flag="$2"

    if _flag_takes_value "$flag"; then
        ((index_ref++)) # Skip the flag's value
    fi
}

# Utility function to check if a file has a markdown extension
# Args: $1 - filename to check
# Returns: 0 if it's a markdown file, 1 otherwise
_is_markdown_file() {
    local file="$1"
    [[ "$file" == *.md ]] || [[ "$file" == *.MD ]] || [[ "$file" == *.Markdown ]] || [[ "$file" == *.markdown ]]
}

# Utility function to find a markdown file in command line arguments
# Returns: the first markdown file found, or empty string if none
_find_markdown_file_in_args() {
    for arg in "${COMP_WORDS[@]}"; do
        if _is_markdown_file "$arg"; then
            echo "$arg"
            return 0
        fi
    done
    echo ""
}

# Utility function to check if we're currently completing a flag value
# Returns: 0 if completing a flag value, 1 otherwise
_is_completing_flag_value() {
    if [[ $COMP_CWORD -gt 0 ]]; then
        local prev_word="${COMP_WORDS[COMP_CWORD-1]}"
        _flag_takes_value "$prev_word"
    else
        return 1
    fi
}

# Utility function to apply nospace for file@ completions
# Checks if any completion ends with '@' and applies compopt -o nospace if so
_apply_nospace_for_file_at() {
    local has_file_at=false
    for completion in "${COMPREPLY[@]}"; do
        if [[ "$completion" == *@ ]]; then
            has_file_at=true
            break
        fi
    done

    if [[ "$has_file_at" == true ]]; then
        compopt -o nospace 2>/dev/null || true
    fi
}

# Function to show enhanced completion with short descriptions
_markdown_runner_show_enhanced_help() {
    # Create completions with combined short/long format to avoid duplicates
    local help_items=(
        # Modes
        "-d, --dry-run:(dry-run)"
        "-l, --list:(list files)"

        # Execution Control  
        "-i, --interactive:(interactive)"
        "-s, --start-from:(start from stage)"
        "-B, --break-at:(break at stage)"
        "-t, --timeout:(timeout)"
        "-u, --update-files:(update files)"
        "--ignore-breakpoints:(ignore breakpoints)"

        # File Selection
        "-f, --filter:(filter)"
        "-r, --recursive:(recursive)"

        # Output & Logging
        "--view:(view mode)"
        "-v, --verbose:(verbose)"
        "-q, --quiet:(quiet)"
        "--no-styling:(no styling)"

        # Help
        "-h, --help:(help)"
    )

    # Filter out incompatible flags
    local available_opts
    available_opts=$(_get_available_flags "$opts")

    # Build filtered help items
    local filtered_help=()
    for item in "${help_items[@]}"; do
        local flags_part="${item%%:*}"
        local desc_part="${item#*:}"

        # Check if this item should be included by checking both short and long flags
        local should_include=false

        # Handle combined format like "-d, --dry-run" or single flags like "--ignore-breakpoints"
        if [[ "$flags_part" == *", "* ]]; then
            # Combined format: check both flags
            local short_flag="${flags_part%%, *}"
            local long_flag="${flags_part##*, }"
            if [[ " $available_opts " =~ " $short_flag " ]] || [[ " $available_opts " =~ " $long_flag " ]]; then
                should_include=true
            fi
        else
            # Single flag: check just this flag
            if [[ " $available_opts " =~ " $flags_part " ]]; then
                should_include=true
            fi
        fi

        if [[ "$should_include" == true ]]; then
            filtered_help+=("$item")
        fi
    done

    # Set completions to the filtered help items
    COMPREPLY=("${filtered_help[@]}")

    # Use compopt to prevent adding space after completion
    compopt -o nospace 2>/dev/null || true
}

# Function to show full categorized help (like --help output)
_markdown_runner_show_full_help() {
    # Show the same categorized format as --help but as completions
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

    # Use compopt to prevent adding space after completion
    compopt -o nospace 2>/dev/null || true
}

# Handle flag value completions
# Args: $1 - previous word (flag), $2 - current word
# Returns: 0 if handled, 1 if not a flag value
_handle_flag_value_completion() {
    local prev="$1"
    local cur="$2"

    case "$prev" in
        -B|--break-at)
            _markdown_runner_break_at_completion "$cur"
            _apply_nospace_for_file_at
            return 0
            ;;
        -s|--start-from)
            _markdown_runner_start_from_completion "$cur"
            _apply_nospace_for_file_at
            return 0
            ;;
        -t|--timeout)
            # Numeric completion for timeout
            COMPREPLY=( $(compgen -W "1 5 10 30 60" -- "$cur") )
            return 0
            ;;
        --view)
            COMPREPLY=( $(compgen -W "default ci" -- "$cur") )
            return 0
            ;;
        -f|--filter)
            # No specific completion for regex patterns
            return 0
            ;;
    esac

    return 1  # Not a flag value
}

# Handle flag completions (when current word starts with -)
# Args: $1 - current word, $2 - available options
# Returns: 0 always (always handles flag completion)
_handle_flag_completion() {
    local cur="$1"
    local opts="$2"

    # Check if user wants enhanced help (just "-")
    if [[ "$cur" == "-" ]]; then
        # Show enhanced help format for bare "-" 
        _markdown_runner_show_enhanced_help
    else
        # Standard flag completion
        local available_opts
        available_opts=$(_get_available_flags "$opts")
        COMPREPLY=( $(compgen -W "${available_opts}" -- "$cur") )
    fi

    return 0
}

# Handle help command completion
# Args: $1 - current word
# Returns: 0 if handled, 1 if not help
_handle_help_completion() {
    local cur="$1"

    # Check for special help command
    if [[ "$cur" == "help" ]] || [[ "$cur" == "hel" ]] || [[ "$cur" == "he" ]]; then
        # Show full categorized help
        _markdown_runner_show_full_help
        return 0
    fi

    return 1  # Not help
}

# Check if a file or directory is already provided as positional argument
# Returns: 0 if file/dir provided, 1 otherwise
_has_file_or_dir_argument() {
    for ((i=1; i<COMP_CWORD; i++)); do
        local arg="${COMP_WORDS[i]}"
        # Skip flags and their values
        if [[ "$arg" == -* ]]; then
            _skip_flag_and_value i "$arg"
            continue
        fi
        # Check if this argument is a markdown file or directory
        if _is_markdown_file "$arg" || [[ -d "$arg" ]]; then
            return 0
        fi
    done
    return 1
}

# Handle positional argument completion (files and directories)
# Args: $1 - current word, $2 - available options
# Returns: 0 always
_handle_positional_completion() {
    local cur="$1"
    local opts="$2"

    if _has_file_or_dir_argument; then
        # Markdown file or directory already provided, show flags
        local available_opts
        available_opts=$(_get_available_flags "$opts")
        COMPREPLY=( $(compgen -W "${available_opts}" -- "$cur") )
    else
        # No markdown file or directory yet, complete with .md files and directories
        _markdown_runner_file_completion "$cur"

        # Check if any completion ends with '/' (directory)
        # If so, disable space suffix to allow chaining
        local has_dirs=false
        for completion in "${COMPREPLY[@]}"; do
            if [[ "$completion" == */ ]]; then
                has_dirs=true
                break
            fi
        done

        if [[ "$has_dirs" == true ]]; then
            compopt -o nospace 2>/dev/null || true
        fi
    fi

    return 0
}

_markdown_runner_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # Basic options
    opts="-d --dry-run -l --list -i --interactive -s --start-from -B --break-at -t --timeout -u --update-files --ignore-breakpoints -f --filter -r --recursive --view -v --verbose -q --quiet --no-styling -h --help"

    # Try flag value completion first
    if _handle_flag_value_completion "$prev" "$cur"; then
        return 0
    fi

    # Handle flag completion (current word starts with -)
    if [[ "$cur" == -* ]]; then
        _handle_flag_completion "$cur" "$opts"
        return 0
    fi

    # Handle help command completion
    if _handle_help_completion "$cur"; then
        return 0
    fi

    # Handle positional argument completion (files and directories)
    _handle_positional_completion "$cur" "$opts"
}

# File completion for markdown files
_markdown_runner_file_completion() {
    local cur="$1"

    # Use bash's built-in file completion but filter for .md files and directories
    # This handles directory traversal automatically
    local files=($(compgen -f -- "${cur}"))
    local dirs=($(compgen -d -- "${cur}"))

    # Filter files to only include .md files
    local md_files=()
    for file in "${files[@]}"; do
        if _is_markdown_file "$file"; then
            md_files+=("$file")
        fi
    done

    # Combine directories (with trailing slash) and .md files
    local completions=()
    for dir in "${dirs[@]}"; do
        completions+=("${dir}/")
    done
    for file in "${md_files[@]}"; do
        completions+=("$file")
    done

    COMPREPLY=("${completions[@]}")
}

# Find markdown files like markdown-runner does
_find_markdown_files() {
    local path="$1"
    local recursive="$2"
    local files=()

    if [[ -f "$path" ]]; then
        echo "$path"
        return
    fi

    if [[ ! -d "$path" ]]; then
        return
    fi

    # Find files in directory
    if [[ "$recursive" == "true" ]]; then
        # Recursive search
        files=($(find "$path" -name "*.md" -o -name "*.MD" -o -name "*.Markdown" -o -name "*.markdown" 2>/dev/null | sort))
    else
        # Non-recursive search
        files=($(find "$path" -maxdepth 1 -name "*.md" -o -name "*.MD" -o -name "*.Markdown" -o -name "*.markdown" 2>/dev/null | sort))
    fi

    printf '%s\n' "${files[@]}"
}

# Get files that would be executed by markdown-runner
_get_executable_files() {
    local target_path=""
    local recursive=false

    # Check if a specific path is provided in the command line
    # But skip arguments that are flag values
    for ((i=1; i<${#COMP_WORDS[@]}; i++)); do
        local arg="${COMP_WORDS[i]}"

        # Skip flags
        if [[ "$arg" == -* ]]; then
            _skip_flag_and_value i "$arg"
            continue
        fi

        # Skip arguments that contain @ (these are flag values like README.md@stage)
        if [[ "$arg" == *@* ]]; then
            continue
        fi

        # Skip .md files (these are usually flag values)
        if _is_markdown_file "$arg"; then
            continue
        fi

        # This should be the target path - only use if it's actually a valid directory or file
        if [[ -d "$arg" ]] || [[ -f "$arg" ]]; then
            target_path="$arg"
            break
        fi
    done

    # Check for recursive flag
    if [[ " ${COMP_WORDS[@]} " =~ " -r " ]] || [[ " ${COMP_WORDS[@]} " =~ " --recursive " ]]; then
        recursive=true
    fi

    # Default to current directory if no path specified
    if [[ -z "$target_path" ]]; then
        target_path="."
    fi

    _find_markdown_files "$target_path" "$recursive"
}

# Helper function to check if we're completing for a directory context
_is_directory_context() {
    # First check if we're completing a flag value
    if _is_completing_flag_value; then
        local prev_word="${COMP_WORDS[COMP_CWORD-1]}"
        # Special cases where flag value completion should be treated as directory context:
        if [[ "$prev_word" == "-B" ]] || [[ "$prev_word" == "--break-at" ]] || 
           [[ "$prev_word" == "-s" ]] || [[ "$prev_word" == "--start-from" ]]; then
                # Case 1: Recursive mode
                if [[ " ${COMP_WORDS[@]} " =~ " -r " ]] || [[ " ${COMP_WORDS[@]} " =~ " --recursive " ]]; then
                    return 0  # Recursive mode: treat as directory context
                fi

                # Case 2: Explicit directory argument present AND we're not completing a complete stage name
                local cur="${COMP_WORDS[COMP_CWORD]}"
                local is_complete_stage=false

                # Check if current word is a complete stage name
                if [[ -n "$cur" ]]; then
                    # Quick check if it looks like a complete stage name (no special chars, reasonable length)
                    if [[ "$cur" =~ ^[a-zA-Z0-9_-]+$ ]] && _is_valid_stage_name "$cur"; then
                        is_complete_stage=true
                    fi
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
                        if [[ -d "$word" ]]; then
                            return 0  # Directory argument present + not complete stage: treat as directory context
                        fi
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
    if [[ -z "$target_path" ]]; then
        target_path="."
    fi

    # Check if it's a directory
    [[ -d "$target_path" ]]
}

# Completion for --start-from flag
_markdown_runner_start_from_completion() {
    local cur="$1"

    # Handle file@stage format
    if [[ "$cur" == *@* ]]; then
        local file_part="${cur%@*}"
        local stage_part="${cur#*@}"

        # file@stage format - complete stages from that specific file
        _markdown_runner_file_at_stage_completion "$file_part" "$stage_part"
    else
        if _is_directory_context; then
            # For directories, only show file@ completions (no individual stage names)
            local executable_files
            executable_files=$(_get_executable_files)
            local file_completions=()
            while IFS= read -r file; do
                if [[ -n "$file" ]]; then
                    local basename=$(basename "$file")
                    file_completions+=("${basename}@")
                fi
            done <<< "$executable_files"

            COMPREPLY=( $(compgen -W "${file_completions[*]}" -- "${cur}") )
        else
            # Check if a specific markdown file is already provided in the command line
            local specific_file_provided=false
            for arg in "${COMP_WORDS[@]}"; do
                if [[ "$arg" == *.md ]] || [[ "$arg" == *.MD ]] || [[ "$arg" == *.Markdown ]] || [[ "$arg" == *.markdown ]]; then
                    specific_file_provided=true
                    break
                fi
            done

            if [[ "$specific_file_provided" == true ]]; then
                # Specific file provided, only show stage names from that file
                _markdown_runner_stage_completion "$cur"
            else
                # No specific file, show only stage names (no file@ completions in non-directory context)
                _markdown_runner_stage_completion "$cur"
            fi
        fi
    fi
}

# Completion for --break-at flag
_markdown_runner_break_at_completion() {
    local cur="$1"

    # Handle file@stage/chunk format
    if [[ "$cur" == *@* ]]; then
        local file_part="${cur%@*}"
        local debug_part="${cur#*@}"

        if [[ "$debug_part" == */* ]]; then
            # file@stage/chunk format
            local stage="${debug_part%/*}"
            local chunk_prefix="${debug_part##*/}"
            _markdown_runner_file_at_chunk_completion "$file_part" "$stage" "$chunk_prefix"
        else
            # file@stage format - check if stage is complete or needs completion
            # First check if this is a complete stage name that should show chunks
            local is_complete_stage=false

            # Check if the stage exists in the specified file
            local target_file=""
            local executable_files
            executable_files=$(_get_executable_files)
            local candidate_files=()
            while IFS= read -r file; do
                if [[ -n "$file" ]] && [[ "$(basename "$file")" == "$file_part" ]]; then
                    candidate_files+=("$file")
                fi
            done <<< "$executable_files"

            # If multiple files found, prefer the one with the shortest path (closest to root)
            if [[ ${#candidate_files[@]} -gt 0 ]]; then
                target_file="${candidate_files[0]}"
                for file in "${candidate_files[@]}"; do
                    if [[ ${#file} -lt ${#target_file} ]]; then
                        target_file="$file"
                    fi
                done
            fi

            if [[ -n "$target_file" ]]; then
                # Check if the debug_part is a complete stage name
                local stage_exists=$(grep -h '```[a-zA-Z0-9_. -]*{.*"stage"[[:space:]]*:[[:space:]]*"'"$debug_part"'"' "$target_file" 2>/dev/null)
                if [[ -n "$stage_exists" ]]; then
                    is_complete_stage=true
                fi
            fi

            if [[ "$is_complete_stage" == true ]]; then
                # Complete stage name, show chunks for this file@stage
                _markdown_runner_file_at_chunk_completion "$file_part" "$debug_part" ""
            else
                # Incomplete stage name, complete stages from that specific file
                _markdown_runner_file_at_stage_completion "$file_part" "$debug_part"
            fi
        fi
    elif [[ "$cur" == */* ]]; then
        # stage/chunk format (no file specified)
        local stage="${cur%/*}"
        local chunk_prefix="${cur##*/}"
        _markdown_runner_chunk_completion "$stage" "$chunk_prefix"
    else
        # Check if the current word is a valid stage name
        if _is_valid_stage_name "$cur"; then
            # Current word is a stage name, show chunks for this stage
            _markdown_runner_chunk_completion "$cur" ""
        elif _is_directory_context; then
            # For directories, only show file@ completions (no individual stage names)
            local executable_files
            executable_files=$(_get_executable_files)
            local file_completions=()
            while IFS= read -r file; do
                if [[ -n "$file" ]]; then
                    local basename=$(basename "$file")
                    file_completions+=("${basename}@")
                fi
            done <<< "$executable_files"

            COMPREPLY=( $(compgen -W "${file_completions[*]}" -- "${cur}") )
        else
            # Check if a specific markdown file is already provided in the command line
            local specific_file_provided=false
            for arg in "${COMP_WORDS[@]}"; do
                if [[ "$arg" == *.md ]] || [[ "$arg" == *.MD ]] || [[ "$arg" == *.Markdown ]] || [[ "$arg" == *.markdown ]]; then
                    specific_file_provided=true
                    break
                fi
            done

            if [[ "$specific_file_provided" == true ]]; then
                # Specific file provided, only show stage names from that file
                _markdown_runner_stage_completion "$cur"
            else
                # No specific file, show only stage names (no file@ completions in non-directory context)
                _markdown_runner_stage_completion "$cur"
            fi
        fi
    fi
}

# Check if a given word is a valid stage name
_is_valid_stage_name() {
    local stage_name="$1"
    local md_files

    # Get files that would actually be executed by markdown-runner
    md_files=$(_get_executable_files | tr '\n' ' ')

    if [[ -n "$md_files" ]]; then
        # Check if the stage exists in any of the files
        local stage_exists=$(echo "$md_files" | xargs grep -l '```[a-zA-Z0-9_. -]*{.*"stage"[[:space:]]*:[[:space:]]*"'"$stage_name"'"' 2>/dev/null)
        if [[ -n "$stage_exists" ]]; then
            return 0  # Stage exists
        fi
    fi
    return 1  # Stage doesn't exist
}

# Get stages from markdown files
_markdown_runner_stage_completion() {
    local cur="$1"
    local stages

    # Check if a specific markdown file is provided in the command line
    local target_file
    target_file=$(_find_markdown_file_in_args)

    local md_files
    if [[ -n "$target_file" ]]; then
        # Use only the specified file
        md_files="$target_file"
    else
        # Get files that would actually be executed by markdown-runner
        md_files=$(_get_executable_files | tr '\n' ' ')
    fi

    # Extract stages from markdown files
    if [[ -n "$md_files" ]]; then
        stages=$(echo "$md_files" | xargs grep -h '```[a-zA-Z0-9_. -]*{.*"stage"' 2>/dev/null | \
                 sed -n 's/.*"stage"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | \
                 sort -u)
    fi

    COMPREPLY=( $(compgen -W "${stages}" -- ${cur}) )

    # Smart nospace behavior: only disable space suffix for stages with multiple chunks
    if [[ ${#COMPREPLY[@]} -gt 0 ]]; then
        local has_multi_chunk_stage=false

        # Check if any of the completed stages has multiple chunks
        for stage in "${COMPREPLY[@]}"; do
            # Get chunk count for this stage
            local chunk_count=0
            if [[ -n "$md_files" ]]; then
                chunk_count=$(echo "$md_files" | xargs grep -h "^\`\`\`[a-zA-Z0-9_. -]*{.*\"stage\"[[:space:]]*:[[:space:]]*\"$stage\"" 2>/dev/null | wc -l)
            fi

            if [[ $chunk_count -gt 1 ]]; then
                has_multi_chunk_stage=true
                break
            fi
        done

        # Only use nospace if there are stages with multiple chunks
        if [[ "$has_multi_chunk_stage" == true ]]; then
            compopt -o nospace 2>/dev/null || true
        fi
    fi
}

# Get chunk IDs and indices for a specific stage
_markdown_runner_chunk_completion() {
    local stage="$1"
    local chunk_prefix="$2"
    local chunks=""

    # Check if a specific markdown file is provided in the command line
    local target_file
    target_file=$(_find_markdown_file_in_args)

    local md_files
    if [[ -n "$target_file" ]]; then
        # Use only the specified file
        md_files="$target_file"
    else
        # Get files that would actually be executed by markdown-runner
        md_files=$(_get_executable_files | tr '\n' ' ')
    fi

    if [[ -n "$md_files" ]]; then
        # Extract all chunks for the specific stage from all files
        local all_chunks=$(echo "$md_files" | xargs grep -h '```[a-zA-Z0-9_. -]*{.*"stage"[[:space:]]*:[[:space:]]*"'"$stage"'"' 2>/dev/null)

        # Process chunks and collect IDs and indices
        local index=0
        local chunk_list=""
        while IFS= read -r line; do
            if [[ -n "$line" ]]; then
                # Extract ID if present
                local chunk_id=$(echo "$line" | sed -n 's/.*"id"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
                if [[ -n "$chunk_id" ]]; then
                    chunk_list="$chunk_list $chunk_id"
                else
                    chunk_list="$chunk_list $index"
                fi
                ((index++))
            fi
        done <<< "$all_chunks"

        chunks="$chunk_list"
    fi

    # Format completions as stage/chunk and remove duplicates
    local completions=""
    for chunk in $chunks; do
        completions="$completions ${stage}/${chunk}"
    done

    # Remove duplicates by converting to array and back
    local unique_completions=($(echo "$completions" | tr ' ' '\n' | sort -u | tr '\n' ' '))

    COMPREPLY=( $(compgen -W "${unique_completions[*]}" -- "${stage}/${chunk_prefix}") )
}

# File-specific stage completion for file@stage format
_markdown_runner_file_at_stage_completion() {
    local file_part="$1"
    local stage_prefix="$2"
    local stages

    # Find the actual file path - prefer files closer to root when multiple matches exist
    local target_file=""
    local executable_files
    executable_files=$(_get_executable_files)
    local candidate_files=()
    while IFS= read -r file; do
        if [[ -n "$file" ]] && [[ "$(basename "$file")" == "$file_part" ]]; then
            candidate_files+=("$file")
        fi
    done <<< "$executable_files"

    # If multiple files found, prefer the one with the shortest path (closest to root)
    if [[ ${#candidate_files[@]} -gt 0 ]]; then
        target_file="${candidate_files[0]}"
        for file in "${candidate_files[@]}"; do
            if [[ ${#file} -lt ${#target_file} ]]; then
                target_file="$file"
            fi
        done
    fi

    if [[ -n "$target_file" ]]; then
        # Extract stages from the specific file
        stages=$(grep -h '```[a-zA-Z0-9_. -]*{.*"stage"' "$target_file" 2>/dev/null | \
                 sed -n 's/.*"stage"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | \
                 sort -u)

        # Format completions as file@stage
        local completions=()
        for stage in $stages; do
            completions+=("${file_part}@${stage}")
        done

        COMPREPLY=( $(compgen -W "${completions[*]}" -- "${file_part}@${stage_prefix}") )

        # Disable space suffix for file@stage completions since they can be chained with /chunk
        if [[ ${#COMPREPLY[@]} -gt 0 ]]; then
            compopt -o nospace 2>/dev/null || true
        fi
    fi
}

# File-specific chunk completion for file@stage/chunk format
_markdown_runner_file_at_chunk_completion() {
    local file_part="$1"
    local stage="$2"
    local chunk_prefix="$3"

    # Find the actual file path - prefer files closer to root when multiple matches exist
    local target_file=""
    local executable_files
    executable_files=$(_get_executable_files)
    local candidate_files=()
    while IFS= read -r file; do
        if [[ -n "$file" ]] && [[ "$(basename "$file")" == "$file_part" ]]; then
            candidate_files+=("$file")
        fi
    done <<< "$executable_files"

    # If multiple files found, prefer the one with the shortest path (closest to root)
    if [[ ${#candidate_files[@]} -gt 0 ]]; then
        target_file="${candidate_files[0]}"
        for file in "${candidate_files[@]}"; do
            if [[ ${#file} -lt ${#target_file} ]]; then
                target_file="$file"
            fi
        done
    fi

    if [[ -n "$target_file" ]]; then
        # Extract chunks for the specific stage from the specific file
        local all_chunks=$(grep -h '```[a-zA-Z0-9_. -]*{.*"stage"[[:space:]]*:[[:space:]]*"'"$stage"'"' "$target_file" 2>/dev/null)

        # Process chunks and collect IDs and indices
        local index=0
        local chunk_list=""
        while IFS= read -r line; do
            if [[ -n "$line" ]]; then
                # Extract ID if present
                local chunk_id=$(echo "$line" | sed -n 's/.*"id"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
                if [[ -n "$chunk_id" ]]; then
                    chunk_list="$chunk_list $chunk_id"
                else
                    chunk_list="$chunk_list $index"
                fi
                ((index++))
            fi
        done <<< "$all_chunks"

        # Format completions as file@stage/chunk
        local completions=()
        for chunk in $chunk_list; do
            completions+=("${file_part}@${stage}/${chunk}")
        done

        COMPREPLY=( $(compgen -W "${completions[*]}" -- "${file_part}@${stage}/${chunk_prefix}") )
    fi
}

# Register the completion function
complete -F _markdown_runner_completion markdown-runner
