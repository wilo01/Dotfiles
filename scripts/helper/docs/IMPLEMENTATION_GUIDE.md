# Google Sheets TUI Integration - Implementation Guide

## Quick Start

This guide provides step-by-step instructions for implementing the Google Sheets TUI integration system.

## Prerequisites

- Python 3.8+
- Google Cloud Project with Sheets API enabled
- Jira API access
- Helper CLI installed

## Phase 1: Foundation Setup

### Step 1.1: Google Cloud Setup

1. **Create Google Cloud Project**
   ```bash
   # Visit: https://console.cloud.google.com
   # Create new project: "helper-cli-sheets"
   ```

2. **Enable Google Sheets API**
   ```bash
   # In Google Cloud Console:
   # APIs & Services → Enable APIs → Search "Google Sheets API" → Enable
   ```

3. **Create Service Account**
   ```bash
   # APIs & Services → Credentials → Create Credentials → Service Account
   # Download JSON key file to ~/.config/helper-cli/google-service-account.json
   ```

4. **Share Sheet with Service Account**
   ```
   # Copy service account email (e.g., helper-cli@project.iam.gserviceaccount.com)
   # Open your Google Sheet
   # Share → Add service account email with Editor permission
   ```

### Step 1.2: Project Structure

```
scripts/helper/
├── src/
│   └── helper_cli/
│       ├── services/
│       │   ├── google_sheets_service.py  # NEW
│       │   ├── sync_coordinator.py       # NEW
│       │   └── jira_monitor.py          # NEW
│       ├── tui/
│       │   ├── standup_app.py           # EXISTING
│       │   └── sheets_tui.py            # NEW
│       └── models/
│           └── sync_models.py           # NEW
├── docs/
│   ├── GOOGLE_SHEETS_TUI_INTEGRATION.md
│   └── IMPLEMENTATION_GUIDE.md
└── config/
    └── sheets-sync.yaml                 # NEW
```

### Step 1.3: Dependencies Installation

```bash
pip install google-api-python-client google-auth google-auth-oauthlib google-auth-httplib2
pip install aiohttp asyncio
pip install watchdog  # For file system monitoring
```

## Phase 2: Core Services Implementation

### Step 2.1: Google Sheets Service

```python
# src/helper_cli/services/google_sheets_service.py

from google.oauth2 import service_account
from googleapiclient.discovery import build
from typing import List, Dict, Any, Optional
import logging

class GoogleSheetsService:
    """Google Sheets API wrapper with caching and error handling."""

    def __init__(self, credentials_path: str, sheet_id: str):
        self.sheet_id = sheet_id
        self.logger = logging.getLogger(__name__)

        # Initialize credentials
        self.creds = service_account.Credentials.from_service_account_file(
            credentials_path,
            scopes=['https://www.googleapis.com/auth/spreadsheets']
        )

        # Build service
        self.service = build('sheets', 'v4', credentials=self.creds)
        self.sheets = self.service.spreadsheets()

    def read_range(self, range_name: str) -> List[List[Any]]:
        """Read data from specified range."""
        try:
            result = self.sheets.values().get(
                spreadsheetId=self.sheet_id,
                range=range_name
            ).execute()
            return result.get('values', [])
        except Exception as e:
            self.logger.error(f"Error reading range {range_name}: {e}")
            raise

    def write_cell(self, cell: str, value: Any) -> bool:
        """Write single value to cell."""
        try:
            body = {'values': [[value]]}
            self.sheets.values().update(
                spreadsheetId=self.sheet_id,
                range=cell,
                valueInputOption='USER_ENTERED',
                body=body
            ).execute()
            return True
        except Exception as e:
            self.logger.error(f"Error writing to cell {cell}: {e}")
            return False

    def batch_update(self, updates: List[Dict[str, Any]]) -> bool:
        """Perform batch update of multiple cells."""
        try:
            body = {
                'valueInputOption': 'USER_ENTERED',
                'data': updates
            }
            self.sheets.values().batchUpdate(
                spreadsheetId=self.sheet_id,
                body=body
            ).execute()
            return True
        except Exception as e:
            self.logger.error(f"Error in batch update: {e}")
            return False

    def append_row(self, values: List[Any], sheet_name: str = 'Sheet1') -> int:
        """Append new row to sheet."""
        try:
            body = {'values': [values]}
            result = self.sheets.values().append(
                spreadsheetId=self.sheet_id,
                range=f'{sheet_name}!A:Z',
                valueInputOption='USER_ENTERED',
                insertDataOption='INSERT_ROWS',
                body=body
            ).execute()

            # Extract row number from response
            updated_range = result.get('updates', {}).get('updatedRange', '')
            row_num = int(updated_range.split(':')[0].split('!')[-1][1:])
            return row_num
        except Exception as e:
            self.logger.error(f"Error appending row: {e}")
            return -1
```

