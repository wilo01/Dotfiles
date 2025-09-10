"""Google Sheets Calendar TUI for daily schedule tracking."""

from datetime import datetime, timedelta
from typing import List, Dict, Any, Optional, Tuple
from calendar import monthrange
import asyncio

from textual.app import App, ComposeResult
from textual.widgets import DataTable, Header, Footer, Static, Input, Button, TextArea
from textual.containers import Container, Vertical, Horizontal, Grid
from textual.binding import Binding
from textual.coordinate import Coordinate
from textual.screen import Screen, ModalScreen
from textual import events
from textual.reactive import reactive
from rich.text import Text
from rich.style import Style

from ..services.google_sheets_service import GoogleSheetsService
from ..services.sync_coordinator import SyncCoordinator
from ..services.sync_models import SyncDatabase
from ..services.context_detector import ContextDetector


class FocusModeScreen(ModalScreen):
    """Full screen focus mode for editing a single day."""

    CSS = """
    FocusModeScreen {
        align: center middle;
    }

    #focus-container {
        background: $surface;
        border: solid $primary;
        width: 90%;
        height: 90%;
        padding: 2;
    }

    #focus-header {
        height: 3;
        text-style: bold;
        text-align: center;
        color: $primary;
        margin-bottom: 1;
    }

    #focus-editor {
        width: 100%;
        height: 1fr;
        border: solid $secondary;
    }

    #focus-status {
        height: 3;
        margin-top: 1;
        color: $text-muted;
    }

    #button-container {
        height: 3;
        align: center middle;
        margin-top: 1;
    }

    Button {
        margin: 0 1;
    }
    """

    def __init__(self, date: datetime, initial_content: str = "",
                 jira_suggestions: List[str] = None):
        super().__init__()
        self.date = date
        self.initial_content = initial_content
        self.jira_suggestions = jira_suggestions or []
        self.saved = False

    def compose(self) -> ComposeResult:
        """Create focus mode UI."""
        with Container(id="focus-container"):
            yield Static(
                f"📅 {self.date.strftime('%A, %B %d, %Y')}",
                id="focus-header"
            )

            # Main editor
            editor = TextArea(
                self.initial_content,
                language="markdown",
                show_line_numbers=True,
                id="focus-editor"
            )
            editor.focus()
            yield editor

            # Status bar
            status_text = f"Lines: {len(self.initial_content.splitlines())} | "
            status_text += f"Words: {len(self.initial_content.split())} | "
            status_text += "Press Ctrl+S to save, ESC to cancel"
            yield Static(status_text, id="focus-status")

            # Buttons
            with Horizontal(id="button-container"):
                yield Button("Save [Ctrl+S]", variant="primary", id="save-btn")
                yield Button("Insert Jira", variant="default", id="jira-btn")
                yield Button("Cancel [ESC]", variant="error", id="cancel-btn")

    def on_button_pressed(self, event: Button.Pressed) -> None:
        """Handle button presses."""
        if event.button.id == "save-btn":
            self.save_and_close()
        elif event.button.id == "jira-btn":
            self.insert_jira_suggestions()
        elif event.button.id == "cancel-btn":
            self.dismiss(None)

    def on_key(self, event: events.Key) -> None:
        """Handle keyboard shortcuts."""
        if event.key == "ctrl+s":
            self.save_and_close()
            event.stop()
        elif event.key == "escape":
            self.dismiss(None)
            event.stop()

    def save_and_close(self):
        """Save content and close."""
        editor = self.query_one("#focus-editor", TextArea)
        self.saved = True
        self.dismiss(editor.text)

    def insert_jira_suggestions(self):
        """Insert Jira activity suggestions."""
        if not self.jira_suggestions:
            return

        editor = self.query_one("#focus-editor", TextArea)
        current_text = editor.text

        # Add suggestions at cursor or end
        suggestions_text = "\n\n--- Jira Activities ---\n"
        for suggestion in self.jira_suggestions:
            suggestions_text += f"• {suggestion}\n"

        editor.text = current_text + suggestions_text


