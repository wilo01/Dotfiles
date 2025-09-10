#!/usr/bin/env python3
"""Better Google Sheets Calendar TUI - Full width, readable content."""

from datetime import datetime, timedelta
from typing import Dict, Optional
from calendar import monthrange

from textual.app import App, ComposeResult
from textual.widgets import DataTable, Static, TextArea
from textual.containers import Container, Vertical, Horizontal
from textual.binding import Binding
from textual.screen import Screen, ModalScreen
from textual import events
from rich.text import Text
from rich.style import Style


class EditModal(ModalScreen):
    """Simple edit modal that works."""
    
    CSS = """
    EditModal {
        align: center middle;
    }
    
    #modal-content {
        background: $surface;
        border: thick $accent;
        width: 70;
        height: 25;
        padding: 1;
    }
    
    #modal-header {
        text-align: center;
        text-style: bold;
        color: $primary;
        padding: 1;
    }
    
    #text-editor {
        width: 100%;
        height: 18;
        border: solid $border;
        background: $background;
    }
    
    #hint {
        text-align: center;
        color: $text-muted;
        padding: 1;
    }
    """
    
    def __init__(self, cell_info: str, content: str = ""):
        super().__init__()
        self.cell_info = cell_info
        self.content = content
        self.saved = False
    
    def compose(self) -> ComposeResult:
        with Container(id="modal-content"):
            yield Static(self.cell_info, id="modal-header")
            yield TextArea(self.content, id="text-editor")
            yield Static("Ctrl+S to save | ESC to cancel", id="hint")
    
    def on_mount(self):
        self.query_one("#text-editor", TextArea).focus()
    
    async def on_key(self, event: events.Key) -> None:
        if event.key == "ctrl+s":
            editor = self.query_one("#text-editor", TextArea)
            self.dismiss(editor.text.strip())
        elif event.key == "escape":
            self.dismiss(None)


