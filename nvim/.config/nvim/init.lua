require("DarO")
vim.g.startup_time = vim.uv.hrtime()
vim.api.nvim_command('autocmd VimEnter * lua require("DarO.utils").print_startup_time()')
