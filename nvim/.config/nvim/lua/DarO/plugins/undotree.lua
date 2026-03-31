local gh = function(x) return 'https://github.com/' .. x end
local loaded = false

local undo_cmds = { "UndotreeToggle", "UndotreeShow", "UndotreeHide", "UndotreeFocus" }

local function ensure_loaded()
   if loaded then return end
   loaded = true
   for _, c in ipairs(undo_cmds) do
      pcall(vim.api.nvim_del_user_command, c)
   end
   vim.pack.add({ gh('mbbill/undotree') })
end

for _, cmd in ipairs(undo_cmds) do
   vim.api.nvim_create_user_command(cmd, function(info)
      vim.api.nvim_del_user_command(cmd)
      ensure_loaded()
      vim.cmd({ cmd = cmd, args = { info.args }, bang = info.bang })
   end, { nargs = "*", bang = true })
end

vim.keymap.set("n", "<leader>u", function()
   ensure_loaded()
   vim.cmd.UndotreeToggle()
end, { desc = "Toggle UndoTree" })
