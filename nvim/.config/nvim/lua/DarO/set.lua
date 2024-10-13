vim.opt.nu = true
vim.opt.relativenumber = true

vim.opt.tabstop = 3
vim.opt.softtabstop = 3
vim.opt.shiftwidth = 3
vim.opt.expandtab = true
vim.o.list = true
vim.o.listchars = "tab:>-,trail:~,extends:>,precedes:<,space:·"

vim.opt.smartindent = true
vim.opt.ignorecase = true

vim.opt.wrap = false

vim.opt.swapfile = false
vim.opt.backup = false
vim.opt.undodir = os.getenv("HOME") .. "/.vim/undodir"
vim.opt.undofile = true
-- vim.opt.undofile = false

vim.opt.hlsearch = false
vim.opt.incsearch = true

vim.opt.termguicolors = true

vim.opt.scrolloff = 8
vim.opt.signcolumn = "yes"
vim.opt.isfname:append("@-@")

vim.opt.updatetime = 50

-- Wrapping
vim.o.cursorline = true
vim.o.cursorcolumn = true
-- vim.o.statuscolumn = "%s %l %r"

-- Netrw
vim.g.netrw_browse_split = 0
-- vim.g.netrw_banner = 0
vim.g.netrw_winsize = 25
vim.g.netrw_browsex_viewer = "xdg-open"
vim.g.netrw_special_syntax = 3
vim.g.netrw_liststyle = 3
vim.g.netrw_altv = 1
vim.g.netrw_keepdir = 0
vim.g.netrw_liststyle = 3

