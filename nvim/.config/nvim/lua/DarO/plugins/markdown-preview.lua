local defer = require("DarO.defer")

local ensure = defer.on_cmd({
   repo = "iamcco/markdown-preview.nvim",
   cmds = { "MarkdownPreviewToggle", "MarkdownPreview", "MarkdownPreviewStop" },
   setup = function()
      vim.cmd([[
         function OpenMarkdownPreview (url)
            let cmd = "google-chrome-stable --new-window " . shellescape(a:url) . " &"
            silent call system(cmd)
         endfunction
      ]])
      vim.g.mkdp_browserfunc = "OpenMarkdownPreview"
   end,
})

vim.api.nvim_create_autocmd("FileType", {
   pattern = "markdown",
   callback = function()
      if ensure() then return true end
   end,
})
