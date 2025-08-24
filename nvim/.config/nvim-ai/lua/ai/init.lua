-- Core module loader
-- Loads configuration in optimal order for performance

-- Critical settings that affect other modules
require("ai.options")

-- Keymaps (loaded early for immediate availability)
require("ai.keymaps") 

-- Plugin manager and plugins (lazy-loaded)
require("ai.plugins")

-- Autocommands (deferred to reduce startup impact)
vim.schedule(function()
   require("ai.autocmds")
end)

-- Statusline (lightweight built-in)
vim.schedule(function()
   require("ai.statusline")
end)