return {
   "folke/which-key.nvim",
   event = "VeryLazy",
   cond = false,
   -- cond = function() return vim.g.which_key_ignore end,
   opts = {},
   keys = {
      {
         "<leader>?",
         function()
            require("which-key").show({ global = false })
         end,
         desc = "Buffer Local Keymaps (which-key)",
      },
   },
}
