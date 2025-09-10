# Google Sheets Calendar TUI Setup Guide

## Overview
The Google Sheets Calendar TUI provides a terminal interface for managing your daily schedule directly in Google Sheets, with automatic Jira integration and offline support.

## Prerequisites

1. **Python Dependencies**
```bash
pip install textual google-api-python-client google-auth pyperclip
```

2. **Google Service Account**
   - Go to [Google Cloud Console](https://console.cloud.google.com)
   - Create a new project or select existing
   - Enable Google Sheets API
   - Create a service account
   - Download credentials JSON file

3. **Google Sheet Setup**
   - Create a new Google Sheet
   - Share it with your service account email
   - Note the Sheet ID from the URL

## Configuration

### Environment Variables
Add to your shell config (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
export GOOGLE_SHEET_ID="your-sheet-id-here"
export GOOGLE_CREDENTIALS_PATH="~/.config/helper-cli/google-credentials.json"
```

### Service Account Credentials
Place your Google service account credentials at:
```bash
~/.config/helper-cli/google-credentials.json
```

### Example Configuration File
Create `~/.config/helper-cli/sheets-config.json`:

```json
{
  "sheet_id": "your-google-sheet-id",
  "credentials_path": "~/.config/helper-cli/google-credentials.json",
  "sync_interval": 60,
  "conflict_resolution": "local_wins",
  "offline_mode": true,
  "jira_integration": {
    "enabled": true,
    "auto_populate": true,
    "default_duration": "1h"
  }
}
```

## Usage

### Launch the TUI
```bash
# Using environment variables
helper jira sheets-tui

# With explicit parameters
helper jira sheets-tui --sheet-id="your-id" --credentials="path/to/creds.json"
```

### Keyboard Shortcuts

#### Month View
- `←` `→` - Navigate between months
- `↑` `↓` - Navigate between weeks
- `Enter` - Edit selected day
- `F` - Focus mode for current day
- `W` - Switch to week view
- `S` - Sync with Google Sheets
- `J` - Auto-populate from Jira
- `R` - Refresh view
- `Q` - Quit

#### Focus Mode
- `Ctrl+S` - Save and exit
- `ESC` - Cancel without saving
- Standard text editor commands

## Features

### 1. Calendar Grid View
- Monthly calendar layout matching Google Sheets
- Visual indicators for today, weekends
- Quick navigation between months
- Cell-by-cell editing

### 2. Focus Mode
- Full-screen editor for single day
- Markdown syntax support
- Line numbers and word count
- Jira activity suggestions

### 3. Bidirectional Sync
- Real-time sync with Google Sheets
- Offline support with local cache
- Conflict resolution strategies
- Background sync every 60 seconds

### 4. Jira Integration
- Auto-detect current ticket from git
- Populate entries from Jira activity
- Track time spent on tickets
- Generate standup notes

## Google Sheets Structure

The TUI expects sheets named by month/year (e.g., "January 2025") with this structure:

| Week # | MONDAY | TUESDAY | WEDNESDAY | THURSDAY | FRIDAY | SATURDAY | SUNDAY |
|--------|--------|---------|-----------|----------|--------|----------|---------|
| W1     | ...    | ...     | ...       | ...      | ...    | ...      | ...     |
| W2     | ...    | ...     | ...       | ...      | ...    | ...      | ...     |

Each cell contains daily activities in this format:
```
TICKET-123 (2h) - Description
TICKET-456 (1.5h) - Another task
• Meeting: Sprint planning
```

## Troubleshooting

### Authentication Issues
```bash
# Check credentials file exists
ls -la ~/.config/helper-cli/google-credentials.json

# Verify sheet is shared with service account
# The service account email is in the credentials JSON
```

### Sync Problems
```bash
# Check sync status
sqlite3 ~/.config/helper-cli/sheets_sync.db "SELECT * FROM sync_operations ORDER BY started_at DESC LIMIT 10;"

# Clear cache if needed
rm ~/.config/helper-cli/sheets_sync.db
```

### Missing Dependencies
```bash
# Install all required packages
pip install -r requirements.txt

# Or install individually
pip install textual>=0.40.0
pip install google-api-python-client>=2.0.0
pip install google-auth>=2.0.0
```

## Advanced Configuration

### Custom Sync Intervals
Edit the sync coordinator initialization:
```python
sync.start_background_sync(interval_seconds=30)  # Sync every 30 seconds
```

### Conflict Resolution Strategies
- `local_wins` - Local changes override remote
- `remote_wins` - Remote changes override local
- `merge` - Attempt to merge both changes
- `manual` - Prompt user for resolution

### Offline Mode
When offline, all changes are queued locally and synced when connection is restored.

## Development

### Running Tests
```bash
pytest tests/test_sheets_service.py
pytest tests/test_sync_coordinator.py
```

### Debug Mode
```bash
# Enable debug logging
export HELPER_DEBUG=1
helper jira sheets-tui
```

## Example Workflow

1. **Morning Standup**
```bash
# Launch TUI
helper jira sheets-tui

# Press 'J' to auto-populate from Jira
# Press 'F' for focus mode on today
# Add standup notes
# Ctrl+S to save
```

2. **Throughout the Day**
```bash
# Quick log entry
helper jira log 2h "Fixed authentication bug"

# TUI will sync automatically
```

3. **End of Day Review**
```bash
# Launch TUI
helper jira sheets-tui

# Press 'W' for week view
# Review the week's activities
# Press 'S' to force sync
```

## Integration with Helper CLI

The sheets-tui command integrates with other helper commands:

- `helper jira log` - Logs to both Jira and Sheets
- `helper jira standup` - Can export to Sheets
- `helper jira sync` - Syncs all unsynced logs

## Support

For issues or questions:
- Check the logs: `~/.config/helper-cli/logs/`
- Run diagnostics: `helper jira config`
- File issues on GitHub