class CalendarScreen(Screen):
    """Main calendar screen with wide cells."""
    
    CSS = """
    CalendarScreen {
        background: $background;
    }
    
    #top-bar {
        height: 3;
        background: $surface;
        border-bottom: solid $border;
        padding: 0 2;
    }
    
    #header-content {
        width: 100%;
        height: 100%;
        align: center middle;
    }
    
    #month-label {
        text-align: center;
        text-style: bold;
        color: $primary;
        width: auto;
    }
    
    #demo-indicator {
        background: $warning;
        color: $background;
        text-style: bold;
        padding: 0 2;
        margin: 0 2;
        width: auto;
        border: solid $border;
    }
    
    #calendar-container {
        padding: 1;
        height: 1fr;
    }
    
    DataTable {
        scrollbar-size: 1 1;
        width: 100%;
        height: 100%;
        scrollbar-background: $surface;
        scrollbar-color: $border;
    }
    
    DataTable > .datatable--header {
        background: $surface;
        color: $primary;
        text-style: bold;
    }
    
    DataTable > .datatable--cursor {
        background: $accent 30%;
    }
    
    DataTable > .datatable--odd-row {
        background: $background;
    }
    
    DataTable > .datatable--even-row {
        background: $surface 50%;
    }
    
    #bottom-bar {
        height: 2;
        background: $surface;
        border-top: solid $border;
        padding: 0 2;
        dock: bottom;
    }
    
    #keys-hint {
        text-align: center;
        color: $text-muted;
    }
    """
    
    BINDINGS = [
        Binding("q", "quit_app", "Quit", priority=True),
        Binding("ctrl+c", "quit_app", "Quit", show=False),
        Binding("left", "prev_month", "← Prev"),
        Binding("right", "next_month", "Next →"),
        Binding("enter", "edit_cell", "Edit"),
        Binding("e", "edit_cell", show=False),
        Binding("t", "today", "Today"),
        Binding("r", "refresh", "Refresh"),
    ]
    
    def __init__(self, demo_mode=False, oauth_service=None):
        super().__init__()
        self.current_month = datetime.now().replace(day=1)
        self.demo_mode = demo_mode
        self.oauth_service = oauth_service
        self.data: Dict[str, str] = {}
        
        # Load demo data if in demo mode
        if demo_mode:
            self.load_demo_data()
        
        # Tokyo Night inspired colors
        self.style_week = Style(color="rgb(187,154,247)", bold=True)  # Purple for week numbers
        self.style_today = Style(bgcolor="rgb(65,66,89)", color="rgb(255,158,100)", bold=True)  # Orange on dark purple
        self.style_weekend = Style(color="rgb(86,95,137)")  # Muted blue-gray for weekends
        self.style_normal = Style(color="rgb(169,177,214)")  # Light blue-gray for normal days
        self.style_content = Style(color="rgb(125,207,255)")  # Bright cyan for content
        self.style_separator = Style(color="rgb(65,66,89)")  # Border color for separators
    
    def load_demo_data(self):
        """Load sample data for demo mode."""
        self.data = {
            # Week 1
            "1_0": "VIS-1234 Authentication\nImplement JWT tokens\n2h estimated",  # Monday W1
            "1_1": "Daily standup 9:30am\nDiscuss blockers",  # Tuesday W1
            "1_2": "Sprint planning 10am\nReview backlog items\nAssign tasks",  # Wednesday W1
            "1_3": "VIS-1235 API endpoints\nCreate user CRUD\n4h work",  # Thursday W1
            "1_4": "Deploy to production\nMonitor metrics\nSmoke testing",  # Friday W1
            # Week 2
            "2_0": "Code review session\nPR feedback\nMerge approved",  # Monday W2
            "2_1": "VIS-1236 Bug fixes\nInvestigate issues\nWrite tests",  # Tuesday W2
            "2_2": "Architecture meeting\nDesign decisions\n2pm - 4pm",  # Wednesday W2
            "2_3": "Team standup\nBlockers discussion\nUpdate tickets",  # Thursday W2
            "2_4": "Documentation day\nUpdate README\nAPI docs",  # Friday W2
            # Week 3
            "3_0": "VIS-1237 Refactoring\nClean up tech debt",  # Monday W3
            "3_1": "Security audit\nReview permissions\nFix vulnerabilities",  # Tuesday W3
            "3_2": "Performance testing\nLoad testing\nOptimizations",  # Wednesday W3
            "3_3": "Client demo prep\nPrepare presentation",  # Thursday W3
            "3_4": "Sprint retrospective\n2pm team meeting\nLessons learned",  # Friday W3
            # Week 4
            "4_0": "New sprint kickoff\nPlanning poker",  # Monday W4
            "4_2": "VIS-1238 Feature flag\nImplement toggles",  # Wednesday W4
            "4_4": "End of sprint\nRelease notes",  # Friday W4
        }
    
    def compose(self) -> ComposeResult:
        with Container(id="top-bar"):
            with Horizontal(id="header-content"):
                yield Static(
                    f"📅 {self.current_month.strftime('%B %Y')}",
                    id="month-label"
                )
                if self.demo_mode:
                    yield Static("🔶 DEMO MODE 🔶", id="demo-indicator")
        
        with Container(id="calendar-container"):
            table = DataTable(
                show_header=True,
                show_row_labels=False,
                cursor_type="cell",
                zebra_stripes=True,
                id="calendar"
            )
            yield table
        
        with Container(id="bottom-bar"):
            yield Static(
                "Q:Quit | ←→:Navigate | Enter:Edit | T:Today | R:Refresh",
                id="keys-hint"
            )
    
    async def on_mount(self):
        await self.build_calendar()
    
    async def build_calendar(self):
        """Build calendar with full-width cells."""
        table = self.query_one("#calendar", DataTable)
        table.clear(columns=True)
        
        # Wider columns to fill screen
        columns = [
            ("Week", 6),
            ("Monday", 28),
            ("Tuesday", 28),
            ("Wednesday", 28),
            ("Thursday", 28),
            ("Friday", 28),
            ("Saturday", 28),
            ("Sunday", 28),
        ]
        
        for name, width in columns:
            table.add_column(name, width=width)
        
        # Calculate month
        year = self.current_month.year
        month = self.current_month.month
        first_day = self.current_month.replace(day=1)
        days_in_month = monthrange(year, month)[1]
        first_weekday = first_day.weekday()
        today = datetime.now().date()
        
        # Build weeks with multi-line cells
        week_rows = []
        week_num = 1
        
        for week_start in range(1 - first_weekday, days_in_month + 1, 7):
            row_cells = [Text(f"W{week_num}", style=self.style_week)]
            
            for day_offset in range(7):
                day = week_start + day_offset
                
                if 1 <= day <= days_in_month:
                    date = self.current_month.replace(day=day)
                    cell_key = f"{week_num}_{day_offset}"
                    content = self.data.get(cell_key, "")
                    
                    # Create multi-line cell with more content
                    cell_text = Text()
                    
                    # Day number with style
                    day_str = f" {day:2d} "
                    if date.date() == today:
                        cell_text.append(day_str, style=self.style_today)
                    elif day_offset >= 5:  # Weekend
                        cell_text.append(day_str, style=self.style_weekend)
                    else:
                        cell_text.append(day_str, style=self.style_normal)
                    
                    cell_text.append("\n")
                    cell_text.append("─" * 25, style=self.style_separator)
                    cell_text.append("\n")
                    
                    if content:
                        # Show first 3 lines of content with content style
                        content_lines = content.split('\n')[:3]
                        for i, line in enumerate(content_lines):
                            if i > 0:
                                cell_text.append("\n")
                            if len(line) > 25:
                                cell_text.append(line[:24] + "…", style=self.style_content)
                            else:
                                cell_text.append(line, style=self.style_content)
                        # Pad remaining lines
                        for _ in range(3 - len(content_lines)):
                            cell_text.append("\n")
                    else:
                        # Empty cell - just add spacing
                        cell_text.append("\n\n")
                    
                    row_cells.append(cell_text)
                else:
                    # Empty cell
                    row_cells.append(Text(""))
            
            week_rows.append(row_cells)
            week_num += 1
        
        # Add rows to table (ensure at least 10 weeks to fill screen)
        while len(week_rows) < 10:
            empty_row = [Text(f"W{week_num}", style=self.style_week)]
            empty_row.extend([Text("") for _ in range(7)])
            week_rows.append(empty_row)
            week_num += 1
        
        for row in week_rows:
            table.add_row(*row, height=4)  # Taller rows for better visibility
    
    def action_quit_app(self):
        """Quit the application."""
        self.app.exit()
    
    def action_prev_month(self):
        """Previous month."""
        self.current_month = (self.current_month.replace(day=1) - timedelta(days=1)).replace(day=1)
        self.query_one("#month-label").update(f"📅 {self.current_month.strftime('%B %Y')}")
        self.run_worker(self.build_calendar())
    
    def action_next_month(self):
        """Next month."""
        if self.current_month.month == 12:
            self.current_month = self.current_month.replace(year=self.current_month.year + 1, month=1)
        else:
            self.current_month = self.current_month.replace(month=self.current_month.month + 1)
        self.query_one("#month-label").update(f"📅 {self.current_month.strftime('%B %Y')}")
        self.run_worker(self.build_calendar())
    
    def action_today(self):
        """Jump to current month."""
        self.current_month = datetime.now().replace(day=1)
        self.query_one("#month-label").update(f"📅 {self.current_month.strftime('%B %Y')}")
        self.run_worker(self.build_calendar())
    
    def action_refresh(self):
        """Refresh the calendar."""
        self.run_worker(self.build_calendar())
    
    def action_edit_cell(self):
        """Edit selected cell."""
        table = self.query_one("#calendar", DataTable)
        cursor = table.cursor_coordinate
        
        if cursor and cursor.column > 0:  # Skip week column
            week_row = cursor.row + 1  # Week number
            day_col = cursor.column - 1  # Day of week
            
            # Calculate the actual date
            first_day = self.current_month.replace(day=1)
            first_weekday = first_day.weekday()
            
            # Calculate day of month
            days_from_start = ((week_row - 1) * 7) + day_col - first_weekday
            
            if days_from_start >= 0:
                try:
                    date = first_day + timedelta(days=days_from_start)
                    
                    if date.month == self.current_month.month:
                        cell_key = f"{week_row}_{day_col}"
                        content = self.data.get(cell_key, "")
                        
                        days = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"]
                        cell_info = f"{days[day_col]}, {date.strftime('%B %d, %Y')}"
                        
                        def handle_result(result: Optional[str]):
                            if result is not None:
                                if result:
                                    self.data[cell_key] = result
                                elif cell_key in self.data:
                                    del self.data[cell_key]
                                self.run_worker(self.build_calendar())
                        
                        self.app.push_screen(EditModal(cell_info, content), handle_result)
                except ValueError:
                    pass


