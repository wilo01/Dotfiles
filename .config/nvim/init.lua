require("DarO")
vim.g.startup_time = vim.loop.hrtime()
vim.api.nvim_command('autocmd VimEnter * lua print_startup_time()')

function print_startup_time()
    local elapsed_time = (vim.loop.hrtime() - vim.g.startup_time) / 1e6
    print(string.format("Hello DarO, Neovim startup time: %.2f ms", elapsed_time))
end
