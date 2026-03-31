local gh = function(x) return 'https://github.com/' .. x end
local loaded = false

local docker_cmds = { "Lazydocker", "LazyDocker" }

local function ensure_loaded()
   if loaded then return end
   loaded = true
   for _, c in ipairs(docker_cmds) do
      pcall(vim.api.nvim_del_user_command, c)
   end
   require("toggleterm").setup({})
   vim.pack.add({ gh('mgierada/lazydocker.nvim') })
   require("lazydocker").setup({
      border = "curved",
   })
end

-- Stub commands
for _, cmd in ipairs(docker_cmds) do
   vim.api.nvim_create_user_command(cmd, function(info)
      vim.api.nvim_del_user_command(cmd)
      ensure_loaded()
      vim.cmd({ cmd = cmd, args = { info.args }, bang = info.bang })
   end, { nargs = "*", bang = true })
end

vim.keymap.set("n", "<leader>ld", function()
   ensure_loaded()
   require("lazydocker").open()
end, { desc = "Open Lazydocker floating window" })
