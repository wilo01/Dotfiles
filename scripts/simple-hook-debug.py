#!/usr/bin/env python3
"""
Simple debug hook to see what environment variables are available.
"""

import os
import sys
from datetime import datetime

def main():
    """Debug hook environment variables."""
    
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    
    debug_file = "/tmp/hook-debug.log"
    
    with open(debug_file, "a") as f:
        f.write(f"\n=== Simple Hook Debug - {timestamp} ===\n")
        f.write(f"Script called: {sys.argv[0]}\n")
        f.write(f"Arguments: {sys.argv[1:]}\n")
        f.write(f"Working directory: {os.getcwd()}\n")
        
        # List all environment variables
        f.write("\nAll environment variables:\n")
        for key in sorted(os.environ.keys()):
            value = os.environ[key]
            # Truncate long values
            if len(value) > 200:
                value = value[:200] + "..."
            f.write(f"  {key}={value}\n")
        
        f.write(f"=== End Hook Debug ===\n\n")
    
    print(f"Hook debug info written to {debug_file}")
    return 0

if __name__ == "__main__":
    sys.exit(main())