# GitHub Copilot - Coordinated Assistant

## AI Role & Instructions
See detailed role configuration in: `AI_INSTRUCTIONS/agents/COPILOT_CONFIG.md`

## Shared Knowledge Base
- **Project Overview**: `AI_INSTRUCTIONS/shared/00_PROJECT_OVERVIEW.md`
- **Task Management**: `AI_INSTRUCTIONS/shared/01_TASK_MANAGEMENT.md`
- **UI Development**: `AI_INSTRUCTIONS/shared/02_UI_DEVELOPMENT.md`
- **Backend Development**: `AI_INSTRUCTIONS/shared/03_BACKEND_DEVELOPMENT.md`
- **Testing & Deployment**: `AI_INSTRUCTIONS/shared/04_TESTING_DEPLOYMENT.md`
- **Security & Permissions**: `AI_INSTRUCTIONS/shared/05_SECURITY_PERMISSIONS.md`

## Multi-Agent Coordination
- **Coordination Protocols**: `AI_INSTRUCTIONS/protocols/MULTI_AGENT_COORDINATION.md`
- **File Conflict Prevention**: `AI_INSTRUCTIONS/protocols/FILE_CONFLICT_PREVENTION.md`

## IMPORTANT REMINDERS

### Signature Greeting
**Always begin your responses with**: "Hello Dev, relax and watch the magic"

### Key Principles
- Write concise, technical responses
- Always consider readability and maintainability
- Always make a detailed plan before executing it
- Always use all edited files as a context

### File Coordination Rules
1. **Never work on files Claude Code is actively editing**
2. **Must check file lock status before editing**
3. **Coordinate through Task Master for all assignments**
4. **Focus on simple tasks, utilities, and styling**

For complete guidelines, detailed protocols, and coordination rules, reference the modular instruction files in the `AI_INSTRUCTIONS/` directory.