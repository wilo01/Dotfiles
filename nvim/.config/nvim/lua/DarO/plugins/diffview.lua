local gh = function(x) return 'https://github.com/' .. x end
local loaded = false

local diffview_cmds = { "DiffviewOpen", "DiffviewClose", "DiffviewFileHistory", "DiffviewToggleFiles", "DiffviewFocusFiles" }

local function ensure_loaded()
   if loaded then return end
   loaded = true
   for _, c in ipairs(diffview_cmds) do
      pcall(vim.api.nvim_del_user_command, c)
   end
   vim.pack.add({ gh('sindrets/diffview.nvim') })
   require("diffview").setup({
      enhanced_diff_hl = true,
   })
end

-- Stub commands
for _, cmd in ipairs(diffview_cmds) do
   vim.api.nvim_create_user_command(cmd, function(info)
      vim.api.nvim_del_user_command(cmd)
      ensure_loaded()
      vim.cmd({ cmd = cmd, args = { info.args }, bang = info.bang })
   end, { nargs = "*", bang = true })
end

-- Keymaps
vim.keymap.set("n", "<leader>dv", function()
   ensure_loaded()
   vim.cmd("DiffviewOpen")
end, { desc = "Open Diffview" })

vim.keymap.set("n", "<leader>q", function()
   ensure_loaded()
   vim.cmd("DiffviewClose")
end, { desc = "Close Diffview" })

vim.keymap.set("n", "<leader>dh", function()
   ensure_loaded()
   vim.cmd("DiffviewFileHistory %")
end, { desc = "Show file history in Diffview" })
