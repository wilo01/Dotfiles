return {
   "tpope/vim-fugitive",
   config = function()
      vim.keymap.set("n", "<leader>gs", vim.cmd.Git, { desc = "Open Git status" })

      local TheDarOFugitive = vim.api.nvim_create_augroup("TheDarOFugitive", {})

      local autocmd = vim.api.nvim_create_autocmd
      autocmd("BufWinEnter", {
         group = TheDarOFugitive,
         pattern = "*",
         callback = function()
            if vim.bo.ft ~= "fugitive" then
               return
            end
         end,
      })

      vim.keymap.set("n", "gu", "<cmd>diffget //2<CR>", { desc = "Get changes from left side in diff" })
      vim.keymap.set("n", "gh", "<cmd>diffget //3<CR>", { desc = "Get changes from right side in diff" })
   end
}
