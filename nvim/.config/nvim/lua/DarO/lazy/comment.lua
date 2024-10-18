return {
   "b3nj5m1n/kommentary",
   config = function()
      vim.g.kommentary_create_default_mappings = false

      vim.api.nvim_set_keymap("n", "<C-_>", "<Plug>kommentary_line_default",
         { desc = "Toggle comment for current line" })
      vim.api.nvim_set_keymap("x", "<C-_>", "<Plug>kommentary_visual_default",
         { desc = "Toggle comment for visual selection" })


      require('kommentary.config').configure_language("default", {
         prefer_single_line_comments = true,
      })
   end
}
