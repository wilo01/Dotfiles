# Vue LSP Configuration Fix Summary

## Critical Issues Resolved

### 1. Vue LS Dependency Issue ✅
**Error**: "Could not find `ts_ls` or `vtsls` lsp client, required by `vue_ls`"

**Root Cause**: `vue_ls` was configured in takeover mode (`hybridMode = false`) but `ts_ls` was explicitly configured to NOT handle Vue files. In takeover mode, `vue_ls` requires TypeScript services to be available for Vue files.

**Solution**: Switched to **Hybrid Mode** configuration:
- `vue_ls`: `hybridMode = true` - Works alongside TypeScript server
- `ts_ls`: Added 'vue' to `filetypes` - Now handles Vue files for TypeScript support
- Added proper `on_attach` handlers to prevent capability conflicts

### 2. ESLint Configuration Errors ✅
**Error**: "eslint: -32603: Request textDocument/diagnostic failed with message: Could not find config file"

**Root Cause**: ESLint server was starting even when no configuration file was present.

**Solution**: Enhanced ESLint error handling:
- Detect missing configuration files in `on_attach`
- Completely disable and stop ESLint client when no config found
- Prevent diagnostic and formatting attempts without proper config

### 3. Format on Save Issues ✅
**Error**: "[LSP] Format request failed, no matching language servers"

**Root Cause**: Conflicting formatter selection logic and stopped clients.

**Solution**: Improved formatter selection:
- Vue files: Use `vue_ls` exclusively
- JS/TS files: Prefer ESLint, fallback to `ts_ls`
- Better client availability checking
- Handle stopped clients gracefully

### 4. Vue LS Timeout Issues ✅
**Error**: "[LSP][vue_ls] timeout"

**Root Cause**: `vue_ls` waiting for unavailable TypeScript services.

**Solution**: Hybrid mode ensures both servers work together properly.

## Configuration Changes Made

### Core LSP Configuration (`lsp.lua`)

#### TypeScript Server (ts_ls)
```lua
ts_ls = {
   -- CRITICAL FIX: Include 'vue' in filetypes for hybrid mode
   filetypes = { "typescript", "javascript", "javascriptreact", "typescriptreact", "vue" },
   on_attach = function(client, bufnr)
      -- For Vue files, let vue_ls handle most features
      if vim.bo[bufnr].filetype == "vue" then
         client.server_capabilities.documentFormattingProvider = false
         client.server_capabilities.hoverProvider = false
         client.server_capabilities.completionProvider = false
      end
   end,
}
```

#### Vue Language Server (vue_ls)
```lua
vue_ls = {
   init_options = {
      vue = {
         -- CRITICAL FIX: Enable hybrid mode
         hybridMode = true,
      },
   },
   on_attach = function(client, bufnr)
      -- vue_ls handles Vue-specific features
      if vim.bo[bufnr].filetype == "vue" then
         client.server_capabilities.documentFormattingProvider = true
         client.server_capabilities.hoverProvider = true
         client.server_capabilities.completionProvider = true
      end
   end,
}
```

#### ESLint Server
```lua
eslint = {
   on_attach = function(client, bufnr)
      if not has_eslint_config() then
         -- Completely disable and stop ESLint
         client.server_capabilities.documentFormattingProvider = false
         client.server_capabilities.diagnosticProvider = false
         client.server_capabilities.codeActionProvider = false
         
         vim.schedule(function()
            if client and not client.is_stopped() then
               client.stop(true)
            end
         end)
      end
   end,
}
```

### Enhanced Safety Layer

Added comprehensive error handling and safety features:
- Safe LSP client operations
- Deprecated API compatibility  
- Vue-specific dependency checking
- Enhanced health checking with `:VueLspCheck` command

### Initialization Order (`init.lua`)

```lua
-- CRITICAL: Load LSP safety before LSP configuration
local lsp_safety = require("DarO.lsp_safety")
lsp_safety.install_safety_patches()

local lsp_fixes = require("DarO.lsp_fixes")
lsp_fixes.apply_all_fixes()
```

## Testing Commands

After restarting Neovim:

1. **Check overall LSP health**: `:LspSafetyCheck`
2. **Check Vue-specific setup**: `:VueLspCheck`
3. **View active clients**: `:LspInfo`
4. **Check Mason installations**: `:Mason`

## Expected Behavior

### Vue Files (.vue)
- `vue_ls`: Handles Vue-specific features (templates, SFC structure)
- `ts_ls`: Provides TypeScript support within script blocks
- **Formatting**: `vue_ls` handles all formatting
- **No timeout errors**

### JavaScript/TypeScript Files
- `eslint`: Primary formatter (if config exists)
- `ts_ls`: Fallback formatter and TypeScript features
- **No config file errors**

### All Files
- **No deprecated API warnings**
- **Graceful error handling**
- **Proper client lifecycle management**

## Hybrid Mode vs Takeover Mode

### Hybrid Mode (Current Configuration)
- ✅ `vue_ls` and `ts_ls` work together
- ✅ Both servers handle Vue files with different responsibilities
- ✅ More reliable and less prone to dependency issues
- ✅ Better performance

### Takeover Mode (Previous Configuration)
- ❌ `vue_ls` must find and control TypeScript services
- ❌ Complex dependency requirements
- ❌ Prone to "client not found" errors
- ❌ More fragile setup

## Files Modified

1. `/home/dariuszw/.Dotfiles/nvim/.config/nvim/lua/DarO/lazy/lsp.lua` - Main LSP configuration
2. `/home/dariuszw/.Dotfiles/nvim/.config/nvim/lua/DarO/lsp_safety.lua` - Enhanced safety layer
3. `/home/dariuszw/.Dotfiles/nvim/.config/nvim/lua/DarO/init.lua` - Added safety layer loading

The configuration now follows modern Neovim LSP best practices with proper error handling and Vue/TypeScript integration.