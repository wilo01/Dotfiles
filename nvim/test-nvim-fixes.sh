#!/bin/bash

# Neovim Error Fix Testing Script
# This script helps test all the fixes we implemented

echo "🧪 Neovim Error Fix Testing Suite"
echo "================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TOTAL_TESTS=0
PASSED_TESTS=0

# Function to run a test
run_test() {
    local test_name="$1"
    local test_command="$2"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${YELLOW}Test $TOTAL_TESTS: $test_name${NC}"
    echo "Command: $test_command"
    
    if eval "$test_command"; then
        echo -e "${GREEN}✓ PASSED${NC}\n"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}✗ FAILED${NC}\n"
    fi
}

# Function to check log for errors
check_log_errors() {
    local error_pattern="$1"
    local description="$2"
    
    if grep -q "$error_pattern" ~/.local/state/nvim/log 2>/dev/null; then
        echo -e "${RED}✗ Found: $description${NC}"
        return 1
    else
        echo -e "${GREEN}✓ Not found: $description${NC}"
        return 0
    fi
}

echo "1. Pre-test Setup"
echo "-----------------"
echo "Backing up current log..."
cp ~/.local/state/nvim/log ~/.local/state/nvim/log.backup 2>/dev/null || echo "No existing log to backup"

echo "Clearing log for fresh test..."
echo "" > ~/.local/state/nvim/log

echo -e "\n2. Testing Treesitter with Various File Types"
echo "----------------------------------------------"

# Create test files
mkdir -p /tmp/nvim-test

# Large file test
echo "Creating large test file (>100KB)..."
head -c 150000 /dev/urandom | base64 > /tmp/nvim-test/large.txt

# Minified JS test
echo "Creating minified JS file..."
echo 'function test(){var a=1,b=2,c=3;return a+b+c}for(var i=0;i<1000;i++){console.log(i)}' > /tmp/nvim-test/test.min.js

# Normal files
echo "Creating normal test files..."
cat > /tmp/nvim-test/test.lua << 'EOF'
local function test()
    local value = 42
    return value * 2
end
EOF

cat > /tmp/nvim-test/test.js << 'EOF'
function testFunction() {
    const value = 42;
    return value * 2;
}
EOF

echo -e "\n3. Running Neovim Tests"
echo "------------------------"

# Test 1: Open and close large file
run_test "Large file handling" "nvim /tmp/nvim-test/large.txt -c 'sleep 100m' -c 'qa!' 2>/dev/null"

# Test 2: Open minified JS
run_test "Minified JS file handling" "nvim /tmp/nvim-test/test.min.js -c 'sleep 100m' -c 'qa!' 2>/dev/null"

# Test 3: Open normal Lua file
run_test "Normal Lua file" "nvim /tmp/nvim-test/test.lua -c 'sleep 100m' -c 'qa!' 2>/dev/null"

# Test 4: Open normal JS file
run_test "Normal JS file" "nvim /tmp/nvim-test/test.js -c 'sleep 100m' -c 'qa!' 2>/dev/null"

# Test 5: Multiple buffers
run_test "Multiple buffers" "nvim /tmp/nvim-test/test.lua /tmp/nvim-test/test.js -c 'bnext' -c 'bprev' -c 'qa!' 2>/dev/null"

# Test 6: Terminal test
run_test "Terminal handling" "nvim -c 'terminal' -c 'sleep 100m' -c 'qa!' 2>/dev/null"

# Test 7: Quick open/close (TUI race condition test)
run_test "Quick exit (TUI test)" "nvim -c 'qa!' 2>/dev/null"

echo -e "\n4. Checking Error Log"
echo "---------------------"
echo "Analyzing ~/.local/state/nvim/log for errors..."
echo ""

# Check for specific errors
check_log_errors "Invalid 'end_col': out of range" "Treesitter end_col errors"
check_log_errors "Invalid 'end_row': out of range" "Treesitter end_row errors"
check_log_errors "Invalid 'col': out of range" "Treesitter col errors"
check_log_errors "Bad file descriptor" "File descriptor errors"
check_log_errors "TUI already stopped" "TUI race conditions"
check_log_errors "stream write failed" "RPC channel errors"
check_log_errors "Index out of bounds" "Index errors"

echo -e "\n5. LSP Test (if applicable)"
echo "---------------------------"
if command -v npm &> /dev/null; then
    echo "Creating test TypeScript file..."
    cat > /tmp/nvim-test/test.ts << 'EOF'
interface TestInterface {
    value: number;
    getName(): string;
}

class TestClass implements TestInterface {
    value: number = 42;
    
    getName(): string {
        return "test";
    }
}
EOF
    
    run_test "TypeScript LSP test" "nvim /tmp/nvim-test/test.ts -c 'sleep 500m' -c 'qa!' 2>/dev/null"
else
    echo "npm not found, skipping LSP tests"
fi

echo -e "\n6. Stress Test"
echo "--------------"
echo "Running rapid file switching test..."

cat > /tmp/nvim-test/stress-test.vim << 'EOF'
" Open multiple files
edit /tmp/nvim-test/test.lua
split /tmp/nvim-test/test.js
vsplit /tmp/nvim-test/test.ts

" Switch buffers rapidly
for i in range(10)
    bnext
    bprev
endfor

" Force treesitter operations
TSBufDisable highlight
TSBufEnable highlight

" Exit
qa!
EOF

run_test "Stress test" "nvim -S /tmp/nvim-test/stress-test.vim 2>/dev/null"

echo -e "\n7. Test Summary"
echo "==============="
echo -e "Tests Passed: ${GREEN}$PASSED_TESTS${NC}/$TOTAL_TESTS"

if [ $PASSED_TESTS -eq $TOTAL_TESTS ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
else
    echo -e "${YELLOW}⚠ Some tests failed. Check the log for details.${NC}"
fi

echo -e "\n8. Log Analysis"
echo "---------------"
LOG_SIZE=$(wc -c < ~/.local/state/nvim/log 2>/dev/null || echo 0)
echo "Log file size: $LOG_SIZE bytes"

if [ "$LOG_SIZE" -gt 1000 ]; then
    echo -e "${YELLOW}⚠ Log file contains entries. Showing last 10 lines:${NC}"
    tail -n 10 ~/.local/state/nvim/log
else
    echo -e "${GREEN}✓ Log file is minimal (good!)${NC}"
fi

echo -e "\n9. Cleanup"
echo "----------"
echo "Cleaning up test files..."
rm -rf /tmp/nvim-test

echo -e "\n${GREEN}Testing complete!${NC}"
echo ""
echo "Additional manual tests you can perform:"
echo "1. Open Neovim normally and edit some files"
echo "2. Use LSP features (go to definition, hover, etc.)"
echo "3. Open very large files or binary files"
echo "4. Use terminal mode (:terminal)"
echo "5. Exit Neovim in various ways (ZZ, :q, :qa!, etc.)"
echo ""
echo "Monitor the log file with: tail -f ~/.local/state/nvim/log"