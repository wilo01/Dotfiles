"""Dark mode optimized Google Sheets Grid TUI - Compact and readable."""

from datetime import datetime, timedelta
from typing import List, Dict, Any, Optional, Tuple
from calendar import monthrange

from textual.app import App, ComposeResult
from textual.widgets import DataTable, Static, Input, Button, TextArea, Label
from textual.containers import Container, Vertical, Horizontal
from textual.binding import Binding
from textual.screen import Screen, ModalScreen
from textual import events
from rich.text import Text
from rich.style import Style


class CompactEditModal(ModalScreen):
    """Compact edit modal for dark terminals."""
    
    CSS = """
    CompactEditModal {
        align: center middle;
    }
    
    #edit-box {
        background: $boost;
        border: heavy $primary;
        width: 60;
        height: 20;
        padding: 1;
    }
    
    #edit-header {
        text-align: center;
        text-style: bold;
        color: $warning;
        margin-bottom: 1;
    }
    
    #edit-content {
        height: 12;
        background: $background;
        border: solid $primary-lighten-2;
    }
    
    #edit-buttons {
        height: 3;
        align: center middle;
        margin-top: 1;
    }
    
    Button {
        margin: 0 1;
    }
    """
    
    def __init__(self, cell_ref: str, date: datetime, content: str = ""):
        super().__init__()
        self.cell_ref = cell_ref
        self.date = date
        self.initial_content = content
    
    def compose(self) -> ComposeResult:
        """Create edit UI."""
        with Container(id="edit-box"):
            yield Static(
                f"{self.cell_ref} - {self.date.strftime('%b %d')}",
                id="edit-header"
            )
            yield TextArea(
                self.initial_content,
                id="edit-content"
            )
            with Horizontal(id="edit-buttons"):
                yield Button("Save [^S]", variant="primary", id="save")
                yield Button("Cancel [ESC]", variant="default", id="cancel")
    
    def on_mount(self):
        """Focus editor on mount."""
        self.query_one("#edit-content", TextArea).focus()
    
    def on_button_pressed(self, event: Button.Pressed) -> None:
        """Handle buttons."""
        if event.button.id == "save":
            self.save_and_close()
        elif event.button.id == "cancel":
            self.dismiss(None)
    
    def on_key(self, event: events.Key) -> None:
        """Handle keys."""
        if event.key == "ctrl+s":
            self.save_and_close()
            event.stop()
        elif event.key == "escape":
            self.dismiss(None)
            event.stop()
    
    def save_and_close(self):
        """Save and close."""
        editor = self.query_one("#edit-content", TextArea)
        self.dismiss(editor.text.strip())


