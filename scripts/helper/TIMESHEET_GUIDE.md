# Timesheet Tracking Plugin CLI Guide

A CLI tool to log time directly to Jira Cloud's Timesheet Tracking plugin from your terminal.

## Setup

### 1. Get Your Authentication Tokens

Open your browser with Jira Cloud open and get your authentication cookies:

1. Press **F12** to open DevTools
2. Go to **Application** tab (Chrome) or **Storage** tab (Firefox)
3. Navigate to **Cookies** → Select your Jira domain
4. Find and copy these three cookie values:
   - `tenant.session.token`
   - `atlassian.xsrf.token`
   - `JSESSIONID`

### 2. Configure the CLI

Run the setup command:

```bash
helper timesheet setup
```

You'll be prompted to enter:
- **Jira Cloud URL**: Your Jira base URL (e.g., `https://acre-identity.atlassian.net`)
- **Session Token**: Paste the `tenant.session.token` value
- **XSRF Token**: Paste the `atlassian.xsrf.token` value
- **JSESSIONID**: Paste the `JSESSIONID` value

Configuration is saved to: `~/.config/helper-cli/timesheet.json`

## Usage

### Log Time to an Issue

Basic usage:
```bash
helper timesheet log TDT-123 30m
helper timesheet log TDT-123 2h
helper timesheet log TDT-123 1h30m
```

With a comment:
```bash
helper timesheet log TDT-123 30m --comment "Fixed authentication bug"
helper timesheet log TDT-123 2h -c "Implemented new feature"
```

Backdate to a specific date/time:
```bash
helper timesheet log TDT-123 1h --date 2025-11-24
helper timesheet log TDT-123 2h --date 2025-11-24 --time 09:00
```

### Time Format

You can use any Jira-compatible time format:
- `30m` = 30 minutes
- `2h` = 2 hours
- `1d` = 1 day (8 hours)
- `1h30m` = 1 hour 30 minutes
- `2d 4h` = 2 days 4 hours

## Examples

```bash
# Log 30 minutes to today
helper timesheet log VIS-1234 30m

# Log 2 hours with a comment
helper timesheet log VIS-1234 2h --comment "Code review and testing"

# Backdate to last Friday at 2pm
helper timesheet log VIS-1234 3h --date 2025-11-21 --time 14:00
```

## Troubleshooting

### Authentication Failed

If you get authentication errors:
1. Your session tokens may have expired
2. Run `helper timesheet setup` again to update them
3. Session tokens typically expire after a few weeks

### Connection Errors

Make sure:
- You're connected to the internet
- Your Jira URL is correct (no trailing slash)
- The Timesheet Tracking plugin is installed in your Jira instance

### Finding Issue Keys

You can use the existing Jira commands to find issue keys:
```bash
# See your assigned tickets
helper jira assigned

# Search for specific tickets
helper jira search "project = TDT AND status = 'In Progress'"
```

## Integration with Existing Features

This works alongside your existing helper CLI features:

```bash
# Create a branch for an issue
helper branch "VIS-1234 fix login bug"

# Log time when you're done
helper timesheet log VIS-1234 2h --comment "Fixed login bug"

# Create standup notes (includes logged time)
helper jira standup
```

## Configuration File Location

Config is stored at: `~/.config/helper-cli/timesheet.json`

The file is automatically secured with 600 permissions (read/write for user only).

## API Details

This tool uses:
- **Endpoint**: `/gateway/api/graphql`
- **Operation**: `forge_ui_invokeExtension` GraphQL mutation
- **Plugin**: Timesheet Tracking for Jira by Cappsule (v4.9.0+)

The tool replicates the exact API calls that the Timesheet plugin UI makes when you log time through the browser.
