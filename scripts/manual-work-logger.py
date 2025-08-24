#!/usr/bin/env python3
"""
Manual work logger - creates work logs when hooks don't work.
Usage: manual-work-logger.py "VIS-<number> description" "agent-name"
"""

import sys
import os
import re
import subprocess
from datetime import datetime
from pathlib import Path

def get_filename_from_helper(ticket_text):
    """Generate filename using helper dash command."""
    try:
        result = subprocess.run(
            ['helper', 'dash', ticket_text],
            capture_output=True,
            text=True,
            check=True
        )
        return result.stdout.strip() + '.md'
    except subprocess.CalledProcessError as e:
        print(f"Error calling helper command: {e}", file=sys.stderr)
        return None

def create_work_log(filename, agent_name, ticket_text, agent_output):
    """Create the work log file."""
    
    # Ensure log directory exists
    log_dir = Path.home() / "Dev" / "Private" / "AI" / "logger"
    log_dir.mkdir(parents=True, exist_ok=True)
    
    file_path = log_dir / filename
    
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    
    content = f"""# {filename.replace('.md', '').replace('-', ' ').title()}

**Agent:** {agent_name}  
**Date:** {timestamp}  
**Task:** {ticket_text}  

---

{agent_output}
"""
    
    try:
        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(content)
        print(f"✅ Work log created: {file_path}")
        return True
    except Exception as e:
        print(f"❌ Error creating work log: {e}", file=sys.stderr)
        return False

def main():
    """Main function."""
    
    if len(sys.argv) < 3:
        print("Usage: manual-work-logger.py \"VIS-<number> description\" \"agent-name\" [\"agent-output\"]")
        print("Example: manual-work-logger.py \"VIS-5187 SWIFT uidtoken validation\" \"manual-tester\"")
        return 1
    
    ticket_text = sys.argv[1]
    agent_name = sys.argv[2]
    agent_output = sys.argv[3] if len(sys.argv) > 3 else "Agent output not provided - please copy and paste the agent's response here."
    
    # Generate filename
    filename = get_filename_from_helper(ticket_text)
    if not filename:
        print("❌ Could not generate filename", file=sys.stderr)
        return 1
    
    print(f"📁 Generated filename: {filename}")
    
    # Create work log
    if create_work_log(filename, agent_name, ticket_text, agent_output):
        print(f"📝 Work log successfully created")
        return 0
    else:
        return 1

if __name__ == "__main__":
    sys.exit(main())