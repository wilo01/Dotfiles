return {
   "iamcco/markdown-preview.nvim",
   cmd = { "MarkdownPreviewToggle", "MarkdownPreview", "MarkdownPreviewStop" },
   ft = { "markdown" },
   build = function() vim.fn["mkdp#util#install"]() end,
   config = function()
      local plugin_dir = vim.fn.stdpath("data") .. "/lazy/markdown-preview.nvim/app"
      if vim.fn.isdirectory(plugin_dir .. "/node_modules") == 0 then
         vim.fn["mkdp#util#install"]()
      end
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
