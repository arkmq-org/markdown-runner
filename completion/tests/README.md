# Bash Completion Test Suite

Declarative JSON-based tests for the bash completion system.

## Quick Start

```bash
# Run all tests
./run_all_tests.sh
```

**Requirements:** `jq` (JSON processor)
```bash
sudo dnf install jq      # Fedora/RHEL
sudo apt install jq      # Debian/Ubuntu
brew install jq          # macOS
```

## Architecture

### Files

- **`test_data.json`** - All test cases in JSON format (~170 lines)
- **`test_framework.sh`** - Test engine and assertion functions (~320 lines)
- **`run_all_tests.sh`** - Main entry point (~50 lines)

**Total: ~540 lines** (previously ~4,700 lines - **89% reduction!**)

### Test Data Format

Each test is a JSON object:

```json
{
  "name": "test description",
  "comp_words": ["markdown-runner", "-t", "1"],
  "assertion": "exact",
  "expected": ["1", "10"]
}
```

### Assertions

| Type | Description | Expected Type |
|------|-------------|---------------|
| `exact` | Exact array match (order matters) | array |
| `contains` | Contains all specified items | array |
| `excludes` | Excludes all specified items | array |
| `count_gt` | Count > N | number |
| `count_eq` | Count == N | number |
| `all_match` | All completions match pattern | array with pattern |
| `any_match` | At least one matches pattern | array with pattern |

## Test Suites

Current test suites (auto-discovered from JSON keys):

- **`filter_tests`** - Filter flag behavior (no completion for `-f`)
- **`flag_equiv_tests`** - Flag equivalents (short/long forms)
- **`incompatible_flag_tests`** - Mutually exclusive flags
- **`positional_tests`** - Positional argument completion
- **`recursive_tests`** - Recursive flag behavior
- **`stage_completion_tests`** - Stage/chunk completion
- **`timeout_tests`** - Timeout value completion
- **`view_tests`** - View mode completion

## Adding Tests

Simply add a new object to the appropriate suite in `test_data.json`:

```json
{
  "timeout_tests": [
    {
      "name": "new timeout test",
      "comp_words": ["markdown-runner", "-t", ""],
      "assertion": "exact",
      "expected": ["1", "5", "10", "30", "60"]
    }
  ]
}
```

Or create a new test suite:

```json
{
  "my_new_tests": [
    {
      "name": "first test",
      "comp_words": ["markdown-runner", "-d", "-"],
      "assertion": "contains",
      "expected": ["-l", "--list"]
    }
  ]
}
```

## Example Test Output

```
╔════════════════════════════════════════════════════════════════╗
║       Bash Completion Test Suite                              ║
╚════════════════════════════════════════════════════════════════╝

=== Timeout Tests ===
✓ timeout shows all values
✓ timeout filters '1' prefix
✓ timeout filters '3' prefix
✓ timeout filters '5' prefix
✓ timeout filters '6' prefix

=== View Tests ===
✓ view mode shows options
✓ view mode filters 'c'
✓ view mode filters 'd'

...

============================================
Test Results:
  Total:  31
  Passed: 31
  Failed: 0
============================================
```