class DarkGridScreen(Screen):
    """Dark mode optimized grid screen."""
    
    CSS = """
    DarkGridScreen {
        background: $background;
    }
    
    #header {
        height: 3;
        padding: 0 2;
        background: $boost;
        border-bottom: heavy $primary;
    }
    
    #month-display {
        text-align: center;
        text-style: bold;
        color: $text;
    }
    
    #grid-wrapper {
        height: 1fr;
        padding: 0;
        margin: 0;
    }
    
    DataTable {
        scrollbar-size: 1 1;
        scrollbar-background: $boost;
        scrollbar-corner-color: $boost;
    }
    
    DataTable > .datatable--header {
        background: $boost;
        color: $primary-lighten-2;
        text-style: bold;
    }
    
    DataTable > .datatable--cursor {
        background: $primary-darken-3;
    }
    
    DataTable > .datatable--odd-row {
        background: $background;
    }
    
    DataTable > .datatable--even-row {
        background: $boost;
    }
    
    #status-bar {
        height: 2;
        padding: 0 2;
        background: $boost;
        border-top: solid $primary-darken-2;
        color: $text-muted;
    }
    
    #help-text {
        text-align: center;
    }
    """
    
    BINDINGS = [
        Binding("left", "prev_month", "← Month"),
        Binding("right", "next_month", "→ Month"),
        Binding("up", "move_up", "↑", show=False),
        Binding("down", "move_down", "↓", show=False),
        Binding("enter", "edit", "Edit"),
        Binding("e", "edit", "Edit", show=False),
        Binding("t", "today", "Today"),
        Binding("q", "quit", "Quit"),
    ]
    
    def __init__(self):
        super().__init__()
        self.current_month = datetime.now().replace(day=1)
        self.data: Dict[str, str] = {}  # No demo data!
        
        # Dark mode optimized colors
        self.style_week = Style(color="cyan", bold=True)
        self.style_today = Style(color="yellow", bold=True, bgcolor="rgb(40,40,0)")
        self.style_weekend = Style(color="rgb(100,100,100)")
        self.style_content = Style(color="white")
        self.style_empty = Style(color="rgb(60,60,60)")
    
    def compose(self) -> ComposeResult:
        """Create UI."""
        # Simple header
        with Container(id="header"):
            yield Static(
                self.current_month.strftime("%B %Y"),
                id="month-display"
            )
        
        # Grid
        with Container(id="grid-wrapper"):
            table = DataTable(
                show_header=True,
                show_row_labels=False,
                cursor_type="cell",
                zebra_stripes=True,
                id="grid"
            )
            yield table
        
        # Status bar
        with Container(id="status-bar"):
            yield Static(
                "↵ Edit | ←→ Navigate | T Today | Q Quit",
                id="help-text"
            )
    
    async def on_mount(self) -> None:
        """Build grid on mount."""
        await self.build_grid()
    
    async def build_grid(self):
        """Build compact grid."""
        table = self.query_one("#grid", DataTable)
        table.clear(columns=True)
        
        # Compact columns
        columns = [
            ("W", 3),
            ("Mon", 10),
            ("Tue", 10),
            ("Wed", 10),
            ("Thu", 10),
            ("Fri", 10),
            ("Sat", 10),
            ("Sun", 10),
        ]
        
        for name, width in columns:
            table.add_column(name, width=width)
        
        # Build month
        year = self.current_month.year
        month = self.current_month.month
        first_day = self.current_month.replace(day=1)
        days_in_month = monthrange(year, month)[1]
        first_weekday = first_day.weekday()
        today = datetime.now().date()
        
        # Create weeks
        week_rows = []
        week_num = 1
        current_week = [""] * 8
        current_week[0] = Text(str(week_num), style=self.style_week)
        
        for day in range(1, days_in_month + 1):
            date = self.current_month.replace(day=day)
            weekday = date.weekday()
            
            # New week on Monday
            if weekday == 0 and day > 1:
                week_rows.append(current_week)
                week_num += 1
                current_week = [""] * 8
                current_week[0] = Text(str(week_num), style=self.style_week)
            
            # Cell content
            cell_key = f"{week_num}_{weekday}"
            content = self.data.get(cell_key, "")
            
            # Format cell
            if content:
                # Show first line only
                first_line = content.split('\n')[0][:8]
                cell_text = Text(f"{day}: {first_line}")
                cell_text.stylize(self.style_content)
            else:
                cell_text = Text(str(day))
                cell_text.stylize(self.style_empty)
            
            # Highlight today
            if date.date() == today:
                cell_text.stylize(self.style_today)
            elif weekday >= 5:  # Weekend
                cell_text.stylize(self.style_weekend)
            
            current_week[weekday + 1] = cell_text
        
        # Add last week
        if any(current_week[1:]):
            week_rows.append(current_week)
        
        # Fill to 6 weeks for consistency
        while len(week_rows) < 6:
            week_num += 1
            empty_week = [""] * 8
            empty_week[0] = Text(str(week_num), style=self.style_week)
            week_rows.append(empty_week)
        
        # Add rows
        for row in week_rows:
            table.add_row(*row)
    
    def action_prev_month(self):
        """Previous month."""
        self.current_month = (self.current_month.replace(day=1) - timedelta(days=1)).replace(day=1)
        self.query_one("#month-display").update(self.current_month.strftime("%B %Y"))
        self.run_worker(self.build_grid())
    
    def action_next_month(self):
        """Next month."""
        if self.current_month.month == 12:
            self.current_month = self.current_month.replace(year=self.current_month.year + 1, month=1)
        else:
            self.current_month = self.current_month.replace(month=self.current_month.month + 1)
        self.query_one("#month-display").update(self.current_month.strftime("%B %Y"))
        self.run_worker(self.build_grid())
    
    def action_today(self):
        """Jump to today."""
        self.current_month = datetime.now().replace(day=1)
        self.query_one("#month-display").update(self.current_month.strftime("%B %Y"))
        self.run_worker(self.build_grid())
    
    def action_edit(self):
        """Edit selected cell."""
        table = self.query_one("#grid", DataTable)
        cursor = table.cursor_coordinate
        
        if cursor and cursor.column > 0:  # Skip week column
            week_row = cursor.row
            day_col = cursor.column - 1
            
            # Calculate date
            first_day = self.current_month.replace(day=1)
            first_weekday = first_day.weekday()
            days_from_start = (week_row * 7) + day_col - first_weekday
            
            if days_from_start >= 0:
                try:
                    date = first_day + timedelta(days=days_from_start)
                    if date.month == self.current_month.month:
                        # Get cell content
                        cell_key = f"{week_row + 1}_{day_col}"
                        content = self.data.get(cell_key, "")
                        
                        # Cell reference
                        days = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]
                        cell_ref = f"{days[day_col]} W{week_row + 1}"
                        
                        # Open edit modal
                        def handle_result(result: Optional[str]):
                            if result is not None:
                                if result:
                                    self.data[cell_key] = result
                                elif cell_key in self.data:
                                    del self.data[cell_key]
                                self.run_worker(self.build_grid())
                        
                        modal = CompactEditModal(cell_ref, date, content)
                        self.app.push_screen(modal, handle_result)
                except ValueError:
                    pass


class DarkGridApp(App):
    """Dark mode optimized Google Sheets Grid TUI."""
    
    CSS = """
    /* Dark mode optimized colors */
    $background: rgb(0,0,0);
    $boost: rgb(20,20,20);
    $surface: rgb(30,30,30);
    $primary: rgb(0,120,212);
    $primary-lighten-2: rgb(100,180,255);
    $primary-darken-2: rgb(0,80,150);
    $primary-darken-3: rgb(0,40,80);
    $secondary: rgb(160,160,160);
    $warning: rgb(255,200,0);
    $error: rgb(255,100,100);
    $success: rgb(100,255,100);
    $text: rgb(230,230,230);
    $text-muted: rgb(130,130,130);
    
    Screen {
        background: $background;
    }
    """
    
    def on_mount(self) -> None:
        """Set up app."""
        self.title = "Calendar Grid"
        self.sub_title = "Dark Mode"
        self.push_screen(DarkGridScreen())
    
    def action_quit(self) -> None:
        """Quit."""
        self.exit()


if __name__ == "__main__":
    app = DarkGridApp()
    app.run()