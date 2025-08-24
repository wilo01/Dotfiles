#!/usr/bin/env python3
"""
Debug script to see what environment variables are available in hooks.
"""

import os
import sys
from datetime import datetime

def main():
    """Debug hook environment variables."""
    
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    
    # Write debug info to a file
    debug_file = os.path.expanduser("~/hook-debug.log")
    
    with open(debug_file, "a") as f:
        f.write(f"\n=== Hook Debug - {timestamp} ===\n")
        
        # List all environment variables that might be related to Claude
        claude_vars = {}
        for key, value in os.environ.items():
            if 'CLAUDE' in key.upper() or 'AGENT' in key.upper() or 'HOOK' in key.upper():
                claude_vars[key] = value
        
        if claude_vars:
            f.write("Claude-related environment variables:\n")
            for key, value in claude_vars.items():
                f.write(f"  {key}={value[:200]}{'...' if len(value) > 200 else ''}\n")
        else:
            f.write("No Claude-related environment variables found.\n")
        
        f.write("\nAll environment variables:\n")
        for key, value in os.environ.items():
            f.write(f"  {key}={value[:100]}{'...' if len(value) > 100 else ''}\n")
        
        f.write(f"=== End Hook Debug ===\n\n")
    
    print(f"Hook debug info written to {debug_file}")
    return 0

if __name__ == "__main__":
    sys.exit(main())