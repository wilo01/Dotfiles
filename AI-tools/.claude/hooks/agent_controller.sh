#!/bin/bash

# Agent Controller Script - Manages the complete agent workflow system
# This script provides commands to start, stop, and manage the agent delegation system

HOOKS_DIR="/home/dariuszw/.claude/hooks"
COMMS_FILE="/home/dariuszw/Dev/branch-opener/branches/tds-suite/comms.md"
LOG_FILE="/home/dariuszw/.claude/hooks/agent_controller.log"
STOP_FILE="/tmp/workflow_stop"

# Function to log messages
log_message() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >> "$LOG_FILE"
    echo "$1"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 {start|stop|status|restart|reset|logs}"
    echo ""
    echo "Commands:"
    echo "  start   - Start the agent workflow system"
    echo "  stop    - Stop the agent workflow system"
    echo "  status  - Show current system status"
    echo "  restart - Restart the agent workflow system"
    echo "  reset   - Reset all agent states and clear logs"
    echo "  logs    - Show recent log entries"
    echo ""
    echo "The system will automatically delegate AI agents based on changes to comms.md"
    echo "and manage dependencies between agents until the goal is completed."
}

# Function to start the workflow system
start_workflow() {
    log_message "Starting agent workflow system..."
    
    # Remove stop file if it exists
    rm -f "$STOP_FILE"
    
    # Start the workflow monitor in background
    nohup bash "$HOOKS_DIR/workflow_monitor.sh" start > /dev/null 2>&1 &
    
    # Give it a moment to start
    sleep 2
    
    # Check if it's running
    if [ -f "/tmp/workflow_monitor.lock" ]; then
        log_message "Agent workflow system started successfully"
        return 0
    else
        log_message "ERROR: Failed to start agent workflow system"
        return 1
    fi
}

# Function to stop the workflow system
stop_workflow() {
    log_message "Stopping agent workflow system..."
    
    # Create stop file
    touch "$STOP_FILE"
    
    # Kill any running monitors
    pkill -f "workflow_monitor.sh" 2>/dev/null
    
    # Remove lock files
    rm -f "/tmp/workflow_monitor.lock" "/tmp/agent_delegator.lock"
    
    # Remove trigger files
    rm -f "/tmp/agent_"*"_trigger"
    
    log_message "Agent workflow system stopped"
}

# Function to check system status
check_status() {
    echo "Agent Workflow System Status:"
    echo "============================="
    
    # Check monitor status
    if [ -f "/tmp/workflow_monitor.lock" ]; then
        local pid=$(cat "/tmp/workflow_monitor.lock")
        if ps -p "$pid" > /dev/null 2>&1; then
            echo "✓ Workflow monitor: Running (PID: $pid)"
        else
            echo "✗ Workflow monitor: Lock file exists but process not running"
        fi
    else
        echo "✗ Workflow monitor: Not running"
    fi
    
    # Check for stop file
    if [ -f "$STOP_FILE" ]; then
        echo "⚠ Stop file exists - system will halt"
    else
        echo "✓ Stop file: Not present"
    fi
    
    # Check comms.md
    if [ -f "$COMMS_FILE" ]; then
        echo "✓ Communications file: Present"
        local next_items=$(grep -E "^\- \[NEXT\]" "$COMMS_FILE" | wc -l)
        echo "  - Pending NEXT items: $next_items"
        
        local recent_activity=$(grep -E "^\- \[[0-9]{2}:[0-9]{2}:[0-9]{2}\]" "$COMMS_FILE" | tail -1)
        if [ -n "$recent_activity" ]; then
            echo "  - Last activity: $recent_activity"
        fi
    else
        echo "✗ Communications file: Not found"
    fi
    
    # Check for agent triggers
    local active_triggers=0
    for agent_num in 1 2 3; do
        if [ -f "/tmp/agent_${agent_num}_trigger" ]; then
            active_triggers=$((active_triggers + 1))
        fi
    done
    echo "  - Active agent triggers: $active_triggers"
    
    # Check recent logs
    echo ""
    echo "Recent Log Entries:"
    echo "==================="
    if [ -f "$LOG_FILE" ]; then
        tail -5 "$LOG_FILE"
    else
        echo "No log file found"
    fi
}

# Function to restart the system
restart_workflow() {
    log_message "Restarting agent workflow system..."
    stop_workflow
    sleep 2
    start_workflow
}

# Function to reset the system
reset_workflow() {
    log_message "Resetting agent workflow system..."
    
    # Stop everything
    stop_workflow
    
    # Clear all logs
    rm -f "$LOG_FILE" "/home/dariuszw/.claude/hooks/agent_delegator.log" "/home/dariuszw/.claude/hooks/workflow_monitor.log"
    
    # Clear all lock and trigger files
    rm -f "/tmp/workflow_monitor.lock" "/tmp/agent_delegator.lock" "/tmp/agent_"*"_trigger"
    
    # Reset agent status in comms.md
    if [ -f "$COMMS_FILE" ]; then
        # This would ideally reset the "Current Status" section
        log_message "Comms file preserved - manual reset may be needed"
    fi
    
    log_message "Agent workflow system reset complete"
}

# Function to show logs
show_logs() {
    echo "Agent Controller Logs:"
    echo "====================="
    if [ -f "$LOG_FILE" ]; then
        tail -20 "$LOG_FILE"
    else
        echo "No controller log file found"
    fi
    
    echo ""
    echo "Agent Delegator Logs:"
    echo "===================="
    if [ -f "/home/dariuszw/.claude/hooks/agent_delegator.log" ]; then
        tail -20 "/home/dariuszw/.claude/hooks/agent_delegator.log"
    else
        echo "No delegator log file found"
    fi
    
    echo ""
    echo "Workflow Monitor Logs:"
    echo "====================="
    if [ -f "/home/dariuszw/.claude/hooks/workflow_monitor.log" ]; then
        tail -20 "/home/dariuszw/.claude/hooks/workflow_monitor.log"
    else
        echo "No monitor log file found"
    fi
}

# Main script logic
case "${1:-}" in
    "start")
        start_workflow
        ;;
    "stop")
        stop_workflow
        ;;
    "status")
        check_status
        ;;
    "restart")
        restart_workflow
        ;;
    "reset")
        reset_workflow
        ;;
    "logs")
        show_logs
        ;;
    *)
        show_usage
        exit 1
        ;;
esac