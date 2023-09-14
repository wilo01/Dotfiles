require('kommentary.config').use_extended_mappings()

vim.g.kommentary_create_default_mappings = false

vim.api.nvim_set_keymap("n", "<C-_>", "<Plug>kommentary_line_default", {})
-- vim.api.nvim_set_keymap("n", "<C-_>", "<Plug>kommentary_motion_default", {})
vim.api.nvim_set_keymap("x", "<C-_>", "<Plug>kommentary_visual_default", {})

require('kommentary.config').configure_language("default", {
    prefer_single_line_comments = true,
})

