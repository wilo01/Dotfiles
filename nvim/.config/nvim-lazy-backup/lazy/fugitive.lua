return {
   "tpope/vim-fugitive",
   config = function()
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

      vim.keymap.set("n", "<leader>gs", vim.cmd.Git, { desc = "Fugitive Open Git status (Git control)" })
      vim.keymap.set("n", "gu", "<cmd>diffget //2<CR>", { desc = "Fugitive Get changes from left side in git diff" })
      vim.keymap.set("n", "gh", "<cmd>diffget //3<CR>", { desc = "Fugitive Get changes from right side in git diff" })
   end
}
