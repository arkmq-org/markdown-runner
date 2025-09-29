# Bash Completion Test Suite

This directory contains tests for the markdown-runner bash completion functionality.

## Test Files

### Main Test Suites

#### `run_all_tests.sh`
Master test runner that executes all test suites and provides a comprehensive summary.

#### `test_integration.sh`
Integration test that verifies the specific issue reported by users:
- Tests that `markdown-runner -B help<TAB>` shows chunks (`help/0`)
- Verifies that `markdown-runner -B setup<TAB>` shows chunks (`setup/init setup/1`)
- Confirms that partial matches work (`hel` → `help`)

#### `test_flag_completions.sh`
Comprehensive test suite for flag completions:
- Timeout flag completion (`-t` → `1 5 10 30 60`)
- View flag completion (`--view` → `default ci`)
- Filter flag completion (should show no completions)
- Flag name completion (`-<TAB>` → all flags, `--<TAB>` → long flags)
- Flag equivalence (short vs long forms)
- Invalid flag handling

#### `test_flag_filtering.sh`
Test suite for flag filtering (excluding already-used flags):
- Single flag exclusion (`markdown-runner -r file.md -<TAB>` excludes `-r` and `--recursive`)
- Multiple flag exclusion (multiple used flags excluded)
- Equivalent flag exclusion (short form excludes long form and vice versa)
- Flags with values properly excluded
- Edge cases (different flag positions, partial completions)
- Comprehensive flag pair testing

#### `test_incompatible_flags.sh`
Test suite for incompatible flag filtering (excluding logically incompatible flags):
- **Unidirectional**: CI mode (`--view ci`) excludes interactive flags (`-i`, `-B`, `-s`)
- **Bidirectional**: Interactive flags (`-i`, `-B`, `-s`) exclude CI mode (`--view`)
- List mode (`-l`) excludes execution flags (`-d`, `-t`, `-u`, `-i`, `-B`, `-s`)
- Interactive flags exclude list mode (`-l`)
- Help mode (`-h`) excludes all other flags
- Dry run (`-d`) excludes file modification flags (`-u`)
- Update-files (`-u`) excludes dry run (`-d`)
- Break-at (`-B`) excludes ignore-breakpoints (`--ignore-breakpoints`)
- Ignore-breakpoints (`--ignore-breakpoints`) excludes break-at (`-B`)
- Verbose (`-v`) and quiet (`-q`) are mutually exclusive
- Complex combinations and edge cases

#### `test_enhanced_completion.sh`
Test suite for enhanced completion features (categorized help and descriptions):
- **Enhanced help**: `markdown-runner -<TAB>` shows flags with short descriptions
- **Full categorized help**: `markdown-runner help<TAB>` shows complete categorized format
- **Standard compatibility**: Partial flags (`-d<TAB>`) work normally without descriptions
- **Incompatible filtering**: Enhanced help respects incompatible flag filtering
- **Context awareness**: File/value completion still works when appropriate
- **Edge cases**: Enhanced features work with existing arguments and flags

#### `test_main_file_completion.sh`
Test suite for main file argument completion:
- Basic file completion (`.md` files and directories)
- Directory traversal (`path/<TAB>`)
- File filtering (excludes non-`.md` files)
- Nospace behavior for directories
- Path traversal and nested directories
- Edge cases (non-existent paths, case sensitivity)

#### `test_positional_arguments.sh`
Test suite for positional argument completion logic:
- No file provided shows file completions (`markdown-runner <TAB>`)
- Markdown file provided switches to flag completions (`markdown-runner file.md <TAB>`)
- Directory provided switches to flag completions (`markdown-runner test/cases/ <TAB>`)
- File + flag combinations work correctly
- Flag value completion unaffected by file presence
- Edge cases (non-existent files, different extensions, nested directories)

#### `test_start_from_completion.sh`
Test suite for `-s/--start-from` flag completion:
- Basic start-from completion (stage names, file@stage)
- Directory context behavior
- Equivalence with `-B/--break-at` flag
- Specific file context scenarios
- Complex multi-flag scenarios
- Edge case handling

#### `test_completion.sh`
Original comprehensive test suite that covers all completion scenarios:
- Basic stage completion (`markdown-runner -B README.md@`)
- Partial stage completion (`markdown-runner -B README.md@test`)
- Chunk completion (`markdown-runner -B README.md@setup/`)
- Stage completion without file prefix (`markdown-runner -B setup`)
- Chunk completion without file prefix (`markdown-runner -B setup/`)
- Internal function testing (`_get_executable_files`)

#### `test_stage_chunk_completion.sh`
Focused test suite for stage-to-chunk completion functionality:
- Exact stage names showing chunks (`help` → `help/0`)
- Partial stage names showing matching stages (`hel` → `help`)
- Verification that existing formats still work

#### `test_file_at_completion.sh`
Test suite for file@stage completion functionality:
- `README.md@` shows all stages from the main README.md file
- Correct file selection when multiple files with same name exist
- File@stage/chunk format completion
- Partial file@stage completion
- Proper handling of files not in current directory

