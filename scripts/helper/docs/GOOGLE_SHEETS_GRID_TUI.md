# Google Sheets Grid TUI - User Guide

## Overview

The Google Sheets Grid TUI is a modern terminal interface that **exactly replicates** the Google Sheets calendar structure with significant UX/UI improvements. It provides a professional, visually appealing way to manage your daily schedule directly from the terminal.

## Key Features

### 📊 Exact Google Sheets Grid Layout
- **Week Rows**: W1, W2, W3, etc. matching your actual sheets
- **Day Columns**: MONDAY through SUNDAY
- **Cell-by-cell editing**: Just like Google Sheets
- **Bidirectional sync**: Changes sync both ways

### 🎨 Modern Visual Design
- **Material Design color scheme** with proper visual hierarchy
- **Color-coded cells**:
  - 🟡 Today highlighted in yellow
  - 🔵 Cells with content in blue
  - ⚫ Weekends in muted gray
  - 🏷️ Week labels with blue badges
- **Smooth animations** for month transitions
- **Visual feedback** for all actions

### ✏️ Enhanced Edit Modal
- **Beautiful modal design** with 70% width
- **Cell indicator badge** showing what you're editing
- **Markdown editor** with syntax highlighting and line numbers
- **Jira integration** for auto-populating activities
- **Keyboard shortcuts** for efficiency

## Installation

```bash
# Ensure dependencies are installed
pip install textual google-api-python-client google-auth

# Navigate to helper directory
cd ~/.Dotfiles/scripts/helper

# The TUI is integrated into the helper CLI
```

## Usage

### Launching the TUI

```bash
# With OAuth authentication (recommended)
helper jira sheets-tui

# With specific sheet ID
helper jira sheets-tui --sheet-id="your-sheet-id"

# In demo mode (no authentication)
python test_grid_tui.py
```

### Keyboard Shortcuts

#### Main Grid View
| Key | Action | Description |
|-----|--------|-------------|
| `←` / `→` | Navigate months | Move between months |
| `↑` / `↓` | Move cursor | Navigate between cells |
| `Enter` / `E` | Edit cell | Open edit modal for selected cell |
| `S` | Sync | Sync with Google Sheets |
| `J` | Jira populate | Auto-fill from Jira activities |
| `R` | Refresh | Refresh current view |
| `H` | Help | Show help dialog |
| `Q` | Quit | Exit application |

#### Edit Modal
| Key | Action | Description |
|-----|--------|-------------|
| `Ctrl+S` | Save | Save and close editor |
| `Ctrl+J` | Insert Jira | Insert Jira activity suggestions |
| `Esc` | Cancel | Close without saving |

### Quick Actions Buttons
- **📅 Today**: Jump to current month
- **🔄 Sync All**: Force full synchronization
- **📊 Week View**: Show detailed week view

## Google Sheets Structure

The TUI expects sheets with this exact structure:

```
| Week # | MONDAY | TUESDAY | WEDNESDAY | THURSDAY | FRIDAY | SATURDAY | SUNDAY |
|--------|--------|---------|-----------|----------|--------|----------|---------|
| W1     | ...    | ...     | ...       | ...      | ...    | ...      | ...     |
| W2     | ...    | ...     | ...       | ...      | ...    | ...      | ...     |
| W3     | ...    | ...     | ...       | ...      | ...    | ...      | ...     |
| W4     | ...    | ...     | ...       | ...      | ...    | ...      | ...     |
| W5     | ...    | ...     | ...       | ...      | ...    | ...      | ...     |
```

Each worksheet should be named by month and year (e.g., "January 2025", "February 2025").

## Features in Detail

### 1. Month Navigation
- Use arrow keys to smoothly transition between months
- Visual notifications show the month you're navigating to
- Loading animations during data fetch
- Automatic grid rebuilding for each month

