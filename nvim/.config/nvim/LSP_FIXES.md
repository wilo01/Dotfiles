# LSP Configuration Fixes Applied

## Issues Resolved

### 1. Vue Language Server Timeout
**Problem**: `[LSP][vue_ls] timeout` - vue_ls was timing out during initialization

**Root Cause**: 
- Missing or incorrect `init_options` configuration
- No proper TypeScript SDK path detection
- Insufficient timeout values

**Solution**:
- Added robust `before_init` function to detect TypeScript SDK path
- Increased timeout to 10 seconds (`timeout_ms = 10000`)
- Enhanced error handling with detailed error messages
- Added memory limit configuration (`maxOldSpaceSize = 4096`)

### 2. ESLint Configuration Detection
**Problem**: `eslint: -32603: Request textDocument/diagnostic failed with message: No ESLint configuration found`

**Root Cause**:
- No ESLint configuration file in the project
- ESLint server trying to lint files without proper configuration

**Solution**:
- Created `.eslintrc.json` with Vue 3 + TypeScript configuration
- Installed required packages: `eslint`, `@typescript-eslint/parser`, `@typescript-eslint/eslint-plugin`, `eslint-plugin-vue`
- Enhanced ESLint server config to check for configuration files before starting
- Added graceful fallback when no configuration is found

### 3. TypeScript Server Conflicts
**Problem**: `Could not find 'vtsls' lsp client, required by 'vue_ls'` and conflicts between ts_ls and vtsls

**Root Cause**:
- Multiple TypeScript servers (ts_ls, vtsls) conflicting
- vue_ls dependency on vtsls not properly configured

**Solution**:
- Removed `ts_ls` from configuration to eliminate conflicts
- Made `vtsls` the primary TypeScript server
- Properly configured vue_ls to work with vtsls
- Added server capability detection and conflict resolution

## Configuration Enhancements

### Enhanced Error Handling
- Added `safe_require()` function for module loading
- Added `safe_setup()` wrapper for server configuration
- Enhanced error messages with actionable information

### Server-Specific Improvements

#### VTSLS (Primary TypeScript Server)
- Enhanced settings for better performance
- Inlay hints configuration
- Auto-import preferences
- Workspace TypeScript SDK detection

#### Vue Language Server
- Proper TypeScript SDK path detection
- Memory limit configuration to prevent timeouts
- Conflict resolution with other servers
- Enhanced error reporting

#### ESLint
- Configuration file detection before starting
- Graceful handling of missing configuration
- Support for both legacy and flat config formats
- Project-specific rules for Vue + TypeScript

### Health Check System
- Created `lsp_health.lua` module for diagnostics
- Added `:LspHealthCheck` command
- Automated server status checking
- Project configuration validation

## Files Modified

1. `/home/dariuszw/.Dotfiles/nvim/.config/nvim/lua/DarO/lazy/lsp.lua` - Main LSP configuration
2. `/home/dariuszw/.Dotfiles/nvim/.config/nvim/lua/DarO/lsp_health.lua` - Health check module
3. `/home/dariuszw/.Dotfiles/nvim/.config/nvim/lua/DarO/init.lua` - Added health check loading
4. `/home/dariuszw/Dev/Private/saas/frontend/.eslintrc.json` - ESLint configuration for project

## Diagnostic Commands

### For Users
- `:LspHealthCheck` - Comprehensive LSP status check
- `:LspInfo` - View active LSP clients
- `:LspRestart` - Restart all LSP servers
- `:Mason` - View/install LSP servers

### For Debugging
- `:lua vim.lsp.set_log_level('DEBUG')` - Enable detailed logging
- `:checkhealth lsp` - Neovim's built-in LSP health check
- `tail -f ~/.local/state/nvim/lsp.log` - Monitor LSP logs in real-time

## Performance Optimizations

1. **Lazy Loading**: All LSP servers are properly lazy-loaded
2. **Conflict Resolution**: Eliminated server conflicts that caused crashes
3. **Memory Management**: Increased memory limits for Vue LS to prevent timeouts
4. **Timeout Configuration**: Appropriate timeouts for different server types

## Testing

Use the included test script to validate configuration:
```bash
/home/dariuszw/test_lsp.sh
```

## Future Maintenance

1. **Regular Updates**: Keep Mason packages updated with `:MasonUpdate`
2. **Configuration Validation**: Run `:LspHealthCheck` when encountering issues
3. **Log Monitoring**: Check LSP logs when experiencing problems
4. **Version Compatibility**: Update configurations when upgrading Neovim

## Expected Behavior After Fixes

1. **Vue Files**: Should have full TypeScript support, syntax highlighting, and completion
2. **TypeScript/JavaScript**: Handled by vtsls with enhanced features
3. **ESLint**: Proper linting and code actions in configured projects
4. **No Timeouts**: Servers should start reliably without timeout errors
5. **Clean Logs**: Minimal error messages in LSP logs

## Notes

- ESLint warnings about `registerCapability` are harmless and expected
- Vue LS may show warnings about TypeScript SDK if not in a TypeScript project
- First startup may be slower due to server initialization
- All servers respect the configured root markers for proper workspace detection