-- Optimized Neovim Options
-- Using built-in vim.opt for better performance

local opt = vim.opt
local g = vim.g

-- Leader key (must be set before plugins)
g.mapleader = " "
g.maplocalleader = " "

-- Performance optimizations
opt.updatetime = 100          -- Faster completion and diagnostics
opt.timeoutlen = 300          -- Faster key sequence completion
opt.redrawtime = 1500         -- Time for redrawing during searches
opt.lazyredraw = false        -- Don't redraw during macros (can cause issues with modern plugins)

-- Display
opt.number = true             -- Show line numbers
opt.relativenumber = true     -- Relative line numbers
opt.signcolumn = "yes:1"      -- Always show sign column (prevent shifting)
opt.colorcolumn = ""          -- No color column by default
opt.wrap = false              -- Don't wrap lines
opt.scrolloff = 8             -- Keep 8 lines visible above/below cursor
opt.sidescrolloff = 8         -- Keep 8 columns visible left/right of cursor
opt.cursorline = true         -- Highlight current line
opt.cursorcolumn = true       -- Highlight current column
opt.laststatus = 3            -- Global statusline
opt.cmdheight = 1             -- Command line height
opt.showcmd = false           -- Don't show command in status line
opt.ruler = false             -- Don't show cursor position (statusline handles this)

-- Indentation
opt.tabstop = 3               -- Tab width
opt.softtabstop = 3           -- Soft tab width
opt.shiftwidth = 3            -- Indent width
opt.expandtab = true          -- Use spaces instead of tabs
opt.smartindent = true        -- Smart indentation
opt.autoindent = true         -- Copy indent from current line

-- Search
opt.ignorecase = true         -- Ignore case in search
opt.smartcase = true          -- Override ignorecase if search contains capitals
opt.hlsearch = false          -- Don't highlight search results by default
opt.incsearch = true          -- Show search matches as you type
opt.magic = true              -- Use magic patterns (default)

-- Files
opt.backup = false            -- Don't create backup files
opt.writebackup = false       -- Don't create backup before overwriting
opt.swapfile = false          -- Don't create swap files
opt.undofile = true           -- Persistent undo
opt.undodir = vim.fn.stdpath("data") .. "/undo"  -- Undo directory
opt.undolevels = 10000        -- Maximum undo levels

-- Behavior
opt.hidden = true             -- Allow hidden buffers
opt.splitbelow = true         -- Horizontal splits below
opt.splitright = true         -- Vertical splits to the right
opt.splitkeep = "screen"      -- Keep screen position when splitting (Neovim 0.9+)
opt.mouse = "a"               -- Enable mouse support
opt.clipboard = "unnamedplus" -- Use system clipboard
opt.completeopt = "menu,menuone,noselect"  -- Better completion experience
opt.wildmode = "longest:full,full"          -- Command line completion mode
opt.pumheight = 10            -- Maximum items in popup menu
opt.pumblend = 10             -- Popup menu transparency

-- Visual
opt.termguicolors = true      -- True color support
opt.list = true               -- Show whitespace characters
opt.listchars = {
   tab = ">-",
   trail = "~",
   extends = ">",
   precedes = "<",
   space = "·",
   nbsp = "␣",
}
opt.fillchars = {
   fold = " ",
   foldopen = "▾",
   foldclose = "▸",
   foldsep = " ",
   diff = "╱",
   eob = " ",
}

-- Folding (using built-in expr folding)
opt.foldmethod = "expr"
opt.foldexpr = ""             -- Will be set by treesitter when available
opt.foldlevel = 99            -- Start with all folds open
opt.foldenable = true

-- Diff
opt.diffopt:append("linematch:60")  -- Better diff algorithm (Neovim 0.9+)

-- Spelling
opt.spelllang = "en_gb"
opt.spell = false             -- Disable by default, enable per filetype

-- Session options
opt.sessionoptions = "blank,buffers,curdir,folds,help,tabpages,winsize,winpos,terminal,localoptions"

-- Netrw (built-in file explorer)
g.netrw_browse_split = 0
g.netrw_banner = 0
g.netrw_winsize = 25
g.netrw_liststyle = 3
g.netrw_altv = 1

-- Disable built-in plugins for faster startup
local disabled_built_ins = {
   "gzip",
   "zip",
   "zipPlugin",
   "tar",
   "tarPlugin",
   "getscript",
   "getscriptPlugin",
   "vimball",
   "vimballPlugin",
   "2html_plugin",
   "logipat",
   "rrhelper",
   "spellfile_plugin",
   "matchit",
   "matchparen",
   "netrwPlugin",  -- Disable if using a file explorer plugin
}

for _, plugin in pairs(disabled_built_ins) do
   g["loaded_" .. plugin] = 1
end

-- Python provider (disable if not needed)
g.loaded_python3_provider = 0
g.loaded_ruby_provider = 0
g.loaded_node_provider = 0
g.loaded_perl_provider = 0

-- Create undo directory if it doesn't exist
local undo_dir = vim.fn.stdpath("data") .. "/undo"
if vim.fn.isdirectory(undo_dir) == 0 then
   vim.fn.mkdir(undo_dir, "p", 0700)
end