# TDS Suite - Claude Code Configuration

## Project Overview
**TDS Suite** is a comprehensive enterprise visitor and access management system built on Oracle APEX with multiple ExtJS front-end applications. This configuration extends the global Claude Code settings with TDS-specific guidance.

## Technology Stack & Architecture

### Core Technologies
- **ExtJS 6.0.2 - 7.4.0** - Primary UI framework (Classic and Modern toolkits)
- **JavaScript (ES5/ES6)** - Core application logic
- **Oracle APEX** - Application development platform
- **Oracle Database** - Primary data storage with PL/SQL procedures
- **Java (JDK 8)** - Server-side runtime and APEX integration
- **GlassFish** - Java EE application server

### Development Stack
- **Liquibase** - Database version control and migration
- **Cypress 13.13.0** - End-to-end testing framework
- **Sencha CMD** - ExtJS application building and development
- **Docker** - Containerization for development environments

## Project Structure

### Main Projects Location: `~/Dev/branch-opener/branches/`

```
tds-suite/
├── source/
│   ├── server/              # Backend services and APIs
│   │   ├── auth/           # Authentication modules
│   │   ├── database/       # Database scripts, migrations
│   │   ├── rt/             # Resource templates (.apex files)
│   │   └── glassfish/      # Java EE configurations
│   ├── ui/                 # Main web interface
│   ├── ui-kiosk/          # Kiosk interface (ExtJS 6.2.1)
│   ├── ui-student-portal/ # Student portal (ExtJS 7.4.0)
│   └── ui-muster/         # Emergency muster interface
├── test/
│   ├── Cypress/           # E2E test suite
│   └── load/              # Performance testing
└── selfhost/              # Self-hosting tools
```

## Development Workflow

### TDS-Specific Commands
```bash
# Database Operations
npm run liquibaseLocalDockerUpdate  # Database updates via Docker
npm run lldupdate                   # Shorthand for database updates
make db-update                      # Direct Liquibase updates

# Testing
npm run devDocker                   # Interactive Cypress testing
npm run devDockerAll                # Run all Cypress tests
npm run devDockerAccess             # Access-specific tests
npm run devDockerVisitor            # Visitor-specific tests

# Build & Development
sencha app build development        # ExtJS development build
sencha app build production         # ExtJS production build
sencha app watch                    # Development server with live reload
```

## File Ownership & Responsibilities

### Primary File Ownership
- **ExtJS Components**: `source/ui/app/` (portlets, main UI components)
- **APEX Resource Templates**: `source/server/rt/` (.apex files)
- **UI Styling**: `source/ui/css/` (CSS files for backoffice)
- **Main Application Logic**: Core application files and configurations

### Secondary Responsibilities
- **Database Operations**: Complex database queries and procedures
- **API Development**: REST API endpoint implementation
- **Performance Optimization**: Code optimization and performance tuning
- **Security Implementation**: Security features and authentication

## Technical Guidelines

### ExtJS Development
- Follow ExtJS MVC/MVVM patterns
- Use proper component lifecycle management
- Implement data stores and models correctly
- Leverage existing portlet patterns
- Components extend ExtJS base classes

### Oracle APEX Development
- Follow .apex file format standards
- Implement proper SQL and PL/SQL procedures
- Use parameterized queries for security
- Include permission checks in all endpoints
- Use `ca_menu_pak.manage_menu_item()` for menu management

### Database Integration
- Oracle Database with PL/SQL stored procedures for business logic
- Liquibase for database version control and migrations
- Database changes controlled through structured changesets
- Use parameterized queries to prevent SQL injection

### Security Implementation
- Always implement CSRF protection
- Use proper input validation and sanitization
- Follow APEX permission patterns
- Implement proper error handling
- Token-based authentication with user session management
- Role-based access control for user permissions

## Task Management Principles

### Core Task Rules
1. **Use Task Master instead of TODOs**
2. **JIRA ID Required**: Every task must start with VIS-<number> identifier (see `shared/06_JIRA_INTEGRATION.md`)
3. **Prompt for JIRA ID**: If user doesn't provide VIS-<number>, always ask before proceeding
4. **Mark tasks in-progress before starting**
5. **Mark tasks for review when completed (never "done")**
6. **Always look for next task when current task is complete**
7. **When a task is reopened, set the main task to 'pending', create new subtasks for the required work, and set all existing subtasks to 'pending'.**

### Task Status Management
- **pending**: Ready to work on
- **in-progress**: Currently being worked on (only one at a time)
- **review**: Completed and ready for review
- **blocked**: Waiting on external factors

## Multi-Application Architecture

The TDS Suite supports multiple specialized applications:
- **Main UI**: Primary back-office application with full feature set
- **Kiosk Application**: Touch-enabled interface for public access
- **Muster Application**: Emergency procedures and headcount functionality
- **Student Portal**: Self-service student interface

## Project-Specific Features

### TDS Suite Functionality
- **Visitor Management**: Registration, approval workflows, badge printing
- **Access Control**: Security levels, time profiles, terminal management
- **Student Portal**: Academic tracking, attendance, course management
- **Kiosk Interface**: Self-service visitor registration
- **Emergency Management**: Muster points, evacuation procedures
- **Multi-tenant**: Location-based configurations with role-based access

## Knowledge Base References

### Shared Knowledge Modules
- `shared/00_PROJECT_OVERVIEW.md` - Complete project overview
- `shared/01_TASK_MANAGEMENT.md` - Task management guidelines
- `shared/02_UI_DEVELOPMENT.md` - UI development patterns
- `shared/03_BACKEND_DEVELOPMENT.md` - Backend development guidelines
- `shared/04_TESTING_DEPLOYMENT.md` - Testing and deployment procedures
- `shared/05_SECURITY_PERMISSIONS.md` - Security implementation guides
- `shared/06_JIRA_INTEGRATION.md` - JIRA integration requirements
- `shared/08_COMPREHENSIVE_TEST_REPORT.md` - Testing standards

### Implementation Patterns
- `patterns/TEXT_TRANSLATION_PATTERNS.md` - Text translation implementations
- `patterns/UNIFIED_CONTENT_DISPLAY_PATTERN.md` - Content display patterns
- `context/patterns/` - ExtJS components, database integration patterns

### Coordination Protocols
- `protocols/MULTI_AGENT_COORDINATION.md` - Multi-agent coordination rules
- `protocols/FILE_CONFLICT_PREVENTION.md` - File conflict prevention

## Context Management

### Session Management
- Use `/clear` between different major tasks
- Keep focused context on current implementation
- Reference shared knowledge modules when needed
- Use Task Master for detailed task context

### Knowledge Base Usage
- Reference `shared/` modules for guidelines
- Use `protocols/` for coordination rules
- Follow patterns established in existing codebase
- Consult documentation for complex implementations

## Multi-Agent Coordination

### Agent Roles
- **Claude Code**: Senior Software Developer (complex implementation)
- **Gemini**: Task Manager & Planner (planning mode only)
- **GitHub Copilot**: Coordinated Assistant (simple tasks)

### Communication Protocol
- **Use Task Master** for all task-related communication
- **Provide detailed implementation notes** in task updates
- **Coordinate with Gemini** for planning and prioritization
- **Report blockers and issues** through Task Master

### File Coordination Protocol
1. **Check Task Master** for file assignments before editing
2. **Communicate file usage** through Task Master
3. **Avoid simultaneous editing** of the same files
4. **Use version control** effectively for collaboration

This configuration works in conjunction with the global Claude Code settings at `~/AI/` and provides TDS Suite-specific guidance for effective development.
