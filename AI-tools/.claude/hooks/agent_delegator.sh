#!/bin/bash

# Agent Delegator Script - Monitors comms.md changes and delegates AI agents
# This script automatically triggers appropriate agents based on NEXT items in execution chain

COMMS_FILE="/home/dariuszw/Dev/branch-opener/branches/tds-suite/comms.md"
LOCK_FILE="/tmp/agent_delegator.lock"
LOG_FILE="/home/dariuszw/.claude/hooks/agent_delegator.log"

# Function to log messages
log_message() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >> "$LOG_FILE"
}

# Function to check if another instance is running
check_lock() {
    if [ -f "$LOCK_FILE" ]; then
        local pid=$(cat "$LOCK_FILE")
        if ps -p "$pid" > /dev/null 2>&1; then
            log_message "Another instance is already running (PID: $pid)"
            exit 0
        else
            log_message "Removing stale lock file"
            rm -f "$LOCK_FILE"
        fi
    fi
}

# Function to create lock file
create_lock() {
    echo $$ > "$LOCK_FILE"
}

# Function to remove lock file
remove_lock() {
    rm -f "$LOCK_FILE"
}

# Function to get next agent from comms.md
get_next_agent() {
    if [ ! -f "$COMMS_FILE" ]; then
        log_message "ERROR: comms.md not found at $COMMS_FILE"
        return 1
    fi
    
    # Extract the last NEXT item from execution chain
    local next_item=$(grep -E "^\- \[NEXT\]" "$COMMS_FILE" | tail -1 | sed 's/^- \[NEXT\] //')
    
    if [ -z "$next_item" ]; then
        # If no NEXT items, look for timestamp entries that need continuation
        local last_entry=$(grep -E "^\- \[[0-9]{2}:[0-9]{2}:[0-9]{2}\]" "$COMMS_FILE" | tail -1)
        if [[ "$last_entry" =~ NEXT:\ (.+)$ ]]; then
            next_item="${BASH_REMATCH[1]}"
        fi
    fi
    
    echo "$next_item"
}

# Function to determine which agent should handle the task
determine_agent() {
    local task="$1"
    
    if [[ "$task" =~ Agent\ 1|ExtJS|frontend|UI ]]; then
        echo "1"
    elif [[ "$task" =~ Agent\ 2|Oracle|database|\.apex|backend ]]; then
        echo "2"
    elif [[ "$task" =~ Agent\ 3|test|review|integration ]]; then
        echo "3"
    else
        # Default to Agent 1 if unclear
        echo "1"
    fi
}

# Function to check if agent is busy
is_agent_busy() {
    local agent_num="$1"
    local status_line=$(grep -E "^- Agent $agent_num" "$COMMS_FILE" | grep -E "\[Status\]")
    
    if [[ "$status_line" =~ Working|In\ Progress|Busy ]]; then
        return 0  # Agent is busy
    else
        return 1  # Agent is available
    fi
}

# Function to trigger agent
trigger_agent() {
    local agent_num="$1"
    local task="$2"
    
    log_message "Triggering Agent $agent_num for task: $task"
    
    # Create a trigger file that the agent monitoring system can pick up
    local trigger_file="/tmp/agent_${agent_num}_trigger"
    echo "$task" > "$trigger_file"
    
    # Log the delegation
    local timestamp=$(date '+%H:%M:%S')
    local delegation_entry="- [$timestamp] DELEGATED to Agent $agent_num: $task"
    echo "$delegation_entry" >> "$COMMS_FILE"
    
    log_message "Agent $agent_num triggered successfully"
}

# Function to check for dependencies
check_dependencies() {
    local task="$1"
    
    # Check if task mentions waiting for another agent
    if [[ "$task" =~ Agent\ ([0-9])\ needs\ to ]]; then
        local required_agent="${BASH_REMATCH[1]}"
        if is_agent_busy "$required_agent"; then
            log_message "Task depends on Agent $required_agent who is still busy"
            return 1  # Dependencies not met
        fi
    fi
    
    return 0  # Dependencies met
}

# Function to check if goal is completed
is_goal_completed() {
    # Check if all TODO items are completed or if there's a completion marker
    local incomplete_todos=$(grep -E "^\- \[NEXT\]|EXECUTING:" "$COMMS_FILE" | wc -l)
    
    if [ "$incomplete_todos" -eq 0 ]; then
        return 0  # Goal completed
    else
        return 1  # Goal not completed
    fi
}

# Main execution
main() {
    log_message "Agent delegator started"
    
    check_lock
    create_lock
    
    # Trap to ensure cleanup on exit
    trap remove_lock EXIT
    
    # Check if goal is already completed
    if is_goal_completed; then
        log_message "Goal appears to be completed, no delegation needed"
        exit 0
    fi
    
    # Get next task
    local next_task=$(get_next_agent)
    
    if [ -z "$next_task" ]; then
        log_message "No next task found in comms.md"
        exit 0
    fi
    
    log_message "Next task found: $next_task"
    
    # Determine which agent should handle it
    local agent_num=$(determine_agent "$next_task")
    log_message "Task assigned to Agent $agent_num"
    
    # Check if agent is available
    if is_agent_busy "$agent_num"; then
        log_message "Agent $agent_num is busy, will retry later"
        exit 0
    fi
    
    # Check dependencies
    if ! check_dependencies "$next_task"; then
        log_message "Dependencies not met, will retry later"
        exit 0
    fi
    
    # Trigger the agent
    trigger_agent "$agent_num" "$next_task"
    
    log_message "Agent delegator completed successfully"
}

# Run main function
main "$@"