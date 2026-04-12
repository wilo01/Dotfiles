return {

   "neovim/nvim-lspconfig",
   event = { "BufReadPre", "BufNewFile" },
   config = function()
      local on_attach = function(_, bufnr)
         local nmap = function(keys, func, desc)
            if desc then
               desc = "LSP: " .. desc
            end
            vim.keymap.set("n", keys, func, { buffer = bufnr, desc = desc })
         end

         nmap("gd", vim.lsp.buf.definition, "Go to Definition")
         nmap("K", vim.lsp.buf.hover, "Hover Documentation")
         nmap("<leader>rn", vim.lsp.buf.rename, "Rename Symbol")
         nmap("<leader>ca", vim.lsp.buf.code_action, "Code Action")
         nmap("gr", vim.lsp.buf.references, "Find References")
         nmap("gi", vim.lsp.buf.implementation, "Go to Implementation")
         nmap("<leader>ds", vim.lsp.buf.document_symbol, "Document Symbols")
         nmap("<leader>ws", vim.lsp.buf.workspace_symbol, "Workspace Symbols")

         vim.api.nvim_buf_create_user_command(bufnr, "Format", function()
            vim.lsp.buf.format()
         end, { desc = "Format current buffer with LSP" })
      end

      -- Enable LSP servers using lspconfig
      local servers = { "lua_ls", "gopls", "bashls", "html", "cssls" }
      local lspconfig = require("lspconfig")

      for _, server in ipairs(servers) do
         lspconfig[server].setup({
            on_attach = on_attach,
            capabilities = vim.lsp.protocol.make_client_capabilities(),
         })
      end
   end,
}
