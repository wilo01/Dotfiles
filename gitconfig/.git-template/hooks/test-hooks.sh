#!/bin/bash
# Test script for git hooks

echo "🧪 Testing Git Hooks Configuration"
echo "=================================="
echo ""

# Test sourcing common library
echo "1. Testing hooks-common.sh library..."
HOOKS_DIR=$(dirname "$0")
if source "${HOOKS_DIR}/hooks-common.sh"; then
    echo "   ✅ Successfully loaded common library"
else
    echo "   ❌ Failed to load common library"
    exit 1
fi

# Test configuration loading
echo ""
echo "2. Testing configuration loading..."
load_hook_config
echo "   AI_MAX_TIMEOUT: $AI_MAX_TIMEOUT"
echo "   AI_INACTIVITY_TIMEOUT: $AI_INACTIVITY_TIMEOUT"
echo "   AI_SHOW_PROGRESS: $AI_SHOW_PROGRESS"
echo "   AI_PARALLEL_MODE: $AI_PARALLEL_MODE"
echo "   ENABLE_GLOBAL_HOOKS: $ENABLE_GLOBAL_HOOKS"
echo "   ENABLE_AI_COMMIT: $ENABLE_AI_COMMIT"

# Test logging functions
echo ""
echo "3. Testing logging functions..."
log_info "   ℹ️  Info message test"
log_success "   Success message test"
log_warning "   Warning message test"
log_error "   Error message test"
log_debug "   Debug message test (only visible if AI_DEBUG=true)"

# Test git utilities
echo ""
echo "4. Testing git utilities..."
echo "   Current branch: $(get_current_branch)"
echo "   JIRA tag: $(get_jira_tag)"
echo "   Project root: $(get_project_root)"
echo "   Has staged changes: $(has_staged_changes && echo "Yes" || echo "No")"

# Test progress indicator
echo ""
echo "5. Testing progress indicator (3 seconds)..."
for i in {1..3}; do
    show_progress "Testing progress" "$i"
    sleep 1
done
clear_progress
echo "   ✅ Progress indicator test complete"

# Test command validation
echo ""
echo "6. Testing command validation..."
if validate_commands git bash mktemp; then
    echo "   ✅ Required commands are available"
else
    echo "   ❌ Some required commands are missing"
fi

# Test AI command availability
echo ""
echo "7. Testing AI tool availability..."
ai_tools_found=""
if command_exists claude; then
    echo "   ✅ Claude CLI found at: $(which claude)"
    ai_tools_found="yes"
else
    echo "   ⚠️  Claude CLI not found"
fi

if command_exists gemini; then
    echo "   ✅ Gemini CLI found at: $(which gemini)"
    ai_tools_found="yes"
else
    echo "   ⚠️  Gemini CLI not found"
fi

if [[ -z "$ai_tools_found" ]]; then
    echo "   ⚠️  No AI tools found - AI commit generation will not work"
fi

# Summary
echo ""
echo "=================================="
echo "🎯 Test Summary"
echo "=================================="
echo ""
echo "Configuration Settings:"
echo "  - Hooks enabled: $ENABLE_GLOBAL_HOOKS"
echo "  - AI commits: $ENABLE_AI_COMMIT"
echo "  - Parallel mode: $AI_PARALLEL_MODE"
echo "  - Timeout: ${AI_INACTIVITY_TIMEOUT}s"
echo ""
echo "To enable AI commit messages:"
echo "  git config --local hooks.enableAiCommit true"
echo ""
echo "To enable parallel AI racing:"
echo "  git config --local hooks.aiParallelMode true"
echo ""
echo "To adjust timeout (default 30s):"
echo "  git config --local hooks.aiInactivityTimeout 45"
echo ""
echo "✨ Hook system is ready to use!"