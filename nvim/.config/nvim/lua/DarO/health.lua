local M = {}

function M.check()
   local health = vim.health

   health.start("DarO Configuration")

   -- Check Neovim version
   local version = vim.version()
   if version.major >= 0 and version.minor >= 11 then
      health.ok(string.format("Neovim version: %d.%d.%d", version.major, version.minor, version.patch))
   else
      health.error("Neovim version is too old. Please upgrade to 0.11+")
   end

   -- Check core modules
   local modules = {
      "DarO.utils",
      "mason",
      "mason-lspconfig",
      "nvim-lspconfig",
      "nvim-cmp",
      "cmp_nvim_lsp",
      "nvim-treesitter.configs"
   }

   for _, module in ipairs(modules) do
      local ok, _ = pcall(require, module)
      if ok then
         health.ok("Module " .. module .. " loaded successfully")
      else
         health.warn("Module " .. module .. " failed to load")
      end
   end

   -- Check LSP servers
   local servers = { 'gopls', 'lua_ls', 'eslint', 'ts_ls', 'dockerls', 'yamlls', 'zls', 'bashls' }
   for _, server in ipairs(servers) do
      local clients = vim.lsp.get_clients({ name = server })
      if #clients > 0 then
         health.ok("LSP server " .. server .. " is running")
      else
         health.info("LSP server " .. server .. " is not running (may start when opening relevant files)")
      end
   end

   -- Check Mason installations
   local mason_registry = require("mason-registry")
   if mason_registry then
      health.ok("Mason registry available")
      local installed_packages = mason_registry.get_installed_packages()
      health.info(string.format("Mason has %d packages installed", #installed_packages))
   else
      health.error("Mason registry not available")
   end
end

return M

