#!/usr/bin/env python3
"""Test script for the refactored string formatting functions."""

import sys
sys.path.insert(0, '/home/dariuszw/.Dotfiles/scripts/helper/src')

from helper_cli.core import HelperCore

def test_formatting():
    core = HelperCore()
    
    # Test cases
    test_inputs = [
        "VIS-1234 Fix Dashboard Issues",
        "SUITE-5678: Update UI Components",
        "bookr feature/new-api",
        "[Special] Characters! Test@#$%",
        "Multiple   Spaces   Test",
        "vis-999 lowercase vis test"
    ]
    
    print("Testing refactored string formatting functions:\n")
    print("=" * 60)
    
    for test_input in test_inputs:
        print(f"\nInput: '{test_input}'")
        print("-" * 40)
        
        # Test JIRA branch name
        jira_result = core.generate_jira_branch_name(test_input)
        print(f"JIRA Branch: {jira_result}")
        
        # Test stash command
        stash_result = core.generate_stash_command(test_input)
        print(f"Stash Command: {stash_result}")
        
        # Test dash format
        dash_result = core.convert_to_dash_format(test_input)
        print(f"Dash Format: {dash_result}")
    
    print("\n" + "=" * 60)
    print("All tests completed successfully!")

if __name__ == "__main__":
    test_formatting()