### Step 2.2: Sync Coordinator

```python
# src/helper_cli/services/sync_coordinator.py

import asyncio
from datetime import datetime
from typing import List, Dict, Any, Optional
import sqlite3
import json
import logging
from enum import Enum

class SyncStatus(Enum):
    IDLE = "idle"
    SYNCING = "syncing"
    CONFLICT = "conflict"
    ERROR = "error"
    OFFLINE = "offline"

class SyncCoordinator:
    """Manages bidirectional sync between local cache and Google Sheets."""

    def __init__(self, sheets_service, db_path: str):
        self.sheets = sheets_service
        self.db_path = db_path
        self.logger = logging.getLogger(__name__)
        self.status = SyncStatus.IDLE
        self.sync_queue = []
        self._init_db()

    def _init_db(self):
        """Initialize sync metadata tables."""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()

        cursor.execute('''
            CREATE TABLE IF NOT EXISTS sync_queue (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                operation TEXT,
                cell_ref TEXT,
                value TEXT,
                timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
                status TEXT DEFAULT 'pending',
                retry_count INTEGER DEFAULT 0
            )
        ''')

        cursor.execute('''
            CREATE TABLE IF NOT EXISTS sync_metadata (
                key TEXT PRIMARY KEY,
                value TEXT,
                updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
            )
        ''')

        conn.commit()
        conn.close()

    async def start_sync_loop(self, interval: int = 30):
        """Start the sync loop that runs every interval seconds."""
        while True:
            try:
                await self.sync_cycle()
                await asyncio.sleep(interval)
            except Exception as e:
                self.logger.error(f"Sync cycle error: {e}")
                self.status = SyncStatus.ERROR
                await asyncio.sleep(interval * 2)  # Back off on error

    async def sync_cycle(self):
        """Perform one sync cycle."""
        self.status = SyncStatus.SYNCING

        # 1. Process offline queue
        await self.process_offline_queue()

        # 2. Pull changes from Sheets
        remote_changes = await self.pull_remote_changes()

        # 3. Get local changes
        local_changes = self.get_local_changes()

        # 4. Detect and resolve conflicts
        conflicts = self.detect_conflicts(local_changes, remote_changes)
        for conflict in conflicts:
            self.resolve_conflict(conflict)

        # 5. Apply changes
        await self.apply_remote_changes(remote_changes)
        await self.push_local_changes(local_changes)

        # 6. Update metadata
        self.update_sync_metadata()

        self.status = SyncStatus.IDLE

    def queue_change(self, operation: str, cell_ref: str, value: Any):
        """Queue a change for sync."""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()

        cursor.execute('''
            INSERT INTO sync_queue (operation, cell_ref, value)
            VALUES (?, ?, ?)
        ''', (operation, cell_ref, json.dumps(value)))

        conn.commit()
        conn.close()

    async def process_offline_queue(self):
        """Process queued changes when online."""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()

        cursor.execute('''
            SELECT id, operation, cell_ref, value
            FROM sync_queue
            WHERE status = 'pending'
            ORDER BY timestamp
        ''')

        for row in cursor.fetchall():
            queue_id, operation, cell_ref, value = row

            try:
                if operation == 'write':
                    success = self.sheets.write_cell(cell_ref, json.loads(value))
                    if success:
                        cursor.execute(
                            "UPDATE sync_queue SET status = 'completed' WHERE id = ?",
                            (queue_id,)
                        )
                    else:
                        cursor.execute(
                            "UPDATE sync_queue SET retry_count = retry_count + 1 WHERE id = ?",
                            (queue_id,)
                        )
            except Exception as e:
                self.logger.error(f"Error processing queue item {queue_id}: {e}")

        conn.commit()
        conn.close()

    def detect_conflicts(self, local: List, remote: List) -> List:
        """Detect conflicts between local and remote changes."""
        conflicts = []

        for local_change in local:
            for remote_change in remote:
                if local_change['cell'] == remote_change['cell']:
                    if local_change['timestamp'] != remote_change['timestamp']:
                        conflicts.append({
                            'cell': local_change['cell'],
                            'local': local_change,
                            'remote': remote_change
                        })

        return conflicts

    def resolve_conflict(self, conflict: Dict):
        """Resolve a sync conflict using configured strategy."""
        strategy = self.get_config('conflict_resolution', 'last-write-wins')

        if strategy == 'last-write-wins':
            if conflict['local']['timestamp'] > conflict['remote']['timestamp']:
                return conflict['local']
            else:
                return conflict['remote']
        elif strategy == 'remote-wins':
            return conflict['remote']
        elif strategy == 'local-wins':
            return conflict['local']
        else:
            # Manual resolution required
            self.status = SyncStatus.CONFLICT
            return None
```

