# Bash Completion for markdown-runner

This directory contains bash completion scripts for markdown-runner that provide intelligent autocompletion for command-line arguments.

## Features

- **Flag completion**: Complete all command-line flags (`-d`, `--dry-run`, etc.)
- **Stage completion**: Complete stage names from your markdown files when using `-B` or `-s`
- **Chunk completion**: Complete chunk IDs and indices when using `-B stage/` format
- **File completion**: Complete `.md` files and directories for the main argument
- **Directory traversal**: Navigate through subdirectories with Tab completion (e.g., `test/cases/`)
- **Context-aware completion**: When targeting directories, only shows `file@` completions to avoid ambiguous stage names
- **Smart discovery**: Automatically discovers stages and chunks from markdown files in your current directory
- **Smart chainable completions**: Intelligent trailing space behavior - single-chunk stages get trailing spaces (complete workflow), multi-chunk stages enable chaining (e.g., `setup<TAB>` → `setup/<TAB>` → `setup/init`)

## Installation

### Quick Install (User-only)

```bash
./install.sh
```

This installs completion for the current user only and adds it to your `.bashrc`.

### System-wide Install

```bash
./install.sh system
```

This installs completion system-wide (requires `sudo`) for all users.

### Manual Install

You can also manually source the completion script:

```bash
source ./bash_completion.sh
```

Add this line to your `.bashrc` to enable it permanently.

## Usage Examples

### Stage Completion

**Context-aware completion**: When targeting directories, only `file@` completions are shown to avoid ambiguous stage names:

```bash
$ markdown-runner test/cases/ -B <TAB><TAB>
happy.md@  parallel.md@  teardown.md@  writer.md@
```

**File-specific completion**: When you specify a markdown file, completion shows stages from that file:

```bash
$ markdown-runner -B <TAB><TAB>
help  init  main  setup  teardown  test
```

**Smart file-specific completion**: When you specify a markdown file, completion only shows stages from that file:

```bash
$ markdown-runner test/cases/happy.md -B <TAB><TAB>
test

$ markdown-runner test/cases/teardown.md -B <TAB><TAB>  
main  teardown
```

### Chunk Completion

When you type `markdown-runner -B setup/` and press Tab twice:

```bash
$ markdown-runner -B setup/<TAB><TAB>
setup/0     setup/1     setup/cleanup     setup/init
```

**File-specific chunk completion**: When a file is specified, chunks are only from that file:

```bash
$ markdown-runner test/cases/teardown.md -B main/<TAB><TAB>
main/failing  main/succeeding
```

- **Named chunks**: Shows actual chunk IDs (like `cleanup`, `init`)
- **Unnamed chunks**: Shows indices (like `0`, `1`, `2`)

### File Path Completion

Navigate through directories and complete markdown files:

```bash
$ markdown-runner test<TAB>
$ markdown-runner test/

$ markdown-runner test/<TAB>
$ markdown-runner test/cases/

$ markdown-runner test/cases/<TAB><TAB>
test/cases/happy.md  test/cases/parallel.md  test/cases/teardown.md

$ markdown-runner test/cases/h<TAB>
$ markdown-runner test/cases/happy.md
```

**Efficient chaining**: Directory completions and file@ completions don't add trailing spaces, allowing you to immediately press Tab again to continue completing:

```bash
$ markdown-runner test/cases/ -B parallel.m<TAB>
$ markdown-runner test/cases/ -B parallel.md@

$ markdown-runner test/cases/ -B parallel.md@<TAB>
$ markdown-runner test/cases/ -B parallel.md@parallel_write
```

### Partial Completion

Type part of a stage or chunk name and press Tab:

```bash
$ markdown-runner -B set<TAB>
$ markdown-runner -B setup/

$ markdown-runner -B setup/i<TAB>
$ markdown-runner -B setup/init
```

## How It Works

The completion script:

1. **Discovers markdown files** in your current directory (or recursively with `-r` flag)
2. **Parses stage definitions** by looking for code blocks with `{"stage":"name"}` metadata
3. **Extracts chunk information** including IDs and calculates indices for unnamed chunks
4. **Provides intelligent completion** based on what you've typed so far

### File Discovery

- **Default**: Searches `.md` files in current directory only
- **With `-r` flag**: Recursively searches subdirectories for `.md` files

### Stage Discovery

