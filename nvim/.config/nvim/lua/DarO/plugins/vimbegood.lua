local gh = function(x) return 'https://github.com/' .. x end
local loaded = false

vim.api.nvim_create_user_command("VimBeGood", function(info)
   if not loaded then
      loaded = true
      vim.pack.add({ gh('theprimeagen/vim-be-good') })
   end
   vim.api.nvim_del_user_command("VimBeGood")
   vim.cmd({ cmd = "VimBeGood", args = { info.args }, bang = info.bang })
end, { nargs = "*", bang = true })
