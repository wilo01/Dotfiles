"""Google Sheets Calendar TUI with OAuth authentication."""

import logging
from datetime import datetime, timedelta
from pathlib import Path
from typing import Optional, Dict, Any, List

from textual.app import App, ComposeResult
from textual.widgets import Header, Footer, Static, Button, Input, Label
from textual.containers import Container, Horizontal, Vertical, ScrollableContainer
from textual.screen import Screen
from textual import events
from textual.reactive import reactive
from rich.table import Table
from rich.panel import Panel
from rich.text import Text

logger = logging.getLogger(__name__)


class CalendarView(Static):
    """Calendar grid view for the current month."""
    
    def __init__(self, sheet_service, year: int, month: int):
        super().__init__()
        self.sheet_service = sheet_service
        self.year = year
        self.month = month
        self.month_name = datetime(year, month, 1).strftime("%B %Y")
        
    def compose(self) -> ComposeResult:
        """Create the calendar grid."""
        yield Static(self.render_calendar())
    
    def render_calendar(self) -> Panel:
        """Render the calendar as a Rich panel."""
        table = Table(title=self.month_name, show_header=True, header_style="bold")
        
        # Add columns
        headers = ["Week", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]
        for header in headers:
            table.add_column(header, justify="center", width=12)
        
        # Calculate calendar
        from calendar import monthrange
        first_day = datetime(self.year, self.month, 1)
        days_in_month = monthrange(self.year, self.month)[1]
        first_weekday = first_day.weekday()
        
        # Build calendar grid
        week_num = 1
        current_week = [f"W{week_num}"] + [""] * 7
        
        for day in range(1, days_in_month + 1):
            date = datetime(self.year, self.month, day)
            weekday = date.weekday()
            
            if weekday == 0 and day > 1:
                table.add_row(*current_week)
                week_num += 1
                current_week = [f"W{week_num}"] + [""] * 7
            
            current_week[weekday + 1] = str(day)
        
        # Add last week
        table.add_row(*current_week)
        
        return Panel(table, title="📅 Calendar View", border_style="blue")


class MonthScreen(Screen):
    """Main month view screen."""
    
    def __init__(self, sheet_service):
        super().__init__()
        self.sheet_service = sheet_service
        self.current_date = datetime.now()
        
    def compose(self) -> ComposeResult:
        """Create the month view layout."""
        yield Header()
        
        with Container():
            yield CalendarView(
                self.sheet_service,
                self.current_date.year,
                self.current_date.month
            )
            
            with Horizontal():
                yield Button("◀ Previous", id="prev_month")
                yield Button("Today", id="today")
                yield Button("Next ▶", id="next_month")
                yield Button("Edit Cell", id="edit_cell", variant="primary")
        
        yield Footer()
    
    def on_button_pressed(self, event: Button.Pressed) -> None:
        """Handle button presses."""
        if event.button.id == "prev_month":
            self.current_date = self.current_date.replace(day=1) - timedelta(days=1)
            self.update_calendar()
        elif event.button.id == "next_month":
            next_month = self.current_date.replace(day=28) + timedelta(days=4)
            self.current_date = next_month.replace(day=1)
            self.update_calendar()
        elif event.button.id == "today":
            self.current_date = datetime.now()
            self.update_calendar()
        elif event.button.id == "edit_cell":
            self.app.push_screen(EditScreen(self.sheet_service))
    
    def update_calendar(self) -> None:
        """Update the calendar view."""
        # Update calendar
        try:
            calendar_view = self.query_one(CalendarView)
            calendar_view.year = self.current_date.year
            calendar_view.month = self.current_date.month
            calendar_view.month_name = self.current_date.strftime("%B %Y")
            calendar_view.update(calendar_view.render_calendar())
        except Exception:
            # Widget not mounted yet
            pass


