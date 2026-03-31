local gh = function(x) return 'https://github.com/' .. x end
local loaded = false

local mkdp_cmds = { "MarkdownPreviewToggle", "MarkdownPreview", "MarkdownPreviewStop" }

local function ensure_loaded()
   if loaded then return end
   loaded = true
   for _, c in ipairs(mkdp_cmds) do
      pcall(vim.api.nvim_del_user_command, c)
   end
   vim.pack.add({ gh('iamcco/markdown-preview.nvim') })

   vim.cmd([[
      function OpenMarkdownPreview (url)
         let cmd = "google-chrome-stable --new-window " . shellescape(a:url) . " &"
         silent call system(cmd)
      endfunction
   ]])
   vim.g.mkdp_browserfunc = "OpenMarkdownPreview"
end

-- Stub commands
for _, cmd in ipairs(mkdp_cmds) do
   vim.api.nvim_create_user_command(cmd, function(info)
      vim.api.nvim_del_user_command(cmd)
      ensure_loaded()
      vim.cmd({ cmd = cmd, args = { info.args }, bang = info.bang })
   end, { nargs = "*", bang = true })
end

-- Auto-load on markdown files
vim.api.nvim_create_autocmd("FileType", {
   pattern = "markdown",
   once = true,
   callback = function()
      ensure_loaded()
   end,
})
