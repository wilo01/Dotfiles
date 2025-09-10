# Google Sheets OAuth Integration Guide

This guide consolidates all Google Sheets OAuth documentation for the Helper CLI.

## Quick Start (Recommended)

The Helper CLI uses **gspread's simple OAuth flow** - the easiest way to integrate with Google Sheets.

### Step 1: Install Dependencies

```bash
cd ~/.Dotfiles/scripts/helper
pip install -e .
```

### Step 2: Authenticate

```bash
# First-time authentication (opens browser)
helper jira sheets-auth

# Your browser will open automatically:
# 1. Sign in with your Google account
# 2. Click "Allow" to grant permissions
# 3. Done! Token saved for future use
```

That's it! The authentication is now saved and will auto-refresh when needed.

## Authentication Options

### Option 1: Default OAuth (Simplest) ✨

**No Google Cloud Console setup required!**

```bash
# Just run the auth command
helper jira sheets-auth

# The tool uses gspread's test OAuth app by default
# This is perfect for personal use and development
```

**Pros:**
- Zero configuration
- Works immediately
- Auto-refreshes tokens
- Perfect for personal projects

**Cons:**
- Shows "unverified app" warning (normal for test apps)
- Rate limits for heavy usage

### Option 2: Custom OAuth App (Production)

For production use or to avoid warnings, create your own OAuth app:

1. **Create Google Cloud Project**
   ```
   1. Go to https://console.cloud.google.com
   2. Create new project or select existing
   3. Enable Google Sheets API
   ```

2. **Create OAuth 2.0 Credentials**
   ```
   1. Go to APIs & Services > Credentials
   2. Click "Create Credentials" > "OAuth 2.0 Client IDs"
   3. Application type: "Desktop app"
   4. Name: "Helper CLI"
   5. Download the JSON file
   ```

3. **Install Credentials**
   ```bash
   # Save credentials to config directory
   cp ~/Downloads/credentials.json ~/.config/helper-cli/google_credentials.json
   
   # Authenticate with custom credentials
   helper jira sheets-auth
   ```

### Option 3: Service Account (Automation)

For server/automation use without browser interaction:

1. **Create Service Account**
   ```
   1. Google Cloud Console > IAM & Admin > Service Accounts
   2. Create service account
   3. Download JSON key file
   ```

2. **Share Sheet with Service Account**
   ```
   1. Open your Google Sheet
   2. Share with service account email
   3. Grant editor permissions
   ```

3. **Configure Helper CLI**
   ```bash
   # Save service account key
   cp ~/Downloads/service-account-key.json ~/.config/helper-cli/
   
   # Configure in settings
   helper jira config --google-service-account
   ```

## Commands Reference

### Authentication Management

```bash
# Initial authentication
helper jira sheets-auth

# Check authentication status
helper jira sheets-auth --status

# Force re-authentication
helper jira sheets-auth --force

# Revoke access
helper jira sheets-auth --revoke

# Use custom credentials
helper jira sheets-auth --credentials path/to/credentials.json
```

### Working with Sheets

```bash
# Open TUI interface
helper jira sheets-tui

# Sync work logs to sheets
helper jira sync --sheets

# Write standup notes
helper jira standup --sheets

# Read today's schedule
helper jira today --from-sheets
```

## File Structure

```
~/.config/helper-cli/
├── google_token.json          # OAuth token (auto-created)
├── google_credentials.json    # Custom OAuth app (optional)
└── service_account.json       # Service account (optional)
```

## Security Best Practices

1. **Token Security**
   - Tokens are stored with 600 permissions (owner-only)
   - Never commit tokens to version control
   - Tokens auto-expire and refresh

2. **Credential Management**
   - Use the secure credential manager
   - Enable encryption for sensitive data
   - Rotate tokens periodically

3. **Sheet Permissions**
   - Grant minimum required permissions
   - Use read-only access where possible
   - Audit sheet sharing regularly

## Troubleshooting

### "Authentication failed"
```bash
# Check internet connection
ping google.com

# Force re-authentication
helper jira sheets-auth --force

# Check for proxy/firewall issues
export HTTPS_PROXY=your-proxy:port
```

### "Sheet not found"
```bash
# Verify sheet ID in config
helper jira config --show

# Check sheet permissions
# Ensure sheet is shared with your Google account

# Test with sheet URL
helper jira sheets-test https://docs.google.com/spreadsheets/d/YOUR_SHEET_ID
```

### "Token expired"
```bash
# Tokens auto-refresh, but if issues persist:
helper jira sheets-auth --force
```

### Browser doesn't open (SSH/Remote)
```bash
# For remote/SSH sessions, copy the URL manually:
helper jira sheets-auth --no-browser

# The command will print a URL like:
# https://accounts.google.com/o/oauth2/auth?...
# Copy and open in your local browser
```

### Rate limiting
```bash
# If hitting rate limits with default app:
# 1. Create custom OAuth app (Option 2 above)
# 2. Or add delays between requests:
helper jira config --set-rate-limit 2  # 2 requests per second
```

## API Scopes

The Helper CLI requests these Google API scopes:
- `https://www.googleapis.com/auth/spreadsheets` - Read/write access to sheets

## Advanced Configuration

### Environment Variables
```bash
# Override config file settings
export GOOGLE_SHEET_ID="your-sheet-id"
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"
```

### Programmatic Access
```python
from helper_cli.services.unified_google_auth import get_google_auth

# Get auth instance
auth = get_google_auth()

# Authenticate
auth.authenticate()

# Read data
data = auth.read_range(sheet_id, "Sheet1!A1:B10")

# Write data
auth.write_range(sheet_id, "Sheet1!A1", [["Hello", "World"]])
```

## Migration Guide

### From Service Account to OAuth
```bash
# Backup existing config
cp ~/.config/helper-cli/config.json ~/.config/helper-cli/config.backup.json

# Remove service account config
helper jira config --remove-service-account

# Set up OAuth
helper jira sheets-auth
```

### From Multiple OAuth Files
```bash
# The new unified auth automatically migrates from:
# - gspread_token.json
# - authorized_user.json
# - oauth_token.json

# Just run:
helper jira sheets-auth
```

## Support

For issues or questions:
1. Check the troubleshooting section above
2. Run diagnostics: `helper jira sheets-auth --diagnose`
3. Check logs: `~/.config/helper-cli/logs/auth.log`
4. Open an issue on GitHub with diagnostic output

## Summary

- **For personal use**: Just run `helper jira sheets-auth` - no setup needed!
- **For production**: Create custom OAuth app for better rate limits
- **For automation**: Use service account for browserless auth
- **Security**: Tokens are encrypted and auto-refresh
- **Simple**: One command to authenticate, works forever after

The Helper CLI makes Google Sheets integration as simple as possible while maintaining security and flexibility for advanced use cases.