# File Conflict Prevention Protocol

## Overview

This protocol establishes clear rules and procedures to prevent file editing conflicts between Claude Code and GitHub Copilot, while ensuring efficient collaboration and maintaining code quality.

## File Ownership Matrix

### Primary File Ownership

#### Claude Code - Primary Areas
```
source/ui/app/                    # ExtJS portlets and main UI components
├── portlets/                     # Individual screen components
├── viewmodels/                   # ExtJS view models
├── views/                        # ExtJS views and panels
└── controllers/                  # ExtJS controllers

source/server/rt/                 # APEX resource templates
├── *.apex                        # All APEX endpoint files

source/ui/css/                    # Main application CSS
├── safe.css                      # Primary stylesheet
├── dashboard.css                 # Dashboard styles
└── *.css                         # Component-specific styles
```

#### Copilot - Secondary Areas
```
source/ui/lib/                    # Third-party libraries and utilities
├── utils/                        # Utility functions
├── helpers/                      # Helper functions
└── shared/                       # Shared components

source/ui/css/icons/              # Icon management
├── new_icons/                    # New icon additions
└── flags/                        # Flag icon management

build/                            # Build and configuration files
├── package.json                  # Dependencies
├── webpack.config.js             # Build configuration
└── .eslintrc.js                  # Linting configuration
```

#### Shared Areas (Coordination Required)
```
test/                             # Testing files (coordination needed)
├── Cypress/                      # E2E tests
└── unit/                         # Unit tests

documentation/                    # Documentation files
├── README.md                     # Project documentation
└── guides/                       # User guides
```

### File Status Tracking

#### Task Master File Status
```bash
# Check current file assignments
task-master list | grep "in-progress"

# Look for file mentions in task descriptions
task-master show <task-id> | grep -i "source/"

# Check recent updates for file activity
task-master show <task-id>
```

#### File Lock Communication
Each agent must communicate file usage through Task Master updates:

```bash
# Starting work on files
task-master update-subtask --id=<task-id> --prompt="
FILES_EDITING: source/ui/app/portlets/ndaList.js, source/ui/css/safe.css
STATUS: Starting implementation of NDA list features
ESTIMATED_TIME: 2 hours
"

# Completing work on files
task-master update-subtask --id=<task-id> --prompt="
FILES_COMPLETED: source/ui/app/portlets/ndaList.js, source/ui/css/safe.css
STATUS: Implementation complete, files available for other work
CHANGES: Added NDA filtering, updated styling
"
```

## Pre-Edit Verification Protocol

### Before Starting Any File Edit

#### 1. Check Task Master Status
```bash
# Verify task assignment
task-master show <task-id>

# Check for conflicting assignments
task-master list | grep "in-progress"
```

#### 2. File Conflict Check
```bash
# Check if files are mentioned in other active tasks
# Search task descriptions for file paths
task-master list | grep -E "(source/|test/|\.js|\.css|\.apex)"
```

#### 3. Communication Protocol
```bash
# Announce intention to edit files
task-master update-subtask --id=<task-id> --prompt="
INTENT: About to edit [specific file paths]
PURPOSE: [Brief description of changes]
DURATION: [Estimated time]
COORDINATION: Checking for conflicts before proceeding
"
```

#### 4. Wait Period
- **Wait 5 minutes** after announcing intent
- **Check for responses** or conflicts
- **Proceed only if no conflicts detected**

## Active Conflict Detection

### Real-Time Monitoring

#### File Edit Signals
When an agent begins editing files, they must:

1. **Update Task Master immediately** with file list
2. **Set clear timestamps** for start and estimated completion
3. **Monitor for conflict warnings** from other agents
4. **Respond to coordination requests** promptly

#### Conflict Warning System
```bash
# If another agent needs to edit the same files
task-master update-subtask --id=<conflicting-task-id> --prompt="
CONFLICT_WARNING: Need to edit files currently assigned to task <other-task-id>
FILES_NEEDED: [specific file paths]
PRIORITY: [high/medium/low]
COORDINATION_REQUEST: Please coordinate timing
"
```

### Conflict Resolution Process

#### Immediate Response Protocol
1. **Stop editing immediately** when conflict detected
2. **Communicate through Task Master** within 5 minutes
3. **Assess priority and urgency** of both tasks
4. **Negotiate resolution** through Task Master
5. **Document agreed resolution** in both tasks

#### Priority Resolution Matrix
| Claude Task Priority | Copilot Task Priority | Resolution |
|---------------------|----------------------|------------|
| High | High | Coordinate timing |
| High | Medium/Low | Claude proceeds |
| Medium | High | Negotiate |
| Medium | Medium | First-come, first-served |
| Low | High/Medium | Copilot proceeds |
| Low | Low | Coordinate timing |

