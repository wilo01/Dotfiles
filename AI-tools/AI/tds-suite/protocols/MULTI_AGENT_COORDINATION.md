# Multi-Agent Coordination Protocol

## Overview

This protocol defines how Claude Code, Gemini, and GitHub Copilot coordinate their work on the TDS Suite project to maximize efficiency while preventing conflicts and ensuring quality.

## Agent Roles & Responsibilities

### Claude Code - Senior Software Developer
- **Primary Focus**: Complex coding tasks, architecture decisions
- **File Ownership**: `source/ui/app/`, `source/server/rt/`
- **Responsibilities**: ExtJS components, APEX development, code reviews
- **Status**: Active implementation and development

### Gemini - Task Manager & Planner
- **Primary Focus**: Task management, planning, and coordination
- **File Ownership**: `test/`, documentation files, task planning
- **Responsibilities**: Task breakdown, complexity analysis, strategic planning
- **Status**: Planning mode only (no code implementation)

### GitHub Copilot - Coordinated Assistant
- **Primary Focus**: Simple tasks, utility functions, styling
- **File Ownership**: Utility files, configuration files, simple implementations
- **Responsibilities**: Supporting tasks, minor fixes, coordination
- **Status**: Coordinated implementation with conflict avoidance

## Task Assignment Protocol

### Task Flow Process

```
1. Requirements/PRD → Gemini (Planning)
2. Gemini → Task Master (Task Creation)
3. Task Master → Agent Assignment
4. Agent → Implementation
5. Agent → Review Status
6. Gemini → Next Task Planning
```

### Task Master Workflow

#### 1. Task Creation (Gemini)
```bash
# Parse requirements into tasks
task-master parse-prd .taskmaster/docs/prd.txt

# Analyze complexity
task-master analyze-complexity --research

# Expand into subtasks
task-master expand --all --research

# Set dependencies and priorities
task-master add-dependency --id=<task-id> --depends-on=<dependency-id>
```

#### 2. Task Assignment (Gemini)
- **Complex Implementation** → Claude Code
- **Planning & Coordination** → Gemini
- **Simple Tasks** → Copilot (when Claude unavailable)

#### 3. Task Execution (Assigned Agent)
```bash
# Start task
task-master set-status --id=<task-id> --status=in-progress

# Update progress
task-master update-subtask --id=<task-id> --prompt="Progress notes..."

# Complete task
task-master set-status --id=<task-id> --status=review
```

### Assignment Guidelines

#### Claude Code Assignments
- **ExtJS Component Development**
- **APEX Resource Template Creation**
- **Complex Database Operations**
- **Architecture Design Implementation**
- **Security Feature Implementation**
- **Performance Optimization**

#### Gemini Assignments
- **Task Planning and Breakdown**
- **Complexity Analysis**
- **Dependency Management**
- **Progress Monitoring**
- **Resource Coordination**
- **Strategic Planning**

#### Copilot Assignments
- **Utility Function Creation**
- **CSS/SCSS Styling Tasks**
- **Configuration File Updates**
- **Simple Bug Fixes**
- **Test Case Implementation**
- **Documentation Updates**

## Coordination Rules

### Task Status Protocol

#### Single Task Rule
- **Only ONE main task** can be "in-progress" at any time
- **Only ONE subtask** can be "in-progress" at any time
- **All other tasks** must be in "pending" status
- **Reopening Tasks**: When a task is reopened, set the main task to 'pending', create new subtasks for the required work, and set all existing subtasks to 'pending'.

#### Status Definitions
- **pending**: Ready to work on
- **in-progress**: Currently being worked on
- **review**: Completed, ready for review
- **done**: Fully completed and tested (set by reviewer only)
- **blocked**: Waiting on external dependencies
- **cancelled**: No longer needed

### Communication Protocol

#### Task Master Communication
- **All task-related communication** goes through Task Master
- **Use detailed progress updates** in task notes
- **Report blockers immediately** with context
- **Provide implementation feedback** for planning improvement

#### Escalation Process
1. **Agent encounters issue** → Update task with details
2. **Set task to blocked** if cannot proceed
3. **Notify coordinating agent** through Task Master
4. **Continue with other available tasks**
5. **Resolution** → Update task and resume work

### File Coordination Rules

#### File Ownership
- **Claude Code**: Primary ownership of core implementation files
- **Gemini**: Documentation and planning files
- **Copilot**: Utility and configuration files

