-- LSP Safety Layer for Neovim 0.11+
-- Prevents critical runtime errors during LSP client initialization

local M = {}

-- Original vim.lsp.start function backup
local original_start = vim.lsp.start

-- Enhanced vim.lsp.start with error handling
function M.safe_lsp_start(config, opts)
   -- Validate config
   if not config then
      vim.notify("LSP start: No config provided", vim.log.levels.ERROR)
      return nil
   end
   
   -- Add safety wrapper for on_error callback
   local original_on_error = config.on_error
   config.on_error = function(code, err)
      vim.notify("LSP Error [" .. tostring(code) .. "]: " .. tostring(err), vim.log.levels.ERROR)
      if original_on_error then
         pcall(original_on_error, code, err)
      end
   end
   
   -- Add safety wrapper for before_init callback
   local original_before_init = config.before_init
   if original_before_init then
      config.before_init = function(params, config_inner)
         local ok, result = pcall(original_before_init, params, config_inner)
         if not ok then
            vim.notify("LSP before_init error: " .. tostring(result), vim.log.levels.ERROR)
         end
      end
   end
   
   -- Call original start with error handling
   local ok, client_id = pcall(original_start, config, opts)
   if not ok then
      vim.notify("LSP start failed: " .. tostring(client_id), vim.log.levels.ERROR)
      return nil
   end
   
   return client_id
end

-- Safe wrapper for deprecated client methods
function M.safe_is_stopped(client)
   if not client then return true end
   
   -- Use modern API if available
   if type(client.is_stopped) == "function" then
      local ok, result = pcall(client.is_stopped, client)
      return ok and result
   end
   
   -- Fallback to internal state
   if client._is_stopping ~= nil then
      return client._is_stopping
   end
   
   -- Final fallback - check if client exists in global list
   local clients = vim.lsp.get_clients and vim.lsp.get_clients() or vim.lsp.get_active_clients()
   for _, active_client in ipairs(clients) do
      if active_client.id == client.id then
         return false
      end
   end
   
   return true
end

function M.safe_supports_method(client, method)
   if not client or not method then return false end
   
   -- Use modern API if available
   if type(client.supports_method) == "function" then
      local ok, result = pcall(client.supports_method, client, method)
      return ok and result
   end
   
   -- Fallback to server capabilities check
   if client.server_capabilities then
      local capabilities = client.server_capabilities
      
      -- Map methods to capabilities
      local method_map = {
         ["textDocument/formatting"] = capabilities.documentFormattingProvider,
         ["textDocument/rangeFormatting"] = capabilities.documentRangeFormattingProvider,
         ["textDocument/hover"] = capabilities.hoverProvider,
         ["textDocument/completion"] = capabilities.completionProvider,
         ["textDocument/definition"] = capabilities.definitionProvider,
         ["textDocument/references"] = capabilities.referencesProvider,
         ["textDocument/rename"] = capabilities.renameProvider,
         ["textDocument/codeAction"] = capabilities.codeActionProvider,
      }
      
      local capability = method_map[method]
      return capability ~= nil and capability ~= false
   end
   
   return false
end

-- Enhanced client request wrapper
function M.safe_client_request(client, method, params, callback, bufnr)
   -- Validate client
   if not client then
      vim.notify("Client request: No client provided", vim.log.levels.ERROR)
      return false, "No client"
   end
   
   -- Check if client is stopped
   if M.safe_is_stopped(client) then
      vim.notify("Client request: Client is stopped", vim.log.levels.WARN)
      return false, "Client stopped"
   end
   
   -- Check if client supports method
   if not M.safe_supports_method(client, method) then
      vim.notify("Client request: Method " .. method .. " not supported", vim.log.levels.DEBUG)
      return false, "Method not supported"
   end
   
   -- Check if client has request function
   if not client.request then
      vim.notify("Client request: Client does not support requests", vim.log.levels.ERROR)
      return false, "No request function"
   end
   
   -- Validate method
   if not method or type(method) ~= "string" then
      vim.notify("Client request: Invalid method", vim.log.levels.ERROR)
      return false, "Invalid method"
   end
   
   -- Make request with error handling
   local ok, result = pcall(function()
      return client.request(method, params, callback, bufnr or 0)
   end)
   
   if not ok then
      vim.notify("Client request failed for " .. method .. ": " .. tostring(result), vim.log.levels.ERROR)
      return false, result
   end
   
   return true, result
end

-- Install LSP safety patches
function M.install_safety_patches()
   -- Patch vim.lsp.start
   vim.lsp.start = M.safe_lsp_start
   
   -- Add global error handler for unhandled LSP errors
   vim.api.nvim_create_autocmd("User", {
      pattern = "LspError",
      callback = function(args)
         vim.notify("Unhandled LSP Error: " .. vim.inspect(args.data), vim.log.levels.ERROR)
      end,
      desc = "Global LSP error handler"
   })
   
   -- Prevent LSP operations on special buffers
   vim.api.nvim_create_autocmd({ "BufEnter", "FileType" }, {
      pattern = { "snacks_dashboard", "dashboard", "alpha", "lazy", "mason", "TelescopePrompt" },
      callback = function()
         -- Disable LSP for special buffer types
         vim.b.lsp_disabled = true
         local bufnr = vim.api.nvim_get_current_buf()
         
         -- Stop any active LSP clients for this buffer
         local clients = vim.lsp.get_clients({ bufnr = bufnr })
         for _, client in ipairs(clients) do
            if client and client.stop then
               pcall(client.stop, true)
            end
         end
      end,
      desc = "Disable LSP for special buffers"
   })
   
   -- Create health check command
   M.create_health_command()
   
   vim.notify("LSP safety patches installed", vim.log.levels.INFO)
