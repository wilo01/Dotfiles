# TreeSitter & Highlight-on-Yank Fix Plan
*Generated: 2025-08-20*  
*Context: Neovim 0.11.3 TreeSitter range errors preventing highlight-on-yank*

## 🚨 Current Problem Summary

### Issue Description
Neovim 0.11.3 generates TreeSitter highlighter errors that flood the log file and conflict with highlight-on-yank functionality. The TextYankPost autocmd was disabled as a temporary workaround.

### Error Pattern in Log
```
ERR 2025-08-20T07:39:02.831 nvim.2925730.0 decor_provider_error:36: Error in decoration provider "line" (ns=nvim.treesitter.highlighter):
Error executing lua: ...al/share/nvim/runtime/lua/vim/treesitter/highlighter.lua:370: Invalid 'end_row': out of range
```

**Common Variations:**
- `Invalid 'end_col': out of range`
- `Invalid 'col': out of range`
- `Invalid 'end_row': out of range`

### Current Status
- ✅ **Temporary Fix Applied**: Highlight-on-yank disabled for Neovim 0.11+
- ✅ **Backwards Compatible**: Works on Neovim < 0.11
- ❌ **Missing Feature**: No highlight-on-yank for Neovim 0.11.3+
- ❌ **Root Cause**: TreeSitter range errors still occur

## 🔍 Root Cause Analysis

### Primary Causes
1. **Neovim 0.11 Breaking Change**: TreeSitter highlighting became asynchronous
2. **Range Calculation Bug**: TreeSitter calculates invalid buffer ranges during async operations  
3. **Timing Conflicts**: Race conditions between highlight operations and TreeSitter
4. **Stricter Validation**: Neovim 0.11+ has tighter extmark range validation

### Trigger Points Identified
1. **TextYankPost autocmd** - Triggers TreeSitter recalculation
2. **statusline.lua:473** - `vim.cmd('redrawstatus')` forces redraws
3. **LSP Semantic Tokens** - Interaction with TreeSitter highlighting
4. **File Editing Operations** - Any buffer changes can trigger range errors

### Technical Details
- TreeSitter uses `nvim_buf_set_extmark()` with default priority 100
- Asynchronous highlighting causes timing issues with other highlight operations
- Invalid ranges occur when buffer changes faster than TreeSitter can recalculate
- Error happens at `highlighter.lua:370` in `nvim_buf_set_extmark` function

## 📁 Current File Locations & Status

### Files Modified
1. **`/home/dariuszw/.Dotfiles/nvim/.config/nvim/lua/DarO/autocmds.lua`**
   - Lines 69-81: TextYankPost autocmd wrapped in version check
   - Status: Disabled for nvim-0.11+

2. **`/home/dariuszw/.Dotfiles/nvim/.config/nvim/lua/DarO/statusline.lua`**
   - Line 473: Contains `vim.cmd('redrawstatus')` - potential trigger
   - Status: Unchanged, potential source of errors

3. **`/home/dariuszw/.Dotfiles/nvim/.config/nvim/lua/DarO/lazy/treesitter.lua`**
   - Current config: Basic setup with file size limits
   - Status: Could use stability improvements

### Version Detection Pattern
Current pattern used in LSP config:
```lua
if vim.fn.has('nvim-0.11') == 1 then
   -- Modern behavior
else
   -- Legacy behavior  
end
```

## 🎯 Comprehensive Fix Plan

### Strategy Overview
1. **Keep Current Workaround**: Don't break existing functionality
2. **Implement Protected Highlight-on-Yank**: Add safe version for 0.11.3+
3. **Fix TreeSitter Root Causes**: Address range validation and timing
4. **Enhance Statusline Safety**: Prevent redraw-triggered errors

### Implementation Phases

#### Phase 1: Enhanced autocmds.lua
**Goal**: Re-enable highlight-on-yank for 0.11.3+ with protection

**Current Code** (lines 69-81):
```lua
-- Skip TextYankPost autocmd for NVIM v0.11+ due to TreeSitter highlighter conflicts
if vim.fn.has('nvim-0.11') == 0 then
   autocmd('TextYankPost', {
      group = augroup('HighlightYank', {}),
      pattern = '*',
      callback = function()
         vim.highlight.on_yank({
            higroup = 'IncSearch',
            timeout = 40,
         })
      end,
   })
end
```