### Step 2.3: Jira Activity Monitor

```python
# src/helper_cli/services/jira_monitor.py

import asyncio
from datetime import datetime, timedelta
from typing import List, Dict, Any
import logging

class JiraActivityMonitor:
    """Monitors Jira for activity and auto-populates standup sheets."""

    def __init__(self, jira_service, sheets_service, sync_coordinator):
        self.jira = jira_service
        self.sheets = sheets_service
        self.sync = sync_coordinator
        self.logger = logging.getLogger(__name__)
        self.last_check = datetime.now()

    async def start_monitoring(self, interval: int = 300):
        """Start monitoring Jira activities."""
        while True:
            try:
                await self.check_activities()
                await asyncio.sleep(interval)
            except Exception as e:
                self.logger.error(f"Monitoring error: {e}")
                await asyncio.sleep(interval * 2)

    async def check_activities(self):
        """Check for new Jira activities."""
        # Get activities since last check
        activities = self.get_recent_activities()

        for activity in activities:
            if self.should_track(activity):
                await self.add_to_sheet(activity)

        self.last_check = datetime.now()

    def get_recent_activities(self) -> List[Dict]:
        """Get Jira activities since last check."""
        activities = []

        # Check status changes
        jql = f"""
            assignee = currentUser() AND
            updated >= -{int((datetime.now() - self.last_check).seconds / 60)}m
        """

        tickets = self.jira.search_tickets(jql)

        for ticket in tickets:
            # Check changelog for status changes
            changelog = self.get_ticket_changelog(ticket['key'])

            for change in changelog:
                if change['field'] == 'status':
                    activities.append({
                        'type': 'status_change',
                        'ticket': ticket['key'],
                        'summary': ticket['fields']['summary'],
                        'from_status': change['from'],
                        'to_status': change['to'],
                        'timestamp': change['created']
                    })

        return activities

    def should_track(self, activity: Dict) -> bool:
        """Determine if activity should be tracked."""
        tracked_types = ['status_change', 'comment', 'time_log']

        if activity['type'] not in tracked_types:
            return False

        # Don't track minor status changes
        if activity['type'] == 'status_change':
            minor_statuses = ['Reopened', 'On Hold']
            if activity['to_status'] in minor_statuses:
                return False

        return True

    async def add_to_sheet(self, activity: Dict):
        """Add activity to Google Sheet."""
        today = datetime.now().strftime('%Y-%m-%d')

        # Format activity for sheet
        row_data = [
            today,
            activity['ticket'],
            activity['summary'],
            activity.get('to_status', ''),
            self.format_activity_description(activity),
            '',  # Notes (empty)
            datetime.now().isoformat(),
            'Jira Bot',
            'synced',
            1  # Version
        ]

        # Add to sheet
        row_num = self.sheets.append_row(row_data)

        if row_num > 0:
            self.logger.info(f"Added activity to row {row_num}: {activity['ticket']}")

    def format_activity_description(self, activity: Dict) -> str:
        """Format activity into readable description."""
        if activity['type'] == 'status_change':
            return f"Status changed from {activity['from_status']} to {activity['to_status']}"
        elif activity['type'] == 'comment':
            return f"Added comment: {activity['comment'][:50]}..."
        elif activity['type'] == 'time_log':
            return f"Logged {activity['duration']} - {activity['description']}"
        else:
            return "Activity tracked"
```

