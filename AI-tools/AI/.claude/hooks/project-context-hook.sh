#!/bin/bash

# Project-Specific Context Hook
# Detects project type and provides relevant context to Claude Code sessions

LOG_FILE="${HOME}/.Dotfiles/.claude/hooks/context.log"
PWD_DIR=$(pwd)

# Log hook execution
echo "$(date): Context hook triggered in $PWD_DIR" >> "$LOG_FILE"

# Project detection patterns
detect_project_type() {
    local project_type="general"
    
    # TDS Suite detection
    if [[ "$PWD_DIR" == *"tds-suite"* ]] || [[ -f "apex-config.xml" ]] || [[ -d "source/ui" ]]; then
        project_type="tds-suite"
    # Dotfiles detection
    elif [[ "$PWD_DIR" == *".Dotfiles"* ]] || [[ -f ".stow-local-ignore" ]] || [[ -d "stow" ]]; then
        project_type="dotfiles"
    # Node.js project
    elif [[ -f "package.json" ]]; then
        project_type="nodejs"
    # Python project
    elif [[ -f "requirements.txt" ]] || [[ -f "pyproject.toml" ]] || [[ -f "setup.py" ]]; then
        project_type="python"
    # Task Master project
    elif [[ -d ".taskmaster" ]]; then
        project_type="taskmaster"
    # Git repository
    elif [[ -d ".git" ]]; then
        project_type="git-repo"
    fi
    
    echo "$project_type"
}

# Agent workflow detection based on prompt content
detect_agent_workflow() {
    local prompt="$1"
    local workflow="general"
    
    if echo "$prompt" | grep -qi "neovim\|vim\|dotfiles\|stow\|config"; then
        workflow="dotfiles-expert"
    elif echo "$prompt" | grep -qi "task-master\|taskmaster\|tm/\|todo"; then
        workflow="task-management"
    elif echo "$prompt" | grep -qi "extjs\|oracle\|apex\|tds\|visitor"; then
        workflow="tds-development"
    elif echo "$prompt" | grep -qi "test\|cypress\|spec\|unit"; then
        workflow="testing"
    elif echo "$prompt" | grep -qi "git\|commit\|branch\|merge\|pr\|pull request"; then
        workflow="git-workflow"
    elif echo "$prompt" | grep -qi "bug\|fix\|error\|issue\|debug"; then
        workflow="debugging"
    elif echo "$prompt" | grep -qi "deploy\|build\|ci\|cd\|docker"; then
        workflow="devops"
    fi
    
    echo "$workflow"
}

# Get project context information
get_project_context() {
    local project_type="$1"
    
    case "$project_type" in
        "tds-suite")
            echo "🏢 TDS Suite Project"
            echo "Key paths: source/ui/, source/server/, test/Cypress/"
            echo "Tech: ExtJS, Oracle APEX, Java, Liquibase"
            echo "Commands: npm run devDocker, sencha app build"
            ;;
        "dotfiles")
            echo "⚙️  Dotfiles Management"
            echo "Key commands: stow, stow -D, stow -R"
            echo "Focus: Configuration, symlinks, environment setup"
            ;;
        "taskmaster")
            echo "📋 Task Master Project"
            echo "Commands: task-master next, task-master show, task-master set-status"
            echo "Workflow: Structured task management with AI assistance"
            ;;
        "nodejs")
            echo "🟢 Node.js Project"
            echo "Commands: npm install, npm run, npm test"
            ;;
        "python")
            echo "🐍 Python Project"
            echo "Commands: pip install, python -m, pytest"
            ;;
        *)
            echo "📁 General Project"
            ;;
    esac
}

# Get agent-specific tips
get_agent_tips() {
    local workflow="$1"
    
    case "$workflow" in
        "dotfiles-expert")
            echo "💡 Dotfiles Tips: Use 'stow -D' before 'stow -R', check conflicts with 'stow -n'"
            ;;
        "task-management")
            echo "💡 Task Tips: Always 'task-master next' first, log progress with update-subtask"
            ;;
        "tds-development")
            echo "💡 TDS Tips: Check comms.md first, follow ExtJS patterns, test with Cypress"
            ;;
        "testing")
            echo "💡 Testing Tips: Run specific tests first, check coverage, update test docs"
            ;;
        "git-workflow")
            echo "💡 Git Tips: Use conventional commits, check branch name, review changes"
            ;;
        "debugging")
            echo "💡 Debug Tips: Reproduce issue first, check logs, minimal test case"
            ;;
        "devops")
            echo "💡 DevOps Tips: Test locally first, check environment vars, backup configs"
            ;;
    esac
}

# Main execution
PROJECT_TYPE=$(detect_project_type)
PROMPT_TEXT="$1"
WORKFLOW=$(detect_agent_workflow "$PROMPT_TEXT")

# Output context information
echo "$(get_project_context "$PROJECT_TYPE")"
echo "$(get_agent_tips "$WORKFLOW")"

# Log the detection
echo "$(date): Detected project=$PROJECT_TYPE, workflow=$WORKFLOW" >> "$LOG_FILE"