#!/bin/bash

# Installation script for markdown-runner bash completion

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPLETION_FILE="$SCRIPT_DIR/bash_completion.sh"

# Check if completion file exists
if [[ ! -f "$COMPLETION_FILE" ]]; then
    echo "Error: Completion file not found at $COMPLETION_FILE"
    exit 1
fi

# Function to install system-wide (requires sudo)
install_system_wide() {
    local system_dir

    # Try different system completion directories
    if [[ -d "/usr/share/bash-completion/completions" ]]; then
        system_dir="/usr/share/bash-completion/completions"
    elif [[ -d "/etc/bash_completion.d" ]]; then
        system_dir="/etc/bash_completion.d"
    else
        echo "Error: No system bash completion directory found"
        echo "Please install bash-completion package first"
        exit 1
    fi

    echo "Installing completion to $system_dir/markdown-runner"
    sudo cp "$COMPLETION_FILE" "$system_dir/markdown-runner"
    echo "System-wide installation complete!"
    echo "You may need to restart your shell or run: source $system_dir/markdown-runner"
}

# Function to install for current user
install_user() {
    local user_dir="$HOME/.bash_completion.d"

    # Create user completion directory if it doesn't exist
    mkdir -p "$user_dir"

    # Copy completion file
    echo "Installing completion to $user_dir/markdown-runner"
    cp "$COMPLETION_FILE" "$user_dir/markdown-runner"

    # Add to .bashrc if not already present
    local bashrc="$HOME/.bashrc"
    local completion_source="source $user_dir/markdown-runner"

    if [[ -f "$bashrc" ]] && ! grep -q "$completion_source" "$bashrc"; then
        echo "" >> "$bashrc"
        echo "# markdown-runner completion" >> "$bashrc"
        echo "$completion_source" >> "$bashrc"
        echo "Added completion source to $bashrc"
    fi

    echo "User installation complete!"
    echo "Run: source $user_dir/markdown-runner"
    echo "Or restart your shell to enable completion"
}

# Parse command line arguments
case "${1:-user}" in
    "system"|"--system")
        install_system_wide
        ;;
    "user"|"--user"|"")
        install_user
        ;;
    "--help"|"-h")
        echo "Usage: $0 [user|system]"
        echo ""
        echo "Install bash completion for markdown-runner"
        echo ""
        echo "Options:"
        echo "  user    Install for current user only (default)"
        echo "  system  Install system-wide (requires sudo)"
        echo "  --help  Show this help message"
        ;;
    *)
        echo "Error: Unknown option '$1'"
        echo "Run '$0 --help' for usage information"
        exit 1
        ;;
esac
