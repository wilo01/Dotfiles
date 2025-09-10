# Ultra-Simple OAuth2 Setup for Google Sheets 🚀

## The SIMPLEST Way - No Google Cloud Console! 

### Option 1: Use gspread's Default OAuth (Recommended) ✨

**Zero setup required!** Just run:

```bash
# Authenticate (opens browser)
helper jira sheets-auth

# That's it! Your browser will open:
# 1. Sign in with Google
# 2. Click "Allow" 
# 3. Done forever!
```

### Option 2: Use Your Own OAuth App (Optional)

If you want more control, you can create your own OAuth app:

1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Create/select a project
3. Enable Google Sheets API
4. Create OAuth 2.0 credentials (Desktop app)
5. Download and save as: `~/.config/helper-cli/oauth_credentials.json`
6. Run `helper jira sheets-auth`

## How It Works

The ultra-simple approach uses `gspread.oauth()` which:
- ✅ Uses Google's test OAuth app by default (no setup!)
- ✅ Opens browser automatically
- ✅ Saves refresh token (one-time auth)
- ✅ Auto-refreshes when needed
- ✅ Works with your personal Google account

## Commands

```bash
# Authenticate
helper jira sheets-auth

# Check status
helper jira sheets-auth --status

# Force re-authentication
helper jira sheets-auth --force

# Revoke access
helper jira sheets-auth --revoke

# Use the TUI
helper jira sheets-tui
```

## File Locations

```
~/.config/helper-cli/
├── gspread_token.json       # Your auth token (auto-created)
└── oauth_credentials.json   # Optional custom OAuth app
```

## Comparison

| Method | Setup Required | Complexity |
|--------|---------------|------------|
| gspread default | None! | ⭐ |
| Custom OAuth | Create app in Google Cloud | ⭐⭐ |
| Service Account | Create service account + share sheets | ⭐⭐⭐ |

## Troubleshooting

### "Authentication failed"
- Make sure you have internet connection
- Check if browser opened
- Try `helper jira sheets-auth --force`

### "Can't access sheet"
- Verify Sheet ID in config
- Make sure you're using the right Google account
- Sheet must exist and you must have access

### Browser didn't open
- Check if you're in SSH/remote session
- May need to copy the URL manually

## Why This Is Better

Traditional OAuth2 setup requires:
- ❌ Creating a Google Cloud project
- ❌ Enabling APIs
- ❌ Creating OAuth credentials
- ❌ Downloading JSON files
- ❌ Complex configuration

With gspread's default OAuth:
- ✅ Just run the command
- ✅ Click "Allow" in browser
- ✅ Done!

## Security Notes

- Token is stored locally in `~/.config/helper-cli/`
- File permissions set to 600 (owner only)
- Token can be revoked anytime
- No sensitive keys to manage

## Next Steps

After authentication:
1. Make sure your Sheet ID is in the config
2. Run `helper jira sheets-tui` to start using the TUI
3. Edit cells directly in the TUI
4. Changes sync automatically to Google Sheets

Enjoy the simplest Google Sheets integration ever! 🎉