return {
   "iamcco/markdown-preview.nvim",
   cmd = { "MarkdownPreviewToggle", "MarkdownPreview", "MarkdownPreviewStop" },
   build = "cd app && yarn install",
   init = function()
      vim.g.mkdp_filetypes = { "markdown" }
   end,
   ft = { "markdown" },
}

-- Markdown keymap in remap.lua
-- Markdown Preview
-- vim.keymap.set("n", "<leader>m", "<CMD>MarkdownPreview<CR>")      -- Markdown preview
-- vim.keymap.set("n", "<leader>mn", "<CMD>MarkdownPreviewStop<CR>") -- Markdown preview stop