class EditScreen(Screen):
    """Screen for editing a cell."""
    
    def __init__(self, sheet_service):
        super().__init__()
        self.sheet_service = sheet_service
        
    def compose(self) -> ComposeResult:
        """Create the edit form."""
        yield Header()
        
        with Container():
            yield Label("Edit Cell", classes="title")
            
            with Vertical():
                yield Label("Date (YYYY-MM-DD):")
                yield Input(placeholder="2025-01-15", id="date_input")
                
                yield Label("Content:")
                yield Input(placeholder="Meeting with team", id="content_input")
                
                with Horizontal():
                    yield Button("Save", id="save", variant="primary")
                    yield Button("Cancel", id="cancel")
        
        yield Footer()
    
    def on_button_pressed(self, event: Button.Pressed) -> None:
        """Handle button presses."""
        if event.button.id == "save":
            date_str = self.query_one("#date_input", Input).value
            content = self.query_one("#content_input", Input).value
            
            if date_str and content:
                try:
                    # Parse date
                    date = datetime.strptime(date_str, "%Y-%m-%d")
                    
                    # Get worksheet name
                    worksheet_name = date.strftime("%B %Y")
                    
                    # Calculate cell position
                    from calendar import monthrange
                    first_day = datetime(date.year, date.month, 1)
                    first_weekday = first_day.weekday()
                    
                    # Calculate week row (1-based, plus 1 for header)
                    week_row = ((date.day + first_weekday - 1) // 7) + 2
                    
                    # Calculate day column (B=Monday, H=Sunday)
                    day_col = chr(ord('B') + date.weekday())
                    
                    cell_ref = f"{day_col}{week_row}"
                    
                    # Write to sheet
                    try:
                        worksheet = self.sheet_service.get_worksheet(worksheet_name)
                        worksheet.update(cell_ref, content)
                        self.notify(f"✅ Saved to {worksheet_name}!{cell_ref}")
                    except Exception as e:
                        self.notify(f"❌ Error: {e}", severity="error")
                    
                    self.app.pop_screen()
                    
                except ValueError:
                    self.notify("Invalid date format. Use YYYY-MM-DD", severity="error")
            else:
                self.notify("Please fill in all fields", severity="warning")
                
        elif event.button.id == "cancel":
            self.app.pop_screen()


class GoogleSheetsCalendarOAuthApp(App):
    """Google Sheets Calendar TUI Application with OAuth."""
    
    CSS = """
    Container {
        padding: 1;
    }
    
    .title {
        text-style: bold;
        text-align: center;
        padding: 1;
    }
    
    Button {
        margin: 1;
    }
    
    Input {
        margin: 1;
    }
    
    Label {
        margin: 1;
    }
    
    Horizontal {
        align: center middle;
        height: auto;
    }
    
    Vertical {
        width: 100%;
        height: auto;
    }
    """
    
    BINDINGS = [
        ("q", "quit", "Quit"),
        ("e", "edit", "Edit Cell"),
        ("r", "refresh", "Refresh"),
        ("left", "prev_month", "Previous Month"),
        ("right", "next_month", "Next Month"),
    ]
    
    def __init__(self, sheet_id: str, oauth_service):
        super().__init__()
        self.sheet_id = sheet_id
        self.oauth_service = oauth_service
        
        # Ensure authenticated
        if not oauth_service.client:
            oauth_service.authenticate()
        
        # Open the sheet
        try:
            oauth_service.open_sheet(sheet_id)
        except Exception as e:
            logger.error(f"Could not open sheet: {e}")
    
    def on_mount(self) -> None:
        """Initialize the app."""
        self.push_screen(MonthScreen(self.oauth_service))
    
    def action_quit(self) -> None:
        """Quit the application."""
        self.exit()
    
    def action_edit(self) -> None:
        """Open edit screen."""
        self.push_screen(EditScreen(self.oauth_service))
    
    def action_refresh(self) -> None:
        """Refresh current screen."""
        screen = self.screen
        if hasattr(screen, 'update_calendar'):
            screen.update_calendar()
    
    def action_prev_month(self) -> None:
        """Go to previous month."""
        screen = self.screen
        if isinstance(screen, MonthScreen):
            screen.current_date = screen.current_date.replace(day=1) - timedelta(days=1)
            screen.update_calendar()
    
    def action_next_month(self) -> None:
        """Go to next month."""
        screen = self.screen
        if isinstance(screen, MonthScreen):
            next_month = screen.current_date.replace(day=28) + timedelta(days=4)
            screen.current_date = next_month.replace(day=1)
            screen.update_calendar()