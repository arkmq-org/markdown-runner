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
