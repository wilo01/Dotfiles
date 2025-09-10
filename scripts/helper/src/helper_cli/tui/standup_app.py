"""Interactive TUI spreadsheet for daily standup notes."""

from datetime import datetime
from typing import List, Dict, Any, Optional
import pyperclip

from textual.app import App, ComposeResult
from textual.widgets import DataTable, Header, Footer, Static
from textual.containers import Container, Vertical, Horizontal
from textual.binding import Binding
from textual.coordinate import Coordinate
from textual.screen import ModalScreen
from textual.widgets import Input, Label, Button
from textual import events


class CellEditModal(ModalScreen):
    """Modal for editing cell content."""

    CSS = """
    CellEditModal {
        align: center middle;
    }

    #edit-container {
        background: $surface;
        padding: 1 2;
        border: solid $primary;
        width: 60;
        height: 9;
    }

    #edit-title {
        text-style: bold;
        margin-bottom: 1;
    }

    #edit-input {
        width: 100%;
        margin-bottom: 1;
    }

    #button-container {
        align: center middle;
        height: 3;
    }
    """

    def __init__(self, title: str, initial_value: str = ""):
        super().__init__()
        self.title = title
        self.initial_value = initial_value
        self.result = None

    def compose(self) -> ComposeResult:
        with Vertical(id="edit-container"):
            yield Label(self.title, id="edit-title")
            yield Input(value=self.initial_value, id="edit-input")
            with Horizontal(id="button-container"):
                yield Button("Save", variant="primary", id="save")
                yield Button("Cancel", variant="default", id="cancel")

    def on_mount(self):
        self.query_one("#edit-input", Input).focus()

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "save":
            self.result = self.query_one("#edit-input", Input).value
            self.dismiss(self.result)
        else:
            self.dismiss(None)

    def on_key(self, event: events.Key) -> None:
        if event.key == "escape":
            self.dismiss(None)
        elif event.key == "enter":
            self.result = self.query_one("#edit-input", Input).value
            self.dismiss(self.result)


