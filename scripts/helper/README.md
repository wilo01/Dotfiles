# Helper CLI

Development helper CLI for JIRA, badges, URLs, and automation tasks.

## Installation

Install in development mode:

```bash
cd ~/.Dotfiles/scripts/helper
pip install -e .
```

## Usage

### Commands

- `helper jira "commit message"` - Generate JIRA branch name
- `helper badge 4324325553` - Generate badge string (RFID/QR)
- `helper stash "stash message"` - Generate git stash command
- `helper ngrok "https://ngrok-url.com"` - Generate QR code for NGROK URL
- `helper vsc "https://vsc-url.com"` - Generate QR code for VSC URL
- `helper rt "directory-name"` - Setup RT directory
- `helper apex file.zip 8090 auth-token` - Generate APEX upload command
- `helper interactive` - Interactive mode

### Examples

```bash
# Generate JIRA branch name
helper jira "Fix user authentication issue"

# Generate RFID badge
helper badge 4324325553 --type rfid

# Generate QR badge
helper badge 4324325553 --type qr

# Create git stash command
helper stash "work in progress changes"

# Interactive mode
helper interactive
```

## Features

- 🔧 JIRA branch name generation from commit messages
- 🏷️ RFID/QR badge automation with auto-typing
- 📦 Git stash command generation
- 🔗 URL shortening and QR code generation
- 📁 RT directory setup automation
- 🚀 APEX upload command generation
- 🎯 Interactive mode for easy usage
- 📋 Automatic clipboard integration
