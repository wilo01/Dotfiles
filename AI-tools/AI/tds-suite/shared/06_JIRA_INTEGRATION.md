# JIRA Integration & Task Identification

## JIRA Identifier Requirements

### Mandatory VIS-<number> Format
All tasks must begin with a JIRA identifier in the format **VIS-<number>** where:
- **VIS** = Project prefix for the TDS Suite project
- **<number>** = Specific numeric identifier from JIRA (e.g., VIS-2759, VIS-1234)
- This identifier corresponds to a specific STORY or BUG ticket in JIRA

### Task Naming Convention
```
VIS-<number>: [Task Description]

Examples:
✅ VIS-2759: Implement NDA translation functionality
✅ VIS-1234: Fix user authentication bug
✅ VIS-5678: Add visitor badge printing feature
❌ Implement NDA translation (missing JIRA ID)
❌ Task-123: Fix bug (wrong format)
```

## AI Agent Responsibilities

### Prompt for Missing JIRA IDs
**CRITICAL**: If a user requests work without providing a VIS-<number> identifier:

1. **Always ask**: "What is the JIRA identifier for this task? (VIS-<number>)"
2. **Do not proceed** until the JIRA ID is provided
3. **Validate format**: Ensure it follows VIS-<number> pattern
4. **Include in task title**: Prepend all task titles with the JIRA ID

### Example Interaction
```
User: "Please add a new feature for visitor check-in"
AI: "What is the JIRA identifier for this task? (VIS-<number>)"
User: "VIS-3456"
AI: "Perfect! Creating task: VIS-3456: Add new feature for visitor check-in"
```

## Integration with Existing Workflow

### Bugwarrior-Pull Compatibility
This system integrates seamlessly with `bugwarrior-pull` workflow:
- JIRA IDs allow cross-referencing with pulled ticket information
- Maintains consistency between local tasks and JIRA stories/bugs
- Enables automated status synchronization when needed

### Task Master Integration
When using Task Master, all tasks should include JIRA IDs:

```bash
# Task creation with JIRA ID
task-master add-task --prompt="VIS-2759: Implement NDA translation functionality"

# Task titles in Task Master
1. VIS-2759: Implement NDA translation functionality
   1.1 VIS-2759: Create translation API endpoints
   1.2 VIS-2759: Update UI components for language selection
   1.3 VIS-2759: Add database schema for translations
```

### Liquibase Consistency
Similar to Liquibase changeset IDs, JIRA identifiers provide:
- **Traceability**: Link code changes to business requirements
- **Change tracking**: Understand the "why" behind each modification
- **Project management**: Connect development work to project planning

## JIRA ID Validation Rules

### Required Format
- **Pattern**: `VIS-\d+` (VIS- followed by one or more digits)
- **Case sensitive**: Must be uppercase "VIS"
- **Separator**: Must use hyphen (-)
- **Numbers only**: No letters in the numeric portion

### Examples of Valid IDs
```
VIS-1       ✅ (minimum valid ID)
VIS-2759    ✅ (current example)
VIS-12345   ✅ (longer numbers allowed)
```

### Examples of Invalid IDs
```
vis-2759    ❌ (lowercase)
VIS2759     ❌ (missing hyphen)
VIS-ABC     ❌ (letters in number portion)
V-2759      ❌ (wrong prefix)
TASK-2759   ❌ (wrong project prefix)
```

## Implementation Guidelines

### For All AI Agents
1. **Pre-task validation**: Always check for JIRA ID before starting work
2. **Consistent formatting**: Use JIRA ID in all task-related communications
3. **Documentation**: Include JIRA ID in commit messages, comments, and documentation
4. **Task hierarchy**: Maintain JIRA ID in all subtasks and related work

### For Task Management
- All main tasks must start with VIS-<number>
- Subtasks inherit the parent JIRA ID for context
- Use JIRA ID when updating task status or progress
- Include JIRA ID in progress reports and completion notifications

### Error Handling
If user provides invalid JIRA ID format:
1. Explain the correct format: "JIRA IDs must follow VIS-<number> format"
2. Provide examples: "Examples: VIS-2759, VIS-1234"
3. Request correction: "Please provide a valid JIRA identifier"
4. Do not proceed until valid ID is provided

## Benefits

### Project Management
- **Requirement traceability**: Every code change links to business requirement
- **Progress tracking**: Easy to see JIRA story/bug progress in development
- **Sprint planning**: Clear connection between JIRA board and actual work
- **Reporting**: Accurate time tracking and completion reporting

### Development Workflow
- **Context preservation**: Developers always know "why" they're working on something
- **Code review**: Reviewers can reference JIRA for business context
- **Testing**: QA can validate against original requirements
- **Maintenance**: Future developers understand the business purpose

### Team Coordination
- **Shared understanding**: All team members use same identifier system
- **Communication**: Clear reference point for discussions
- **Knowledge management**: Easier to find related work and decisions
- **Handoffs**: Smooth transition between team members on same story/bug