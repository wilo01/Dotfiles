local defer = require("DarO.defer")

local ensure = defer.on_cmd({
   repo = "sindrets/diffview.nvim",
   cmds = { "DiffviewOpen", "DiffviewClose", "DiffviewFileHistory", "DiffviewToggleFiles", "DiffviewFocusFiles" },
   setup = function()
      require("diffview").setup({
         enhanced_diff_hl = true,
      })
   end,
})

vim.keymap.set("n", "<leader>dv", function()
   if not ensure() then return end
   vim.cmd("DiffviewOpen")
end, { desc = "Open Diffview" })

vim.keymap.set("n", "<leader>q", function()
   if not ensure() then return end
   vim.cmd("DiffviewClose")
end, { desc = "Close Diffview" })

vim.keymap.set("n", "<leader>dh", function()
   if not ensure() then return end
   vim.cmd("DiffviewFileHistory %")
end, { desc = "Show file history in Diffview" })
