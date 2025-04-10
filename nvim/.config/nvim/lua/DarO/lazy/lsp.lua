return {
   "neovim/nvim-lspconfig",
   dependencies = {
      "williamboman/mason.nvim",
      "williamboman/mason-lspconfig.nvim",
      "hrsh7th/cmp-nvim-lsp",
      "hrsh7th/cmp-buffer",
      "hrsh7th/cmp-path",
      "hrsh7th/cmp-cmdline",
      "hrsh7th/nvim-cmp",
      "L3MON4D3/LuaSnip",
      "saadparwaiz1/cmp_luasnip",
      "j-hui/fidget.nvim",
      {
         "folke/lazydev.nvim",
         ft = "lua",
         opts = {
            library = {
               { path = "${3rd}/luv/library", words = { "vim%.uv" } },
            }
         }
      }
   },
   config = function()
      local cmp = require('cmp')
      local cmp_lsp = require("cmp_nvim_lsp")

      local capabilities = vim.lsp.protocol.make_client_capabilities()
      capabilities = cmp_lsp.default_capabilities(capabilities)

      require("mason").setup()
      require("mason-lspconfig").setup({
         automatic_installation = true,
         ensure_installed = {
            "eslint",
            "lua_ls",
            "gopls",
            "dockerls",
            "yamlls",
            -- "volar"
         },
         handlers = {
            function(server_name)
               require("lspconfig")[server_name].setup {
                  capabilities = capabilities
               }
            end,

            zls = function()
               local lspconfig = require("lspconfig")
               lspconfig.zls.setup({
                  root_dir = lspconfig.util.root_pattern(".git", "build.zig", "zls.json"),
                  settings = {
                     zls = {
                        enable_inlay_hints = true,
                        enable_snippets = true,
                        warn_style = true,
                     },
                  },
               })
               vim.g.zig_fmt_parse_errors = 0
               vim.g.zig_fmt_autosave = 0
            end,

            ["lua_ls"] = function()
               local lspconfig = require("lspconfig")
               lspconfig.lua_ls.setup {
                  capabilities = capabilities,
                  settings = {
                     Lua = {
                        runtime = { version = "Lua 5.1" },
                        diagnostics = {
                           globals = { "bit", "vim", "it", "describe", "before_each", "after_each" },
                        }
                     }
                  }
               }
            end,

            ["gopls"] = function()
               local lspconfig = require("lspconfig")
               lspconfig.gopls.setup({
                  capabilities = capabilities,
                  settings = {
                     gopls = {
                        analyses = {
                           unusedparams = true,
                        },
                        staticcheck = true,
                     },
                  },
               })
            end,

            ["dockerls"] = function()
               require("lspconfig").dockerls.setup({
                  capabilities = capabilities,
               })
            end,

            ["yamlls"] = function()
               require("lspconfig").yamlls.setup({
                  capabilities = capabilities,
                  settings = {
                     yaml = {
                        schemas = {
                           ["https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json"] = "/docker-compose.yml"
                        }
                     }
                  }
               })
            end,

            -- ["volar"] = function()
            --    local lspconfig = require("lspconfig")
            --    lspconfig.volar.setup({
            --       capabilities = capabilities,
            --       filetypes = { "typescript", "javascript", "javascriptreact", "typescriptreact", "vue", "json" },
            --       on_attach = function(client, bufnr)
            --          if client.supports_method("textDocument/formatting") then
            --             vim.api.nvim_create_autocmd("BufWritePre", {
            --                buffer = bufnr,
            --                callback = function()
            --                   vim.lsp.buf.format({ async = true })
            --                end,
            --             })
            --          end
            --       end,
            --    })
            -- end,
         }
      })

      local cmp_select = { behavior = cmp.SelectBehavior.Select }
      cmp.setup({
         snippet = {
            expand = function(args)
               require('luasnip').lsp_expand(args.body) -- For `luasnip` users.
            end,
         },
         mapping = cmp.mapping.preset.insert({
            ['<C-p>'] = cmp.mapping.select_prev_item(cmp_select),
            ['<C-n>'] = cmp.mapping.select_next_item(cmp_select),
            ['<C-y>'] = cmp.mapping.confirm({ select = true }),
            ["<C-Space>"] = cmp.mapping.complete(),
         }),
         sources = cmp.config.sources({
            { name = 'nvim_lsp' },
            { name = 'luasnip' }, -- For luasnip users.
         }, {
            { name = 'buffer' },
         })
      })

      vim.diagnostic.config({
         float = {
            focusable = false,
            style = "minimal",
            border = "rounded",
            source = "if_many",
            header = "",
            prefix = "",
         },
      })
   end
}
