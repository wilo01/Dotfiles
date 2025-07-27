# Debug Session Template

## Description
Use this template when starting a debugging session for a specific issue.

## Template
```
I need help debugging ${ISSUE_TYPE} in ${PROJECT_COMPONENT}.

**Issue Description:**
${ISSUE_DESCRIPTION}

**Expected Behavior:**
${EXPECTED_BEHAVIOR}

**Actual Behavior:**
${ACTUAL_BEHAVIOR}

**Steps to Reproduce:**
1. ${STEP_1}
2. ${STEP_2}
3. ${STEP_3}

**Environment:**
- OS: ${OS}
- Language/Framework: ${LANGUAGE_FRAMEWORK}
- Version: ${VERSION}

**Relevant Code/Logs:**
${CODE_OR_LOGS}

**What I've Tried:**
- ${ATTEMPT_1}
- ${ATTEMPT_2}

Please help me identify the root cause and provide a solution.
```

## Variables
- ${ISSUE_TYPE} - Type of issue (performance, crash, incorrect output, etc.)
- ${PROJECT_COMPONENT} - Specific component or module having issues
- ${ISSUE_DESCRIPTION} - Brief description of what's wrong
- ${EXPECTED_BEHAVIOR} - What should happen
- ${ACTUAL_BEHAVIOR} - What actually happens
- ${STEP_1}, ${STEP_2}, ${STEP_3} - Steps to reproduce
- ${OS} - Operating system
- ${LANGUAGE_FRAMEWORK} - Programming language or framework
- ${VERSION} - Version information
- ${CODE_OR_LOGS} - Relevant code snippets or error logs
- ${ATTEMPT_1}, ${ATTEMPT_2} - Previous debugging attempts

## Usage
```bash
prompt-manager template debug-session
```