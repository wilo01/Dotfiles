"""Base TUI framework for all Textual applications."""

import logging
from abc import ABC, abstractmethod
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

from textual import events
from textual.app import App, ComposeResult
from textual.binding import Binding
from textual.containers import Container, Horizontal, Vertical
from textual.message import Message
from textual.screen import ModalScreen, Screen
from textual.widgets import Button, DataTable, Footer, Header, Input, Label, Static

logger = logging.getLogger(__name__)


class BaseModal(ModalScreen):
    """Base modal for all TUI modals."""

    BINDINGS = [
        Binding("escape", "dismiss", "Cancel"),
    ]

    def __init__(self, title: str = "Modal", **kwargs):
        """Initialize base modal."""
        super().__init__(**kwargs)
        self.title = title
        self.result = None

    def on_key(self, event: events.Key) -> None:
        """Handle key events."""
        if event.key == "escape":
            self.dismiss(None)

    def dismiss(self, result: Any = None) -> None:
        """Dismiss modal with result."""
        self.result = result
        super().dismiss(result)


class EditModal(BaseModal):
    """Standard edit modal for text input."""

    def __init__(
        self, title: str, initial_value: str = "", placeholder: str = "", **kwargs
    ):
        """Initialize edit modal."""
        super().__init__(title=title, **kwargs)
        self.initial_value = initial_value
        self.placeholder = placeholder

    def compose(self) -> ComposeResult:
        """Compose the modal UI."""
        with Vertical(id="edit-modal"):
            yield Label(self.title, id="modal-title")
            yield Input(
                value=self.initial_value, placeholder=self.placeholder, id="modal-input"
            )
            with Horizontal(id="modal-buttons"):
                yield Button("Save", variant="primary", id="save-button")
                yield Button("Cancel", variant="default", id="cancel-button")

    def on_button_pressed(self, event: Button.Pressed) -> None:
        """Handle button press."""
        if event.button.id == "save-button":
            input_widget = self.query_one("#modal-input", Input)
            self.dismiss(input_widget.value)
        else:
            self.dismiss(None)

    def on_input_submitted(self, event: Input.Submitted) -> None:
        """Handle enter key in input."""
        self.dismiss(event.value)


class BaseScreen(Screen):
    """Base screen for all TUI screens."""

    def __init__(self, name: Optional[str] = None, **kwargs):
        """Initialize base screen."""
        super().__init__(name=name, **kwargs)
        self.data = {}
        self.modified = False

    def mark_modified(self) -> None:
        """Mark screen as modified."""
        self.modified = True
        self.app.title = (
            f"{self.app.title} *"
            if not self.app.title.endswith("*")
            else self.app.title
        )

    def clear_modified(self) -> None:
        """Clear modified flag."""
        self.modified = False
        if self.app.title.endswith(" *"):
            self.app.title = self.app.title[:-2]


