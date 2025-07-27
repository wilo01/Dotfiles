# Code Review Template

## Description
Use this template when requesting a code review for changes you've made.

## Template
```
Please review the following code changes for ${PROJECT_NAME}.

**Purpose:**
${PURPOSE_DESCRIPTION}

**Files Changed:**
${FILES_CHANGED}

**Key Changes:**
- ${CHANGE_1}
- ${CHANGE_2}
- ${CHANGE_3}

**Areas of Focus:**
Please pay special attention to:
- ${FOCUS_AREA_1}
- ${FOCUS_AREA_2}
- ${FOCUS_AREA_3}

**Testing:**
${TESTING_DESCRIPTION}

**Code:**
${CODE_CONTENT}

Please check for:
- Code quality and best practices
- Security vulnerabilities
- Performance implications
- Maintainability
- Test coverage
```

## Variables
- ${PROJECT_NAME} - Name of the project
- ${PURPOSE_DESCRIPTION} - What the changes are meant to accomplish
- ${FILES_CHANGED} - List of modified files
- ${CHANGE_1}, ${CHANGE_2}, ${CHANGE_3} - Key changes made
- ${FOCUS_AREA_1}, ${FOCUS_AREA_2}, ${FOCUS_AREA_3} - Specific areas to review
- ${TESTING_DESCRIPTION} - How the changes were tested
- ${CODE_CONTENT} - The actual code changes

## Usage
```bash
prompt-manager template code-review
```