#### Conflict Prevention
1. **Check Task Master** before editing any file
2. **Verify no active work** on target files
3. **Communicate file usage** through task updates
4. **Use version control** effectively

#### File Lock Protocol
```bash
# Before editing, check current assignments
task-master list | grep "in-progress"

# Verify file not in use by other agent
# Check recent task updates for file mentions

# Proceed only if no conflicts detected
```

## Communication Standards

### Progress Reporting

#### Standard Update Format
```bash
task-master update-subtask --id=<task-id> --prompt="
PROGRESS: [Brief description of what was accomplished]
NEXT: [What will be done next]
BLOCKERS: [Any issues or dependencies blocking progress]
NOTES: [Additional context or observations]
"
```

#### Completion Reporting
```bash
task-master update-subtask --id=<task-id> --prompt="
COMPLETED: [Summary of what was implemented]
TESTING: [Testing performed and results]
NOTES: [Implementation notes and decisions made]
HANDOFF: [Any follow-up work needed]
"
```

### Inter-Agent Communication

#### Claude Code ↔ Gemini
- **Implementation feedback** for future planning
- **Complexity estimates** for task breakdown
- **Technical constraints** and dependencies
- **Architecture decisions** and implications

#### Copilot ↔ Claude Code
- **File coordination** and conflict avoidance
- **Implementation pattern** sharing
- **Code review** requests and feedback
- **Task handoffs** for complex work

#### Gemini ↔ All Agents
- **Task assignments** and priorities
- **Progress monitoring** and status updates
- **Resource allocation** and coordination
- **Strategic direction** and planning updates

## Quality Assurance Protocol

### Review Process

#### Code Review Requirements
1. **Self-review** before marking task as "review"
2. **Peer review** by appropriate agent when possible
3. **Testing validation** before completion
4. **Documentation update** as needed

#### Quality Gates
- **Functionality**: Does it work as intended?
- **Integration**: Does it integrate properly with existing code?
- **Security**: Does it follow security best practices?
- **Performance**: Does it meet performance requirements?
- **Maintainability**: Is it readable and maintainable?

### Testing Coordination

#### Testing Responsibilities
- **Unit Tests**: Implementing agent's responsibility
- **Integration Tests**: Claude Code's responsibility
- **E2E Tests**: Coordinated between all agents
- **Manual Testing**: As specified in task requirements

#### Test Execution Protocol
```bash
# Run relevant tests before marking task complete
npm run test:unit
npm run test:integration
npm run devDocker  # Cypress E2E tests
```

## Continuous Improvement

### Retrospective Process

#### Weekly Coordination Review
1. **Review completed tasks** and outcomes
2. **Identify coordination issues** and improvements
3. **Update protocols** based on lessons learned
4. **Plan process optimizations**

#### Metrics Tracking
- **Task completion time** and accuracy
- **Coordination efficiency** and conflict frequency
- **Quality metrics** and rework rates
- **Agent satisfaction** and process feedback

### Process Optimization

#### Protocol Updates
- **Document successful patterns** and practices
- **Refine coordination rules** based on experience
- **Improve communication** standards and formats
- **Optimize task assignment** criteria and processes

#### Tool Improvements
- **Enhance Task Master** usage and integration
- **Develop coordination scripts** and automation
- **Improve monitoring** and reporting capabilities
- **Create workflow templates** for common scenarios

## Emergency Procedures

### Conflict Resolution

#### File Conflict Protocol
1. **Stop all work** on conflicted files immediately
2. **Communicate through Task Master** about the conflict
3. **Coordinate resolution** with affected agents
4. **Use version control** to resolve technical conflicts
5. **Update protocols** to prevent similar conflicts

#### Task Priority Conflicts
1. **Escalate to Gemini** for priority resolution
2. **Document conflicting requirements** clearly
3. **Seek stakeholder input** if needed
4. **Update task priorities** and dependencies
5. **Communicate resolution** to all affected agents

### Agent Unavailability

#### Backup Procedures
- **Document handoff requirements** in Task Master
- **Assign backup agent** for critical tasks
- **Update task status** and notes appropriately
- **Communicate availability** and return timeline

#### Knowledge Transfer
- **Maintain detailed task notes** for continuity
- **Document implementation decisions** and context
- **Provide status updates** before going offline
- **Ensure critical information** is accessible to team