**Replacement Code**:
```lua
-- Highlight on yank with version-specific optimizations
if vim.fn.has('nvim-0.11') == 1 then
   -- Neovim 0.11+: Protected implementation with async safety
   autocmd('TextYankPost', {
      group = augroup('HighlightYank', {}),
      pattern = '*',
      callback = function()
         -- Defer execution to avoid TreeSitter race conditions
         vim.schedule(function()
            pcall(function()
               vim.highlight.on_yank({
                  higroup = 'IncSearch',
                  timeout = 150,  -- Longer timeout for stability
                  priority = 150, -- Higher than TreeSitter default (100)
               })
            end)
         end)
      end,
   })
else
   -- Neovim < 0.11: Original working implementation
   autocmd('TextYankPost', {
      group = augroup('HighlightYank', {}),
      pattern = '*',
      callback = function()
         vim.highlight.on_yank({
            higroup = 'IncSearch',
            timeout = 40,
         })
      end,
   })
end
```

**Key Improvements**:
- `vim.schedule()`: Defers execution to avoid race conditions
- `pcall()`: Catches and silences any errors
- `priority = 150`: Higher than TreeSitter's default 100
- `timeout = 150`: Longer timeout for stability

#### Phase 2: Enhanced treesitter.lua
**Goal**: Add stability and error handling to TreeSitter config

**Add at top of config function**:
```lua
-- Force synchronous parsing for stability in 0.11+
if vim.fn.has('nvim-0.11') == 1 then
   vim.g._ts_force_sync_parsing = true
end
```

**Enhanced highlight configuration**:
```lua
highlight = {
   enable = true,
   
   -- Disable for problematic cases
   disable = function(lang, buf)
      -- Existing large file check
      local max_filesize = 1000 * 1024 -- 1MB
      local ok, stats = pcall(vim.loop.fs_stat, vim.api.nvim_buf_get_name(buf))
      if ok and stats and stats.size > max_filesize then
         return true
      end
      
      -- Additional checks for 0.11+
      if vim.fn.has('nvim-0.11') == 1 then
         -- Disable for buffers with too many lines
         local line_count = vim.api.nvim_buf_line_count(buf)
         if line_count > 10000 then
            return true
         end
         
         -- Disable for specific problematic filetypes if needed
         local ft = vim.bo[buf].filetype
         if ft == 'log' or ft == 'txt' then
            return true
         end
      end
      
      return false
   end,
   
   -- Reduce conflicts with other highlighting
   additional_vim_regex_highlighting = { "markdown" },
},

-- Add incremental selection safety
incremental_selection = {
   enable = true,
   disable = function(lang, buf)
      -- Same disable logic as highlight
      return vim.fn.has('nvim-0.11') == 1 and vim.api.nvim_buf_line_count(buf) > 5000
   end,
},
```

#### Phase 3: Enhanced statusline.lua 
**Goal**: Prevent redraw operations from triggering TreeSitter errors

**Find line 473** (in timer function):
```lua
vim.cmd('redrawstatus')
```

**Replace with**:
```lua
-- Safer redraw with error protection
vim.schedule(function()
   pcall(vim.api.nvim_command, 'redrawstatus')
end)
```

**Additional Safety** (wrap entire timer callback):
```lua
timer:start(5000, 5000, vim.schedule_wrap(function()
   pcall(function()
      if vim.api.nvim_get_mode().mode == 'n' then
         git_cache.last_update = 0
         -- Protected redraw
         vim.schedule(function()
            pcall(vim.api.nvim_command, 'redrawstatus')
         end)
      end
   end)
end))
```

#### Phase 4: Optional Global Error Handler
**Goal**: Catch any remaining TreeSitter errors gracefully

**Add to init.lua or autocmds.lua**:
```lua
-- Global TreeSitter error handler (optional)
if vim.fn.has('nvim-0.11') == 1 then
   vim.api.nvim_create_autocmd('User', {
      pattern = 'TreesitterError',
      callback = function(args)
         -- Log instead of showing error
         vim.notify('TreeSitter error suppressed: ' .. vim.inspect(args.data), vim.log.levels.DEBUG)
      end,
   })
end
```

## 🧪 Testing Strategy

