"""Simple Google Sheets TUI - Minimal calendar interface."""

from datetime import datetime, timedelta
from typing import Optional
import logging

from textual.app import App, ComposeResult
from textual.widgets import Header, Footer, Static, Button, Input, Label, DataTable
from textual.containers import Container, Horizontal, Vertical
from textual.screen import ModalScreen
from rich.text import Text

logger = logging.getLogger(__name__)


class EditCellModal(ModalScreen):
    """Modal for editing a cell."""
    
    def __init__(self, date: datetime, sheet_service):
        super().__init__()
        self.date = date
        self.sheet_service = sheet_service
        
    def compose(self) -> ComposeResult:
        with Container(id="dialog"):
            yield Label(f"Edit Cell: {self.date.strftime('%Y-%m-%d')}")
            yield Input(placeholder="Enter content", id="content")
            with Horizontal():
                yield Button("Save", variant="primary", id="save")
                yield Button("Cancel", id="cancel")
    
    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "save":
            content = self.query_one("#content", Input).value
            if content:
                self.save_to_sheet(content)
            self.dismiss(True)
        else:
            self.dismiss(False)
    
    def save_to_sheet(self, content: str):
        """Save content to Google Sheets."""
        try:
            worksheet_name = self.date.strftime("%B %Y")
            day = self.date.day
            weekday = self.date.weekday()
            
            # Calculate cell (simplified)
            row = 2 + (day - 1) // 7  # Week row
            col = chr(ord('B') + weekday)  # B=Monday, H=Sunday
            cell_ref = f"{col}{row}"
            
            worksheet = self.sheet_service.get_worksheet(worksheet_name)
            worksheet.update(cell_ref, content)
            logger.info(f"Saved to {worksheet_name}!{cell_ref}")
        except Exception as e:
            logger.error(f"Failed to save: {e}")


class SimpleSheetsApp(App):
    """Simple Google Sheets Calendar TUI."""
    
    CSS = """
    #dialog {
        grid-size: 2;
        grid-gutter: 1 2;
        padding: 0 1;
        width: 60;
        height: 11;
        border: thick $background 80%;
        background: $surface;
    }
    
    #calendar {
        width: 100%;
        height: 100%;
    }
    
    Button {
        width: 100%;
        margin: 1;
    }
    """
    
    BINDINGS = [
        ("q", "quit", "Quit"),
        ("left", "prev_month", "Previous"),
        ("right", "next_month", "Next"),
        ("e", "edit_cell", "Edit"),
    ]
    
    def __init__(self, sheet_service):
        super().__init__()
        self.sheet_service = sheet_service
        self.current_date = datetime.now()
        self.calendar_data = []
        
    def compose(self) -> ComposeResult:
        yield Header()
        
        with Container():
            yield Label(f"📅 {self.current_date.strftime('%B %Y')}", id="month_label")
            
            # Simple calendar table
            yield DataTable(id="calendar")
            
            with Horizontal():
                yield Button("◀ Previous", id="prev", variant="default")
                yield Button("Today", id="today", variant="primary") 
                yield Button("Next ▶", id="next", variant="default")
                
        yield Footer()
    
    def on_mount(self) -> None:
        """Initialize the calendar."""
        table = self.query_one(DataTable)
        
        # Add columns
        table.add_column("Week", width=6)
        for day in ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]:
            table.add_column(day, width=10)
        
        self.update_calendar()
    
    def update_calendar(self):
        """Update calendar display."""
        table = self.query_one(DataTable)
        table.clear()
        
        # Update month label
        self.query_one("#month_label", Label).update(
            f"📅 {self.current_date.strftime('%B %Y')}"
        )
        
        # Calculate calendar
        year = self.current_date.year
        month = self.current_date.month
        
        from calendar import monthrange
        first_day = datetime(year, month, 1)
        days_in_month = monthrange(year, month)[1]
        first_weekday = first_day.weekday()
        
        # Build calendar grid
        week_num = 1
        current_week = [f"W{week_num}"] + [""] * 7
        
        for day in range(1, days_in_month + 1):
            date = datetime(year, month, day)
            weekday = date.weekday()
            
            if weekday == 0 and day > 1:
                table.add_row(*current_week)
                week_num += 1
                current_week = [f"W{week_num}"] + [""] * 7
            
            # Highlight today
            if date.date() == datetime.now().date():
                current_week[weekday + 1] = f"[bold green]{day}[/]"
            else:
                current_week[weekday + 1] = str(day)
        
        # Add last week
        table.add_row(*current_week)
    
    def on_button_pressed(self, event: Button.Pressed) -> None:
        """Handle button presses."""
        if event.button.id == "prev":
            self.action_prev_month()
        elif event.button.id == "next":
            self.action_next_month()
        elif event.button.id == "today":
            self.current_date = datetime.now()
            self.update_calendar()
    
    def action_prev_month(self) -> None:
        """Go to previous month."""
        self.current_date = self.current_date.replace(day=1) - timedelta(days=1)
        self.update_calendar()
    
    def action_next_month(self) -> None:
        """Go to next month."""
        next_month = self.current_date.replace(day=28) + timedelta(days=4)
        self.current_date = next_month.replace(day=1)
        self.update_calendar()
    
    def action_edit_cell(self) -> None:
        """Open edit dialog."""
        # For simplicity, edit today's date
        self.push_screen(EditCellModal(datetime.now(), self.sheet_service))


def run_simple_tui(sheet_service):
    """Run the simple TUI."""
    app = SimpleSheetsApp(sheet_service)
    app.run()


def run_demo_tui():
    """Run a demo TUI without sheet access."""
    class DemoService:
        def get_worksheet(self, name):
            return None
    
    app = SimpleSheetsApp(DemoService())
    app.run()