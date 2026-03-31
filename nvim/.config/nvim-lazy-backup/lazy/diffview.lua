-- ~/.config/nvim/lua/DarO/lazy/diffview.lua
return {
   "sindrets/diffview.nvim",
   dependencies = {
      "nvim-lua/plenary.nvim",
      "nvim-tree/nvim-web-devicons"
   },
   keys = {
      { "<leader>dv", ":DiffviewOpen<CR>",          desc = "Open Diffview" },
      { "<leader>q",  ":DiffviewClose<CR>",         desc = "Close Diffview" },
      { "<leader>dh", ":DiffviewFileHistory %<CR>", desc = "Show file history in Diffview" }
   },
   cmd = { "DiffviewOpen", "DiffviewClose", "DiffviewFileHistory", "DiffviewToggleFiles", "DiffviewFocusFiles" },
   config = function()
      require("diffview").setup({
         enhanced_diff_hl = true,
      })
   end,
}