class BaseSheetsApp(App[None], ABC):
    """Base class for all sheets-related TUI applications."""

    # Common bindings for all TUI apps
    BINDINGS = [
        Binding("ctrl+s", "save_data", "Save"),
        Binding("ctrl+r", "refresh_data", "Refresh"),
        Binding("ctrl+q", "quit", "Quit"),
        Binding("escape", "cancel_edit", "Cancel", show=False),
        Binding("?", "show_help", "Help"),
    ]

    # CSS for common styling
    CSS = """
    #header {
        background: $boost;
        color: $text;
        height: 3;
        padding: 1;
    }

    #status-bar {
        background: $panel;
        color: $text-muted;
        height: 1;
        padding: 0 1;
    }

    DataTable {
        background: $surface;
        color: $text;
    }

    DataTable > .datatable--cursor {
        background: $primary;
        color: $text;
    }

    DataTable > .datatable--header {
        background: $panel;
        color: $text;
        text-style: bold;
    }

    #edit-modal {
        background: $surface;
        border: solid $primary;
        padding: 1 2;
        width: 60;
        height: auto;
    }

    #modal-title {
        text-style: bold;
        margin-bottom: 1;
    }

    #modal-input {
        margin-bottom: 1;
    }

    #modal-buttons {
        align: center middle;
        height: 3;
    }

    Button {
        margin: 0 1;
    }

    .status-saved {
        color: $success;
    }

    .status-error {
        color: $error;
    }

    .status-modified {
        color: $warning;
    }
    """

    def __init__(self, sheet_service: Optional[Any] = None, **kwargs):
        """
        Initialize base TUI app.

        Args:
            sheet_service: Google Sheets service instance
        """
        super().__init__(**kwargs)
        self.sheet_service = sheet_service
        self.current_sheet_id = None
        self.data_cache = {}
        self.modified = False
        self.status_message = "Ready"
        self.last_save = None
        self.auto_save_enabled = True
        self.auto_save_interval = 30  # seconds

    def compose(self) -> ComposeResult:
        """Compose the base UI structure."""
        yield Header()
        yield from self.compose_main_content()
        yield Static(self.status_message, id="status-bar")
        yield Footer()

    @abstractmethod
    def compose_main_content(self) -> ComposeResult:
        """Compose the main content area. Must be implemented by subclasses."""
        pass

    def on_mount(self) -> None:
        """Handle mount event."""
        self.title = self.get_app_title()
        self.load_initial_data()

        if self.auto_save_enabled:
            self.set_interval(self.auto_save_interval, self.auto_save)

    @abstractmethod
    def get_app_title(self) -> str:
        """Get the application title. Must be implemented by subclasses."""
        pass

    @abstractmethod
    def load_initial_data(self) -> None:
        """Load initial data. Must be implemented by subclasses."""
        pass

    def action_save_data(self) -> None:
        """Save data action."""
        if self.modified:
            success = self.save_data()
            if success:
                self.update_status("✅ Data saved successfully")
                self.modified = False
                self.last_save = datetime.now()
            else:
                self.update_status("❌ Failed to save data", error=True)
        else:
            self.update_status("No changes to save")

    @abstractmethod
    def save_data(self) -> bool:
        """Save data to storage. Must be implemented by subclasses."""
        pass

    def action_refresh_data(self) -> None:
        """Refresh data action."""
        if self.modified:
            # Confirm before refreshing if there are unsaved changes
            self.confirm_refresh()
        else:
            self.refresh_data()

    @abstractmethod
    def refresh_data(self) -> None:
        """Refresh data from source. Must be implemented by subclasses."""
        pass

    def confirm_refresh(self) -> None:
        """Confirm refresh when there are unsaved changes."""
        # This should show a confirmation modal
        # For now, just refresh
        self.refresh_data()

    def action_quit(self) -> None:
        """Quit action."""
        if self.modified:
            # Confirm before quitting if there are unsaved changes
            self.confirm_quit()
        else:
            self.exit()

    def confirm_quit(self) -> None:
        """Confirm quit when there are unsaved changes."""
        # This should show a confirmation modal
        # For now, just quit
        self.exit()

    def action_cancel_edit(self) -> None:
        """Cancel current edit action."""
        # Override in subclasses if needed
        pass

    def action_show_help(self) -> None:
        """Show help screen."""
        self.push_screen(self.create_help_screen())

    def create_help_screen(self) -> Screen:
        """Create help screen. Can be overridden by subclasses."""
        return HelpScreen()

    def update_status(
        self, message: str, error: bool = False, temporary: bool = True
    ) -> None:
        """
        Update status bar message.

        Args:
            message: Status message
            error: Whether this is an error message
            temporary: Whether to clear message after timeout
        """
        self.status_message = message
        status_bar = self.query_one("#status-bar", Static)

        if error:
            status_bar.add_class("status-error")
        else:
            status_bar.remove_class("status-error")

        status_bar.update(message)

        if temporary:
            self.set_timer(3, lambda: self.clear_status())

    def clear_status(self) -> None:
        """Clear status message."""
        self.update_status("Ready", error=False, temporary=False)

    def mark_modified(self) -> None:
        """Mark data as modified."""
        if not self.modified:
            self.modified = True
            self.title = f"{self.get_app_title()} *"
            self.update_status("Modified - press Ctrl+S to save", temporary=False)

    def clear_modified(self) -> None:
        """Clear modified flag."""
        if self.modified:
            self.modified = False
            self.title = self.get_app_title()
            self.clear_status()

    def auto_save(self) -> None:
        """Auto-save if data is modified."""
        if self.modified and self.auto_save_enabled:
            self.action_save_data()

    def format_cell_value(self, value: Any) -> str:
        """
        Format cell value for display.

        Args:
            value: Cell value

        Returns:
            Formatted string
        """
        if value is None:
            return ""
        if isinstance(value, bool):
            return "✓" if value else "✗"
        if isinstance(value, (int, float)):
            return str(value)
        if isinstance(value, datetime):
            return value.strftime("%Y-%m-%d %H:%M")
        return str(value)

    def validate_input(
        self, value: str, field_type: str = "text"
    ) -> Tuple[bool, Any, Optional[str]]:
        """
        Validate and convert input value.

        Args:
            value: Input value
            field_type: Expected field type

        Returns:
            Tuple of (is_valid, converted_value, error_message)
        """
        if not value and field_type != "text":
            return True, None, None

        try:
            if field_type == "int":
                return True, int(value), None
            elif field_type == "float":
                return True, float(value), None
            elif field_type == "bool":
                val = value.lower() in ("true", "yes", "1", "✓")
                return True, val, None
            elif field_type == "date":
                from datetime import datetime

                dt = datetime.strptime(value, "%Y-%m-%d")
                return True, dt, None
            elif field_type == "email":
                import re

                if re.match(r"[^@]+@[^@]+\.[^@]+", value):
                    return True, value, None
                return False, None, "Invalid email format"
            else:  # text
                return True, value, None

        except ValueError as e:
            return False, None, f"Invalid {field_type}: {e}"


class HelpScreen(Screen):
    """Default help screen."""

    BINDINGS = [
        Binding("escape", "dismiss", "Close"),
    ]

    def compose(self) -> ComposeResult:
        """Compose help screen."""
        with Container(id="help-container"):
            yield Label("Help", id="help-title")
            yield Static(self.get_help_text(), id="help-content")
            yield Button("Close", id="close-button")

    def get_help_text(self) -> str:
        """Get help text."""
        return """
        Keyboard Shortcuts:

        Ctrl+S  - Save changes
        Ctrl+R  - Refresh data
        Ctrl+Q  - Quit application
        Enter   - Edit selected cell
        Tab     - Move to next cell
        Arrows  - Navigate cells
        ?       - Show this help
        Escape  - Cancel edit / Close dialog

        Tips:
        • Changes are auto-saved every 30 seconds
        • Modified data is marked with * in title
        • Status bar shows current operation status
        """

    def on_button_pressed(self, event: Button.Pressed) -> None:
        """Handle button press."""
        self.dismiss()

    def action_dismiss(self) -> None:
        """Dismiss help screen."""
        self.app.pop_screen()
