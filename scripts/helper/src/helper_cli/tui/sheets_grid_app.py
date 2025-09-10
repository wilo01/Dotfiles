"""Google Sheets Grid TUI - Exact sheets replication with improved UX/UI."""

from datetime import datetime, timedelta
from typing import List, Dict, Any, Optional, Tuple
from calendar import monthrange
import asyncio

from textual.app import App, ComposeResult
from textual.widgets import DataTable, Header, Footer, Static, Input, Button, TextArea, Label
from textual.containers import Container, Vertical, Horizontal, Grid, ScrollableContainer
from textual.binding import Binding
from textual.coordinate import Coordinate
from textual.screen import Screen, ModalScreen
from textual import events
from textual.reactive import reactive
from rich.text import Text
from rich.style import Style
from rich.table import Table
from rich.panel import Panel
from rich.box import ROUNDED, DOUBLE, HEAVY


class EnhancedEditModal(ModalScreen):
    """Enhanced edit modal with better visual hierarchy and UX."""
    
    CSS = """
    EnhancedEditModal {
        align: center middle;
        background: rgba(0, 0, 0, 0.7);
    }
    
    #edit-container {
        background: $surface;
        border: thick $primary;
        width: 70%;
        max-width: 80;
        height: auto;
        max-height: 40;
        padding: 2;
        border-title-color: $primary;
        border-title-background: $surface;
        border-title-style: bold;
    }
    
    #cell-indicator {
        height: 3;
        padding: 1;
        background: $primary-lighten-3;
        color: $primary;
        text-style: bold;
        text-align: center;
        margin-bottom: 1;
        border: solid $primary;
    }
    
    #date-label {
        height: 1;
        color: $text;
        text-style: italic;
        margin-bottom: 1;
    }
    
    #content-label {
        height: 1;
        color: $text;
        text-style: bold;
        margin-top: 1;
        margin-bottom: 1;
    }
    
    #content-editor {
        width: 100%;
        height: 15;
        border: solid $primary;
        background: $background;
        padding: 1;
        margin-bottom: 2;
    }
    
    #button-container {
        height: 3;
        align: center middle;
        margin-top: 1;
    }
    
    .primary-button {
        background: $primary;
        color: $background;
        border: none;
        padding: 0 2;
        margin: 0 1;
        text-style: bold;
    }
    
    .primary-button:hover {
        background: $primary-darken-1;
    }
    
    .secondary-button {
        background: transparent;
        color: $text;
        border: solid $text;
        padding: 0 2;
        margin: 0 1;
    }
    
    .secondary-button:hover {
        background: $surface;
    }
    
    #shortcuts-hint {
        height: 1;
        color: $text-muted;
        text-align: center;
        margin-top: 1;
    }
    """
    
    def __init__(self, cell_ref: str, date: datetime, initial_content: str = "",
                 jira_suggestions: List[str] = None):
        super().__init__()
        self.cell_ref = cell_ref
        self.date = date
        self.initial_content = initial_content
        self.jira_suggestions = jira_suggestions or []
        self.saved = False
    
    def compose(self) -> ComposeResult:
        """Create enhanced edit modal UI."""
        with Container(id="edit-container"):
            # Cell indicator with visual prominence
            yield Static(
                f"📝 Editing {self.cell_ref}",
                id="cell-indicator"
            )
            
            # Date label
            yield Label(
                f"{self.date.strftime('%A, %B %d, %Y')}",
                id="date-label"
            )
            
            # Content label
            yield Label("Daily Activities:", id="content-label")
            
            # Main editor with line numbers
            editor = TextArea(
                self.initial_content,
                language="markdown",
                show_line_numbers=True,
                theme="monokai",
                id="content-editor"
            )
            editor.focus()
            yield editor
            
            # Action buttons with better styling
            with Horizontal(id="button-container"):
                yield Button("💾 Save", variant="primary", id="save-btn", classes="primary-button")
                yield Button("📋 Insert Jira", variant="default", id="jira-btn", classes="secondary-button")
                yield Button("❌ Cancel", variant="default", id="cancel-btn", classes="secondary-button")
            
            # Keyboard shortcuts hint
            yield Static(
                "Ctrl+S: Save | Ctrl+J: Insert Jira | Esc: Cancel",
                id="shortcuts-hint"
            )
    
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
        elif event.key == "ctrl+j":
            self.insert_jira_suggestions()
            event.stop()
        elif event.key == "escape":
            self.dismiss(None)
            event.stop()
    
    def save_and_close(self):
        """Save content and close with animation."""
        editor = self.query_one("#content-editor", TextArea)
        content = editor.text.strip()
        
        # Validate content
        if len(content) > 5000:
            self.notify("Content too long (max 5000 characters)", severity="error")
            return
        
        self.saved = True
        
        # Visual feedback before closing
        save_btn = self.query_one("#save-btn", Button)
        save_btn.variant = "success"
        save_btn.label = "✅ Saved!"
        
        # Small delay for visual feedback
        self.set_timer(0.3, lambda: self.dismiss(content))
    
    def insert_jira_suggestions(self):
        """Insert Jira activity suggestions."""
        if not self.jira_suggestions:
            self.notify("No Jira activities available", severity="warning")
            return
        
        editor = self.query_one("#content-editor", TextArea)
        current_text = editor.text
        
        # Format suggestions
        suggestions_text = "\n\n--- Jira Activities ---\n"
        for suggestion in self.jira_suggestions:
            suggestions_text += f"• {suggestion}\n"
        
        editor.text = current_text + suggestions_text
        self.notify("Jira activities inserted", severity="information")


