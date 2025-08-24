# Features Added to nvim-ai from Main Config

## Enhanced Keymaps (lua/ai/keymaps.lua)
### Navigation & Git
- `J` / `K` - Navigate git hunks with centering
- `gb` - Go back and center
- `<C-j>` / `<C-k>` - Quickfix navigation
- `<leader>j` / `<leader>k` - Location list navigation

### File Operations
- `<leader>fp` - Copy file path to clipboard (pwd)
- `<leader>rp` - Copy relative path to clipboard
- `<leader>ov` - Open current file in VSCode
- `<leader>x` - Make file executable
- `<leader>v` - Open Netrw
- `<C-f>` - Tmux sessionizer

### Clipboard & Text Operations
- `<leader><leader>` - Select all and yank to clipboard
- `<leader>z` - Delete without yanking
- `<leader>zy` - Delete & yank to clipboard
- `x` in visual mode - Replace with yanked text, keep yanked
- Additional bracket/quote wrapping shortcuts

### Git Integration
- `<leader>va` - Preview git hunk inline
- `<leader>vs` - Diff current buffer
- `<leader>bl` - Blame current file
- `<leader>vt` - Toggle deleted lines
- `<leader>vb` - Blame current line
- `<leader>rg` - Reset hunk
- `<leader>sh` - Stage hunk
- `<leader>sf` - Stage entire file
- `<leader>uf` - Unstage entire file
- `<leader>gr` - Open in TDS GitLab
- `<leader>og` - Open in GitHub/GitLab

### Console Snippets
- `<leader>cl` - Insert console.warn with object
- `<leader>cn` - Insert console.warn without object
- `<leader>ck` - Insert if statement
- `<leader>cj` - Insert ca_log_pak.log_warning
- `<leader>ch` - Insert DBMS_OUTPUT

### Utilities
- `<leader>c/` - Remove comment from current line
- `<leader>t` - Toggle CSV formatting
- `<leader>,` - Toggle true/false
- Spelling keymaps (z=, zg, zG, etc.)

## Enhanced Plugins (lua/ai/plugins.lua)

### Harpoon Improvements
- Number shortcuts (`<leader>1-0`) for quick file access
- Telescope integration (`<C-t>`)
- Previous/Next navigation (`<C-h>` / `<C-l>`)
- Auto-add functionality with `HarpoonAutoAdd` command
- File notifications when adding to Harpoon

### Gitsigns Enhancements
- Current line blame enabled by default
- Auto-remove trailing whitespace on changed lines (BufWritePre)
- Enhanced blame formatter with author, date, and summary
- Better sign characters (│, ‾, ~, ┆)

### New Plugins Added
- `sindrets/diffview.nvim` - Better git diff viewer
- `folke/zen-mode.nvim` - Distraction-free writing
- `m4xshen/hardtime.nvim` - Break bad habits

### Plugin Configuration Updates
- Snacks.nvim with zen mode, scratch buffers, git browse
- Enhanced Fugitive keymaps
- Color highlighting always active

## Configuration Optimizations
- All keymaps preserved from main config
- Git remote detection for GitHub/GitLab
- CSV prettification with safe formatting
- Console snippet generation for multiple languages
- Comment removal with filetype detection

All functionality from your main Neovim configuration has been successfully ported to nvim-ai while maintaining the optimized structure!