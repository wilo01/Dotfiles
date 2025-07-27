# Global Claude Code Instructions

## Overview
This file provides global guidance to Claude Code (claude.ai/code) across all projects. Project-specific instructions are located in their respective directories (e.g., `~/AI/tds-suite/CLAUDE.md`).

## Global Configuration System

### Configuration Hierarchy
1. **Global Config**: `~/AI/` (this directory) - Shared across all Claude instances
2. **Project Config**: `~/AI/project-name/` - Project-specific extensions and overrides
3. **Auto-loaded**: Claude Code automatically loads both global and project context

### Directory Structure
```
~/AI/                         # Global AI configurations
├── .claude/                  # Global Claude settings & commands
├── .taskmaster/             # Global Task Master config
├── .gemini/                 # Global Gemini config
├── .mcp.json                # Global MCP configuration
├── CLAUDE.md                # This file - global instructions
└── project-name/            # Project-specific configurations
    ├── .claude/             # Project Claude overrides
    ├── CLAUDE.md            # Project-specific instructions
    └── ...                  # Project-specific AI guidance
```

## Core Development Principles

### Code Quality Standards
- **Follow existing patterns** in the codebase
- **Use established libraries** and frameworks
- **Implement proper error handling**
- **Write self-documenting code**
- **Follow security best practices**

### Task Management
- **Use Task Master** for all task coordination and tracking
- **NEVER** create manual TODO comments - use Task Master instead
- **Mark tasks in-progress** before starting work
- **Update progress** regularly with detailed notes
- **Mark tasks complete** only when fully tested

### Security Standards
- **Always implement permission checks**
- **Use parameterized queries**
- **Validate all input data**
- **Never expose or log secrets**
- **Never commit secrets to repositories**

## Global Tool Configuration

### Task Master Integration
Essential commands available globally:
```bash
# Core workflow
task-master next                    # Get next available task
task-master show <id>              # View task details
task-master set-status --id=<id> --status=done

# Task management
task-master add-task --prompt="description"
task-master update-task --id=<id> --prompt="changes"
task-master list                   # Show all tasks
```

### MCP Integration
Global MCP servers configured in `.mcp.json`:
- **task-master-ai**: Core task management functionality
- Additional servers can be added per project

## Global Slash Commands

Available across all projects via `~/AI/.claude/commands/`:
- `/tm` - Task Master operations
- `/prompt-*` - Prompt management utilities
- Project-specific commands loaded from project directories

## Development Guidelines

### Context Management
- Use `/clear` between different major tasks
- Reference global knowledge for cross-project patterns
- Use Task Master for detailed task context
- Project-specific context automatically loaded

### Multi-Project Workflows
- **Global configs** apply to all projects
- **Project configs** extend/override globals
- **Consistent patterns** across all projects
- **Shared utilities** available everywhere

### File Organization
- **Never** edit files outside your project scope
- **Always** check Task Master before editing files
- **Coordinate** with other agents through Task Master
- **Document** architectural decisions in appropriate locations

## Centralized Management

This configuration system is managed through GNU Stow:
- **Source**: `~/.Dotfiles/AI-tools/AI/`
- **Target**: `~/AI/` (via `stow AI-tools`)
- **Version controlled** in dotfiles repository
- **Easily deployable** across different systems

## Project Integration

When working on specific projects:
1. **Global context** is automatically available
2. **Project context** loaded from `~/AI/project-name/`
3. **Combined configuration** provides complete guidance
4. **Task Master** coordinates across all contexts

For project-specific guidance, refer to the appropriate project directory (e.g., `~/AI/tds-suite/CLAUDE.md`).