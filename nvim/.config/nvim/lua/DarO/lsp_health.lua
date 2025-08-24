-- LSP Health Check Module
local M = {}

-- Check if servers are installed and running
function M.check_servers()
   local mason_registry = require("mason-registry")
   local required_servers = {
      "vtsls", "vue-language-server", "eslint-lsp", "lua-language-server",
      "bash-language-server", "gopls", "dockerfile-language-server", "yaml-language-server"
   }
   
   print("=== LSP Server Status ===")
   for _, server in ipairs(required_servers) do
      local pkg = mason_registry.get_package(server)
      if pkg:is_installed() then
         print("✓ " .. server .. " - Installed")
      else
         print("✗ " .. server .. " - NOT INSTALLED")
      end
   end
   
   -- Check active clients
   print("\n=== Active LSP Clients ===")
   local clients = vim.lsp.get_clients()
   if #clients == 0 then
      print("No active LSP clients")
   else
      for _, client in ipairs(clients) do
         print("✓ " .. client.name .. " (ID: " .. client.id .. ")")
      end
   end
end

-- Check project configuration
function M.check_project_config()
   local cwd = vim.fn.getcwd()
   print("=== Project Configuration ===")
   print("Working directory: " .. cwd)
   
   -- Check for TypeScript
   local ts_config = cwd .. "/tsconfig.json"
   if vim.fn.filereadable(ts_config) == 1 then
      print("✓ TypeScript config found")
   else
      print("✗ No tsconfig.json found")
   end
   
   -- Check for Vue
   local package_json = cwd .. "/package.json"
   if vim.fn.filereadable(package_json) == 1 then
      local content = vim.fn.readfile(package_json)
      local json_str = table.concat(content, "\n")
      if string.find(json_str, '"vue"') then
         print("✓ Vue.js project detected")
      else
         print("- No Vue.js detected")
      end
   end
   
   -- Check for ESLint config
   local eslint_configs = {
      ".eslintrc.js", ".eslintrc.json", ".eslintrc.cjs", ".eslintrc.yaml",
      "eslint.config.js", "eslint.config.mjs", "eslint.config.cjs"
   }
   
   local has_eslint = false
   for _, config_file in ipairs(eslint_configs) do
      if vim.fn.filereadable(cwd .. "/" .. config_file) == 1 then
         print("✓ ESLint config found: " .. config_file)
         has_eslint = true
         break
      end
   end
   
   if not has_eslint then
      -- Check package.json for eslintConfig
      if vim.fn.filereadable(package_json) == 1 then
         local content = vim.fn.readfile(package_json)
         local json_str = table.concat(content, "\n")
         if string.find(json_str, '"eslintConfig"') then
            print("✓ ESLint config found in package.json")
            has_eslint = true
         end
      end
   end
   
   if not has_eslint then
      print("✗ No ESLint configuration found")
   end
   
   -- Check TypeScript installation
   local ts_lib = cwd .. "/node_modules/typescript/lib"
   if vim.fn.isdirectory(ts_lib) == 1 then
      print("✓ Local TypeScript installation found")
   else
      print("⚠ No local TypeScript installation")
   end
end

-- Check log for recent errors
function M.check_logs()
   local log_path = vim.fn.stdpath("log") .. "/lsp.log"
   print("=== Recent LSP Errors ===")
   print("Log file: " .. log_path)
   
   if vim.fn.filereadable(log_path) == 1 then
      local cmd = "tail -20 " .. log_path .. " | grep -E '(ERROR|WARN)' | tail -5"
      local errors = vim.fn.systemlist(cmd)
      
      if #errors > 0 then
         for _, error in ipairs(errors) do
            print("⚠ " .. error)
         end
      else
         print("✓ No recent errors found")
      end
   else
      print("⚠ Log file not found")
   end
end

-- Run comprehensive health check
function M.health_check()
   M.check_servers()
   print("\n")
   M.check_project_config()
   print("\n")
   M.check_logs()
   print("\n=== Recommendations ===")
   print("- If servers are missing: :MasonInstall <server-name>")
   print("- To restart LSP: :LspRestart")
   print("- To check capabilities: :LspInfo")
   print("- For detailed logs: :lua vim.lsp.set_log_level('DEBUG')")
end

-- Create user command
vim.api.nvim_create_user_command('LspHealthCheck', M.health_check, {})

return M