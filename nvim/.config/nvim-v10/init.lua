require("DarO")
vim.g.startup_time = vim.loop.hrtime()
vim.api.nvim_command('autocmd VimEnter * lua require("DarO.utils").print_startup_time()')
