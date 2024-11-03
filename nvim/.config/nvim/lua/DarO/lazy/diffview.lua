-- ~/.config/nvim/lua/DarO/lazy/diffview.lua
return {
   "sindrets/diffview.nvim",
   dependencies = {
      "nvim-lua/plenary.nvim",
      "nvim-tree/nvim-web-devicons"
   },
   config = function()
      require("diffview").setup({
         enhanced_diff_hl = true,
      })
      vim.keymap.set("n", "<leader>dv", ":DiffviewOpen<CR>", { desc = "Open Diffview" })
      vim.keymap.set("n", "<leader>dc", ":DiffviewClose<CR>", { desc = "Close Diffview" })
      vim.keymap.set("n", "<leader>dh", ":DiffviewFileHistory %<CR>", { desc = "Show file history in Diffview" })
   end,
}
