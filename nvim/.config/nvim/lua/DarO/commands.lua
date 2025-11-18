-- Enhanced User Commands for LSP Management

-- LspStatus command - comprehensive LSP status
vim.api.nvim_create_user_command('LspStatus', function()
   print("=== LSP Status Report ===")
   
   -- Get clients safely
   local get_clients = vim.lsp.get_clients or vim.lsp.get_active_clients
   local clients = get_clients()
   
   if #clients == 0 then
      print("No active LSP clients")
      return
   end
   
   for _, client in ipairs(clients) do
      local status = "RUNNING"
      
      -- Check if client has is_stopped method
      if type(client.is_stopped) == "function" then
         local ok, stopped = pcall(client.is_stopped, client)
         if ok and stopped then
            status = "STOPPED"
         end
      elseif client._is_stopping then
         status = "STOPPING"
      end
      
      print(string.format("• %s (ID: %d) - %s", client.name, client.id, status))
      
      -- Show attached buffers
      local attached_count = 0
      if client.attached_buffers then
         for _ in pairs(client.attached_buffers) do
            attached_count = attached_count + 1
         end
      end
      print(string.format("  Attached buffers: %d", attached_count))
      
      -- Show root directory
      if client.config and client.config.root_dir then
         print(string.format("  Root: %s", client.config.root_dir))
      end
   end
end, { desc = "Show comprehensive LSP status" })

-- LspReload command - safe LSP restart
vim.api.nvim_create_user_command('LspReload', function(opts)
   local get_clients = vim.lsp.get_clients or vim.lsp.get_active_clients
   local clients = get_clients()
   
   if opts.args and opts.args ~= "" then
      -- Restart specific client
      for _, client in ipairs(clients) do
         if client.name == opts.args then
            print("Restarting " .. client.name .. "...")
            client.stop()
            vim.defer_fn(function()
               print("Restarted " .. client.name)
            end, 1000)
            return
         end
      end
      print("Client not found: " .. opts.args)
   else
      -- Restart all clients
      print("Restarting all LSP clients...")
      for _, client in ipairs(clients) do
         client.stop()
      end
      vim.defer_fn(function()
         print("All LSP clients restarted")
      end, 1000)
   end
end, {
   nargs = '?',
   complete = function()
      local get_clients = vim.lsp.get_clients or vim.lsp.get_active_clients
      local clients = get_clients()
      local names = {}
      for _, client in ipairs(clients) do
         table.insert(names, client.name)
      end
      return names
   end,
   desc = "Safely restart LSP clients"
})

-- LspDebug command - comprehensive debugging info
vim.api.nvim_create_user_command('LspDebug', function()
   print("=== LSP Debug Information ===")

   -- Check Neovim version
   local version = vim.version()
   print(string.format("Neovim: %d.%d.%d", version.major, version.minor, version.patch))

   -- Check if modern LSP API is available
   print("Modern LSP API:")
   print("  vim.lsp.config:", vim.lsp.config and "✓" or "✗")
   print("  vim.lsp.enable:", vim.lsp.enable and "✓" or "✗")
   print("  vim.lsp.get_clients:", vim.lsp.get_clients and "✓" or "✗")

   -- Check for deprecated API
   if vim.lsp.get_active_clients and not vim.lsp.get_clients then
      print("  WARNING: Using deprecated vim.lsp.get_active_clients")
   end

   -- Check required modules
   print("\nRequired modules:")
   local modules = { "lspconfig", "cmp_nvim_lsp", "mason", "mason-lspconfig" }
   for _, module in ipairs(modules) do
      local ok, _ = pcall(require, module)
      print(string.format("  %s: %s", module, ok and "✓" or "✗"))
   end

   -- Show current buffer LSP info
   local bufnr = vim.api.nvim_get_current_buf()
   local get_clients = vim.lsp.get_clients or vim.lsp.get_active_clients
   local buf_clients = get_clients({ bufnr = bufnr })

   print(string.format("\nCurrent buffer (%d):", bufnr))
   print("  Filetype:", vim.bo.filetype)
   print("  LSP clients:", #buf_clients)
   for _, client in ipairs(buf_clients) do
      print(string.format("    • %s", client.name))
   end

   -- Check log file
   local log_path = vim.lsp.get_log_path()
   local log_stat = vim.loop.fs_stat(log_path)
   if log_stat then
      print(string.format("\nLSP log: %s (%.1f KB)", log_path, log_stat.size / 1024))
   else
      print("\nLSP log: Not found")
   end
end, { desc = "Show comprehensive LSP debug information" })

-- FormatDebug command - show all autocommands that might format
vim.api.nvim_create_user_command('FormatDebug', function()
   print("=== Format Debug Information ===")
   print("Autoformat disabled:", vim.g.disable_autoformat and "YES" or "NO")
   print("\nBufWritePre autocommands:")
   local autocmds = vim.api.nvim_get_autocmds({ event = "BufWritePre" })
   for _, autocmd in ipairs(autocmds) do
      print(string.format("  Group: %s", autocmd.group_name or "none"))
      if autocmd.pattern then
         print(string.format("  Pattern: %s", vim.inspect(autocmd.pattern)))
      end
   end

   print("\nLSP clients with formatting capability:")
   local get_clients = vim.lsp.get_clients or vim.lsp.get_active_clients
   local clients = get_clients()
   for _, client in ipairs(clients) do
      if client.supports_method("textDocument/formatting") then
         print(string.format("  • %s", client.name))
         if client.name == "eslint" and client.config.settings then
            print(string.format("    format.enable: %s", client.config.settings.format and client.config.settings.format.enable or "nil"))
         end
      end
   end
end, { desc = "Show format-related debug information" })