## Phase 3: Enhanced TUI Implementation

### Step 3.1: Sheets TUI Application

```python
# src/helper_cli/tui/sheets_tui.py

from textual.app import App, ComposeResult
from textual.widgets import DataTable, Header, Footer, Static, Label
from textual.containers import Container, Vertical
from textual.binding import Binding
from textual.reactive import reactive
from textual import events
import asyncio
from datetime import datetime

class SheetsTUI(App):
    """Google Sheets integrated TUI for standup notes."""

    CSS = """
    SheetsTUI {
        background: $background;
    }

    #sync-status {
        dock: top;
        height: 1;
        background: $success;
        color: $text;
        text-align: center;
    }

    #sync-status.syncing {
        background: $warning;
    }

    #sync-status.error {
        background: $error;
    }

    DataTable {
        height: 1fr;
    }

    .changed-cell {
        background: $warning 20%;
    }
    """

    BINDINGS = [
        Binding("ctrl+s", "force_sync", "Force Sync"),
        Binding("ctrl+r", "refresh", "Refresh"),
        Binding("ctrl+z", "undo", "Undo"),
        Binding("ctrl+y", "redo", "Redo"),
    ]

    sync_status = reactive("idle")

    def __init__(self, sheets_service, sync_coordinator):
        super().__init__()
        self.sheets = sheets_service
        self.sync = sync_coordinator
        self.undo_stack = []
        self.redo_stack = []
        self.sheet_data = []
        self.changed_cells = set()

    def compose(self) -> ComposeResult:
        yield Header()
        yield Static("⚡ Connected to Google Sheets", id="sync-status")
        yield DataTable(cursor_type="cell", zebra_stripes=True)
        yield Footer()

    async def on_mount(self):
        """Initialize the application."""
        # Load sheet data
        await self.load_sheet_data()

        # Start sync monitoring
        asyncio.create_task(self.sync_monitor())

        # Setup table
        self.setup_table()

    async def load_sheet_data(self):
        """Load data from Google Sheets."""
        self.sheet_data = self.sheets.read_range("A1:J100")

    def setup_table(self):
        """Setup the data table with sheet data."""
        table = self.query_one(DataTable)

        if not self.sheet_data:
            return

        # Add columns from first row
        headers = self.sheet_data[0]
        for header in headers:
            table.add_column(header)

        # Add data rows
        for row_data in self.sheet_data[1:]:
            table.add_row(*row_data)

    async def sync_monitor(self):
        """Monitor sync status and update UI."""
        while True:
            # Update sync status indicator
            status_widget = self.query_one("#sync-status", Static)

            if self.sync.status.value == "syncing":
                status_widget.update("🔄 Syncing...")
                status_widget.add_class("syncing")
            elif self.sync.status.value == "error":
                status_widget.update("❌ Sync Error")
                status_widget.add_class("error")
            else:
                last_sync = datetime.now().strftime("%H:%M:%S")
                status_widget.update(f"✅ Last sync: {last_sync}")
                status_widget.remove_class("syncing", "error")

            # Check for remote changes
            await self.check_remote_changes()

            await asyncio.sleep(5)

    async def check_remote_changes(self):
        """Check for changes in Google Sheets."""
        new_data = self.sheets.read_range("A1:J100")

        if new_data != self.sheet_data:
            # Find changed cells
            for i, row in enumerate(new_data):
                if i < len(self.sheet_data):
                    for j, cell in enumerate(row):
                        if j < len(self.sheet_data[i]):
                            if cell != self.sheet_data[i][j]:
                                self.changed_cells.add((i, j))

            self.sheet_data = new_data
            self.refresh_table()

    def on_data_table_cell_updated(self, event):
        """Handle cell updates."""
        # Save to undo stack
        self.undo_stack.append({
            'cell': event.coordinate,
            'old_value': event.old_value,
            'new_value': event.value
        })

        # Queue sync
        row = event.coordinate.row + 2  # Account for header
        col = chr(65 + event.coordinate.column)  # Convert to column letter
        cell_ref = f"{col}{row}"

        self.sync.queue_change("write", cell_ref, event.value)

    def action_force_sync(self):
        """Force immediate sync."""
        asyncio.create_task(self.sync.sync_cycle())

    def action_undo(self):
        """Undo last change."""
        if self.undo_stack:
            change = self.undo_stack.pop()
            self.redo_stack.append(change)

            # Revert cell
            table = self.query_one(DataTable)
            table.update_cell_at(change['cell'], change['old_value'])
```

