vim.loader.enable()
vim.g.startup_time = vim.uv.hrtime()
require("DarO")
vim.api.nvim_command('autocmd VimEnter * lua require("DarO.utils").print_startup_time()')