end

-- Safe client getter with deprecated API fallback
function M.get_clients(opts)
   opts = opts or {}
   
   -- Use modern API if available
   if vim.lsp.get_clients then
      local ok, clients = pcall(vim.lsp.get_clients, opts)
      if ok then return clients end
   end
   
   -- Fallback to deprecated API
   if vim.lsp.get_active_clients then
      local ok, clients = pcall(vim.lsp.get_active_clients, opts)
      if ok then return clients end
   end
   
   return {}
end

-- Check LSP health and provide diagnostics
function M.check_lsp_health()
   local issues = {}
   
   -- Check if we're using modern LSP API
   if not vim.lsp.config then
      table.insert(issues, "vim.lsp.config not available - using legacy mode")
   end
   
   if not vim.lsp.enable then
      table.insert(issues, "vim.lsp.enable not available - using legacy mode")
   end
   
   -- Check for deprecated API usage warnings
   if vim.lsp.get_active_clients and not vim.lsp.get_clients then
      table.insert(issues, "Using deprecated vim.lsp.get_active_clients")
   end
   
   -- Check for common problematic configurations
   local clients = M.get_clients()
   for _, client in ipairs(clients) do
      if not client.request then
         table.insert(issues, "Client " .. client.name .. " missing request function")
      end
      
      if M.safe_is_stopped(client) then
         table.insert(issues, "Client " .. client.name .. " is stopped but still listed")
      end
      
      -- Check for specific dependency issues
      if client.name == "vue_ls" then
         -- Check if ts_ls/vtsls is available for vue_ls
         local has_ts_support = false
         for _, other_client in ipairs(clients) do
            if (other_client.name == "vtsls" or other_client.name == "ts_ls") and not M.safe_is_stopped(other_client) then
               -- For hybrid mode, ts_ls should handle Vue files
               if other_client.name == "ts_ls" then
                  local filetypes = other_client.config and other_client.config.filetypes or {}
                  local supports_vue = false
                  for _, ft in ipairs(filetypes) do
                     if ft == "vue" then
                        supports_vue = true
                        break
                     end
                  end
                  if supports_vue then
                     has_ts_support = true
                  end
               else
                  has_ts_support = true
               end
               break
            end
         end
         if not has_ts_support then
            table.insert(issues, "vue_ls requires ts_ls/vtsls with Vue support but none found or stopped")
         end
      end
   end
   
   if #issues > 0 then
      vim.notify("LSP Health Issues:\n" .. table.concat(issues, "\n"), vim.log.levels.WARN)
   else
      vim.notify("LSP Health: All checks passed", vim.log.levels.INFO)
   end
   
   return #issues == 0
end

-- User command for health check
function M.create_health_command()
   vim.api.nvim_create_user_command('LspSafetyCheck', function()
      M.check_lsp_health()
   end, { desc = "Check LSP safety and deprecated API usage" })
   
   -- Vue-specific diagnostic command
   vim.api.nvim_create_user_command('VueLspCheck', function()
      local clients = M.get_clients()
      local vue_client = nil
      local ts_client = nil
      
      for _, client in ipairs(clients) do
         if client.name == "vue_ls" then
            vue_client = client
         elseif client.name == "ts_ls" then
            ts_client = client
         end
      end
      
      local status_lines = {
         "=== Vue LSP Status ===",
         "",
      }
      
      if vue_client then
         table.insert(status_lines, "✓ vue_ls: Running")
         if vue_client.config and vue_client.config.init_options then
            local hybrid = vue_client.config.init_options.vue and vue_client.config.init_options.vue.hybridMode
            table.insert(status_lines, "  - Hybrid Mode: " .. tostring(hybrid))
         end
      else
         table.insert(status_lines, "✗ vue_ls: Not found")
      end
      
      if ts_client then
         table.insert(status_lines, "✓ ts_ls: Running")
         if ts_client.config and ts_client.config.filetypes then
            local has_vue = false
            for _, ft in ipairs(ts_client.config.filetypes) do
               if ft == "vue" then
                  has_vue = true
                  break
               end
            end
            table.insert(status_lines, "  - Handles Vue files: " .. tostring(has_vue))
         end
      else
         table.insert(status_lines, "✗ ts_ls: Not found")
      end
      
      table.insert(status_lines, "")
      if vue_client and ts_client then
         table.insert(status_lines, "✓ Vue LSP configuration appears correct")
      else
         table.insert(status_lines, "✗ Vue LSP configuration has issues")
      end
      
      vim.notify(table.concat(status_lines, "\n"), vim.log.levels.INFO)
   end, { desc = "Check Vue LSP configuration status" })
end

return M