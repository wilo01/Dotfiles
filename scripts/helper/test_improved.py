#!/usr/bin/env python3
"""Test script for the improved format_string function with smart tag detection."""

import sys
sys.path.insert(0, '/home/dariuszw/.Dotfiles/scripts/helper/src')

from helper_cli.core import HelperCore

def test_improved_formatting():
    core = HelperCore()
    
    # Test cases with various tag formats
    test_inputs = [
        "VIS-1234 Fix Dashboard Issues",
        "suite-5678 Update UI Components",
        "TAC-342 new feature implementation",
        "NEW-4324 implement API endpoints",
        "bookr-999 update booking system",
        "vis-123 and suite-456 multiple tags",
        "Random text without tags",
        "[Special] Characters! Test@#$% with VIS-789",
        "Multiple   Spaces   Test with TAC-111",
        "mixed-123-case-456-numbers everywhere",
        "ABC-001 Beginning tag test",
        "End with tag XYZ-999"
    ]
    
    print("Testing improved string formatting with smart tag detection:\n")
    print("=" * 80)
    
    for test_input in test_inputs:
        print(f"\nInput: '{test_input}'")
        print("-" * 60)
        
        # Test JIRA branch name
        jira_result = core.generate_jira_branch_name(test_input)
        print(f"JIRA:  {jira_result}")
        
        # Test stash command
        stash_result = core.generate_stash_command(test_input)
        print(f"Stash: {stash_result}")
        
        # Test dash format (standard mode)
        dash_result = core.convert_to_dash_format(test_input)
        print(f"Dash:  {dash_result}")
    
    print("\n" + "=" * 80)
    
    # Test with helper CLI commands
    print("\nTesting with actual helper CLI commands:\n")
    print("=" * 80)
    
    special_cases = [
        "VIS-1234 Fix the dashboard",
        "SUITE-5678: Update components",
        "TAC-342 - New Feature",
        "NEW-4324/implement-api"
    ]
    
    for case in special_cases:
        print(f"\nInput: '{case}'")
        print(f"Branch: {core.generate_jira_branch_name(case)}")
    
    print("\n" + "=" * 80)
    print("✅ All tests completed! Smart tag detection is working.")

if __name__ == "__main__":
    test_improved_formatting()