### 2. Cell Editing
- Click or press Enter on any cell to edit
- Enhanced modal with:
  - Clear cell reference (e.g., "MONDAY W2")
  - Date display for context
  - Multi-line text editor with line numbers
  - Character count validation (max 5000)
  - Visual save confirmation

### 3. Week View
- Press "Week View" button or use shortcut
- Shows detailed view of a single week
- Each day displayed in its own panel
- Easier to see full content for each day
- Quick close to return to month view

### 4. Sync Status
- Real-time sync status in header
- Shows last sync timestamp
- Visual indicators:
  - ⟳ Loading/Syncing
  - ✅ Successfully synced
  - ❌ Sync error
  - 🔶 Demo mode
  - 📝 New month (no data yet)

### 5. Visual Feedback
- Button hover effects
- Cell selection highlighting
- Loading animations during async operations
- Success/error notifications
- Month transition animations

## Configuration

### Environment Variables
```bash
# Google Sheet ID
export GOOGLE_SHEET_ID="your-sheet-id"

# Credentials path (for service account)
export GOOGLE_CREDENTIALS_PATH="~/.config/helper-cli/google-credentials.json"
```

### OAuth Setup
```bash
# First-time OAuth setup
helper jira sheets-auth

# Force re-authentication
helper jira sheets-auth --force
```

## Troubleshooting

### Common Issues

#### 1. Cannot Access Sheet
- Verify sheet ID is correct
- Ensure sheet is shared with your account
- Check authentication status: `helper jira sheets-auth`

#### 2. Sync Errors
- Check internet connection
- Verify Google Sheets API is enabled
- Review error messages in sync status

#### 3. Display Issues
- Ensure terminal supports Unicode
- Check terminal size (minimum 80x24 recommended)
- Try different terminal emulator

### Debug Mode
```bash
# Enable debug logging
export HELPER_DEBUG=1
helper jira sheets-tui

# Check logs
tail -f ~/.config/helper-cli/logs/sheets-sync.log
```

## Demo Mode

The TUI includes a demo mode for testing without Google Sheets:
- Launches automatically if authentication fails
- Shows sample data for UI testing
- All features work except actual sync
- Useful for development and testing

## Architecture

### Components
1. **GoogleSheetsGridApp**: Main application class
2. **GoogleSheetsGridScreen**: Main grid view screen
3. **EnhancedEditModal**: Cell editing modal
4. **WeekViewScreen**: Detailed week view modal
5. **LoadingIndicator**: Animated loading component

### Services
- **OAuth Service**: Handles Google authentication
- **Sheets Service**: Manages sheet read/write operations
- **Sync Coordinator**: Handles bidirectional sync
- **Context Detector**: Provides Jira suggestions

## Best Practices

1. **Regular Syncing**: The TUI auto-syncs every 60 seconds
2. **Offline Work**: Changes are queued when offline
3. **Batch Edits**: Edit multiple cells before syncing
4. **Month Organization**: Create sheets for each month
5. **Content Format**: Use markdown for better formatting

## Integration with Helper CLI

The Grid TUI integrates seamlessly with other helper commands:
- `helper jira log`: Logs work to both Jira and Sheets
- `helper jira standup`: Generates standup notes
- `helper jira fetch`: Updates local Jira cache

## Performance

- **Grid rendering**: < 100ms
- **Cell edit latency**: < 50ms
- **Month navigation**: < 200ms
- **Sync operation**: 1-3 seconds
- **Memory usage**: ~50MB

## Future Enhancements

- [ ] Multi-cell selection and editing
- [ ] Copy/paste between cells
- [ ] Formula support
- [ ] Export to various formats
- [ ] Collaborative real-time editing
- [ ] Mobile companion app

## Support

For issues or questions:
- Check this documentation
- Review logs in `~/.config/helper-cli/logs/`
- Run `helper jira config` for configuration help
- File issues on GitHub

---

**Version 2.0** - Complete redesign with Google Sheets grid structure
*Last updated: September 2025*