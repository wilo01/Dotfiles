"""Core functionality for the helper CLI tool."""

import re
import time
import random
import shutil
import os
from typing import Optional
import pyshorteners
import pyperclip
import segno


class HelperCore:
    """Core helper functionality."""

    def __init__(self):
        self.type_tiny = pyshorteners.Shortener()
        self.rfid_badge_default = '@4324325553#'
        self.qr_badge_default = '$4324325553#'
        self.keyboard = None

    def _get_keyboard(self):
        """Lazy load keyboard controller."""
        if self.keyboard is None:
            try:
                from pynput.keyboard import Controller
                self.keyboard = Controller()
            except ImportError as e:
                print(f"Error: Cannot use keyboard automation: {e}")
                print("Running in CLI-only mode. Text copied to clipboard instead.")
                return None
        return self.keyboard

    def type_string_with_delay(self, string: str, delay: int = 2):
        """Type a string with delay, fallback to clipboard if keyboard unavailable."""
        keyboard = self._get_keyboard()
        if keyboard is None:
            pyperclip.copy(string)
            return

        time.sleep(delay)
        for character in string:
            keyboard.type(character)

    def create_temp_dir(self, path: str) -> str:
        """Create temporary directory."""
        os.makedirs(path, exist_ok=True)
        return path

    def copy_and_override_files(self, source1: str, source2: str, temp_dir: str):
        """Copy files from two source directories, with second overriding first."""
        try:
            file_names1 = os.listdir(source1)
            file_names2 = os.listdir(source2)

            for file_name in file_names1:
                shutil.copy(os.path.join(source1, file_name), temp_dir)

            for file_name in file_names2:
                shutil.copy(os.path.join(source2, file_name), temp_dir)

        except (OSError, IOError) as e:
            print(f"Error copying files: {e}")

    def format_string(self, text: str, mode: str = 'standard', uppercase_tags: bool = True, clean_dashes: bool = True, strip_edges: bool = True) -> str:
        """Unified string formatting function with configurable options.

        Args:
            text: The text to format
            mode: The formatting mode ('jira', 'stash', 'standard', 'dash', etc.)
            uppercase_tags: Whether to uppercase TAG-NUMBER patterns (default: True)
            clean_dashes: Whether to clean up multiple consecutive dashes (default: True)
            strip_edges: Whether to strip leading/trailing dashes (default: True)

        Returns:
            Formatted string based on the specified mode and options
        """
        text = text.strip()

        # Convert to lowercase and replace special chars with dashes
        formatted = re.sub(r"[ !+@#$%^&*(),_.'/:;>\[\]\\-]", "-", text.lower().replace('\n', '-'))

        if uppercase_tags:
            def uppercase_tag_match(match):
                tag = match.group(1).upper()
                number = match.group(2)
                return f"{tag}-{number}"

            formatted = re.sub(r'\b([a-zA-Z]+)-(\d+)\b', uppercase_tag_match, formatted, flags=re.IGNORECASE)

        if clean_dashes:
            formatted = re.sub(r'-+', '-', formatted)

        if strip_edges:
            formatted = formatted.strip('-')

        if mode == 'jira':
            return f'git checkout -b {formatted}'
        elif mode == 'stash':
            return f'git stash push -u -m {formatted}'
        elif mode == 'dash':
            return formatted
        else:
            return formatted

    def generate_jira_branch_name(self, text: str) -> str:
        """Generate JIRA branch name from text."""
        return self.format_string(text, mode='jira')

    def generate_stash_command(self, text: str) -> str:
        """Generate git stash command with formatted message."""
        return self.format_string(text, mode='stash')

    def generate_ngrok_qr(self, ngrok_url: str) -> tuple[str, str]:
        """Generate QR code command for NGROK URL."""
        full_url = f'{ngrok_url}/i/source/ui-kiosk/index.html'
        short_url = self.type_tiny.tinyurl.short(full_url)
        qr_command = f'segno --compact {short_url}'
        return full_url, qr_command

    def generate_vsc_qr(self, vsc_url: str) -> tuple[str, str]:
        """Generate QR code command for VSC URL."""
        full_url = f'{vsc_url}i/source/ui-kiosk/index.html'
        short_url = self.type_tiny.tinyurl.short(full_url)
        qr_command = f'segno --compact {short_url}'
        return full_url, qr_command

    def generate_apex_upload_command(self, apex_file: str, port: str, auth: str) -> str:
        """Generate APEX upload curl command."""
        command = (
            f'curl -v --request POST --url http://127.0.0.1:{port}/apex/_/resources.zip '
            f'--header "Authorization: Basic {auth}" '
            f'--form name==file '
            f'--form "filename=@{apex_file}.zip;type=application/zip" | > {apex_file}.html'
        )
        return command

    def setup_rt_temp_directory(self, rt_name: str):
        """Set up temporary RT directory with file copying."""
        path_to_temp_dir = f'~/Dev/branch-opener2/{rt_name}'
        source_dir1 = '~/Dev/branch-opener2/branches/safe/source/server/rt'
        source_dir2 = '~/Dev/branch-opener2/branches/safe/source/server/rtsp'

        temp_dir = self.create_temp_dir(path_to_temp_dir)
        self.copy_and_override_files(source_dir1, source_dir2, temp_dir)
        print(f'Created RT directory: {temp_dir}')

    def convert_to_dash_format(self, text: str) -> str:
        """Convert text to dash-separated format for filenames."""
        return self.format_string(text, mode='dash', uppercase_tags=False)

    def generate_pr_title(self, text: str) -> str:
        """Generate PR title with TAG-NUMBER uppercased but preserving spaces.
        Example: 'vis-1234 add new feature' -> 'VIS-1234 add new feature'
        """
        text = text.strip()
        def uppercase_tag_match(match):
            tag = match.group(1).upper()
            number = match.group(2)
            return f"{tag}-{number}"

        formatted = re.sub(r'\b([a-zA-Z]+)-(\d+)\b', uppercase_tag_match, text, flags=re.IGNORECASE)
        return formatted

    def generate_filename(self, text: str) -> str:
        """Generate clean filename without tag uppercasing.
        Example: 'VIS-1234 My Report' -> 'vis-1234-my-report.md'
        """
        formatted = self.format_string(text, mode='dash', uppercase_tags=False)
        return f"{formatted}.md"

    def detect_vis_ticket_from_context(self) -> tuple[str, str]:
        """Detect VIS ticket from working directory patterns, environment, and git context."""
        import subprocess
        import os
        import glob

        cwd = os.getcwd()

        vis_match = re.search(r'(vis-\d+)', cwd, re.IGNORECASE)
        if vis_match:
            ticket_number = vis_match.group(1).upper()
            path_parts = cwd.split('/')
            for part in reversed(path_parts):
                if vis_match := re.search(r'(vis-\d+)(?:-(.+))?', part, re.IGNORECASE):
                    description = vis_match.group(2) or ""
                    if description:
                        description = description.replace('-', ' ').title()
                        return f"{ticket_number} {description}", ticket_number
            return ticket_number, ticket_number

        try:
            log_dir = os.path.expanduser("~/Dev/Private/AI/logger")
            if os.path.exists(log_dir):
                vis_files = glob.glob(os.path.join(log_dir, "VIS-*.md"))
                if vis_files:
                    recent_file = max(vis_files, key=os.path.getmtime)
                    filename = os.path.basename(recent_file)
                    vis_match = re.search(r'(VIS-\d+)(?:-(.+?))?(?:-\d{8}-\d{6})?\.md', filename)
                    if vis_match:
                        ticket_number = vis_match.group(1)
                        description = vis_match.group(2) or ""
                        if description:
                            description = description.replace('-', ' ').title()
                            return f"{ticket_number} {description}", ticket_number
                        return ticket_number, ticket_number
        except (OSError, ValueError):
            pass

        for env_var in ['CLAUDE_PROJECT_DIR', 'PWD', 'OLDPWD']:
            path = os.environ.get(env_var, '')
            vis_match = re.search(r'(vis-\d+)', path, re.IGNORECASE)
            if vis_match:
                ticket_number = vis_match.group(1).upper()
                return ticket_number, ticket_number

        try:
            result = subprocess.run(['git', 'branch', '--show-current'],
                                  capture_output=True, text=True, check=True)
            branch_name = result.stdout.strip()

            vis_match = re.search(r'(vis-\d+)(?:-(.+))?', branch_name, re.IGNORECASE)
            if vis_match:
                ticket_number = vis_match.group(1).upper()
                description = vis_match.group(2) or ""
                if description:
                    description = description.replace('-', ' ').title()
                    return f"{ticket_number} {description}", ticket_number
                return ticket_number, ticket_number
        except (subprocess.CalledProcessError, FileNotFoundError):
            pass

        return "", ""

    def create_work_log(self, ticket_text: str, agent_name: str, agent_output: str, auto_timestamp: bool = True, append: bool = False) -> tuple[bool, str]:
        """Create work log file with agent output."""
        try:
            from datetime import datetime
            from pathlib import Path
            import glob

            base_filename = self.convert_to_dash_format(ticket_text)

            log_dir = Path.home() / "Dev" / "Private" / "AI" / "logger"
            log_dir.mkdir(parents=True, exist_ok=True)

            timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            iso_timestamp = datetime.now().isoformat()

            if append:
                pattern = str(log_dir / f"{base_filename}*.md")
                existing_files = glob.glob(pattern)

                if existing_files:
                    existing_files.sort(key=lambda x: Path(x).stat().st_mtime, reverse=True)
                    file_path = Path(existing_files[0])

                    agent_section = f"""

---

## {agent_name.title()} Agent Update

**Date:** {timestamp}
**Timestamp:** {iso_timestamp}

{agent_output}
"""

                    with open(file_path, 'a', encoding='utf-8') as f:
                        f.write(agent_section)

                    return True, f"{file_path} (appended)"
                else:
                    file_path = log_dir / f"{base_filename}.md"

                    content = f"""

**Agent:** {agent_name}
**Date:** {timestamp}
**Timestamp:** {iso_timestamp}
**Task:** {ticket_text}

---

{agent_output}
"""

                    with open(file_path, 'w', encoding='utf-8') as f:
                        f.write(content)

                    return True, str(file_path)

            if not append:
                if auto_timestamp:
                    timestamp_suffix = datetime.now().strftime("%Y%m%d-%H%M%S")
                    filename = f"{base_filename}-{timestamp_suffix}.md"
                else:
                    filename = f"{base_filename}.md"

                file_path = log_dir / filename

                content = f"""

**Agent:** {agent_name}
**Date:** {timestamp}
**Timestamp:** {iso_timestamp}
**Task:** {ticket_text}

---

{agent_output}
"""

                with open(file_path, 'w', encoding='utf-8') as f:
                    f.write(content)

                return True, str(file_path)

        except Exception as e:
            return False, str(e)
