#!/bin/bash
# Git Hooks Analytics Viewer

# Source common functions
HOOKS_DIR=$(dirname "$0")
source "${HOOKS_DIR}/hooks-common.sh" || {
    echo "❌ Failed to load hooks-common.sh" >&2
    exit 1
}

# Initialize analytics
init_analytics

echo "📊 Git Hooks AI Performance Analytics"
echo "======================================"
echo ""

# Check if jq or python is available
if ! command_exists jq && ! command_exists python3; then
    echo "⚠️  Neither jq nor python3 found. Showing raw data:"
    cat "$ANALYTICS_FILE"
    exit 0
fi

# Display overall statistics
if command_exists jq; then
    # Use jq for parsing
    echo "📈 Overall Statistics:"
    echo "----------------------"
    
    jq -r '
        "Total API Calls: " + ((.claude.total_calls + .gemini.total_calls) | tostring) + "\n" +
        "Total Successful: " + ((.claude.successful_calls + .gemini.successful_calls) | tostring) + "\n" +
        "Total Failed: " + ((.claude.failed_calls + .gemini.failed_calls) | tostring) + "\n" +
        "Race Mode Wins: Claude " + (.claude.wins | tostring) + " vs Gemini " + (.gemini.wins | tostring)
    ' "$ANALYTICS_FILE"
    
    echo ""
    echo "🤖 Claude Performance:"
    echo "---------------------"
    jq -r --arg name "Claude" '
        .claude |
        "  Total Calls: " + (.total_calls | tostring) + "\n" +
        "  Success Rate: " + (if .total_calls > 0 then ((.successful_calls / .total_calls * 100) | floor | tostring) + "%" else "N/A" end) + "\n" +
        "  Average Time: " + (if .successful_calls > 0 then ((.total_time / .successful_calls) | floor | tostring) + "s" else "N/A" end) + "\n" +
        "  Race Wins: " + (.wins | tostring) + "\n" +
        "  Status: " + (if .disabled_until then "⚠️  Disabled until " + .disabled_until else "✅ Active" end) + "\n" +
        "  Recent Failures: " + (.recent_failures | length | tostring)
    ' "$ANALYTICS_FILE"
    
    echo ""
    echo "🤖 Gemini Performance:"
    echo "---------------------"
    jq -r --arg name "Gemini" '
        .gemini |
        "  Total Calls: " + (.total_calls | tostring) + "\n" +
        "  Success Rate: " + (if .total_calls > 0 then ((.successful_calls / .total_calls * 100) | floor | tostring) + "%" else "N/A" end) + "\n" +
        "  Average Time: " + (if .successful_calls > 0 then ((.total_time / .successful_calls) | floor | tostring) + "s" else "N/A" end) + "\n" +
        "  Race Wins: " + (.wins | tostring) + "\n" +
        "  Status: " + (if .disabled_until then "⚠️  Disabled until " + .disabled_until else "✅ Active" end) + "\n" +
        "  Recent Failures: " + (.recent_failures | length | tostring)
    ' "$ANALYTICS_FILE"
    
    echo ""
    echo "📜 Recent History (last 10):"
    echo "---------------------------"
    jq -r '
        .history[-10:] | reverse | .[] |
        "  " + .timestamp[11:19] + " - " + .ai + 
        (if .success then " ✅" else " ❌" end) +
        (if .winner then " 🏆" else "" end) +
        " (" + (.time | tostring) + "s)"
    ' "$ANALYTICS_FILE"
    
    echo ""
    echo "🎯 Recommendations:"
    echo "------------------"
    
    # Get best performer
    best_ai=$(get_best_ai)
    echo "  Best Performer: ${best_ai^}"
    
    # Get adaptive timeouts
    claude_timeout=$(get_adaptive_timeout "claude")
    gemini_timeout=$(get_adaptive_timeout "gemini")
    echo "  Optimal Timeouts: Claude ${claude_timeout}s, Gemini ${gemini_timeout}s"
    
    # Check if race mode is beneficial
    jq -r '
        if .claude.wins > 0 or .gemini.wins > 0 then
            if .claude.wins > .gemini.wins then
                "  Race Mode Winner: Claude (wins " + (.claude.wins | tostring) + " times)"
            else
                "  Race Mode Winner: Gemini (wins " + (.gemini.wins | tostring) + " times)"
            end
        else
            "  Race Mode: No data yet"
        end
    ' "$ANALYTICS_FILE"
    
elif command_exists python3; then
    # Use Python for parsing
    python3 << 'EOF'
import json
import sys

with open('${ANALYTICS_FILE}') as f:
    data = json.load(f)

print("📈 Overall Statistics:")
print("----------------------")
total_calls = data['claude']['total_calls'] + data['gemini']['total_calls']
total_success = data['claude']['successful_calls'] + data['gemini']['successful_calls']
total_failed = data['claude']['failed_calls'] + data['gemini']['failed_calls']
print(f"Total API Calls: {total_calls}")
print(f"Total Successful: {total_success}")
print(f"Total Failed: {total_failed}")
print(f"Race Mode Wins: Claude {data['claude']['wins']} vs Gemini {data['gemini']['wins']}")

for ai_name in ['claude', 'gemini']:
    ai = data[ai_name]
    print(f"\n🤖 {ai_name.capitalize()} Performance:")
    print("---------------------")
    print(f"  Total Calls: {ai['total_calls']}")
    
    if ai['total_calls'] > 0:
        success_rate = int(ai['successful_calls'] / ai['total_calls'] * 100)
        print(f"  Success Rate: {success_rate}%")
    else:
        print("  Success Rate: N/A")
    
    if ai['successful_calls'] > 0:
        avg_time = int(ai['total_time'] / ai['successful_calls'])
        print(f"  Average Time: {avg_time}s")
    else:
        print("  Average Time: N/A")
    
    print(f"  Race Wins: {ai['wins']}")
    
    if ai.get('disabled_until'):
        print(f"  Status: ⚠️  Disabled until {ai['disabled_until']}")
    else:
        print("  Status: ✅ Active")
    
    print(f"  Recent Failures: {len(ai['recent_failures'])}")

print("\n📜 Recent History (last 10):")
print("---------------------------")
for entry in data['history'][-10:][::-1]:
    time_str = entry['timestamp'][11:19]
    status = "✅" if entry['success'] else "❌"
    winner = " 🏆" if entry.get('winner') else ""
    print(f"  {time_str} - {entry['ai']} {status}{winner} ({entry['time']}s)")
EOF
fi

echo ""
echo "======================================"
echo "💡 Tips:"
echo "  - Use 'git config --local hooks.aiParallelMode true' for race mode"
echo "  - Check '$ANALYTICS_FILE' for full data"
echo "  - Analytics reset after 50 commits"