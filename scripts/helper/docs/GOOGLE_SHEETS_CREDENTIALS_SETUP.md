# Google Sheets Service Account Setup Guide

## Quick Setup Steps

### 1. Create Google Cloud Project
1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Click "Select a project" → "New Project"
3. Name it (e.g., "Helper CLI Sheets")
4. Click "Create"

### 2. Enable Google Sheets API
1. In your project, go to "APIs & Services" → "Library"
2. Search for "Google Sheets API"
3. Click on it and press "Enable"

### 3. Create Service Account
1. Go to "APIs & Services" → "Credentials"
2. Click "Create Credentials" → "Service Account"
3. Fill in:
   - Service account name: `helper-cli-sheets`
   - Service account ID: (auto-fills)
   - Description: "Helper CLI Google Sheets integration"
4. Click "Create and Continue"
5. Skip the optional steps (roles and users)
6. Click "Done"

### 4. Generate Service Account Key
1. Click on the service account you just created
2. Go to the "Keys" tab
3. Click "Add Key" → "Create new key"
4. Choose "JSON" format
5. Click "Create"
6. **Save the downloaded JSON file** as:
   ```bash
   ~/.config/helper-cli/google-credentials.json
   ```

### 5. Share Your Google Sheet
1. Open your Google Sheet
2. Click "Share" button
3. Add the service account email (found in your JSON file as `client_email`)
4. Give it "Editor" permissions
5. Click "Send"

### 6. Test the Setup
```bash
# Your Sheet ID is already configured
helper jira sheets-tui
```

## Your Current Configuration

Based on your config file, you already have:
- **Sheet ID**: `1JwNOy90j7qM1gYFFhdGDnJ9uB_ShMmjHdQWyvX2nWhk`

You just need to:
1. Create the service account credentials (steps above)
2. Save them to `~/.config/helper-cli/google-credentials.json`
3. Share your sheet with the service account email

## Alternative: Use Existing Credentials

If you already have Google credentials from another project:

```bash
# Copy existing credentials
cp /path/to/existing/credentials.json ~/.config/helper-cli/google-credentials.json

# Or set environment variable
export GOOGLE_CREDENTIALS_PATH="/path/to/existing/credentials.json"
```

## Troubleshooting

### Error: "Google credentials not found"
```bash
# Check if file exists
ls -la ~/.config/helper-cli/google-credentials.json

# Create directory if needed
mkdir -p ~/.config/helper-cli
```

### Error: "Permission denied" from Google
- Make sure you shared the sheet with the service account email
- Check that the service account has "Editor" permissions
- Verify the Sheet ID matches your actual sheet

### Finding Your Sheet ID
Your Google Sheet URL looks like:
```
https://docs.google.com/spreadsheets/d/SHEET_ID_HERE/edit
```

The Sheet ID is the long string between `/d/` and `/edit`.

## Example Credentials File Structure

Your `google-credentials.json` should look like this:
```json
{
  "type": "service_account",
  "project_id": "your-project-id",
  "private_key_id": "...",
  "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
  "client_email": "helper-cli@your-project.iam.gserviceaccount.com",
  "client_id": "...",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "...",
  "client_x509_cert_url": "..."
}
```

## Security Notes

- **Never commit** the credentials JSON file to git
- Add to `.gitignore`: `google-credentials.json`
- Store securely with appropriate file permissions:
  ```bash
  chmod 600 ~/.config/helper-cli/google-credentials.json
  ```