### Test Cases
1. **Basic Functionality**
   ```bash
   # Test highlight on yank
   nvim test.txt
   # Try yanking text (yy) - should see highlight
   ```

2. **Error Log Monitoring**
   ```bash
   # Clear log and monitor
   > ~/.local/state/nvim/log
   tail -f ~/.local/state/nvim/log
   # Use nvim normally, check for new errors
   ```

3. **Version Compatibility**
   ```bash
   # Test on different versions if available
   nvim --version
   # Verify highlight-on-yank works appropriately
   ```

4. **Stress Testing**
   ```bash
   # Open large files
   nvim large-file.log
   # Rapid yank operations
   # File editing with lots of changes
   ```

### Success Criteria
- ✅ No TreeSitter range errors in log
- ✅ Highlight-on-yank works on Neovim 0.11.3+
- ✅ Backwards compatibility maintained
- ✅ Statusline updates without errors
- ✅ Normal editing performance

## 🔄 Implementation Steps

### Preparation
1. **Backup current config**:
   ```bash
   cd ~/.Dotfiles
   git add -A
   git commit -m "Backup before TreeSitter fix"
   ```

2. **Clear error log**:
   ```bash
   > ~/.local/state/nvim/log
   ```

### Implementation Order
1. **Start with TreeSitter stability** (safest first)
   - Modify `treesitter.lua`
   - Test basic functionality
   
2. **Add statusline safety**
   - Modify `statusline.lua`
   - Monitor for redraw-related errors
   
3. **Enable protected highlight-on-yank**
   - Modify `autocmds.lua`
   - Test yank highlighting
   
4. **Optional: Add global error handler**
   - Add to `autocmds.lua` or `init.lua`
   - Final safety net

### Verification After Each Step
```bash
# Check for new errors
tail -20 ~/.local/state/nvim/log

# Test basic nvim functionality
nvim --cmd 'echo "Step completed"' +q
```

## 🚨 Rollback Plan

If any step causes issues:

### Quick Rollback
```bash
cd ~/.Dotfiles
git reset --hard HEAD~1  # Undo last commit
```

### Selective Rollback
Revert specific files to working versions:
```bash
git checkout HEAD~1 -- nvim/.config/nvim/lua/DarO/autocmds.lua
git checkout HEAD~1 -- nvim/.config/nvim/lua/DarO/treesitter.lua  
git checkout HEAD~1 -- nvim/.config/nvim/lua/DarO/statusline.lua
```

### Emergency Disable
If TreeSitter becomes completely broken:
```lua
-- Add to treesitter.lua temporarily
highlight = { enable = false },
```

## 📚 Research References

