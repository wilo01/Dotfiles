-- LSP Configuration
-- Optimized with built-in features and minimal overhead

local lspconfig = require("lspconfig")

-- Diagnostic configuration (using built-in vim.diagnostic)
vim.diagnostic.config({
   virtual_text = {
      prefix = "●",
      source = "if_many",
      spacing = 4,
   },
   float = {
      focusable = false,
      style = "minimal",
      border = "rounded",
      source = "if_many",
      header = "",
      prefix = "",
   },
   signs = {
      text = {
         [vim.diagnostic.severity.ERROR] = "",
         [vim.diagnostic.severity.WARN] = "",
         [vim.diagnostic.severity.HINT] = "",
         [vim.diagnostic.severity.INFO] = "",
      },
   },
   underline = true,
   update_in_insert = false,
   severity_sort = true,
})

-- LSP Attach function (keymaps and settings per buffer)
local on_attach = function(client, bufnr)
   -- Enable completion triggered by <c-x><c-o>
   vim.bo[bufnr].omnifunc = "v:lua.vim.lsp.omnifunc"

   -- Keymaps
   local function map(mode, lhs, rhs, desc)
      vim.keymap.set(mode, lhs, rhs, { buffer = bufnr, desc = desc })
   end

   -- Navigation
   map("n", "gd", vim.lsp.buf.definition, "Go to definition")
   map("n", "gD", vim.lsp.buf.declaration, "Go to declaration")
   map("n", "gi", vim.lsp.buf.implementation, "Go to implementation")
   map("n", "go", vim.lsp.buf.type_definition, "Go to type definition")
   map("n", "gr", vim.lsp.buf.references, "References")
   map("n", "gR", "<cmd>Telescope lsp_references<CR>", "References (Telescope)")
   map("n", "gs", vim.lsp.buf.signature_help, "Signature help")
   map("i", "<C-k>", vim.lsp.buf.signature_help, "Signature help")

   -- Information
   map("n", "K", vim.lsp.buf.hover, "Hover documentation")
   map("n", "<leader>k", vim.lsp.buf.hover, "Hover documentation")

   -- Symbols and navigation
   map("n", "<leader>ss", "<cmd>Telescope lsp_document_symbols<CR>", "Document symbols")
   map("n", "<leader>sS", "<cmd>Telescope lsp_workspace_symbols<CR>", "Workspace symbols")

   -- Actions
   map("n", "<leader>rn", vim.lsp.buf.rename, "Rename")
   map({ "n", "v" }, "<leader>ca", vim.lsp.buf.code_action, "Code action")
   map("n", "<leader>f", function()
      vim.lsp.buf.format({ async = true })
   end, "Format document")

   -- Workspace
   map("n", "<leader>wa", vim.lsp.buf.add_workspace_folder, "Add workspace folder")
   map("n", "<leader>wr", vim.lsp.buf.remove_workspace_folder, "Remove workspace folder")
   map("n", "<leader>wl", function()
      print(vim.inspect(vim.lsp.buf.list_workspace_folders()))
   end, "List workspace folders")

   -- Inlay hints (Neovim 0.10+)
   if vim.lsp.inlay_hint and client.server_capabilities.inlayHintProvider then
      map("n", "<leader>th", function()
         vim.lsp.inlay_hint.enable(not vim.lsp.inlay_hint.is_enabled())
      end, "Toggle inlay hints")
   end

   -- Highlight references on cursor hold
   if client.server_capabilities.documentHighlightProvider then
      vim.api.nvim_create_autocmd({ "CursorHold", "CursorHoldI" }, {
         buffer = bufnr,
         callback = vim.lsp.buf.document_highlight,
      })
      vim.api.nvim_create_autocmd({ "CursorMoved", "CursorMovedI" }, {
         buffer = bufnr,
         callback = vim.lsp.buf.clear_references,
      })
   end

   -- Enable semantic tokens if supported
   if client.server_capabilities.semanticTokensProvider then
      vim.lsp.semantic_tokens.start(bufnr, client.id)
   end
end

