# Helper CLI

Comprehensive development helper CLI with Jira workflow automation, time tracking, Google Sheets integration, and various productivity tools.

## Installation

Install in development mode:

```bash
cd ~/.Dotfiles/scripts/helper
pip install -e .
```

## Features

### 🎫 Jira Workflow Automation
- **Time Tracking**: Log work time with offline caching and automatic sync
- **Ticket Management**: Fetch, search, and filter tickets with JQL support
- **Sprint Tracking**: View current sprint status and progress
- **Standup Notes**: Generate daily standup notes from work logs
- **Google Sheets Integration**: Sync work logs and schedules
- **Offline Mode**: Work offline with automatic sync when connected

### 🔧 Development Tools
- **Branch Names**: Generate Jira-compliant branch names from commit messages
- **Badge Automation**: RFID/QR badge generation with auto-typing
- **Git Helpers**: Stash commands, PR titles with ticket formatting
- **URL Tools**: QR code generation for ngrok/VSCode URLs
- **Name Checking**: Domain, npm, GitHub availability checker

## Usage

### Jira Workflow Commands

```bash
# Configuration (run once)
helper jira config --interactive

# Fetch assigned tickets
helper jira fetch                              # Current sprint tickets
helper jira fetch --jql "priority = High"      # Custom JQL query
helper jira fetch --sprint all --limit 50      # All tickets, max 50

# Log work time
helper jira log                                # Interactive mode
helper jira log 2h "Implemented feature"       # Quick log with auto-detect ticket
helper jira log 1h30m "Code review" --ticket VIS-4703

# View work history
helper jira recent                             # Last 7 days, grouped by day
helper jira recent --days 30 --group-by ticket # Last 30 days by ticket
helper jira today                              # Today's schedule and logged work

# Sync and standup
helper jira sync                               # Sync offline logs to Jira
helper jira standup                            # Generate standup notes
helper jira sprint                             # View current sprint info

# Start work session
helper jira start                              # Auto-detect ticket from git
helper jira start --ticket VIS-4703            # Specific ticket
```

### Development Helper Commands

```bash
# Branch name generation (formerly 'helper jira')
helper branch "Fix user authentication issue"  # Generates: fix-user-authentication-issue

# Badge automation
helper badge "@4324325553#"                    # Type badge string
helper rfid 4324325553                         # RFID badge
helper qr 5678901234                           # QR badge

# Git helpers
helper stash "work in progress"                # Generate stash command
helper pr "vis-4703 implement new feature"     # Format PR title

# URL tools
helper ngrok "https://abc123.ngrok.io"         # Generate QR for ngrok URL
helper vsc "https://vscode.dev/..."            # Generate QR for VSCode URL

# Name/Domain checking
helper check-domain example.com                # Check single domain
helper check-domains myproject                 # Check all TLDs
helper check-npm my-package                    # Check npm availability
helper check-github username                   # Check GitHub username
helper name-finder project1 project2           # Comprehensive availability check

# Other utilities
helper dash "Convert to Dash Format"           # convert-to-dash-format
helper filename "Create Clean Filename"        # create-clean-filename.md
helper log                                     # Create work log for VIS tickets
helper apex file.zip 8090 auth-token           # APEX upload command
helper rt "directory-name"                     # Setup RT directory
```

### Interactive Mode

```bash
helper interactive                              # Enter interactive shell
```

## Configuration

### Jira Setup

1. Get your Jira API token from: https://id.atlassian.com/manage-profile/security/api-tokens
2. Run configuration:
   ```bash
   helper jira config --interactive
   ```
3. Enter your Jira URL, email, and API token

### Google Sheets Setup (Optional)

1. Create a Google Sheets document for daily logs
2. Set up service account credentials (see Google Cloud Console)
3. Configure during `helper jira config --interactive`

### Environment Variables

You can also set credentials via environment:
```bash
export JIRA_URL="https://company.atlassian.net"
export JIRA_EMAIL="your-email@company.com"
export JIRA_API_TOKEN="your-api-token"
```

## Offline Mode

The Jira integration works offline by default:
- Work logs are cached locally when offline
- Automatic sync when connection is restored
- View cached tickets and logs anytime
- Manual sync with `helper jira sync`

## Examples

### Daily Workflow

```bash
# Morning - Start work
helper jira start                     # Auto-detect ticket from git branch
helper jira fetch                     # Get latest tickets
helper jira today                     # View today's plan

# During work - Log time
helper jira log 2h "Implemented user auth"
helper jira log 1h "Code review"

# End of day - Prepare for tomorrow
helper jira sync                      # Sync all work to Jira
helper jira standup --copy            # Generate standup notes

# Check weekly progress
helper jira recent --group-by ticket  # See time per ticket
helper jira status                    # Today's summary
```

### Sprint Management

```bash
# View sprint status
helper jira sprint

# Get all sprint tickets
helper jira fetch --sprint current

# Search specific tickets
helper jira fetch --jql "status = 'In Progress' AND assignee = currentUser()"
```

## Features in Detail

- ⚡ **Offline First**: All commands work offline with automatic sync
- 🔄 **Smart Sync**: Intelligent conflict resolution and retry logic
- 📊 **Google Sheets**: Automatic daily log synchronization
- 🎯 **Auto-Detection**: Detects tickets from git branches
- 📅 **Time Tracking**: Flexible duration formats (2h, 30m, 1h30m)
- 🏃 **Sprint Support**: Current sprint focus with progress tracking
- 📝 **Standup Ready**: Generate formatted standup notes instantly
- 🔍 **Powerful Search**: Full JQL support for ticket queries
- 📋 **Rich Output**: Beautiful terminal tables and formatting
- 🚀 **Fast & Efficient**: Caching and batch operations