### GitHub Issues Consulted
- [neovim/neovim#12861](https://github.com/neovim/neovim/issues/12861) - TreeSitter end_col range errors
- [neovim/neovim#30610](https://github.com/neovim/neovim/issues/30610) - TreeSitter Invalid end_row on CompleteChanged
- [neovim/neovim#29550](https://github.com/neovim/neovim/issues/29550) - TreeSitter Invalid end_col range error
- [nvim-treesitter/nvim-treesitter#651](https://github.com/nvim-treesitter/nvim-treesitter/issues/651) - End_col value outside range when editing

### Key Findings
1. TreeSitter highlighting became asynchronous in 0.11
2. Default TreeSitter priority is 100
3. `vim.g._ts_force_sync_parsing = true` forces synchronous mode
4. Priority can be set higher than 100 to override TreeSitter
5. vim.schedule() defers operations to avoid race conditions

## 💡 Additional Optimizations

### Performance Improvements
- Consider disabling TreeSitter for very large files
- Use incremental parsing where possible
- Monitor memory usage with large files

### Alternative Highlight Groups
If 'IncSearch' conflicts, try:
- `Visual` - Standard visual selection highlight  
- `Search` - Search match highlight
- `CurSearch` - Current search highlight (if available)

### Future Monitoring
- Watch for Neovim 0.12 changes
- Monitor TreeSitter plugin updates
- Keep eye on performance with large codebases

---

## 📝 Implementation Notes

**When you're ready to implement:**
1. Read through this entire document
2. Backup your current configuration
3. Follow the implementation steps in order
4. Test thoroughly at each step
5. Use rollback plan if needed

**Estimated time**: 30-60 minutes for full implementation and testing

**Risk level**: Low (all changes are protected with error handling)

---

## 🔍 Latest Investigation Results (August 20, 2025)

### Key Discovery: TextYankPost Was Not The Root Cause

**Initial Hypothesis**: TextYankPost autocmd was causing TreeSitter conflicts
**Reality Discovered**: TextYankPost was just one trigger among many

### Investigation Timeline

1. **Initial Fix Applied**: Disabled TextYankPost for Neovim 0.11+ 
   - Result: ✅ No more highlight-on-yank specific errors
   - Reality: ❌ TreeSitter errors continued appearing

2. **Continued Error Analysis**: Examined ongoing log errors
   - Found errors still occurring from multiple sources
   - Confirmed TreeSitter range calculation is fundamentally broken

3. **Trigger Confirmation Testing**: User tested with buffer modifications
   - `dd` (delete line) → Immediate TreeSitter range errors
   - Visual selection + delete → Same TreeSitter errors  
   - Any buffer editing operation → TreeSitter recalculation fails

### Current Error Patterns Identified

**From Latest Log Analysis (2025-08-20T08:19:27.927)**:
```
ERR 2025-08-20T08:19:27.927 nvim.2995607.0 decor_provider_error:36: Error in decoration provider "line" (ns=nvim.treesitter.highlighter):
Error executing lua: ...al/share/nvim/runtime/lua/vim/treesitter/highlighter.lua:370: Invalid 'end_col': out of range
```

**Confirmed Triggers**:
1. **Buffer Editing Operations**: `dd`, visual delete, text insertion/deletion
2. **Statusline Operations**: `vim.cmd('redrawstatus')` at lines 220, 473
3. **LSP Semantic Tokens**: Interaction with TreeSitter highlighting
4. **General Buffer Changes**: Any modification that requires syntax re-highlighting

### Root Cause Confirmed

**The Real Problem**: Neovim 0.11.3 TreeSitter highlighter has a fundamental range calculation bug that occurs when:
- Buffer structure changes (lines added/removed)
- Text blocks are modified or deleted
- TreeSitter attempts to recalculate syntax highlighting ranges
- Calculated ranges exceed actual buffer boundaries

### Updated Status Assessment

- ✅ **TextYankPost Fix**: Successfully prevents highlight-on-yank errors
- ❌ **Core TreeSitter Issue**: Range calculation bug remains unaddressed  
- ❌ **Editing Operations**: Still trigger TreeSitter range errors
- ⚠️ **Partial Solution**: Reduced error frequency, but not eliminated

### Implications for Future Fix

The comprehensive fix plan remains valid but is **more critical than initially thought** because:
1. **Multiple Active Triggers**: Not just highlight-on-yank, but all buffer editing
2. **Core Neovim Bug**: This is a widespread issue affecting many 0.11.3 users
3. **Daily Impact**: Every `dd`, visual delete, or text edit triggers errors
4. **Log Pollution**: Continuous error generation affects debugging other issues

### Recommended Immediate Workarounds

While waiting for comprehensive fix implementation:

1. **Disable TreeSitter for problematic filetypes**:
   ```lua
   highlight = {
     disable = { "markdown", "text", "lua" } -- Add problematic types
   }
   ```

2. **Force synchronous TreeSitter parsing**:
   ```lua
   vim.g._ts_force_sync_parsing = true
   ```

3. **Reduce TreeSitter scope**:
   ```lua
   highlight = {
     enable = true,
     disable = function(lang, buf)
       return vim.api.nvim_buf_line_count(buf) > 1000
     end,
   }
   ```

### Conversation Context Preservation

This investigation was conducted through multiple testing sessions where:
- User confirmed TextYankPost disabling didn't eliminate errors
- Real-time testing with `dd` and visual operations confirmed triggers
- Log analysis revealed multiple error sources beyond initial scope
- Root cause traced to fundamental TreeSitter range calculation in Neovim 0.11.3

**Next Steps**: The comprehensive fix plan documented above becomes more urgent given the scope of daily editing operations affected.

---

*This document preserves all research and provides a complete roadmap for fixing TreeSitter/highlight-on-yank issues while maintaining backwards compatibility.*

*Updated with latest investigation findings confirming the scope and urgency of the TreeSitter range calculation bug in Neovim 0.11.3.*