## File Coordination Strategies

### Sequential Coordination

#### Handoff Protocol
```bash
# Completing agent
task-master update-subtask --id=<task-id> --prompt="
HANDOFF_READY: [file paths] available for next agent
CHANGES_MADE: [summary of changes]
TESTING_STATUS: [test status]
NEXT_AGENT_NOTES: [important information for next agent]
"

# Receiving agent
task-master update-subtask --id=<next-task-id> --prompt="
HANDOFF_RECEIVED: Taking over [file paths] from task <previous-task-id>
ACKNOWLEDGED: [summary of previous changes]
NEXT_STEPS: [planned work]
"
```

### Parallel Coordination

#### File Splitting Strategy
When possible, split work to avoid conflicts:

```bash
# Claude works on JavaScript logic
FILES: source/ui/app/portlets/ndaList.js (business logic)

# Copilot works on styling
FILES: source/ui/css/nda-list.css (styling only)

# Coordination through task updates
COORDINATION: Parallel work on related files, no direct conflicts
```

#### Branch-Based Coordination
```bash
# Create feature branches for conflicting work
git checkout -b feature/claude-nda-logic
git checkout -b feature/copilot-nda-styling

# Coordinate merge timing through Task Master
MERGE_COORDINATION: Both agents coordinate PR timing
```

## Emergency Procedures

### Immediate Conflict Resolution

#### When Simultaneous Editing Detected
1. **Both agents stop immediately**
2. **Communicate in Task Master within 2 minutes**
3. **Determine who saves first** based on priority
4. **Use version control** to resolve conflicts
5. **Update protocols** to prevent recurrence

#### Emergency Communication Format
```bash
task-master update-subtask --id=<task-id> --prompt="
EMERGENCY_CONFLICT: Simultaneous editing detected
FILES_AFFECTED: [specific file paths]
MY_CHANGES: [brief description]
STATUS: Stopped editing, awaiting coordination
TIMESTAMP: $(date)
"
```

### Version Control Conflict Resolution

#### Git Conflict Resolution
```bash
# Fetch latest changes
git fetch origin

# Check for conflicts
git status

# If conflicts exist, coordinate resolution
git merge origin/main

# Resolve conflicts with other agent input
# Test thoroughly after resolution
npm run test

# Commit with clear message
git commit -m "Resolve file conflict between agents - [description]"
```

#### Backup and Recovery
```bash
# Create backup before resolving conflicts
git stash push -m "Backup before conflict resolution - $(date)"

# After resolution, verify backup not needed
git stash list
git stash drop stash@{0}  # Only if resolution successful
```

## Monitoring and Reporting

### Conflict Metrics

#### Track and Report
- **Number of conflicts** per week
- **Resolution time** for each conflict
- **File patterns** that cause frequent conflicts
- **Process improvement** opportunities

#### Weekly Conflict Review
```bash
# Review conflict incidents
grep -i "conflict" .taskmaster/tasks/*.txt

# Analyze patterns
# - Which files cause most conflicts?
# - What times of day do conflicts occur?
# - Which task types conflict most often?

# Update protocols based on findings
```

### Process Improvement

#### Continuous Optimization
1. **Identify conflict hotspots** and create specific rules
2. **Develop file-specific protocols** for frequently conflicted files
3. **Create automation** for conflict detection
4. **Improve communication** templates and timing

#### Success Metrics
- **Conflict frequency reduction** over time
- **Faster conflict resolution** times
- **Improved coordination** efficiency
- **Higher code quality** maintenance

## Special Case Protocols

### Large File Restructuring

#### Coordination for Major Changes
```bash
# Announce major restructuring plans
task-master add-task --prompt="
MAJOR_RESTRUCTURING: Planning significant changes to [file/directory]
SCOPE: [detailed description of changes]
COORDINATION_NEEDED: All agents coordinate timing
ESTIMATED_IMPACT: [duration and scope of file locks]
"
```

#### Structured Approach
1. **Plan changes collaboratively** through Task Master
2. **Create detailed timeline** with checkpoints
3. **Assign specific phases** to specific agents
4. **Test extensively** at each phase
5. **Document all changes** thoroughly

### Hotfix Scenarios

#### Emergency Fix Protocol
```bash
# Emergency fix announcement
task-master add-task --prompt="
HOTFIX_EMERGENCY: Critical issue requiring immediate fix
FILES_NEEDED: [specific files]
ESTIMATED_TIME: [very short timeline]
OTHER_WORK: Please pause conflicting work temporarily
COORDINATION: Emergency override - high priority
"
```

#### Post-Hotfix Coordination
1. **Communicate fix completion** immediately
2. **Update all affected agents** about changes
3. **Resume paused work** in coordinated manner
4. **Document emergency** for future reference