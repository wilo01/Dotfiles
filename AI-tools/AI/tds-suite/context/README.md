# AI Context Storage - Learn as You Go

This directory contains automatically generated context notes from the TDS Suite codebase to help AI agents understand project-specific patterns, common solutions, and architectural decisions.

## Directory Structure

```
context/
├── README.md                           # This file
├── patterns/                           # Code patterns and templates
│   ├── extjs-components/              # ExtJS component patterns
│   ├── database-integration/          # Database/APEX patterns  
│   ├── security-implementations/      # Security patterns
│   └── ui-patterns/                   # User interface patterns
├── architecture/                      # High-level architecture docs
│   ├── system-overview/               # System architecture
│   ├── data-flow/                     # Data flow patterns
│   └── integration-patterns/          # Integration approaches
├── common-solutions/                  # Solutions to common problems
│   ├── error-handling/                # Error handling approaches
│   ├── performance-optimizations/     # Performance solutions
│   └── troubleshooting/               # Common issues and fixes
├── project-knowledge/                 # Project-specific information
│   ├── business-logic/                # Business rules and logic
│   ├── naming-conventions/            # Coding standards
│   └── configuration/                 # Configuration patterns
└── generated/                         # Auto-generated context files
    ├── recent-changes/                # Recent code changes analysis
    ├── code-snippets/                 # Useful code snippets
    └── api-patterns/                  # API usage patterns
```

## File Naming Conventions

### Pattern Files
- Format: `{category}-{specific-pattern}.md`
- Examples: 
  - `extjs-portlet-base-pattern.md`
  - `apex-permission-check-pattern.md`
  - `form-validation-pattern.md`

### Architecture Files
- Format: `{system-area}-{architectural-aspect}.md`
- Examples:
  - `authentication-flow-overview.md`
  - `data-persistence-architecture.md`
  - `ui-state-management.md`

### Solution Files
- Format: `{problem-category}-{solution-type}.md`
- Examples:
  - `extjs-memory-leak-solutions.md`
  - `database-performance-optimizations.md`
  - `csrf-protection-implementation.md`

### Generated Files
- Format: `{timestamp}-{analysis-type}.md`
- Examples:
  - `2024-07-23-recent-api-changes.md`
  - `2024-07-23-new-component-patterns.md`
  - `2024-07-23-code-review-insights.md`

## Content Structure Template

Each context file should follow this structure:

```markdown
# Title

## Overview
Brief description of the pattern/solution/concept

## Context
When and why this pattern/solution is used

## Implementation
Code examples and implementation details

## Key Points
- Important considerations
- Common pitfalls
- Best practices

## Related Patterns
Links to related context files

## Examples
Real-world usage examples from the codebase

## Generated On
Timestamp and trigger for auto-generated files
```

## Hook Integration Points

The context hook will be triggered by:
1. File modifications in key directories
2. Task completions and reviews
3. Error resolutions
4. New pattern implementations
5. Architecture changes

This creates a living knowledge base that grows with the project and helps maintain consistency across AI agent interactions.