#!/usr/bin/env python3
"""
Universal VIS Work Log Creator
Automatically captures complete agent output for VIS tickets and saves as work logs.
"""

import os
import re
import sys
import subprocess
from datetime import datetime
from pathlib import Path

def extract_vis_info(message):
    """Extract VIS ticket number and description from user message."""
    if not message:
        return None, None
    
    # Look for VIS-<number> pattern
    vis_pattern = r'VIS-(\d+)(?:\s+(.+?))?(?:\s|$)'
    match = re.search(vis_pattern, message, re.IGNORECASE)
    
    if not match:
        return None, None
    
    ticket_number = match.group(1)
    description = match.group(2) if match.group(2) else ""
    
    # Clean up the description - take everything after VIS-<number>
    # Look for the full context after VIS-<number>
    full_pattern = rf'VIS-{ticket_number}\s*(.*?)(?:\s*(?:use|run|execute)\s+|$)'
    full_match = re.search(full_pattern, message, re.IGNORECASE | re.DOTALL)
    
    if full_match and full_match.group(1).strip():
        description = full_match.group(1).strip()
    
    return ticket_number, description

def get_filename_from_helper(ticket_number, description):
    """Generate filename using helper dash command."""
    if not ticket_number:
        return None
    
    # Construct full ticket string
    if description:
        full_ticket = f"VIS-{ticket_number} {description}"
    else:
        full_ticket = f"VIS-{ticket_number}"
    
    try:
        # Call helper dash command
        result = subprocess.run(
            ['helper', 'dash', full_ticket],
            capture_output=True,
            text=True,
            check=True
        )
        return result.stdout.strip() + '.md'
    except subprocess.CalledProcessError as e:
        print(f"Error calling helper command: {e}", file=sys.stderr)
        # Fallback to simple filename
        clean_desc = re.sub(r'[^\w\s-]', '', description) if description else ''
        clean_desc = re.sub(r'\s+', '-', clean_desc.strip().lower())
        if clean_desc:
            return f"VIS-{ticket_number}-{clean_desc}.md"
        else:
            return f"VIS-{ticket_number}.md"

def create_work_log(filename, agent_name, user_message, agent_output):
    """Create the work log file with complete agent output."""
    
    # Ensure log directory exists
    log_dir = Path.home() / "Dev" / "Private" / "AI" / "logger"
    log_dir.mkdir(parents=True, exist_ok=True)
    
    file_path = log_dir / filename
    
    # Extract task description from user message
    task_desc = "Agent task execution"
    if user_message:
        # Try to extract what the user asked the agent to do
        task_patterns = [
            r'use\s+[\w-]+\s+to\s+(.+)',
            r'run\s+[\w-]+\s+(?:to\s+)?(.+)',
            r'execute\s+[\w-]+\s+(?:to\s+)?(.+)',
            r'[\w-]+\s+(.+)'
        ]
        
        for pattern in task_patterns:
            match = re.search(pattern, user_message, re.IGNORECASE)
            if match:
                task_desc = match.group(1).strip()
                break
    
    # Create markdown content
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    
    content = f"""# {filename.replace('.md', '').replace('-', ' ').title()}

**Agent:** {agent_name}  
**Date:** {timestamp}  
**Task:** {task_desc}  

---

{agent_output}
"""
    
    # Write to file
    try:
        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(content)
        print(f"✅ Work log created: {file_path}")
        return True
    except Exception as e:
        print(f"❌ Error creating work log: {e}", file=sys.stderr)
        return False

def main():
    """Main function to process the hook."""
    
    # Get environment variables from Claude Code hook - try multiple possible names
    user_message = (
        os.environ.get('CLAUDE_USER_MESSAGE', '') or
        os.environ.get('CLAUDE_USER_INPUT', '') or
        os.environ.get('CLAUDE_PROMPT', '') or
        os.environ.get('CLAUDE_REQUEST', '') or
        os.environ.get('USER_MESSAGE', '') or
        os.environ.get('USER_INPUT', '')
    )
    
    agent_output = (
        os.environ.get('CLAUDE_AGENT_OUTPUT', '') or
        os.environ.get('CLAUDE_OUTPUT', '') or
        os.environ.get('CLAUDE_RESPONSE', '') or
        os.environ.get('AGENT_OUTPUT', '') or
        os.environ.get('AGENT_RESPONSE', '') or
        os.environ.get('OUTPUT', '')
    )
    
    agent_name = (
        os.environ.get('CLAUDE_AGENT_NAME', '') or
        os.environ.get('CLAUDE_SUBAGENT', '') or
        os.environ.get('AGENT_NAME', '') or
        os.environ.get('SUBAGENT_NAME', '') or
        'unknown-agent'
    )
    
    # Debug output (can be removed later)
    print(f"🔍 Hook triggered for agent: {agent_name}")
    print(f"🔍 User message: {user_message[:100]}..." if len(user_message) > 100 else f"🔍 User message: {user_message}")
    
    # Extract VIS ticket information
    ticket_number, description = extract_vis_info(user_message)
    
    if not ticket_number:
        print("ℹ️  No VIS ticket found in message, skipping work log creation")
        return 0
    
    print(f"🎫 Found VIS-{ticket_number}: {description}")
    
    # Generate filename
    filename = get_filename_from_helper(ticket_number, description)
    if not filename:
        print("❌ Could not generate filename", file=sys.stderr)
        return 1
    
    print(f"📁 Generated filename: {filename}")
    
    # Create work log
    if create_work_log(filename, agent_name, user_message, agent_output):
        print(f"📝 Work log successfully created for VIS-{ticket_number}")
        return 0
    else:
        return 1

if __name__ == "__main__":
    sys.exit(main())