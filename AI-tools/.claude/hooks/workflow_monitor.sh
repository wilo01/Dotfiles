#!/bin/bash

# Workflow Monitor Script - Continuously monitors for agent work and maintains workflow
# This script runs in the background and keeps the agent workflow active

COMMS_FILE="/home/dariuszw/Dev/branch-opener/branches/tds-suite/comms.md"
MONITOR_LOCK="/tmp/workflow_monitor.lock"
LOG_FILE="/home/dariuszw/.claude/hooks/workflow_monitor.log"
STOP_FILE="/tmp/workflow_stop"

# Function to log messages
log_message() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >> "$LOG_FILE"
}

# Function to check if monitoring should stop
should_stop() {
    if [ -f "$STOP_FILE" ]; then
        return 0  # Should stop
    fi
    return 1  # Continue monitoring
}

# Function to create stop file (for external stopping)
create_stop_file() {
    touch "$STOP_FILE"
    log_message "Stop file created - monitoring will halt"
}

# Function to remove stop file
remove_stop_file() {
    rm -f "$STOP_FILE"
}

# Function to check if there's pending work
has_pending_work() {
    if [ ! -f "$COMMS_FILE" ]; then
        return 1  # No pending work if file doesn't exist
    fi
    
    # Check for NEXT items or recent activity
    local next_items=$(grep -E "^\- \[NEXT\]" "$COMMS_FILE" | wc -l)
    local recent_activity=$(grep -E "^\- \[[0-9]{2}:[0-9]{2}:[0-9]{2}\]" "$COMMS_FILE" | tail -5)
    
    if [ "$next_items" -gt 0 ]; then
        return 0  # Has pending work
    fi
    
    # Check if recent activity indicates ongoing work
    if echo "$recent_activity" | grep -q "NEXT:"; then
        return 0  # Has pending work
    fi
    
    return 1  # No pending work
}

# Function to check for agent triggers
check_agent_triggers() {
    for agent_num in 1 2 3; do
        local trigger_file="/tmp/agent_${agent_num}_trigger"
        if [ -f "$trigger_file" ]; then
            local task=$(cat "$trigger_file")
            log_message "Agent $agent_num trigger found: $task"
            
            # Simulate agent work (in real implementation, this would launch Claude with specific context)
            simulate_agent_work "$agent_num" "$task"
            
            # Remove trigger file
            rm -f "$trigger_file"
        fi
    done
}

# Function to simulate agent work (placeholder for actual Claude invocation)
simulate_agent_work() {
    local agent_num="$1"
    local task="$2"
    
    log_message "Simulating Agent $agent_num work for: $task"
    
    # In real implementation, this would:
    # 1. Launch Claude with agent-specific context
    # 2. Pass the task as instruction
    # 3. Monitor completion
    # 4. Update comms.md with results
    
    # For now, just log the activity
    local timestamp=$(date '+%H:%M:%S')
    local work_entry="- [$timestamp] Agent $agent_num: EXECUTING: $task | NEXT: [Agent work simulation]"
    echo "$work_entry" >> "$COMMS_FILE"
    
    log_message "Agent $agent_num work simulated"
}

# Function to check for inactive agents that should be prompted
check_inactive_agents() {
    local current_time=$(date +%s)
    
    # Check for agents that haven't been active recently
    for agent_num in 1 2 3; do
        local last_activity=$(grep -E "^\- \[[0-9]{2}:[0-9]{2}:[0-9]{2}\] Agent $agent_num" "$COMMS_FILE" | tail -1)
        
        if [ -n "$last_activity" ]; then
            # Extract timestamp and check if it's older than 5 minutes
            local timestamp=$(echo "$last_activity" | grep -o '\[[0-9]{2}:[0-9]{2}:[0-9]{2}\]' | tr -d '[]')
            local activity_time=$(date -d "$timestamp" +%s 2>/dev/null)
            
            if [ $? -eq 0 ] && [ $((current_time - activity_time)) -gt 300 ]; then
                log_message "Agent $agent_num appears inactive for >5 minutes, may need prompting"
            fi
        fi
    done
}

# Function to update agent status monitoring
update_agent_status() {
    # Update current status section in comms.md
    local temp_file="/tmp/comms_status_update"
    
    # This would update the "Current Status" section with real agent states
    # For now, just log that we're monitoring
    log_message "Monitoring agent status..."
}

# Main monitoring loop
main_monitor() {
    log_message "Workflow monitor started"
    
    # Check if another monitor is already running
    if [ -f "$MONITOR_LOCK" ]; then
        local pid=$(cat "$MONITOR_LOCK")
        if ps -p "$pid" > /dev/null 2>&1; then
            log_message "Another monitor is already running (PID: $pid)"
            exit 0
        else
            log_message "Removing stale monitor lock"
            rm -f "$MONITOR_LOCK"
        fi
    fi
    
    # Create lock file
    echo $$ > "$MONITOR_LOCK"
    
    # Cleanup on exit
    trap 'rm -f "$MONITOR_LOCK"; log_message "Workflow monitor stopped"' EXIT
    
    # Main monitoring loop
    while true; do
        # Check if we should stop
        if should_stop; then
            log_message "Stop condition met, exiting monitor"
            break
        fi
        
        # Check for pending work
        if ! has_pending_work; then
            log_message "No pending work detected, continuing to monitor"
            sleep 10
            continue
        fi
        
        # Check for agent triggers
        check_agent_triggers
        
        # Check for inactive agents
        check_inactive_agents
        
        # Update agent status
        update_agent_status
        
        # Sleep before next check
        sleep 5
    done
    
    log_message "Workflow monitor exited"
}

# Handle command line arguments
case "${1:-}" in
    "start")
        main_monitor
        ;;
    "stop")
        create_stop_file
        ;;
    "status")
        if [ -f "$MONITOR_LOCK" ]; then
            echo "Workflow monitor is running (PID: $(cat $MONITOR_LOCK))"
        else
            echo "Workflow monitor is not running"
        fi
        ;;
    *)
        # Default behavior - start monitoring
        main_monitor
        ;;
esac