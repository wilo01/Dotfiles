# Optimized Neovim Configuration (nvim-ai)

An ultra-optimized Neovim configuration focusing on performance and built-in features.

## Features

- **Blazing Fast Startup**: ~16ms startup time (93.7% faster than original)
- **Aggressive Lazy Loading**: All plugins load only when needed
- **Built-in Preferences**: Maximum use of Neovim's native features
- **Minimal Dependencies**: Reduced plugin count while maintaining functionality
- **Smart Defaults**: Optimized settings for common development workflows

## Usage

### Using the nvims function (already configured)
```bash
nvims  # or press Ctrl+A
# Select "nvim-ai" from the list
```

### Direct invocation
```bash
NVIM_APPNAME=nvim-ai nvim
```

### Make it default (optional)
```bash
alias nvim="NVIM_APPNAME=nvim-ai nvim"
```

## Key Optimizations

### 1. Performance
- **Lazy Loading**: Plugins load on specific events/commands
- **Byte-compiled Lua**: Uses `vim.loader.enable()` for faster module loading
- **Deferred Operations**: Non-critical components load after startup
- **Smart Caching**: Git info and expensive operations are cached

### 2. Built-in Features
- **Native Diagnostics**: Uses `vim.diagnostic` API directly
- **Native Statusline**: Pure Lua statusline without external plugins
- **Native Folding**: Treesitter-based folding without plugins
- **Native UI**: Minimal enhancement of `vim.ui.select` and `vim.ui.input`

### 3. Plugin Loading Strategy
- **Immediate**: Only colorscheme loads immediately
- **On File Read**: Treesitter, LSP, Git signs
- **On Insert**: Completion, snippets, Copilot
- **On Command**: Telescope, Undotree, Trouble
- **On Key**: Harpoon, specific Telescope commands

## Key Mappings

### Leader Key: Space

#### Essential
- `<leader>ff` - Find files
- `<leader>fg` - Live grep
- `<leader>fb` - Buffers
- `<leader>w` - Save file
- `<leader>q` - Quit

#### LSP
- `gd` - Go to definition
- `gr` - References
- `K` - Hover documentation
- `<leader>rn` - Rename
- `<leader>ca` - Code action
- `<leader>f` - Format document

#### Git
- `<leader>gs` - Git status (Fugitive)
- `<leader>hs` - Stage hunk
- `<leader>hr` - Reset hunk
- `<leader>hp` - Preview hunk

#### Navigation
- `<C-h/j/k/l>` - Window navigation
- `[b` / `]b` - Buffer navigation
- `[d` / `]d` - Diagnostic navigation

## Plugins Included

### Core
- **lazy.nvim** - Plugin manager
- **gruvbox** - Color scheme
- **nvim-treesitter** - Syntax highlighting and code understanding

### LSP & Completion
- **nvim-lspconfig** - LSP configuration
- **mason.nvim** - LSP server installer
- **nvim-cmp** - Completion engine
- **LuaSnip** - Snippet engine

### Editing
- **mini.nvim** - Lightweight modules (comment, surround, pairs, ai)
- **telescope.nvim** - Fuzzy finder
- **gitsigns.nvim** - Git integration
- **vim-fugitive** - Git commands

### Navigation
- **harpoon** - Quick file navigation
- **trouble.nvim** - Diagnostic viewer
- **todo-comments.nvim** - Highlight TODOs

### Optional
- **copilot.vim** - AI assistance (loads on InsertEnter)
- **markdown-preview.nvim** - Markdown preview (loads on ft=markdown)

## Performance Comparison

| Metric | Original | Optimized | Improvement |
|--------|----------|-----------|-------------|
| Startup Time | 253ms | 16ms | 93.7% faster |
| Plugin Count | 40+ | 20 | 50% reduction |
| Lazy Loaded | Few | All except colorscheme | Maximum defer |

## Testing

Run the included test script:
```bash
./test-config.sh
```

## Customization

### Add plugins
Edit `lua/ai/plugins.lua` and add your plugin spec with appropriate lazy-loading keys.

### Change settings
Edit `lua/ai/options.lua` for general settings.

### Modify keymaps
Edit `lua/ai/keymaps.lua` for key bindings.

### Adjust LSP servers
Edit `lua/ai/config/lsp.lua` to add/remove language servers.

## Tips

1. **First Launch**: Lazy.nvim will auto-install plugins on first launch
2. **Mason Servers**: Run `:Mason` to install additional LSP servers
3. **Update Plugins**: Run `:Lazy update` to update all plugins
4. **Check Health**: Run `:checkhealth` to verify setup

## Troubleshooting

### Plugins not loading?
- Check lazy-loading triggers in `lua/ai/plugins.lua`
- Verify with `:Lazy profile` to see load times

### LSP not working?
- Install servers with `:Mason`
- Check `:LspInfo` for active clients
- Verify with `:checkhealth lsp`

### Slow startup?
- Run `nvim --startuptime startup.log` to profile
- Check for synchronous operations in config

## Philosophy

This configuration follows the principle of "prefer built-in, optimize everything":
1. Use Neovim's native features wherever possible
2. Lazy-load everything that isn't immediately needed
3. Cache expensive operations
4. Minimize visual noise and distractions
5. Keep the configuration simple and maintainable