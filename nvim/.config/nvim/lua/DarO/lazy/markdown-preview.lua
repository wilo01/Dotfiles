return {
   "iamcco/markdown-preview.nvim",
   cmd = { "MarkdownPreviewToggle", "MarkdownPreview", "MarkdownPreviewStop" },
   ft = { "markdown" },
   build = function() vim.fn["mkdp#util#install"]() end,
   config = function()
      vim.cmd([[do FileType]])
      vim.cmd([[
         function OpenMarkdownPreview (url)
            let cmd = "google-chrome-stable --new-window " . shellescape(a:url) . " &"
            silent call system(cmd)
         endfunction
      ]])
      vim.g.mkdp_browserfunc = "OpenMarkdownPreview"
   end,
}

-- Old install using yarn
-- return {
--    "iamcco/markdown-preview.nvim",
--    cmd = { "MarkdownPreviewToggle", "MarkdownPreview", "MarkdownPreviewStop" },
--    build = "cd app && yarn install",
--    init = function()
--       vim.g.mkdp_auto_start = 1              -- Automatically start preview when opening markdown file
--       vim.g.mkdp_auto_close = 1              -- Automatically close preview when switching buffers
--       vim.g.mkdp_refresh_slow = 0            -- Refresh preview as you type (live preview)
--       vim.g.mkdp_command_for_global = 1      -- Allow using preview commands globally
--       vim.g.mkdp_browser = 'google-chrome'   -- Change default browser to Google Chrome
--       vim.g.mkdp_filetypes = { "markdown" }  -- Restrict to markdown files
--    end,
--    ft = { "markdown" },
-- }

-- Markdown keymap in remap.lua
-- Markdown Preview
-- vim.keymap.set("n", "<leader>m", "<CMD>MarkdownPreview<CR>")      -- Markdown preview
-- vim.keymap.set("n", "<leader>mn", "<CMD>MarkdownPreviewStop<CR>") -- Markdown preview stop
