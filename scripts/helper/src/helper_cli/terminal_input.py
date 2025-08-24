#!/usr/bin/env python3
"""Terminal input injection methods for different shells and terminals."""

import sys
import os
import subprocess
import shutil
from typing import Optional


def inject_to_terminal(command: str) -> bool:
    """
    Try various methods to inject command into terminal input buffer.
    Returns True if successful, False otherwise.
    """
    
    # Method 1: Use tmux if we're in a tmux session
    if os.environ.get('TMUX'):
        try:
            subprocess.run(['tmux', 'send-keys', '-t', '.', command], 
                          check=True, capture_output=True)
            return True
        except (subprocess.CalledProcessError, FileNotFoundError):
            pass
    
    # Method 2: Use screen if we're in a screen session
    if os.environ.get('STY'):
        try:
            subprocess.run(['screen', '-X', 'stuff', command], 
                          check=True, capture_output=True)
            return True
        except (subprocess.CalledProcessError, FileNotFoundError):
            pass
    
    # Method 3: Use xdotool for X11 systems
    if shutil.which('xdotool') and os.environ.get('DISPLAY'):
        try:
            # Get the current window ID
            result = subprocess.run(['xdotool', 'getwindowfocus'], 
                                  capture_output=True, text=True, check=True)
            window_id = result.stdout.strip()
            
            # Type to that window
            subprocess.run(['xdotool', 'type', '--window', window_id, command], 
                          check=True, capture_output=True)
            return True
        except (subprocess.CalledProcessError, FileNotFoundError):
            pass
    
    # Method 4: Use ydotool for Wayland (requires ydotoold daemon)
    if shutil.which('ydotool'):
        try:
            subprocess.run(['ydotool', 'type', command], 
                          check=True, capture_output=True)
            return True
        except (subprocess.CalledProcessError, FileNotFoundError):
            pass
    
    # Method 5: Use AppleScript on macOS
    if sys.platform == 'darwin' and shutil.which('osascript'):
        try:
            # Escape special characters for AppleScript
            escaped_command = command.replace('"', '\\"').replace('\\', '\\\\')
            script = f'''
            tell application "System Events"
                keystroke "{escaped_command}"
            end tell
            '''
            subprocess.run(['osascript', '-e', script], 
                          check=True, capture_output=True)
            return True
        except (subprocess.CalledProcessError, FileNotFoundError):
            pass
    
    # Method 6: Use TIOCSTI ioctl (requires specific permissions)
    if sys.platform != 'win32':
        try:
            import fcntl
            import termios
            
            for char in command:
                fcntl.ioctl(sys.stdin, termios.TIOCSTI, char.encode())
            return True
        except (ImportError, OSError, PermissionError):
            pass
    
    return False


def print_ready_to_paste(command: str):
    """
    Print command in a way that's ready to paste/execute.
    This is the fallback when injection isn't possible.
    """
    # Just print the command without newline
    # This makes it appear as if typed, user just presses Enter
    print(command, end='', flush=True)


def inject_or_print(command: str) -> str:
    """
    Try to inject command to terminal, fallback to printing.
    Returns a status message.
    """
    if inject_to_terminal(command):
        return "✓ Command typed to terminal (press Enter to execute)"
    else:
        print_ready_to_paste(command)
        return ""  # No message needed, command is visible