class WeekViewScreen(ModalScreen):
    """Detailed week view showing a single week with more information."""
    
    CSS = """
    WeekViewScreen {
        align: center middle;
        background: rgba(0, 0, 0, 0.8);
    }
    
    #week-container {
        background: $surface;
        border: thick $primary;
        width: 90%;
        height: 80%;
        padding: 2;
    }
    
    #week-header {
        height: 3;
        text-style: bold;
        text-align: center;
        color: $primary;
        background: $primary-lighten-3;
        padding: 1;
        margin-bottom: 1;
    }
    
    #days-grid {
        height: 1fr;
        layout: grid;
        grid-size: 7 1;
        grid-gutter: 1;
        margin: 1;
    }
    
    .day-panel {
        border: solid $primary;
        padding: 1;
        background: white;
    }
    
    .day-panel:hover {
        background: $surface;
        border: solid $primary;
    }
    
    .day-header {
        text-style: bold;
        color: $primary;
        margin-bottom: 1;
    }
    
    .day-content {
        color: $text;
        height: 1fr;
    }
    
    #week-actions {
        height: 3;
        align: center middle;
        margin-top: 1;
    }
    """
    
    def __init__(self, week_num: int, week_data: Dict[str, str], current_month: datetime):
        super().__init__()
        self.week_num = week_num
        self.week_data = week_data
        self.current_month = current_month
        
    def compose(self) -> ComposeResult:
        """Create week view UI."""
        with Container(id="week-container"):
            # Week header
            yield Static(
                f"📅 Week {self.week_num} - {self.current_month.strftime('%B %Y')}",
                id="week-header"
            )
            
            # Days grid
            with Container(id="days-grid"):
                days = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"]
                for idx, day in enumerate(days):
                    with Container(classes="day-panel"):
                        yield Static(day.upper(), classes="day-header")
                        content = self.week_data.get(f"{self.week_num}_{idx}", "No activities")
                        yield Static(content, classes="day-content")
            
            # Actions
            with Horizontal(id="week-actions"):
                yield Button("Close", variant="primary", id="close-week")
    
    def on_button_pressed(self, event: Button.Pressed) -> None:
        """Handle button presses."""
        if event.button.id == "close-week":
            self.dismiss()


class LoadingIndicator(Static):
    """Animated loading indicator."""
    
    DEFAULT_CSS = """
    LoadingIndicator {
        align: center middle;
        width: 100%;
        height: 100%;
        display: none;
    }
    
    LoadingIndicator.visible {
        display: block;
    }
    """
    
    def __init__(self):
        super().__init__()
        self.frames = ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"]
        self.current_frame = 0
        
    def on_mount(self):
        """Start animation when mounted."""
        self.set_interval(0.1, self.update_spinner)
    
    def update_spinner(self):
        """Update spinner animation."""
        if self.display:
            self.update(f"{self.frames[self.current_frame]} Loading...")
            self.current_frame = (self.current_frame + 1) % len(self.frames)
    
    def show(self):
        """Show the loading indicator."""
        self.add_class("visible")
        
    def hide(self):
        """Hide the loading indicator."""
        self.remove_class("visible")


