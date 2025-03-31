require("DarO")
vim.g.startup_time = vim.loop.hrtime()
vim.api.nvim_command('autocmd VimEnter * lua print_startup_time()')

function _G.print_startup_time()
   local elapsed_time = (vim.loop.hrtime() - vim.g.startup_time) / 1e6
   local v = vim.version()
   print(string.format("Hello DarO, Neovim v%d.%d.%d startup time: %.2f ms", v.major, v.minor, v.patch, elapsed_time))
   vim.defer_fn(function()
      vim.cmd('messages')
   end, 50)
end