## Phase 4: Configuration & Deployment

### Step 4.1: Configuration File

```yaml
# config/sheets-sync.yaml

google:
  service_account_file: "~/.config/helper-cli/google-service-account.json"
  sheet_id: "1JwNOy90j7qM1gYFFhdGDnJ9uB_ShMmjHdQWyvX2nWhk"
  worksheet_name: "Daily Standup"

sync:
  interval_seconds: 30
  conflict_resolution: "last-write-wins"
  enable_offline_queue: true
  max_retry_attempts: 3

jira:
  monitor_enabled: true
  check_interval_seconds: 300
  auto_add_activities: true
  tracked_events:
    - status_change
    - comment_added
    - time_logged

tui:
  auto_refresh: true
  show_sync_indicator: true
  enable_undo_redo: true
  max_undo_history: 50

logging:
  level: INFO
  file: "~/.logs/helper-cli/sheets-sync.log"
```

### Step 4.2: CLI Integration

```python
# Update cli.py

@jira_group.command('sheets-standup')
@click.option('--config', default='~/.config/helper-cli/sheets-sync.yaml')
def sheets_standup(config):
    """Launch Google Sheets integrated standup TUI."""
    from helper_cli.services.google_sheets_service import GoogleSheetsService
    from helper_cli.services.sync_coordinator import SyncCoordinator
    from helper_cli.services.jira_monitor import JiraActivityMonitor
    from helper_cli.tui.sheets_tui import SheetsTUI

    # Load configuration
    import yaml
    with open(os.path.expanduser(config)) as f:
        cfg = yaml.safe_load(f)

    # Initialize services
    sheets_service = GoogleSheetsService(
        credentials_path=os.path.expanduser(cfg['google']['service_account_file']),
        sheet_id=cfg['google']['sheet_id']
    )

    sync_coordinator = SyncCoordinator(
        sheets_service=sheets_service,
        db_path=os.path.expanduser("~/.cache/helper-cli/sheets-sync.db")
    )

    # Initialize TUI
    app = SheetsTUI(sheets_service, sync_coordinator)

    # Start background services
    import asyncio
    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)

    # Start sync loop
    loop.create_task(sync_coordinator.start_sync_loop(cfg['sync']['interval_seconds']))

    # Start Jira monitor if enabled
    if cfg['jira']['monitor_enabled']:
        jira_monitor = JiraActivityMonitor(
            jira_service=get_jira_service(),
            sheets_service=sheets_service,
            sync_coordinator=sync_coordinator
        )
        loop.create_task(jira_monitor.start_monitoring(cfg['jira']['check_interval_seconds']))

    # Run TUI
    app.run()
```

