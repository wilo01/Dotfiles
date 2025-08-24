#!/usr/bin/env python3
"""Test different methods for typing commands to terminal."""

import sys
import os
import subprocess

def test_methods():
    command = "git checkout -b test-branch"
    
    print("Testing different methods to 'type' command to terminal:")
    print("=" * 60)
    
    # Method 1: Simple print without newline
    print("\nMethod 1 - Print without newline:")
    print(command, end='', flush=True)
    input("\n(Press Enter to continue)")
    
    # Method 2: Using echo with xdotool (if available)
    print("\nMethod 2 - Using xdotool (if available):")
    try:
        # Check if xdotool is available
        subprocess.run(['which', 'xdotool'], check=True, capture_output=True)
        print("xdotool is available - would type: " + command)
        # subprocess.run(['xdotool', 'type', command])
    except:
        print("xdotool not available")
    
    # Method 3: Using shell aliases or functions
    print("\nMethod 3 - Shell function approach:")
    print(f"You could add this to your shell config:")
    print(f"  alias helper-exec='eval $(helper jira \"$1\" --print-only)'")
    
    # Method 4: Using TIOCSTI ioctl (requires root on modern systems)
    print("\nMethod 4 - TIOCSTI (usually requires root):")
    print("This method injects keystrokes but needs special permissions")
    
    # Method 5: Output for eval
    print("\nMethod 5 - Output for eval:")
    print(f"Run: eval $(echo '{command}')")
    print(f"Or: {command}")

if __name__ == "__main__":
    test_methods()