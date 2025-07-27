# AI Agent Auto-Delegation System

This system automatically delegates AI agents based on changes to the `comms.md` file, managing dependencies and workflow until the goal is completed.

## System Components

### 1. Hooks Configuration (`settings.local.json`)
- **PostToolUse hooks**: Monitor Edit, Write, and MultiEdit operations on comms.md
- **Stop hooks**: Trigger workflow monitoring when Claude finishes responding
- All hooks run in background to avoid blocking Claude

### 2. Agent Delegator (`agent_delegator.sh`)
- Triggered whenever comms.md is modified
- Parses `[NEXT]` items from execution chain
- Determines appropriate agent based on task content
- Checks agent availability and dependencies
- Creates trigger files for agent activation

### 3. Workflow Monitor (`workflow_monitor.sh`)
- Continuously monitors for pending work
- Checks for agent trigger files
- Simulates agent work (placeholder for actual Claude invocation)
- Monitors agent activity and detects stale processes
- Maintains workflow until completion or stop signal

### 4. Agent Controller (`agent_controller.sh`)
- Management interface for the entire system
- Commands: start, stop, status, restart, reset, logs
- Provides system status and monitoring capabilities

## Usage

### Starting the System
```bash
bash /home/dariuszw/.claude/hooks/agent_controller.sh start
```

### Stopping the System
```bash
bash /home/dariuszw/.claude/hooks/agent_controller.sh stop
```

### Checking Status
```bash
bash /home/dariuszw/.claude/hooks/agent_controller.sh status
```

### Viewing Logs
```bash
bash /home/dariuszw/.claude/hooks/agent_controller.sh logs
```

## How It Works

1. **Trigger**: When comms.md is modified, PostToolUse hooks automatically run `agent_delegator.sh`

2. **Analysis**: The delegator script:
   - Parses the latest `[NEXT]` items from the execution chain
   - Determines which agent (1=ExtJS UI, 2=Oracle DB, 3=Reviewer) should handle the task
   - Checks if the agent is available and dependencies are met

3. **Delegation**: If conditions are met:
   - Creates a trigger file (`/tmp/agent_X_trigger`) with the task
   - Logs the delegation in comms.md
   - Updates agent status

4. **Monitoring**: The workflow monitor:
   - Continuously checks for trigger files
   - Simulates agent work (in production, would launch Claude with specific context)
   - Monitors for completion or stop conditions

5. **Dependencies**: The system handles:
   - Sequential agent execution (1→2→3→1...)
   - Agent availability checking
   - Task dependency validation
   - Automatic retry for blocked tasks

## Agent Roles

- **Agent 1 (ExtJS UI)**: Frontend ExtJS component development
- **Agent 2 (Oracle DB)**: Backend database operations and .apex files
- **Agent 3 (Reviewer)**: Code review, testing, and integration verification

## Configuration

The system uses the following patterns to determine agent assignment:
- **Agent 1**: Tasks containing "ExtJS", "frontend", "UI", "component"
- **Agent 2**: Tasks containing "Oracle", "database", ".apex", "backend"
- **Agent 3**: Tasks containing "test", "review", "integration", "verify"

## Logging

All activities are logged to:
- `/home/dariuszw/.claude/hooks/agent_controller.log`
- `/home/dariuszw/.claude/hooks/agent_delegator.log`
- `/home/dariuszw/.claude/hooks/workflow_monitor.log`

## Stop Conditions

The system will automatically stop when:
- A stop file (`/tmp/workflow_stop`) is created
- No `[NEXT]` items are found in comms.md
- The goal appears to be completed (all TODO items done)

## Dependencies

The system checks for dependencies like:
- "Agent X needs to complete Y before Z can start"
- Agent availability (not busy with other tasks)
- Required files or resources exist

## Production Notes

In a production environment, the `simulate_agent_work` function would be replaced with actual Claude invocations that:
1. Launch Claude with agent-specific context
2. Pass the task as instruction
3. Monitor completion status
4. Update comms.md with results
5. Handle errors and retries

The current implementation provides the framework and monitoring infrastructure needed for full automation.