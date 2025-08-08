# Browser Tools & Extensions

Browser-related tools, standalone builds, and custom extensions.

## Chrome Linux Standalone

Standalone Chrome/Chromium build for Linux (February 2023).

### Usage

Run directly from extracted folder:
```bash
~/.Dotfiles/Browser/chrome-linux/chrome
```

Or create an alias:
```bash
alias chrome-standalone='~/.Dotfiles/Browser/chrome-linux/chrome'
```

### Files
- `chrome-linux.zip.part-*` - Split archive files (under 100MB each for GitHub)
- `chrome-linux/chrome.part-*` - Split binary files (under 100MB each for GitHub)
- `chrome-linux/` - Extracted Chrome application directory
- `reassemble.sh` - Script to reassemble split files after cloning

### First Time Setup After Clone
```bash
cd ~/.Dotfiles/Browser
./reassemble.sh  # Reassembles the split files
```

**Note:** Large files are split to fit GitHub's 100MB limit. The reassemble script will restore them.

## Forever Pinned

Chrome extension that saves and restores pinned tabs when Chrome starts up.

### Installation

1. Open Chrome and navigate to `chrome://extensions/`
2. Enable "Developer mode" (toggle in top-right corner)
3. Click "Load unpacked"
4. Select: `~/.Dotfiles/Browser/forever-pinned/`

### Features

- Automatically saves your pinned tabs
- Restores pinned tabs when Chrome starts
- Click extension icon to manually restore pinned tabs
- Configure saved tabs through the options page

### Why Local?

This extension was removed from the Chrome Web Store but is essential for daily workflow (meeting reminders via pinned Google Calendar tabs, etc.)

Modified to work with Manifest V3 for compatibility with modern Chrome versions.