class StandupApp(App):
    """Interactive standup notes TUI application."""

    CSS = """
    StandupApp {
        background: $background;
    }

    #header {
        height: 3;
        background: $primary;
        color: $text;
        text-align: center;
        text-style: bold;
        padding: 1;
    }

    #table-container {
        height: 1fr;
        border: solid $primary;
        margin: 0 1;
    }

    DataTable {
        height: 100%;
    }

    #footer {
        height: 3;
        background: $surface;
        color: $text;
        padding: 1;
    }

    #status {
        text-align: center;
        color: $success;
        height: 1;
        margin: 0 1;
    }
    """

    BINDINGS = [
        Binding("q", "quit", "Quit", priority=True),
        Binding("s", "save", "Save", priority=True),
        Binding("c", "copy", "Copy to Clipboard"),
        Binding("r", "refresh", "Refresh Tickets"),
        Binding("enter", "edit_cell", "Edit Cell", show=False),
        Binding("e", "export", "Export"),
    ]

    def __init__(self, jira_service=None, tickets=None):
        super().__init__()
        self.jira_service = jira_service
        self.tickets = tickets or []
        self.notes = {}  # ticket_key -> description
        self.current_cell = None
        self.title = f"Daily Standup - {datetime.now().strftime('%B %d, %Y')}"

    def compose(self) -> ComposeResult:
        """Create the UI layout."""
        yield Static(self.title, id="header")
        with Container(id="table-container"):
            yield DataTable(cursor_type="cell", zebra_stripes=True)
        yield Static("", id="status")
        yield Footer()

    def on_mount(self) -> None:
        """Initialize the data table when app starts."""
        table = self.query_one(DataTable)

        # Add columns
        table.add_column("Ticket", width=10, key="ticket")
        table.add_column("Summary", width=40, key="summary")
        table.add_column("Status", width=12, key="status")
        table.add_column("Today's Work", width=50, key="today")

        # Load tickets and populate table
        self.load_tickets()
        self.populate_table()

        # Load previous notes if available
        if self.jira_service:
            self.load_previous_notes()

    def load_tickets(self) -> None:
        """Load tickets from Jira service or use provided tickets."""
        if not self.tickets and self.jira_service:
            self.tickets = self.jira_service.get_relevant_tickets_for_standup()

    def populate_table(self) -> None:
        """Populate the data table with ticket information."""
        table = self.query_one(DataTable)

        for ticket in self.tickets[:15]:  # Limit to 15 for UI
            key = ticket['key']
            fields = ticket.get('fields', {})
            summary = fields.get('summary', 'No summary')[:40]
            status = fields.get('status', {}).get('name', 'Unknown')

            # Format status with emoji
            if status == 'Done':
                status_display = "✓ Done"
            elif 'Progress' in status:
                status_display = "⚡ Progress"
            elif status in ['To Do', 'Backlog']:
                status_display = f"○ {status}"
            else:
                status_display = status[:12]

            # Get today's work (empty initially or from saved notes)
            today_work = self.notes.get(key, "")

            table.add_row(key, summary, status_display, today_work, key=key)

    def load_previous_notes(self) -> None:
        """Load previous standup notes as suggestions."""
        if not self.jira_service:
            return

        table = self.query_one(DataTable)

        for row_key in table.rows:
            ticket_key = str(row_key.value)
            # Get previous description
            prev_desc = self.jira_service.get_recent_standup_description(ticket_key)
            if prev_desc:
                # Show as placeholder/hint (dimmed)
                self.notes[ticket_key] = ""  # Keep empty but we know there's history

    def on_data_table_cell_selected(self, event: DataTable.CellSelected) -> None:
        """Handle cell selection for editing."""
        table = self.query_one(DataTable)

        # Only allow editing the "Today's Work" column
        if event.coordinate.column == 3:  # Today's Work column
            row_key = event.row_key
            ticket_key = str(row_key.value)
            current_value = self.notes.get(ticket_key, "")

            # Get ticket info for the modal title
            ticket_row = table.get_row(row_key)
            ticket_num = ticket_row[0]

            # Show edit modal
            self.push_screen(
                CellEditModal(f"Edit notes for {ticket_num}", current_value),
                self.handle_edit_result
            )

    def handle_edit_result(self, result: Optional[str]) -> None:
        """Handle the result from the edit modal."""
        if result is not None:
            table = self.query_one(DataTable)
            selected = table.cursor_coordinate

            if selected and selected.column == 3:
                row_key = table.coordinate_to_cell_key(selected).row_key
                ticket_key = str(row_key.value)

                # Update notes
                self.notes[ticket_key] = result

                # Update table cell
                table.update_cell_at(selected, result)

                # Show status
                self.update_status("✓ Updated")

    def action_edit_cell(self) -> None:
        """Action to edit the current cell."""
        table = self.query_one(DataTable)
        if table.cursor_coordinate:
            # Simulate cell selection
            event = DataTable.CellSelected(
                table,
                table.cursor_coordinate,
                table.coordinate_to_cell_key(table.cursor_coordinate)
            )
            self.on_data_table_cell_selected(event)

    def action_save(self) -> None:
        """Save notes to database."""
        if not self.jira_service:
            self.update_status("❌ No Jira service available")
            return

        # Prepare notes for saving
        standup_notes = []
        table = self.query_one(DataTable)

        for row_key in table.rows:
            ticket_key = str(row_key.value)
            if ticket_key in self.notes and self.notes[ticket_key]:
                row = table.get_row(row_key)
                standup_notes.append({
                    'ticket_key': ticket_key,
                    'ticket_summary': row[1],  # Summary column
                    'description': self.notes[ticket_key]
                })

        if standup_notes:
            self.jira_service.save_standup_notes(standup_notes)
            self.update_status("✅ Saved to database")
        else:
            self.update_status("⚠️ No notes to save")

    def action_copy(self) -> None:
        """Copy notes to clipboard in tab-separated format."""
        # Generate sheets format
        sheets_output = []
        table = self.query_one(DataTable)

        for row_key in table.rows:
            ticket_key = str(row_key.value)
            if ticket_key in self.notes and self.notes[ticket_key]:
                row = table.get_row(row_key)
                sheets_output.append(f"{ticket_key}\t{row[1]}\t{self.notes[ticket_key]}")

        if sheets_output:
            clipboard_text = "\n".join(sheets_output)
            try:
                pyperclip.copy(clipboard_text)
                self.update_status("📋 Copied to clipboard!")
            except:
                self.update_status("❌ Clipboard not available")
        else:
            self.update_status("⚠️ No notes to copy")

    def action_export(self) -> None:
        """Export notes as markdown."""
        output = f"# Standup Notes - {datetime.now().strftime('%B %d, %Y')}\n\n"

        table = self.query_one(DataTable)
        has_notes = False

        for row_key in table.rows:
            ticket_key = str(row_key.value)
            if ticket_key in self.notes and self.notes[ticket_key]:
                row = table.get_row(row_key)
                output += f"- **{ticket_key}**: {self.notes[ticket_key]}\n"
                has_notes = True

        if has_notes:
            # Save to file or copy to clipboard
            try:
                pyperclip.copy(output)
                self.update_status("📋 Exported as markdown to clipboard")
            except:
                self.update_status("✅ Export ready")
        else:
            self.update_status("⚠️ No notes to export")

    def action_refresh(self) -> None:
        """Refresh ticket list."""
        if self.jira_service:
            self.update_status("🔄 Refreshing...")
            self.tickets = self.jira_service.get_relevant_tickets_for_standup()

            # Clear and repopulate table
            table = self.query_one(DataTable)
            table.clear()
            self.populate_table()
            self.update_status("✅ Refreshed")
        else:
            self.update_status("❌ No Jira service available")

    def update_status(self, message: str) -> None:
        """Update the status bar message."""
        status = self.query_one("#status", Static)
        status.update(message)

        # Clear status after 3 seconds
        self.set_timer(3.0, lambda: status.update(""))


def run_standup_tui(jira_service=None, tickets=None):
    """Run the standup TUI application."""
    app = StandupApp(jira_service=jira_service, tickets=tickets)
    return app.run()
