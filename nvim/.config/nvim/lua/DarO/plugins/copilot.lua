vim.g.copilot_no_tab_map = true
vim.g.copilot_assume_mapped = true

vim.keymap.set("i", "<M-l>", 'copilot#Accept("\\<CR>")', {
   expr = true,
   replace_keycodes = false,
   silent = true,
   desc = "Copilot: accept suggestion",
})

vim.keymap.set("i", "<M-w>", "<Plug>(copilot-accept-word)",
   { silent = true, desc = "Copilot: accept word" })

vim.keymap.set("i", "<M-CR>", "<Plug>(copilot-accept-line)",
   { silent = true, desc = "Copilot: accept line" })

vim.keymap.set("i", "<C-]>", "<Plug>(copilot-dismiss)",
   { silent = true, desc = "Copilot: dismiss" })
