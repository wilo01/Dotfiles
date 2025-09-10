# Quickest Google Sheets OAuth Setup (5 minutes!) 🚀

## The Reality
Google requires OAuth credentials for Sheets API access. But we've made it as simple as possible!

## Quick Setup Steps

### Step 1: Get OAuth Credentials (2 minutes)

1. **Click this link**: [Google Cloud Console - Create OAuth Client](https://console.cloud.google.com/apis/credentials/oauthclient)
   
2. **If prompted to create a project**: 
   - Click "Create Project"
   - Name: "Helper CLI" (or anything)
   - Click "Create"

3. **Enable Sheets API**:
   - [Click here to enable Google Sheets API](https://console.cloud.google.com/apis/library/sheets.googleapis.com)
   - Click "Enable"

4. **Create OAuth Client**:
   - Go back to [Create OAuth Client](https://console.cloud.google.com/apis/credentials/oauthclient)
   - **Application type**: Desktop app
   - **Name**: Helper CLI
   - Click "Create"
   - Click "Download JSON"

5. **Save the file**:
```bash
# Move downloaded file to the right location
mv ~/Downloads/client_secret_*.json ~/.config/helper-cli/oauth_credentials.json
```

### Step 2: Authenticate (30 seconds)

```bash
# Run authentication
helper jira sheets-auth

# Your browser opens → Sign in → Click "Allow" → Done!
```

### Step 3: Use the TUI

```bash
helper jira sheets-tui
```

## That's It! 🎉

You're done! The token is saved and you'll never need to authenticate again (unless you revoke it).

## Direct Links for Setup

1. [Create New Project](https://console.cloud.google.com/projectcreate) (if needed)
2. [Enable Sheets API](https://console.cloud.google.com/apis/library/sheets.googleapis.com)
3. [Create OAuth Credentials](https://console.cloud.google.com/apis/credentials/oauthclient)

## Why OAuth is Actually Better

| OAuth | Service Account |
|-------|----------------|
| ✅ Uses YOUR account | ❌ Uses robot account |
| ✅ No sharing needed | ❌ Must share sheets |
| ✅ See sheets in your Drive | ❌ Hidden from Drive |
| ✅ One-time browser auth | ❌ Manage JSON keys |

## Commands Reference

```bash
# First time setup
helper jira sheets-auth

# Check status
helper jira sheets-auth --status

# Use the TUI
helper jira sheets-tui

# Force re-auth (if needed)
helper jira sheets-auth --force

# Revoke access
helper jira sheets-auth --revoke
```

## File Locations

```
~/.config/helper-cli/
├── oauth_credentials.json   # OAuth app credentials (one-time download)
├── gspread_token.json      # Your refresh token (auto-created)
└── jira_config.json        # Your sheet ID and settings
```

## Troubleshooting

**"No such file or directory: credentials.json"**
- You need to download OAuth credentials first (Step 1 above)

**"Authentication failed"**
- Make sure browser opened and you clicked "Allow"
- Try `helper jira sheets-auth --force`

**"Can't access sheet"**
- Check Sheet ID in config: `~/.config/helper-cli/jira_config.json`
- Make sure you're signed in with the right Google account

## Alternative: Use Google Colab OAuth (Experimental)

If you have Google Colab, you can try using their OAuth:
```python
from google.colab import auth
auth.authenticate_user()
```
But this requires running in Colab environment.

## The Simplest Path

1. Click the 3 links above to set up OAuth (2 minutes)
2. Download JSON and save to `~/.config/helper-cli/oauth_credentials.json`
3. Run `helper jira sheets-auth`
4. Click "Allow" in browser
5. Never worry about auth again!

Total time: ~5 minutes, one-time setup! 🚀