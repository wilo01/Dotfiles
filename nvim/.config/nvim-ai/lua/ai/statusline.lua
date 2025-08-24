-- Lightweight built-in statusline
-- Using pure Lua and vim.opt.statusline for maximum performance

local M = {}

-- Cache for expensive operations
local cache = {
   git_branch = "",
   git_status = "",
   last_git_update = 0,
}

-- Update interval for git info (milliseconds)
local GIT_UPDATE_INTERVAL = 5000

-- Mode map
local mode_map = {
   ["n"] = "NORMAL",
   ["no"] = "N-PENDING",
   ["v"] = "VISUAL",
   ["V"] = "V-LINE",
   [""] = "V-BLOCK",
   ["s"] = "SELECT",
   ["S"] = "S-LINE",
   [""] = "S-BLOCK",
   ["i"] = "INSERT",
   ["ic"] = "INSERT",
   ["R"] = "REPLACE",
   ["Rv"] = "V-REPLACE",
   ["c"] = "COMMAND",
   ["cv"] = "VIM-EX",
   ["ce"] = "EX",
   ["r"] = "PROMPT",
   ["rm"] = "MORE",
   ["r?"] = "CONFIRM",
   ["!"] = "SHELL",
   ["t"] = "TERMINAL",
}

-- Get current mode
local function get_mode()
   local mode = vim.api.nvim_get_mode().mode
   return mode_map[mode] or mode:upper()
end

-- Get git branch (cached)
local function get_git_branch()
   local now = vim.loop.now()
   if now - cache.last_git_update > GIT_UPDATE_INTERVAL then
      cache.last_git_update = now
      local handle = io.popen("git branch --show-current 2>/dev/null")
      if handle then
         local branch = handle:read("*l")
         handle:close()
         cache.git_branch = branch or ""
      end
   end
   return cache.git_branch
end

-- Get file icon (simple approach without dependencies)
local function get_file_icon()
   local extension = vim.fn.expand("%:e")
   local icons = {
      lua = "",
      py = "",
      js = "",
      ts = "",
      jsx = "",
      tsx = "",
      json = "",
      md = "",
      vim = "",
      sh = "",
      zsh = "",
      bash = "",
      c = "",
      cpp = "",
      rs = "",
      go = "",
      html = "",
      css = "",
      scss = "",
      yaml = "",
      yml = "",
      toml = "",
      xml = "",
      txt = "",
   }
   return icons[extension] or ""
end

-- Get diagnostic counts
local function get_diagnostics()
   local diagnostics = vim.diagnostic.get(0)
   local counts = { errors = 0, warnings = 0, info = 0, hints = 0 }

   for _, diagnostic in ipairs(diagnostics) do
      if diagnostic.severity == vim.diagnostic.severity.ERROR then
         counts.errors = counts.errors + 1
      elseif diagnostic.severity == vim.diagnostic.severity.WARN then
         counts.warnings = counts.warnings + 1
      elseif diagnostic.severity == vim.diagnostic.severity.INFO then
         counts.info = counts.info + 1
      elseif diagnostic.severity == vim.diagnostic.severity.HINT then
         counts.hints = counts.hints + 1
      end
   end

   local parts = {}
   if counts.errors > 0 then
      table.insert(parts, " " .. counts.errors)
   end
   if counts.warnings > 0 then
      table.insert(parts, " " .. counts.warnings)
   end
   if counts.info > 0 then
      table.insert(parts, " " .. counts.info)
   end
   if counts.hints > 0 then
      table.insert(parts, " " .. counts.hints)
   end

   return table.concat(parts, " ")
end

-- Get LSP status
local function get_lsp_status()
   local clients = vim.lsp.get_clients({ bufnr = 0 })
   if #clients == 0 then
      return ""
   end

   local client_names = {}
   for _, client in ipairs(clients) do
      table.insert(client_names, client.name)
   end

   return " " .. table.concat(client_names, ", ")
end

-- Get file format info
local function get_format_info()
   local format = vim.bo.fileformat
   local encoding = vim.bo.fileencoding ~= "" and vim.bo.fileencoding or vim.o.encoding
   return string.format("%s[%s]", format:upper(), encoding:upper())
end

-- Build the statusline
function M.build()
   local parts = {}

   -- Mode
   table.insert(parts, "%#StatusLineMode# " .. get_mode() .. " %*")

   -- Git branch
   local branch = get_git_branch()
   if branch ~= "" then
      table.insert(parts, "%#StatusLineGit#  " .. branch .. " %*")
   end

   -- File info
   table.insert(parts, "%#StatusLineFile# " .. get_file_icon() .. " %f%m%r %*")

   -- Diagnostics
   local diagnostics = get_diagnostics()
   if diagnostics ~= "" then
      table.insert(parts, "%#StatusLineDiagnostics# " .. diagnostics .. " %*")
   end

   -- Center separator
   table.insert(parts, "%=")

   -- LSP status
   local lsp = get_lsp_status()
   if lsp ~= "" then
      table.insert(parts, "%#StatusLineLSP#" .. lsp .. " %*")
   end

   -- File type
   table.insert(parts, "%#StatusLineFileType# %Y %*")

   -- Format info
   table.insert(parts, "%#StatusLineFormat# " .. get_format_info() .. " %*")

   -- Position
   table.insert(parts, "%#StatusLinePosition# %3l:%-2c  %3p%% %*")

   return table.concat(parts, " ")
end

-- Setup function
function M.setup()
   -- Define highlight groups (tokyonight compatible)
   local highlights = {
      StatusLineMode = { fg = "#1a1b26", bg = "#7aa2f7", bold = true },
      StatusLineGit = { fg = "#c0caf5", bg = "#292e42" },
      StatusLineFile = { fg = "#c0caf5", bg = "#1f2335" },
      StatusLineDiagnostics = { fg = "#f7768e", bg = "#1f2335" },
      StatusLineLSP = { fg = "#7dcfff", bg = "#1f2335" },
      StatusLineFileType = { fg = "#9ece6a", bg = "#1f2335" },
      StatusLineFormat = { fg = "#9ece6a", bg = "#1f2335" },
      StatusLinePosition = { fg = "#1a1b26", bg = "#9ece6a", bold = true },
   }

   for name, opts in pairs(highlights) do
      vim.api.nvim_set_hl(0, name, opts)
   end

   -- Set the statusline
   vim.opt.statusline = "%!v:lua.require('ai.statusline').build()"

   -- Update git info periodically
   vim.api.nvim_create_autocmd({ "BufEnter", "FocusGained", "BufWritePost" }, {
      callback = function()
         cache.last_git_update = 0 -- Force update
      end,
   })
end

return M