#### `test_smart_nospace.sh`
Test suite for smart nospace behavior:
- Single-chunk stages allow trailing spaces (complete workflow)
- Multi-chunk stages use nospace (enable chaining)
- Mixed completions use nospace if any stage has multiple chunks
- Chunk counting logic verification
- Smart behavior based on actual chunk analysis

#### `test_file_at_stage_logic.sh`
Test suite for file@stage complete vs partial logic:
- Complete stage names show chunks (e.g., `README.md@inner_test1` → chunks)
- Partial stage names complete to stage names (e.g., `README.md@inner` → `README.md@inner_test1`)
- Non-existent stages show no completions
- File@stage/chunk format works correctly
- Recursive vs non-recursive consistency

#### `test_directory_partial_completion.sh`
Test suite for directory partial completion fix:
- Directory empty completion shows file@ completions (`test/cases/ -B <TAB>`)
- Directory partial completion shows filtered file@ completions (`test/cases/ -B p<TAB>` → `parallel.md@`)
- Complete stage names in directory context show chunks
- Multi-character partial matches work correctly
- Non-directory and recursive contexts remain unaffected

#### `test_recursive_behavior.sh`
Test suite for recursive vs non-recursive completion behavior:
- Non-recursive mode shows stage names + file@ completions
- Recursive mode shows ONLY file@ completions
- Partial matching behavior in both modes
- Support for both `-r` and `--recursive` flags
- Recursive behavior with different completion flags (`-B`, `-s`)

#### `test_file_discovery.sh`
Test suite for the `_get_executable_files` function:
- File discovery in different completion contexts
- Handling of flag values vs directory arguments
- Edge cases with multiple flags and values

#### `test_stage_validation.sh`
Test suite for the `_is_valid_stage_name` function:
- Valid stage name detection
- Invalid stage name rejection
- Edge cases with partial matches and empty strings

#### `test_directory_context.sh`
Test suite for the `_is_directory_context` function:
- Flag value contexts (should not be directory contexts)
- Actual directory arguments (should be directory contexts)
- Mixed scenarios with flags and paths

### Debug Scripts

#### `debug_get_executable_files.sh`
Debug script for the `_get_executable_files` function. Helps troubleshoot:
- File discovery logic
- Argument parsing from `COMP_WORDS`
- Path resolution
- Recursive flag handling

#### `debug_stage_extraction.sh`
Debug script for stage extraction from markdown files. Helps troubleshoot:
- Stage discovery from markdown files
- Regex patterns for stage extraction
- File@stage completion logic
- Stage counting and sorting

#### `debug_chunk_completion.sh`
Debug script for chunk completion. Helps troubleshoot:
- Chunk discovery within stages
- ID extraction vs index assignment
- File@stage/chunk completion logic
- Chunk list building

## Running Tests

### Run All Tests
```bash
cd completion/tests
./run_all_tests.sh
```

### Run Individual Test Suites
```bash
cd completion/tests
./test_completion.sh                    # Original comprehensive tests
./test_stage_chunk_completion.sh        # Stage-to-chunk completion tests
./test_file_discovery.sh                # File discovery tests
./test_stage_validation.sh              # Stage validation tests
./test_directory_context.sh             # Directory context tests
```

### Run Individual Debug Scripts
```bash
cd completion/tests
./debug_get_executable_files.sh
./debug_stage_extraction.sh
./debug_chunk_completion.sh
```

## Test Output

The main test suite provides colored output:
- 🟢 **Green**: Passed tests
- 🔴 **Red**: Failed tests
- 🟡 **Yellow**: Test names and descriptions

Example output:
```
=== Markdown Runner Bash Completion Test Suite ===

Testing: Basic stage completion
  Command: markdown-runner -B README.md@
  Found 13 completions
  PASS

Testing: Partial stage completion
  Command: markdown-runner -B README.md@test
  Found 6 completions
  PASS

...

=== Test Results ===
Tests run: 6
Tests passed: 6
Tests failed: 0
All tests passed!
```

## Adding New Tests

To add a new test to the main test suite:

1. Use the `run_completion_test` helper function:
   ```bash
   run_completion_test \
       "Test description" \
       "command line to test" \
       "regex pattern for validation" \
       "specific completion that should be present"
   ```

2. For debugging specific functionality, create a new debug script following the pattern of existing ones.

## Troubleshooting

If tests fail:

1. **Check file discovery**: Run `debug_get_executable_files.sh` to ensure markdown files are being found correctly.

2. **Check stage extraction**: Run `debug_stage_extraction.sh` to verify stages are being extracted from markdown files.

3. **Check chunk completion**: Run `debug_chunk_completion.sh` to debug chunk-specific completion issues.

4. **Verify completion script**: Ensure the completion script is sourced correctly and all functions are available.

## Integration with CI

These tests can be integrated into CI/CD pipelines to ensure completion functionality doesn't regress:

```bash
# In your CI script
cd completion/tests
./test_completion.sh
```

The test script exits with code 0 on success and 1 on failure, making it suitable for automated testing.