Finds stages by parsing code blocks like:
```markdown
```bash {"stage":"setup", "id":"init"}
echo "Setting up..."
```
```

### Chunk Discovery

For each stage, discovers:
- **Named chunks**: Uses the `"id"` field (e.g., `init`, `cleanup`)
- **Unnamed chunks**: Assigns indices based on order (e.g., `0`, `1`, `2`)

## Compatibility

- Requires **bash 4.0+**
- Works with **bash-completion** package
- Tested on Fedora 42

## Troubleshooting

### Completion Not Working

1. Make sure bash-completion is installed:
   ```bash
   # Ubuntu/Debian
   sudo apt install bash-completion
   ```

2. Restart your shell or source the completion:
   ```bash
   source ~/.bash_completion.d/markdown-runner
   ```

### No Stages Found

- Make sure you're in a directory with `.md` files
- Check that your markdown files have properly formatted stage definitions
- Use `--recursive` flag if your markdown files are in subdirectories

### Slow Completion

- Large numbers of markdown files can slow completion
- Consider using the `--filter` flag to limit which files are processed
- The completion caches results within a single completion session

## Development

The completion script is located in `bash_completion.sh` and has been refactored into a modular architecture with 25 focused functions:

### Core Architecture

**Main Orchestrator:**
- `_markdown_runner_completion()`: Main completion function that delegates to specialized handlers

**Completion Handlers:**
- `_handle_flag_value_completion()`: Handles flag value completions (`-B`, `-s`, `-t`, `--view`, `-f`)
- `_handle_flag_completion()`: Handles flag name completions and enhanced help
- `_handle_help_completion()`: Handles help command completion
- `_handle_positional_completion()`: Handles file and directory completion

**Flag Management:**
- `_get_available_flags()`: Main flag filtering orchestrator
- `_collect_used_flags()`: Identifies already-used flags
- `_collect_incompatible_flags()`: Identifies incompatible flag combinations
- `_filter_available_flags()`: Applies filtering logic

**Utility Functions:**
- `_get_flag_equivalent()`: Maps short/long flag equivalents
- `_flag_takes_value()`: Determines if a flag expects a value
- `_skip_flag_and_value()`: Standardized flag parsing
- `_is_markdown_file()`: Unified file extension checking
- `_find_markdown_file_in_args()`: Centralized file discovery
- `_is_completing_flag_value()`: Context detection
- `_apply_nospace_for_file_at()`: Consistent nospace behavior
- `_has_file_or_dir_argument()`: Positional argument detection

**Specialized Completion Functions:**
- `_markdown_runner_file_completion()`: Handles file and directory completion for .md files
- `_markdown_runner_break_at_completion()`: Handles `-B` flag completion
- `_markdown_runner_start_from_completion()`: Handles `-s` flag completion
- `_markdown_runner_stage_completion()`: Discovers and completes stage names
- `_markdown_runner_chunk_completion()`: Discovers and completes chunk IDs/indices
- `_markdown_runner_file_at_stage_completion()`: Handles file@stage completion
- `_markdown_runner_file_at_chunk_completion()`: Handles file@stage/chunk completion

**Context Detection:**
- `_is_directory_context()`: Determines if completion is for a directory vs specific file
- `_is_valid_stage_name()`: Validates stage names against markdown files

**Enhanced UX:**
- `_markdown_runner_show_enhanced_help()`: Shows flags with short descriptions
- `_markdown_runner_show_full_help()`: Shows full categorized help

### Architecture Benefits

- **Maintainability**: Single-responsibility functions are easier to understand and modify
- **Testability**: Each component can be tested independently
- **Extensibility**: Clear extension points for new features
- **Debugging**: Issues can be isolated to specific functions

## Testing

The completion functionality is backed by a comprehensive test suite with **191 tests** across **18 test suites**:

### Test Coverage
- **Integration tests**: End-to-end completion scenarios
- **Flag completion tests**: All flag types and combinations
- **Stage/chunk completion**: Stage discovery and chunk completion
- **File completion**: File and directory completion
- **Context detection**: Directory vs file contexts
- **Edge cases**: Error handling and boundary conditions
- **Regression tests**: Prevent future breakage

### Running Tests
```bash
# Run all tests
cd completion/tests
./run_all_tests.sh

# Run specific test suite
./test_flag_completions.sh
./test_stage_chunk_completion.sh
```

See [`tests/README.md`](tests/README.md) for detailed test documentation.
