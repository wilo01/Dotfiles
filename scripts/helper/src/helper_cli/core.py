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

            # Copy files from the first source directory
            for file_name in file_names1:
                shutil.copy(os.path.join(source1, file_name), temp_dir)

            # Copy and overwrite files from the second source directory
            for file_name in file_names2:
                shutil.copy(os.path.join(source2, file_name), temp_dir)

        except (OSError, IOError) as e:
            print(f"Error copying files: {e}")

    def generate_jira_branch_name(self, text: str) -> str:
        """Generate JIRA branch name from text."""
        # Clean up the text and create branch name
        branch_name = re.sub(r"[ !+@#$%^&*(),_.'/:;>-]", "-", text.lower().replace('\n', '-'))
        branch_name = branch_name.replace('suite', 'SUITE').replace('vis-', 'VIS-').replace('bookr', 'BOOKR')
        branch_name = re.sub(r'-+', '-', branch_name)
        return f'git checkout -b {branch_name}'

    def generate_stash_command(self, text: str) -> str:
        """Generate git stash command with formatted message."""
        stash_text_raw = f'[stash]{text}'
        stash_text = re.sub(r"[ !+@#$%^&*(),_.'/:;>-]", "-", stash_text_raw.lower().replace('\n', '-'))
        stash_text = stash_text.replace('suite', 'SUITE').replace('[stash]', 'git stash push -u -m ')
        stash_text = re.sub(r'-+', '-', stash_text)
        return stash_text

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
