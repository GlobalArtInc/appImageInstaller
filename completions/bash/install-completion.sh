#!/bin/bash

# Define completion directories
COMPLETION_DIR="/etc/bash_completion.d"
USER_COMPLETION_DIR="$HOME/.bash_completion.d"
COMPLETION_FILE="appinstaller"

# Function to install in system directory
install_system() {
    if [ "$EUID" -ne 0 ]; then
        echo "Root privileges required for system-wide installation"
        return 1
    fi
    
    cp "$(dirname "$0")/$COMPLETION_FILE" "$COMPLETION_DIR/$COMPLETION_FILE"
    chmod 644 "$COMPLETION_DIR/$COMPLETION_FILE"
    echo "Completion installed system-wide"
}

# Function to install in user directory
install_user() {
    mkdir -p "$USER_COMPLETION_DIR"
    cp "$(dirname "$0")/$COMPLETION_FILE" "$USER_COMPLETION_DIR/$COMPLETION_FILE"
    chmod 644 "$USER_COMPLETION_DIR/$COMPLETION_FILE"
    
    # Add user completions loading to .bashrc if not present
    if ! grep -q "bash_completion.d" "$HOME/.bashrc"; then
        echo '
# Load user bash completions
if [ -d "$HOME/.bash_completion.d" ]; then
    for completion in "$HOME/.bash_completion.d"/*; do
        [ -f "$completion" ] && source "$completion"
    done
fi' >> "$HOME/.bashrc"
    fi
    
    echo "Completion installed for current user"
}

# Main installation logic
if [ "$1" = "--user" ]; then
    install_user
else
    install_system || install_user
fi

echo "To apply changes, either:"
echo "1. Run: source ~/.bashrc"
echo "2. Restart your terminal" 