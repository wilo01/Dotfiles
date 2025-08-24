-- Optimized Neovim Configuration
-- Fast startup with built-in preferences

-- Early optimizations
vim.loader.enable() -- Enable byte-compiled lua module loader (Neovim 0.9+)

-- Load core configuration
require("ai")

-- Startup time tracking (optional, remove for production)
if vim.env.NVIM_STARTUP_TIME then
   vim.g.startup_time = vim.loop.hrtime()
   vim.api.nvim_create_autocmd("VimEnter", {
      callback = function()
         vim.schedule(function()
            local startup_ms = (vim.loop.hrtime() - vim.g.startup_time) / 1000000
            vim.notify(string.format("Startup time: %.2f ms", startup_ms), vim.log.levels.INFO)
         end)
      end,
   })
end