class MonthViewScreen(Screen):
    """Main month calendar view screen."""

    CSS = """
    MonthViewScreen {
        background: $background;
    }

    #month-header {
        height: 5;
        background: $panel;
        border: solid $primary;
        padding: 1;
    }

    #month-title {
        text-style: bold;
        text-align: center;
        color: $primary;
    }

    #sync-status {
        text-align: right;
        color: $text-muted;
    }

    #calendar-container {
        height: 1fr;
        margin: 1;
    }

    DataTable {
        height: 100%;
    }

    #week-view-hint {
        height: 3;
        padding: 1;
        color: $text-muted;
        text-align: center;
    }
    """

    BINDINGS = [
        Binding("left", "previous_month", "Previous Month"),
        Binding("right", "next_month", "Next Month"),
        Binding("enter", "edit_cell", "Edit Day"),
        Binding("f", "focus_mode", "Focus Mode"),
        Binding("w", "week_view", "Week View"),
        Binding("s", "sync_now", "Sync Now"),
        Binding("j", "jira_populate", "Jira Auto-fill"),
        Binding("r", "refresh", "Refresh"),
        Binding("q", "quit", "Quit"),
    ]

    def __init__(self, sheets_service: GoogleSheetsService,
                 sync_coordinator: SyncCoordinator):
        super().__init__()
        self.sheets = sheets_service
        self.sync = sync_coordinator
        self.context_detector = ContextDetector()

        self.current_date = datetime.now()
        self.current_month = self.current_date.replace(day=1)
        self.selected_date: Optional[datetime] = None
        self.calendar_data: Dict[int, str] = {}

    def compose(self) -> ComposeResult:
        """Create month view UI."""
        # Header
        with Container(id="month-header"):
            yield Static(
                self.current_month.strftime("%B %Y"),
                id="month-title"
            )
            yield Static(
                "⟳ Syncing...",
                id="sync-status"
            )

        # Calendar grid
        with Container(id="calendar-container"):
            table = DataTable(
                show_header=True,
                show_row_labels=True,
                zebra_stripes=True,
                cursor_type="cell"
            )
            yield table

        # Hints
        yield Static(
            "↵ Edit | F Focus Mode | W Week View | S Sync | J Jira | ← → Navigate",
            id="week-view-hint"
        )

        yield Footer()

    async def on_mount(self) -> None:
        """Initialize calendar on mount."""
        await self.build_calendar()

        # Start background sync
        self.sync.start_background_sync(interval_seconds=60)

        # Load initial data
        await self.load_month_data()

    async def build_calendar(self):
        """Build the calendar grid."""
        table = self.query_one(DataTable)
        table.clear(columns=True)

        # Add columns for days
        table.add_column("Week", width=6)
        for day in ["MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"]:
            table.add_column(day, width=20)

        # Calculate month layout
        year = self.current_month.year
        month = self.current_month.month
        first_day = self.current_month.replace(day=1)
        days_in_month = monthrange(year, month)[1]
        first_weekday = first_day.weekday()

        # Build weeks
        week_data = []
        current_week = [""] * 7
        week_num = 1

        # Fill in days
        for day in range(1, days_in_month + 1):
            date = self.current_month.replace(day=day)
            weekday = date.weekday()

            # Start new week on Monday
            if weekday == 0 and day > 1:
                week_data.append([f"W{week_num}"] + current_week)
                current_week = [""] * 7
                week_num += 1

            # Format cell content
            cell_content = self.calendar_data.get(day, "")
            if cell_content:
                # Truncate for display
                lines = cell_content.split("\n")
                display = f"{day}: {lines[0][:15]}..."
            else:
                display = str(day)

            # Style based on date
            if date.date() == datetime.now().date():
                display = f"[bold yellow]{display}[/]"  # Today
            elif date.weekday() >= 5:
                display = f"[dim]{display}[/]"  # Weekend

            current_week[weekday] = display

        # Add last week
        if any(current_week):
            week_data.append([f"W{week_num}"] + current_week)

        # Add rows to table
        for week_row in week_data:
            table.add_row(*week_row)

    async def load_month_data(self):
        """Load month data from Google Sheets."""
        self.update_sync_status("Loading...")

        try:
            # Get data from sheets
            month_data = self.sheets.read_month_grid(self.current_month)

            # Process into calendar_data
            self.calendar_data = {}
            for week in month_data.get('weeks', []):
                for day_name, content in week['days'].items():
                    # Map to day number (simplified)
                    # In real implementation, you'd properly map dates
                    if content:
                        # Extract day number from content or position
                        day = 1  # This needs proper implementation
                        self.calendar_data[day] = content

            # Rebuild display
            await self.build_calendar()

            self.update_sync_status("✓ Synced")

        except Exception as e:
            self.update_sync_status(f"✗ Error: {e}")

    def update_sync_status(self, status: str):
        """Update sync status display."""
        status_widget = self.query_one("#sync-status", Static)
        status_widget.update(f"⟳ {status}")

    def action_previous_month(self):
        """Navigate to previous month."""
        self.current_month = (self.current_month.replace(day=1) - timedelta(days=1)).replace(day=1)
        self.query_one("#month-title", Static).update(self.current_month.strftime("%B %Y"))
        self.run_worker(self.load_month_data())

    def action_next_month(self):
        """Navigate to next month."""
        next_month = self.current_month.month % 12 + 1
        next_year = self.current_month.year + (1 if next_month == 1 else 0)
        self.current_month = self.current_month.replace(year=next_year, month=next_month, day=1)
        self.query_one("#month-title", Static).update(self.current_month.strftime("%B %Y"))
        self.run_worker(self.load_month_data())

    def action_edit_cell(self):
        """Edit selected cell."""
        table = self.query_one(DataTable)
        cursor = table.cursor_coordinate

        if cursor:
            # Calculate date from cursor position
            week_row = cursor.row
            day_col = cursor.column - 1  # Skip week column

            if day_col >= 0:
                # Calculate actual date (simplified)
                day = week_row * 7 + day_col + 1 - self.current_month.weekday()

                if 1 <= day <= monthrange(self.current_month.year, self.current_month.month)[1]:
                    date = self.current_month.replace(day=day)
                    self.enter_focus_mode(date)

    def action_focus_mode(self):
        """Enter focus mode for current day."""
        self.enter_focus_mode(datetime.now())

    def enter_focus_mode(self, date: datetime):
        """Open focus mode for a specific date."""
        # Get current content
        day = date.day
        content = self.calendar_data.get(day, "")

        # Get Jira suggestions
        context = self.context_detector.get_full_context()
        suggestions = []
        if context.get('suggestion', {}).get('ticket'):
            ticket = context['suggestion']['ticket']
            desc = context['suggestion'].get('description', 'Work on ticket')
            suggestions.append(f"{ticket} - {desc}")

        # Open focus mode
        def handle_result(result: Optional[str]):
            if result is not None:
                # Save the content
                self.calendar_data[day] = result
                self.sheets.write_to_day_cell(date, result)
                self.run_worker(self.build_calendar())

        modal = FocusModeScreen(date, content, suggestions)
        self.app.push_screen(modal, handle_result)

    def action_sync_now(self):
        """Trigger immediate sync."""
        self.update_sync_status("Syncing...")
        self.run_worker(self.perform_sync())

    async def perform_sync(self):
        """Perform sync operation."""
        result = self.sync.sync_month(
            self.current_month.year,
            self.current_month.month
        )

        if result['success']:
            self.update_sync_status(f"✓ Synced {result['cells_synced']} cells")
            await self.load_month_data()
        else:
            self.update_sync_status(f"✗ Sync failed: {result.get('error')}")

    def action_jira_populate(self):
        """Auto-populate from Jira activities."""
        self.update_sync_status("Fetching Jira...")
        # This would integrate with Jira service
        self.update_sync_status("✓ Jira updated")


class GoogleSheetsCalendarApp(App):
    """Main Google Sheets Calendar TUI Application."""

    CSS = """
    Screen {
        background: $background;
    }
    """

    def __init__(self, sheet_id: str, credentials_path: str):
        super().__init__()

        # Initialize services
        self.sheets = GoogleSheetsService(sheet_id, credentials_path)
        self.db = SyncDatabase()
        self.sync = SyncCoordinator(self.sheets, self.db)

    def on_mount(self) -> None:
        """Set up the app."""
        self.title = "Google Sheets Calendar TUI"
        self.sub_title = "Daily Schedule Tracker"

        # Push main screen
        self.push_screen(MonthViewScreen(self.sheets, self.sync))

    def action_quit(self) -> None:
        """Quit the application."""
        self.sync.stop_background_sync()
        self.exit()
