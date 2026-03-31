require("cloak").setup({
   enabled = true,
   cloak_character = "*",
   highlight_group = "Comment",
   patterns = {
      {
         file_pattern = {
            ".env*",
            "wrangler.toml",
            ".dev.vars",
         },
         cloak_pattern = "=.+"
      },
   },
})
vim.api.nvim_set_keymap("n", "<leader>e", ":CloakToggle<CR>",
   { noremap = true, silent = true, desc = "Toggle Cloak for sensitive text in env files" })
