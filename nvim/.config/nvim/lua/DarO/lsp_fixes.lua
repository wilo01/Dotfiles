-- LSP Configuration Fixes
-- This module contains specific fixes for common LSP issues

local M = {}

-- Fix ESLint configuration detection
function M.fix_eslint_config()
   vim.api.nvim_create_autocmd("FileType", {
      pattern = { "javascript", "javascriptreact", "typescript", "typescriptreact", "vue" },
      callback = function()
         -- Only apply ESLint fix if ESLint client is active
         local clients = vim.lsp.get_clients and vim.lsp.get_clients() or vim.lsp.get_active_clients()
         local eslint_client = nil
         
         for _, client in ipairs(clients) do
            if client.name == "eslint" then
               eslint_client = client
               break
            end
         end
         
         if eslint_client then
            -- Set buffer-local ESLint settings to be more lenient
            vim.b.eslint_config_fallback = true
         end
      end,
      desc = "Fix ESLint configuration detection"
   })
end

-- Fix Vue.js LSP initialization
function M.fix_vue_ls_init()
   -- Create autocmd to ensure proper Vue LS initialization
   vim.api.nvim_create_autocmd("FileType", {
      pattern = "vue",
      callback = function()
         -- Ensure TypeScript is available for Vue files
         local root_dir = vim.fn.getcwd()
         local ts_lib = root_dir .. "/node_modules/typescript/lib"
         
         if vim.fn.isdirectory(ts_lib) == 0 then
            vim.notify("Vue: No local TypeScript found. Run 'npm install typescript' for better Vue support.", 
                      vim.log.levels.INFO)
         end
      end,
      desc = "Vue LS initialization check"
   })
end

-- Fix deprecation warnings
function M.fix_deprecation_warnings()
   -- Monkey patch deprecated functions to reduce noise
   if vim.lsp.get_active_clients and not vim.lsp._get_active_clients_warned then
      local original = vim.lsp.get_active_clients
      vim.lsp.get_active_clients = function(...)
         -- Suppress deprecation warning in logs
         local original_notify = vim.notify
         vim.notify = function(msg, level)
            if type(msg) == "string" and msg:match("get_active_clients.*deprecated") then
               return -- Suppress this specific warning
            end
            original_notify(msg, level)
         end
         
         local result = original(...)
         vim.notify = original_notify -- Restore
         return result
      end
      vim.lsp._get_active_clients_warned = true
   end
end

-- Apply all fixes
function M.apply_all_fixes()
   M.fix_eslint_config()
   M.fix_vue_ls_init()
   M.fix_deprecation_warnings()
   
   vim.notify("LSP fixes applied", vim.log.levels.DEBUG)
end

return M