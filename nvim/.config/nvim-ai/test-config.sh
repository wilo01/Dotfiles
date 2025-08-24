#!/bin/bash
# Test script for nvim-ai configuration

set -e

echo "Testing nvim-ai configuration..."
echo "================================"

# Test 1: Basic startup
echo -n "1. Testing basic startup... "
if NVIM_APPNAME=nvim-ai nvim --headless -c 'quit' 2>/dev/null; then
   echo "✓ PASS"
else
   echo "✗ FAIL"
   exit 1
fi

# Test 2: Check health
echo -n "2. Testing checkhealth... "
if NVIM_APPNAME=nvim-ai nvim --headless -c 'checkhealth' -c 'quit' 2>/dev/null; then
   echo "✓ PASS"
else
   echo "✗ FAIL"
fi

# Test 3: Load a file
echo -n "3. Testing file loading... "
echo "print('test')" > /tmp/test.lua
if NVIM_APPNAME=nvim-ai nvim --headless /tmp/test.lua -c 'quit' 2>/dev/null; then
   echo "✓ PASS"
   rm /tmp/test.lua
else
   echo "✗ FAIL"
   rm /tmp/test.lua
   exit 1
fi

# Test 4: Check plugin installation
echo -n "4. Testing lazy.nvim installation... "
if NVIM_APPNAME=nvim-ai nvim --headless -c 'lua require("lazy")' -c 'quit' 2>/dev/null; then
   echo "✓ PASS"
else
   echo "✗ FAIL"
fi

# Test 5: Startup time
echo -n "5. Measuring startup time... "
START_TIME=$(NVIM_APPNAME=nvim-ai nvim --startuptime /tmp/nvim-startup.log --headless -c 'quit' 2>&1 && tail -1 /tmp/nvim-startup.log | awk '{print $1}')
echo "$START_TIME ms"
rm -f /tmp/nvim-startup.log

echo ""
echo "Configuration test complete!"
echo "============================"
echo "Startup time: $START_TIME ms"

# Compare with original if it exists
if command -v nvim &> /dev/null; then
   echo ""
   echo "Comparison with default config:"
   ORIG_TIME=$(nvim --startuptime /tmp/nvim-startup-orig.log --headless -c 'quit' 2>&1 && tail -1 /tmp/nvim-startup-orig.log | awk '{print $1}')
   echo "Original startup time: $ORIG_TIME ms"
   rm -f /tmp/nvim-startup-orig.log

   # Calculate improvement
   if [ ! -z "$ORIG_TIME" ] && [ ! -z "$START_TIME" ]; then
      IMPROVEMENT=$(echo "scale=2; (($ORIG_TIME - $START_TIME) / $ORIG_TIME) * 100" | bc 2>/dev/null || echo "N/A")
      if [ "$IMPROVEMENT" != "N/A" ]; then
         echo "Improvement: ${IMPROVEMENT}%"
      fi
   fi
fi