# OAuth2 Setup for Google Sheets - The Simple Way! 🎉

## Why OAuth2 is Better Than Service Account

| Feature | OAuth2 (Recommended) | Service Account |
|---------|---------------------|-----------------|
| Setup Complexity | ⭐ Simple | ⭐⭐⭐ Complex |
| GCP Project | Optional (can use default) | Required |
| Sheet Sharing | Uses YOUR account | Must share with service account |
| Authentication | One-time browser login | JSON key file management |
| Security | Refresh token (safer) | Private key (sensitive) |
| Best For | Personal use | Server/automation |

## Quick Start (5 Minutes!)

### Step 1: Get OAuth2 Credentials
```bash
# Check current auth status
helper jira sheets-auth --status
```

If you don't have OAuth credentials yet:

1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Select or create a project (or use the default)
3. Enable Google Sheets API:
   - Go to "APIs & Services" → "Library"
   - Search "Google Sheets API"
   - Click "Enable"

4. Create OAuth2 credentials:
   - Go to "APIs & Services" → "Credentials"
   - Click "Create Credentials" → "OAuth 2.0 Client ID"
   - Configure consent screen if needed (just basic info)
   - Application type: **Desktop app**
   - Name: "Helper CLI"
   - Download the JSON file

5. Save the credentials:
```bash
# Save to the expected location
mv ~/Downloads/client_secret_*.json ~/.config/helper-cli/oauth_credentials.json
```

### Step 2: Authenticate Once
```bash
# Run the auth command
helper jira sheets-auth

# Your browser will open for one-time authorization
# After this, you never need to do it again!
```

### Step 3: Use the TUI
```bash
# Launch with OAuth2 (default)
helper jira sheets-tui

# Or explicitly use OAuth2
helper jira sheets-tui --use-oauth
```

## That's It! 🚀

No more:
- ❌ Service account creation
- ❌ Sharing sheets with weird emails
- ❌ Managing private keys
- ❌ Complex GCP setup

Just:
- ✅ Your Google account
- ✅ One-time browser auth
- ✅ Works forever with refresh token

## Advanced Usage

### Force Re-authentication
```bash
# If you need to switch Google accounts
helper jira sheets-auth --force
```

### Revoke Access
```bash
# Remove stored token
helper jira sheets-auth --revoke
```

### Check Status
```bash
# See current auth status
helper jira sheets-auth --status
```

## File Locations

OAuth2 uses these files:
```
~/.config/helper-cli/
├── oauth_credentials.json  # OAuth2 client ID (from Google)
└── token.json             # Your refresh token (created after auth)
```

## Troubleshooting

### "OAuth credentials not found"
Download OAuth2 credentials from Google Cloud Console and save as:
```bash
~/.config/helper-cli/oauth_credentials.json
```

### "Token expired" or authentication errors
```bash
# Force re-authentication
helper jira sheets-auth --force
```

### "Cannot access sheet"
- Make sure the sheet exists
- You're using the correct Google account
- The Sheet ID in config is correct

## Security Notes

- `token.json` contains your refresh token - keep it secure
- File is automatically set to 600 permissions (owner-only)
- Never commit these files to git
- Token can be revoked anytime from Google Account settings

## Comparison with Service Account

### Service Account (Old Way - Complex)
1. Create GCP project ❌
2. Enable APIs ❌
3. Create service account ❌
4. Generate private key ❌
5. Download JSON key ❌
6. Share sheet with service account email ❌
7. Manage sensitive key file ❌

### OAuth2 (New Way - Simple)
1. Download OAuth credentials ✅
2. Run `helper jira sheets-auth` ✅
3. Click "Allow" in browser ✅
4. Done forever! ✅

## Migration from Service Account

If you were using Service Account before:

1. Set up OAuth2 as described above
2. Run: `helper jira sheets-auth`
3. Use: `helper jira sheets-tui --use-oauth`
4. Once working, you can delete the service account credentials

## Why We Added OAuth2

Based on user feedback:
- Service Account setup is too complex for personal tools
- Managing GCP projects is overkill for single-user CLI
- Sharing sheets with service accounts is confusing
- OAuth2 is how most Google tools work (gsheets, gdrive, etc.)

## Next Steps

Now that you have OAuth2 set up:
- Use `helper jira sheets-tui` to manage your daily schedule
- Sync with `helper jira log` for time tracking
- Auto-populate from Jira with keyboard shortcuts

Enjoy the simpler authentication! 🎉