-- Capabilities (enhanced with cmp_nvim_lsp)
local capabilities = vim.lsp.protocol.make_client_capabilities()
local has_cmp, cmp_nvim_lsp = pcall(require, "cmp_nvim_lsp")
if has_cmp then
   capabilities = vim.tbl_deep_extend("force", capabilities, cmp_nvim_lsp.default_capabilities())
end

-- Enhance capabilities for specific features
capabilities.textDocument.foldingRange = {
   dynamicRegistration = false,
   lineFoldingOnly = true,
}
capabilities.textDocument.completion.completionItem.snippetSupport = true
capabilities.textDocument.completion.completionItem.resolveSupport = {
   properties = {
      "documentation",
      "detail",
      "additionalTextEdits",
   },
}

-- Server configurations
local servers = {
   -- Lua
   lua_ls = {
      settings = {
         Lua = {
            runtime = {
               version = "LuaJIT",
            },
            diagnostics = {
               globals = { "vim", "it", "describe", "before_each", "after_each" },
            },
            workspace = {
               library = {
                  vim.env.VIMRUNTIME,
               },
               checkThirdParty = false,
            },
            telemetry = {
               enable = false,
            },
            format = {
               enable = true,
               defaultConfig = {
                  indent_style = "space",
                  indent_size = "3",
               },
            },
         },
      },
   },

   -- Python
   pyright = {
      settings = {
         python = {
            analysis = {
               autoSearchPaths = true,
               useLibraryCodeForTypes = true,
               diagnosticMode = "workspace",
               typeCheckingMode = "basic",
            },
         },
      },
   },

   -- TypeScript/JavaScript
   ts_ls = {
      settings = {
         typescript = {
            inlayHints = {
               includeInlayParameterNameHints = "all",
               includeInlayParameterNameHintsWhenArgumentMatchesName = false,
               includeInlayFunctionParameterTypeHints = true,
               includeInlayVariableTypeHints = true,
               includeInlayPropertyDeclarationTypeHints = true,
               includeInlayFunctionLikeReturnTypeHints = true,
               includeInlayEnumMemberValueHints = true,
            },
         },
      },
   },

   -- Rust
   rust_analyzer = {
      settings = {
         ["rust-analyzer"] = {
            checkOnSave = {
               command = "clippy",
            },
            cargo = {
               allFeatures = true,
            },
            inlayHints = {
               enable = true,
            },
         },
      },
   },

   -- Go
   gopls = {
      settings = {
         gopls = {
            analyses = {
               unusedparams = true,
               shadow = true,
            },
            staticcheck = true,
            gofumpt = true,
            hints = {
               assignVariableTypes = true,
               compositeLiteralFields = true,
               compositeLiteralTypes = true,
               constantValues = true,
               functionTypeParameters = true,
               parameterNames = true,
               rangeVariableTypes = true,
            },
         },
      },
   },

   -- JSON
   jsonls = {
      settings = {
         json = {
            validate = { enable = true },
            format = { enable = true },
         },
      },
   },

   -- YAML
   yamlls = {
      settings = {
         yaml = {
            format = {
               enable = true,
            },
            validate = true,
            hover = true,
            completion = true,
         },
      },
   },

   -- HTML/CSS
   html = {},
   cssls = {},
   emmet_ls = {},

   -- Bash
   bashls = {},

   -- Docker
   dockerls = {},
   docker_compose_language_service = {},
}

-- Setup servers
for server, config in pairs(servers) do
   config.on_attach = on_attach
   config.capabilities = capabilities
   lspconfig[server].setup(config)
end

-- Additional servers that might not have mason-lspconfig integration
-- but can be set up if they're installed
local additional_servers = {
   "clangd",
   "cmake",
   "vimls",
   "marksman",
   "sqlls",
   "terraformls",
}

for _, server in ipairs(additional_servers) do
   if vim.fn.executable(server) == 1 then
      lspconfig[server].setup({
         on_attach = on_attach,
         capabilities = capabilities,
      })
   end
end

-- UI enhancements
vim.lsp.handlers["textDocument/hover"] = vim.lsp.with(vim.lsp.handlers.hover, {
   border = "rounded",
})

vim.lsp.handlers["textDocument/signatureHelp"] = vim.lsp.with(vim.lsp.handlers.signature_help, {
   border = "rounded",
})