## Phase 5: Testing & Validation

### Step 5.1: Unit Tests

```python
# tests/test_sheets_sync.py

import pytest
from unittest.mock import Mock, patch
from helper_cli.services.google_sheets_service import GoogleSheetsService
from helper_cli.services.sync_coordinator import SyncCoordinator

class TestGoogleSheetsService:
    def test_read_range(self):
        """Test reading sheet range."""
        service = Mock()
        sheets = GoogleSheetsService("test.json", "sheet123")
        sheets.service = service

        # Mock response
        service.spreadsheets().values().get().execute.return_value = {
            'values': [['A1', 'B1'], ['A2', 'B2']]
        }

        result = sheets.read_range("A1:B2")
        assert result == [['A1', 'B1'], ['A2', 'B2']]

    def test_write_cell(self):
        """Test writing to cell."""
        # Test implementation
        pass

class TestSyncCoordinator:
    def test_conflict_resolution(self):
        """Test conflict resolution logic."""
        sync = SyncCoordinator(Mock(), ":memory:")

        conflict = {
            'local': {'timestamp': '2025-09-06T10:00:00', 'value': 'local'},
            'remote': {'timestamp': '2025-09-06T09:00:00', 'value': 'remote'}
        }

        result = sync.resolve_conflict(conflict)
        assert result == conflict['local']  # Local is newer
```

### Step 5.2: Integration Tests

```python
# tests/test_integration.py

import asyncio
import pytest
from helper_cli.services.google_sheets_service import GoogleSheetsService
from helper_cli.services.sync_coordinator import SyncCoordinator

@pytest.mark.asyncio
async def test_full_sync_cycle():
    """Test complete sync cycle."""
    # Setup test sheet
    sheets = GoogleSheetsService("test.json", "test_sheet")
    sync = SyncCoordinator(sheets, ":memory:")

    # Queue some changes
    sync.queue_change("write", "A1", "Test Value")

    # Run sync cycle
    await sync.sync_cycle()

    # Verify changes were applied
    assert sync.status.value == "idle"
```

## Deployment Checklist

- [ ] Google Cloud Project created
- [ ] Sheets API enabled
- [ ] Service Account created and key downloaded
- [ ] Sheet shared with Service Account
- [ ] Dependencies installed
- [ ] Configuration file created
- [ ] Unit tests passing
- [ ] Integration tests passing
- [ ] Documentation updated
- [ ] Backup of existing data created

## Troubleshooting

### Common Issues

1. **Authentication Errors**
   - Verify service account key file path
   - Check sheet is shared with service account email
   - Ensure Sheets API is enabled

2. **Sync Conflicts**
   - Check conflict resolution strategy in config
   - Review sync logs for conflict details
   - Manually resolve if needed

3. **Performance Issues**
   - Reduce sync interval if too frequent
   - Implement pagination for large sheets
   - Use batch operations instead of single cell updates

## Next Steps

1. Monitor system performance
2. Gather user feedback
3. Implement additional features:
   - Formula support
   - Conditional formatting
   - Multi-sheet support
   - Real-time collaboration indicators

---

*Implementation Guide v1.0 - September 2025*
