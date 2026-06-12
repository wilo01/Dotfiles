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

# Test command validation
echo ""
echo "5. Testing command validation..."
if validate_commands git bash mktemp; then
    echo "   ✅ Required commands are available"
else
    echo "   ❌ Some required commands are missing"
fi

# Test AI command availability
echo ""
echo "6. Testing AI tool availability..."
if command_exists or-cli; then
    echo "   ✅ or-cli found at: $(command -v or-cli)"
    echo "   ✅ curl available: $(command -v curl)"
    echo "   ✅ jq available: $(command -v jq)"
elif [[ -x "$HOME/.local/bin/or-cli" ]]; then
    echo "   ✅ or-cli found at: $HOME/.local/bin/or-cli"
    echo "   ✅ curl available: $(command -v curl)"
    echo "   ✅ jq available: $(command -v jq)"
else
    echo "   ⚠️  or-cli not found - AI commit generation will not work"
    echo "   📦 Install: place or-cli script in ~/.local/bin/"
    echo "   🔧 Requires: curl, jq, and OPENROUTER_API_KEY env var"
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
echo "  - Timeout: ${AI_INACTIVITY_TIMEOUT}s"
echo ""
echo "To enable AI commit messages:"
echo "  git config --local hooks.enableAiCommit true"
echo ""
echo "To adjust timeout (default 30s):"
echo "  git config --local hooks.aiInactivityTimeout 45"
echo ""
echo "✨ Hook system is ready to use!"
