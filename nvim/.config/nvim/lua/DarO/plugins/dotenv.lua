local gh = function(x) return 'https://github.com/' .. x end
local loaded = false

local dotenv_cmds = { "Dotenv", "DotenvGet" }

local function ensure_loaded()
   if loaded then return end
   loaded = true
   for _, c in ipairs(dotenv_cmds) do
      pcall(vim.api.nvim_del_user_command, c)
   end
   vim.pack.add({ gh('ellisonleao/dotenv.nvim') })
   require("dotenv").setup({
      enable_on_load = true,
      verbose = false,
      file_name = vim.fn.expand("~/.env"),
   })
end

for _, cmd in ipairs(dotenv_cmds) do
   vim.api.nvim_create_user_command(cmd, function(info)
      vim.api.nvim_del_user_command(cmd)
      ensure_loaded()
      vim.cmd({ cmd = cmd, args = { info.args }, bang = info.bang })
   end, { nargs = "*", bang = true })
end
