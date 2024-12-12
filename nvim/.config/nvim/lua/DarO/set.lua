-- Line Numbers
vim.opt.nu = true
vim.opt.relativenumber = true

-- Tabs and Indentation
vim.opt.tabstop = 3
vim.opt.softtabstop = 3
vim.opt.shiftwidth = 3
vim.opt.expandtab = true
vim.opt.smartindent = true

-- Whitespace and Formatting
vim.o.list = true
vim.o.listchars = "tab:>-,trail:~,extends:>,precedes:<,space:·"

-- Search Settings
vim.opt.ignorecase = true
vim.opt.hlsearch = false
vim.opt.incsearch = true

-- Spelling
vim.opt.spelllang = 'en_gb'
vim.opt.spell = true
vim.cmd("set spell syntax=off")
vim.api.nvim_create_user_command('Spellcolor', function()
   vim.cmd([[highlight SpellBad cterm=underline, bold ctermfg=Red guifg=#ff0000 guibg=NONE gui=underline, bold]])
end, {})

-- File Handling
vim.opt.swapfile = false
vim.opt.backup = false
vim.opt.undodir = os.getenv("HOME") .. "/.vim/undodir"
vim.opt.undofile = true

-- Visual Settings
vim.o.statuscolumn = ""
vim.opt.termguicolors = true
vim.opt.wrap = false
vim.o.cursorline = true
vim.o.cursorcolumn = true
vim.opt.scrolloff = 8
vim.opt.signcolumn = "yes"
vim.opt.isfname:append("@-@")

-- Update Time
vim.opt.updatetime = 50

-- Netrw Configuration
vim.g.netrw_browse_split = 0
vim.g.netrw_winsize = 25
vim.g.netrw_browsex_viewer = "xdg-open"
vim.g.netrw_special_syntax = 3
vim.g.netrw_liststyle = 3
vim.g.netrw_altv = 1
vim.g.netrw_keepdir = 1

-- Other
vim.g.loaded_perl_provider = 0