class GoogleSheetsGridScreen(Screen):
    """Main Google Sheets grid view - exact sheets replication."""
    
    CSS = """
    GoogleSheetsGridScreen {
        background: $background;
    }
    
    #header-container {
        height: 7;
        background: $primary;
        border: solid $primary;
        padding: 1;
        margin-bottom: 1;
    }
    
    #month-title {
        text-style: bold;
        text-align: center;
        color: white;
        background: $primary;
        padding: 1;
        border: solid $primary-lighten-1;
    }
    
    #sync-status-bar {
        height: 3;
        background: $surface;
        border: solid $secondary;
        padding: 0 1;
        margin-top: 1;
    }
    
    #sync-status {
        color: $success;
        text-align: right;
    }
    
    #last-sync {
        color: $text-muted;
        text-align: left;
    }
    
    #grid-container {
        height: 1fr;
        padding: 1;
        background: $surface;
        border: solid $border;
        margin: 0 1;
    }
    
    DataTable {
        height: 100%;
        background: white;
    }
    
    DataTable > .datatable--header {
        background: $primary;
        color: white;
        text-style: bold;
    }
    
    DataTable > .datatable--cursor {
        background: $accent;
        color: $text;
    }
    
    DataTable > .datatable--hover {
        background: $primary;
    }
    
    DataTable > .datatable--odd-row {
        background: $surface;
    }
    
    DataTable > .datatable--even-row {
        background: white;
    }
    
    #action-bar {
        height: 5;
        background: $panel;
        border: solid $border;
        padding: 1;
        margin-top: 1;
    }
    
    #navigation-hints {
        text-align: center;
        color: $text-muted;
    }
    
    #quick-actions {
        height: 3;
        align: center middle;
        margin-top: 1;
    }
    
    .action-button {
        margin: 0 1;
        min-width: 15;
    }
    """
    
    BINDINGS = [
        Binding("left", "previous_month", "← Previous Month", priority=True),
        Binding("right", "next_month", "→ Next Month", priority=True),
        Binding("up", "move_up", "↑ Move Up", show=False),
        Binding("down", "move_down", "↓ Move Down", show=False),
        Binding("enter", "edit_cell", "Edit Cell", priority=True),
        Binding("e", "edit_cell", "Edit", show=False),
        Binding("s", "sync_now", "Sync Now", priority=True),
        Binding("j", "jira_populate", "Auto-fill Jira", priority=True),
        Binding("r", "refresh", "Refresh", priority=True),
        Binding("h", "show_help", "Help", priority=True),
        Binding("q", "quit", "Quit", priority=True),
    ]
    
    def __init__(self, sheets_service=None, sync_coordinator=None):
        super().__init__()
        self.sheets_service = sheets_service
        self.sync_coordinator = sync_coordinator
        
        self.current_date = datetime.now()
        self.current_month = self.current_date.replace(day=1)
        self.selected_cell: Optional[Tuple[int, int]] = None
        self.grid_data: Dict[str, str] = {}  # Key: "week_day", Value: content
        
        # Colors for visual enhancement
        self.color_today = Style(bgcolor="rgb(255,235,59)", color="rgb(0,0,0)", bold=True)
        self.color_weekend = Style(color="rgb(158,158,158)")
        self.color_has_content = Style(color="rgb(25,118,210)", bold=True)
        self.color_week_label = Style(bgcolor="rgb(33,150,243)", color="white", bold=True)
    
    def compose(self) -> ComposeResult:
        """Create Google Sheets grid UI."""
        # Header with month and sync status
        with Container(id="header-container"):
            yield Static(
                self.current_month.strftime("📅 %B %Y"),
                id="month-title"
            )
            with Horizontal(id="sync-status-bar"):
                yield Static("Last sync: Never", id="last-sync")
                yield Static("⟳ Ready", id="sync-status")
        
        # Main grid container
        with Container(id="grid-container"):
            table = DataTable(
                show_header=True,
                show_row_labels=False,
                zebra_stripes=True,
                cursor_type="cell",
                show_cursor=True,
                id="sheets-grid"
            )
            yield table
        
        # Action bar with navigation and quick actions
        with Container(id="action-bar"):
            yield Static(
                "Navigation: ←→ Months | ↑↓ Cells | Enter: Edit | S: Sync | J: Jira | H: Help",
                id="navigation-hints"
            )
            with Horizontal(id="quick-actions"):
                yield Button("📅 Today", variant="primary", id="today-btn", classes="action-button")
                yield Button("🔄 Sync All", variant="success", id="sync-all-btn", classes="action-button")
                yield Button("📊 Week View", variant="default", id="week-view-btn", classes="action-button")
        
        yield Footer()
    
    async def on_mount(self) -> None:
        """Initialize grid on mount."""
        await self.build_sheets_grid()
        
        # Start background sync if coordinator available
        if self.sync_coordinator:
            self.sync_coordinator.start_background_sync(interval_seconds=60)
        
        # Load initial data
        await self.load_month_data()
    
    async def build_sheets_grid(self):
        """Build the Google Sheets style grid with weeks as rows."""
        table = self.query_one("#sheets-grid", DataTable)
        table.clear(columns=True)
        
        # Add columns exactly as in Google Sheets
        columns = [
            ("Week #", 8),
            ("MONDAY", 15),
            ("TUESDAY", 15),
            ("WEDNESDAY", 15),
            ("THURSDAY", 15),
            ("FRIDAY", 15),
            ("SATURDAY", 15),
            ("SUNDAY", 15)
        ]
        
        for col_name, width in columns:
            if col_name == "Week #":
                # Week column with special styling
                table.add_column(col_name, width=width, key=col_name)
            else:
                table.add_column(col_name, width=width, key=col_name)
        
        # Calculate month layout
        year = self.current_month.year
        month = self.current_month.month
        first_day = self.current_month.replace(day=1)
        days_in_month = monthrange(year, month)[1]
        first_weekday = first_day.weekday()
        
        # Build week rows
        week_data = []
        current_week = [""] * 8  # Week # + 7 days
        week_num = 1
        current_week[0] = Text(f"W{week_num}", style=self.color_week_label)
        
        # Fill in days
        for day in range(1, days_in_month + 1):
            date = self.current_month.replace(day=day)
            weekday = date.weekday()
            
            # Start new week on Monday
            if weekday == 0 and day > 1:
                week_data.append(current_week)
                week_num += 1
                current_week = [""] * 8
                current_week[0] = Text(f"W{week_num}", style=self.color_week_label)
            
            # Format cell content
            cell_key = f"{week_num}_{weekday}"
            cell_content = self.grid_data.get(cell_key, "")
            
            # Create display text with styling
            if cell_content:
                # Show day number and truncated content
                lines = cell_content.split("\n")
                display = f"{day}: {lines[0][:20]}..."
                display_text = Text(display, style=self.color_has_content)
            else:
                display_text = Text(str(day))
            
            # Apply special styling
            if date.date() == datetime.now().date():
                display_text.stylize(self.color_today)
            elif weekday >= 5:  # Weekend
                display_text.stylize(self.color_weekend)
            
            current_week[weekday + 1] = display_text
        
        # Add last week
        if any(current_week[1:]):
            week_data.append(current_week)
        
        # Add empty weeks to maintain grid structure (minimum 5 weeks)
        while len(week_data) < 5:
            week_num += 1
            empty_week = [""] * 8
            empty_week[0] = Text(f"W{week_num}", style=self.color_week_label)
            week_data.append(empty_week)
        
        # Add rows to table
        for week_row in week_data:
            table.add_row(*week_row)
    
    async def load_month_data(self):
        """Load month data from Google Sheets."""
        self.update_sync_status("📡 Loading...")
        
        try:
            if self.sheets_service:
                # Check if using OAuth service
                if hasattr(self.sheets_service, 'get_worksheet'):
                    # OAuth service - read from worksheet
                    worksheet_name = self.current_month.strftime("%B %Y")
                    try:
                        worksheet = self.sheets_service.get_worksheet(worksheet_name)
                        if worksheet:
                            # Read all values from worksheet
                            values = worksheet.get_all_values()
                            
                            # Process into grid_data
                            self.grid_data = {}
                            for row_idx, row in enumerate(values[1:], 1):  # Skip header
                                for col_idx, cell_value in enumerate(row[1:], 0):  # Skip week column
                                    if cell_value:
                                        self.grid_data[f"{row_idx}_{col_idx}"] = cell_value
                            
                            self.update_sync_status("✅ Synced")
                            self.update_last_sync(datetime.now())
                    except Exception as ws_error:
                        # Worksheet doesn't exist, use demo data
                        self.load_demo_data()
                        self.update_sync_status("📝 New Month")
                else:
                    # Regular sheets service
                    month_data = self.sheets_service.read_month_grid(self.current_month)
                    
                    # Process into grid_data
                    self.grid_data = {}
                    for week_num, week in enumerate(month_data.get('weeks', []), 1):
                        for day_idx, (day_name, content) in enumerate(week['days'].items()):
                            if content:
                                self.grid_data[f"{week_num}_{day_idx}"] = content
                    
                    self.update_sync_status("✅ Synced")
                    self.update_last_sync(datetime.now())
                
                # Rebuild display
                await self.build_sheets_grid()
            else:
                # Demo mode with sample data
                self.load_demo_data()
                await self.build_sheets_grid()
                self.update_sync_status("🔶 Demo Mode")
        
        except Exception as e:
            self.update_sync_status(f"❌ Error: {str(e)[:30]}")
            self.load_demo_data()
            await self.build_sheets_grid()
    
    def load_demo_data(self):
        """Load demo data for testing."""
        self.grid_data = {
            "1_0": "Sprint Planning\nCode Review",
            "1_2": "Feature Implementation\nTesting",
            "2_1": "Team Meeting\nDocumentation",
            "2_4": "Deploy to Production",
        }
    
    def update_sync_status(self, status: str):
        """Update sync status display."""
        status_widget = self.query_one("#sync-status", Static)
        status_widget.update(status)
    
    def update_last_sync(self, timestamp: datetime):
        """Update last sync timestamp."""
        last_sync_widget = self.query_one("#last-sync", Static)
        last_sync_widget.update(f"Last sync: {timestamp.strftime('%H:%M:%S')}")
    
    def action_previous_month(self):
        """Navigate to previous month with animation."""
        # Animate the transition
        table = self.query_one("#sheets-grid", DataTable)
        table.loading = True
        
        self.current_month = (self.current_month.replace(day=1) - timedelta(days=1)).replace(day=1)
        
        # Update title with fade effect
        title_widget = self.query_one("#month-title", Static)
        title_widget.update(f"📅 {self.current_month.strftime('%B %Y')}")
        
        # Add visual feedback
        self.notify(f"← {self.current_month.strftime('%B %Y')}", severity="information", timeout=1)
        
        self.run_worker(self.load_month_data())
    
    def action_next_month(self):
        """Navigate to next month with animation."""
        # Animate the transition
        table = self.query_one("#sheets-grid", DataTable)
        table.loading = True
        
        next_month = self.current_month.month % 12 + 1
        next_year = self.current_month.year + (1 if next_month == 1 else 0)
        self.current_month = self.current_month.replace(year=next_year, month=next_month, day=1)
        
        # Update title with fade effect
        title_widget = self.query_one("#month-title", Static)
        title_widget.update(f"📅 {self.current_month.strftime('%B %Y')}")
        
        # Add visual feedback
        self.notify(f"→ {self.current_month.strftime('%B %Y')}", severity="information", timeout=1)
        
        self.run_worker(self.load_month_data())
    
    def action_edit_cell(self):
        """Edit selected cell."""
        table = self.query_one("#sheets-grid", DataTable)
        cursor = table.cursor_coordinate
        
        if cursor and cursor.column > 0:  # Skip week column
            week_row = cursor.row
            day_col = cursor.column - 1
            
            # Calculate actual date
            days_of_week = ["MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY", "SUNDAY"]
            
            # Find the date for this cell
            first_day = self.current_month.replace(day=1)
            first_weekday = first_day.weekday()
            
            # Calculate day of month
            days_from_start = (week_row * 7) + day_col - first_weekday
            
            if days_from_start >= 0:
                try:
                    date = first_day + timedelta(days=days_from_start)
                    
                    # Only edit if within current month
                    if date.month == self.current_month.month:
                        cell_ref = f"{days_of_week[day_col]} W{week_row + 1}"
                        cell_key = f"{week_row + 1}_{day_col}"
                        content = self.grid_data.get(cell_key, "")
                        
                        # Get Jira suggestions (if available)
                        suggestions = []
                        if hasattr(self, 'context_detector'):
                            context = self.context_detector.get_full_context()
                            if context.get('suggestion', {}).get('ticket'):
                                ticket = context['suggestion']['ticket']
                                desc = context['suggestion'].get('description', 'Work on ticket')
                                suggestions.append(f"{ticket} - {desc}")
                        
                        # Open enhanced edit modal
                        def handle_result(result: Optional[str]):
                            if result is not None:
                                # Save the content
                                self.grid_data[cell_key] = result
                                if self.sheets_service:
                                    # Save to Google Sheets
                                    if hasattr(self.sheets_service, 'get_worksheet'):
                                        # OAuth service
                                        try:
                                            worksheet_name = self.current_month.strftime("%B %Y")
                                            worksheet = self.sheets_service.get_worksheet(worksheet_name)
                                            if not worksheet:
                                                # Create worksheet if it doesn't exist
                                                worksheet = self.sheets_service.create_worksheet(worksheet_name)
                                            
                                            # Calculate cell position (B2 = Monday W1, etc.)
                                            col_letter = chr(ord('B') + day_col)  # B=Monday, H=Sunday
                                            cell_address = f"{col_letter}{week_row + 2}"  # +2 for header and 0-index
                                            worksheet.update(cell_address, result)
                                        except Exception as save_error:
                                            self.notify(f"Error saving: {save_error}", severity="error")
                                    else:
                                        # Regular service
                                        self.sheets_service.write_to_day_cell(date, result)
                                
                                self.run_worker(self.build_sheets_grid())
                                self.notify(f"Saved to {cell_ref}", severity="success")
                        
                        modal = EnhancedEditModal(cell_ref, date, content, suggestions)
                        self.app.push_screen(modal, handle_result)
                
                except ValueError:
                    pass  # Date out of range
    
    def action_sync_now(self):
        """Trigger immediate sync."""
        self.update_sync_status("🔄 Syncing...")
        self.run_worker(self.perform_sync())
    
    async def perform_sync(self):
        """Perform sync operation with error handling."""
        try:
            if self.sync_coordinator:
                result = self.sync_coordinator.sync_month(
                    self.current_month.year,
                    self.current_month.month
                )
                
                if result['success']:
                    self.update_sync_status(f"✅ Synced {result['cells_synced']} cells")
                    self.update_last_sync(datetime.now())
                    await self.load_month_data()
                else:
                    error_msg = result.get('error', 'Unknown error')
                    self.update_sync_status(f"❌ Sync failed")
                    self.notify(f"Sync error: {error_msg}", severity="error")
            else:
                self.update_sync_status("🔶 Demo Mode")
                self.notify("Running in demo mode - no sync available", severity="warning")
        except Exception as e:
            self.update_sync_status("❌ Sync error")
            self.notify(f"Unexpected error: {str(e)[:100]}", severity="error")
    
    def action_jira_populate(self):
        """Auto-populate from Jira activities."""
        self.update_sync_status("🔍 Fetching Jira...")
        # This would integrate with Jira service
        self.notify("Jira integration would populate current day's activities", severity="information")
        self.update_sync_status("✅ Jira updated")
    
    def action_show_help(self):
        """Show help dialog."""
        help_text = """
        🎯 Google Sheets Grid TUI - Help
        
        Navigation:
        • ← → : Navigate between months
        • ↑ ↓ : Move between cells
        • Enter/E : Edit selected cell
        
        Actions:
        • S : Sync with Google Sheets
        • J : Auto-populate from Jira
        • R : Refresh current view
        • H : Show this help
        • Q : Quit application
        
        Edit Mode:
        • Ctrl+S : Save and close
        • Ctrl+J : Insert Jira activities
        • Esc : Cancel without saving
        
        This TUI replicates your Google Sheets calendar
        with week rows and day columns for easy tracking.
        """
        self.notify(help_text, title="Help", severity="information", timeout=10)
    
    def show_week_view(self):
        """Show detailed week view for current week."""
        # Calculate current week number
        today = datetime.now()
        first_day = self.current_month.replace(day=1)
        first_weekday = first_day.weekday()
        
        # Find which week today is in
        if today.month == self.current_month.month:
            days_from_start = today.day + first_weekday - 1
            week_num = (days_from_start // 7) + 1
        else:
            week_num = 1  # Default to first week if not current month
        
        # Gather week data
        week_data = {}
        for day_idx in range(7):
            key = f"{week_num}_{day_idx}"
            week_data[key] = self.grid_data.get(key, "")
        
        # Show week view modal
        modal = WeekViewScreen(week_num, week_data, self.current_month)
        self.app.push_screen(modal)
    
    def on_button_pressed(self, event: Button.Pressed) -> None:
        """Handle button presses."""
        if event.button.id == "today-btn":
            self.current_month = datetime.now().replace(day=1)
            self.query_one("#month-title", Static).update(
                f"📅 {self.current_month.strftime('%B %Y')}"
            )
            self.run_worker(self.load_month_data())
        elif event.button.id == "sync-all-btn":
            self.action_sync_now()
        elif event.button.id == "week-view-btn":
            self.show_week_view()


class GoogleSheetsGridApp(App):
    """Main Google Sheets Grid TUI Application."""
    
    CSS = """
    $primary: rgb(25,118,210);
    $primary-lighten-1: rgb(66,165,245);
    $primary-lighten-3: rgb(144,202,249);
    $primary-darken-1: rgb(21,101,192);
    $secondary: rgb(156,39,176);
    $accent: rgb(255,193,7);
    $success: rgb(76,175,80);
    $error: rgb(244,67,54);
    $warning: rgb(255,152,0);
    $surface: rgb(250,250,250);
    $background: rgb(245,245,245);
    $panel: rgb(238,238,238);
    $border: rgb(224,224,224);
    $text: rgb(33,33,33);
    $text-muted: rgb(117,117,117);
    
    Screen {
        background: $background;
    }
    
    Header {
        background: $primary;
        color: white;
        text-style: bold;
    }
    
    Footer {
        background: $panel;
        color: $text-muted;
    }
    """
    
    def __init__(self, sheet_id: str = None, credentials_path: str = None):
        super().__init__()
        
        self.sheet_id = sheet_id
        self.credentials_path = credentials_path
        self.oauth_service = None  # Will be set externally if using OAuth
        
        # Initialize services
        self.sheets_service = None
        self.sync_coordinator = None
        
        # Try to initialize services if credentials provided
        if sheet_id and credentials_path:
            try:
                from ..services.google_sheets_service import GoogleSheetsService
                from ..services.sync_coordinator import SyncCoordinator
                from ..services.sync_models import SyncDatabase
                
                self.sheets_service = GoogleSheetsService(sheet_id, credentials_path)
                self.db = SyncDatabase()
                self.sync_coordinator = SyncCoordinator(self.sheets_service, self.db)
            except ImportError:
                # Services not available, will run in demo mode
                pass
    
    def on_mount(self) -> None:
        """Set up the app."""
        self.title = "Google Sheets Grid TUI"
        self.sub_title = "Calendar Tracking with Sheets Integration"
        
        # Use OAuth service if available, otherwise sheets service
        service_to_use = self.oauth_service or self.sheets_service
        
        # Push main screen
        self.push_screen(GoogleSheetsGridScreen(service_to_use, self.sync_coordinator))
    
    def action_quit(self) -> None:
        """Quit the application."""
        if self.sync_coordinator:
            self.sync_coordinator.stop_background_sync()
        self.exit()


if __name__ == "__main__":
    # Demo mode
    app = GoogleSheetsGridApp()
    app.run()