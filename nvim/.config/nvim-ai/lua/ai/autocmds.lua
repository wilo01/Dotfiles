-- Autocommands
-- Organized and optimized for performance

local augroup = vim.api.nvim_create_augroup
local autocmd = vim.api.nvim_create_autocmd

-- Clear existing autocmds in our groups
local groups = {
   General = augroup("General", { clear = true }),
   FileType = augroup("FileType", { clear = true }),
   Windows = augroup("Windows", { clear = true }),
   Yank = augroup("YankHighlight", { clear = true }),
   Terminal = augroup("Terminal", { clear = true }),
   Performance = augroup("Performance", { clear = true }),
}

-- General autocommands
autocmd("TextYankPost", {
   group = groups.Yank,
   pattern = "*",
   callback = function()
      vim.highlight.on_yank({ timeout = 200 })
   end,
   desc = "Highlight yanked text",
})

autocmd("BufReadPost", {
   group = groups.General,
   pattern = "*",
   callback = function()
      local mark = vim.api.nvim_buf_get_mark(0, '"')
      local lcount = vim.api.nvim_buf_line_count(0)
      if mark[1] > 0 and mark[1] <= lcount then
         pcall(vim.api.nvim_win_set_cursor, 0, mark)
      end
   end,
   desc = "Restore cursor position",
})

autocmd("FileType", {
   group = groups.General,
   pattern = { "qf", "help", "man", "lspinfo", "spectre_panel", "startuptime", "checkhealth" },
   callback = function()
      vim.keymap.set("n", "q", "<cmd>close<CR>", { buffer = true })
      vim.opt_local.wrap = false
      vim.opt_local.spell = false
      vim.opt_local.buflisted = false
   end,
   desc = "Close with q and disable wrap/spell",
})

autocmd({ "FocusGained", "TermClose", "TermLeave" }, {
   group = groups.General,
   callback = function()
      if vim.o.buftype ~= "nofile" then
         vim.cmd("checktime")
      end
   end,
   desc = "Check if file changed outside of Neovim",
})

autocmd("VimResized", {
   group = groups.Windows,
   callback = function()
      local current_tab = vim.fn.tabpagenr()
      vim.cmd("tabdo wincmd =")
      vim.cmd("tabnext " .. current_tab)
   end,
   desc = "Resize windows on terminal resize",
})

autocmd("BufWritePre", {
   group = groups.General,
   pattern = "*",
   callback = function()
      -- Create directory if it doesn't exist
      local dir = vim.fn.expand("<afile>:p:h")
      if vim.fn.isdirectory(dir) == 0 then
         vim.fn.mkdir(dir, "p")
      end
      -- Remove trailing whitespace (optional, remove if not wanted)
      vim.cmd([[%s/\s\+$//e]])
   end,
   desc = "Create directories and trim whitespace on save",
})

-- Filetype specific settings
autocmd("FileType", {
   group = groups.FileType,
   pattern = { "gitcommit", "markdown", "txt" },
   callback = function()
      vim.opt_local.wrap = true
      vim.opt_local.spell = true
      vim.opt_local.textwidth = 80
   end,
   desc = "Enable wrap and spell for text files",
})

autocmd("FileType", {
   group = groups.FileType,
   pattern = { "json", "jsonc" },
   callback = function()
      vim.opt_local.conceallevel = 0
      vim.opt_local.tabstop = 2
      vim.opt_local.shiftwidth = 2
   end,
   desc = "JSON specific settings",
})

autocmd("FileType", {
   group = groups.FileType,
   pattern = { "python" },
   callback = function()
      vim.opt_local.tabstop = 4
      vim.opt_local.shiftwidth = 4
      vim.opt_local.expandtab = true
   end,
   desc = "Python specific settings",
})

autocmd("FileType", {
   group = groups.FileType,
   pattern = { "go" },
   callback = function()
      vim.opt_local.tabstop = 4
      vim.opt_local.shiftwidth = 4
      vim.opt_local.expandtab = false
   end,
   desc = "Go specific settings",
})

autocmd("FileType", {
   group = groups.FileType,
   pattern = { "javascript", "javascriptreact", "typescript", "typescriptreact", "vue", "css", "scss", "html" },
   callback = function()
      vim.opt_local.tabstop = 2
      vim.opt_local.shiftwidth = 2
   end,
   desc = "Web development settings",
})

autocmd("FileType", {
   group = groups.FileType,
   pattern = { "make", "makefile" },
   callback = function()
      vim.opt_local.expandtab = false
      vim.opt_local.tabstop = 4
      vim.opt_local.shiftwidth = 4
   end,
   desc = "Makefile settings",
})

-- Terminal settings
autocmd("TermOpen", {
   group = groups.Terminal,
   pattern = "*",
   callback = function()
      vim.opt_local.number = false
      vim.opt_local.relativenumber = false
      vim.opt_local.scrolloff = 0
      vim.opt_local.signcolumn = "no"
      vim.cmd("startinsert")
   end,
   desc = "Terminal settings",
})

-- Performance optimizations
autocmd({ "CursorHold", "CursorHoldI" }, {
   group = groups.Performance,
   pattern = "*",
   callback = function()
      -- Trigger CursorHold more frequently for better responsiveness
      vim.diagnostic.open_float(nil, { focus = false, scope = "cursor" })
   end,
   desc = "Show diagnostics on cursor hold",
})

-- Large file handling
autocmd("BufReadPre", {
   group = groups.Performance,
   pattern = "*",
   callback = function()
      local max_filesize = 100 * 1024 -- 100 KB
      local ok, stats = pcall(vim.loop.fs_stat, vim.api.nvim_buf_get_name(0))
      if ok and stats and stats.size > max_filesize then
         vim.b.large_file = true
         vim.cmd("syntax off")
         vim.opt_local.foldmethod = "manual"
         vim.opt_local.spell = false
         vim.opt_local.swapfile = false
         vim.opt_local.undofile = false
         vim.opt_local.relativenumber = false
         vim.schedule(function()
            vim.notify("Large file detected, disabling some features for performance", vim.log.levels.WARN)
         end)
      end
   end,
   desc = "Disable features for large files",
})

-- Auto-save (optional, uncomment if desired)
-- autocmd({ "InsertLeave", "TextChanged" }, {
--    group = groups.General,
--    pattern = "*",
--    callback = function()
--       if vim.bo.modified and not vim.bo.readonly and vim.fn.expand("%") ~= "" and vim.bo.buftype == "" then
--          vim.cmd("silent! write")
--       end
--    end,
--    desc = "Auto-save on insert leave and text change",
-- })

-- Disable automatic comment insertion
autocmd("BufEnter", {
   group = groups.General,
   pattern = "*",
   callback = function()
      vim.opt.formatoptions:remove({ "c", "r", "o" })
   end,
   desc = "Disable automatic comment insertion",
})

-- Quickfix window settings
autocmd("FileType", {
   group = groups.FileType,
   pattern = "qf",
   callback = function()
      vim.opt_local.colorcolumn = ""
      vim.opt_local.signcolumn = "no"
      vim.keymap.set("n", "<CR>", "<CR><cmd>cclose<CR>", { buffer = true })
   end,
   desc = "Quickfix settings",
})