class BetterSheetsApp(App):
    """Better Google Sheets Calendar TUI."""
    
    CSS = """
    /* Tokyo Night color scheme */
    $background: rgb(26,27,38);
    $surface: rgb(36,40,59);
    $primary: rgb(125,207,255);
    $accent: rgb(187,154,247);
    $warning: rgb(255,158,100);
    $error: rgb(247,118,142);
    $success: rgb(158,206,106);
    $text: rgb(195,212,233);
    $text-muted: rgb(86,95,137);
    $border: rgb(65,66,89);
    """
    
    def __init__(self, demo_mode=True, oauth_service=None):
        super().__init__()
        self.demo_mode = demo_mode
        self.oauth_service = oauth_service
    
    def on_mount(self):
        self.title = "Sheets Calendar"
        # Push the calendar screen with proper parameters
        screen = CalendarScreen(demo_mode=self.demo_mode, oauth_service=self.oauth_service)
        self.push_screen(screen)
    
    def action_quit(self):
        """Quit and clean up terminal."""
        import sys
        # Reset terminal state
        sys.stdout.write("\033[?1000l")  # Disable mouse tracking
        sys.stdout.write("\033[?1003l")  # Disable mouse motion tracking
        sys.stdout.write("\033[?1015l")  # Disable urxvt mouse mode
        sys.stdout.write("\033[?1006l")  # Disable SGR mouse mode
        sys.stdout.flush()
        self.exit()


if __name__ == "__main__":
    import sys
    try:
        # Run in demo mode for testing
        app = BetterSheetsApp(demo_mode=True, oauth_service=None)
        app.run()
    finally:
        # Clean up terminal on exit
        sys.stdout.write("\033[?1000l\033[?1003l\033[?1015l\033[?1006l")
        sys.stdout.flush()