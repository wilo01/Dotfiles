if vim.fn.has 'nvim-0.11' == 1 then
   return {
      {
         "williamboman/mason.nvim",
         config = function()
            require("mason").setup()
         end
      },
      {
         "stevearc/conform.nvim",
         config = function()
            require("conform").setup({
               formatters_by_ft = {
                  go = { "gofmt" },
                  javascript = { "prettierd" },
                  typescript = { "prettierd" },
                  json = { "prettierd" },
                  yaml = { "prettierd" },
                  html = { "prettierd" },
                  markdown = { "prettierd" },
               },
               format_on_save = {
                  timeout_ms = 500,
                  lsp_fallback = true,
               },
            })
         end
      },
      {
         "hrsh7th/nvim-cmp",
         dependencies = {
            "hrsh7th/cmp-nvim-lsp",
            "hrsh7th/cmp-buffer",
            "hrsh7th/cmp-path",
            "L3MON4D3/LuaSnip",
            "saadparwaiz1/cmp_luasnip",
         },
         config = function()
            local cmp = require('cmp')
            local cmp_select = { behavior = cmp.SelectBehavior.Select }

            cmp.setup({
               snippet = {
                  expand = function(args)
                     require('luasnip').lsp_expand(args.body)
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
                  { name = 'luasnip' },
               }, {
                  { name = 'buffer' },
               })
            })
         end
      },
      {
         "folke/lazydev.nvim",
         ft = "lua",
         opts = {
            library = {
               { path = "${3rd}/luv/library", words = { "vim%.uv" } },
            }
         }
      },
      config = function()
         local capabilities = require('cmp_nvim_lsp').default_capabilities()

         vim.lsp.enable({
            {
               name = "gopls",
               cmd = { "gopls" },
               root_markers = { "go.mod", ".git" },
               capabilities = capabilities,
               settings = {
                  gopls = {
                     analyses = {
                        unusedparams = true,
                     },
                     staticcheck = true,
                  },
               },
            },
            {
               name = "lua_ls",
               cmd = { "lua-language-server" },
               root_markers = { ".luarc.json", ".git" },
               capabilities = capabilities,
               settings = {
                  Lua = {
                     runtime = { version = "Lua 5.1" },
                     diagnostics = {
                        globals = { "bit", "vim", "it", "describe", "before_each", "after_each" },
                     }
                  }
               }
            },
            {
               name = "eslint",
               cmd = { "vscode-eslint-language-server", "--stdio" },
               root_markers = { ".eslintrc.js", ".eslintrc.json", "package.json", ".git" },
               capabilities = capabilities,
            },
            {
               name = "dockerls",
               cmd = { "docker-langserver", "--stdio" },
               root_markers = { "Dockerfile", ".git" },
               capabilities = capabilities,
            },
            {
               name = "yamlls",
               cmd = { "yaml-language-server", "--stdio" },
               root_markers = { ".git" },
               capabilities = capabilities,
               settings = {
                  yaml = {
                     schemas = {
                        ["https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json"] = "/docker-compose.yml"
                     }
                  }
               }
            },
            {
               name = "zls",
               cmd = { "zls" },
               root_markers = { ".git", "build.zig", "zls.json" },
               capabilities = capabilities,
               settings = {
                  zls = {
                     enable_inlay_hints = true,
                     enable_snippets = true,
                     warn_style = true,
                  },
               },
            }
         })

         vim.diagnostic.config({
            virtual_text = true,
            underline = true,
            update_in_insert = false,
            severity_sort = true,
            float = {
               focusable = false,
               style = "minimal",
               border = "rounded",
               source = true,
               header = "",
               prefix = "",
            },
            signs = {
               text = {
                  [vim.diagnostic.severity.ERROR] = "",
                  [vim.diagnostic.severity.WARN] = "",
                  [vim.diagnostic.severity.INFO] = "",
                  [vim.diagnostic.severity.HINT] = ""
               },
               numhl = {
                  [vim.diagnostic.severity.ERROR] = "ErrorMsg",
                  [vim.diagnostic.severity.WARN] = "WarningMsg",
               }
            }
         })

         vim.g.zig_fmt_parse_errors = 0
         vim.g.zig_fmt_autosave = 0
      end
   }
else
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
         "jose-elias-alvarez/null-ls.nvim",
         "jay-babu/mason-null-ls.nvim",
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
         local null_ls = require("null-ls")

         local capabilities = vim.tbl_deep_extend(
            "force",
            {},
            vim.lsp.protocol.make_client_capabilities(),
            cmp_lsp.default_capabilities()
         )

         require("mason").setup()
         require("mason-null-ls").setup({
            automatic_installation = true,
            ensure_installed = { "gofmt", "prettierd" },
         })
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
               function(server_name) -- default handler (optional)
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
                              ["https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json"] =
                              "/docker-compose.yml"
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

         null_ls.setup({
            sources = {
               null_ls.builtins.formatting.gofmt,
               null_ls.builtins.formatting.prettierd.with({
                  -- filetypes = { "json", "yaml", "typescript", "html", "vue", "markdown" },
                  filetypes = { "json", "yaml", "typescript", "html", "markdown" },
                  extra_args = {
                     "--ignore-path", "/dev/null",
                     "--ignore-patterns", "%[(.-)%]"
                  },
               }),
            },
            on_attach = function(client, bufnr)
               if client.supports_method("textDocument/formatting") then
                  vim.api.nvim_create_autocmd("BufWritePre", {
                     buffer = bufnr,
                     callback = function()
                        vim.lsp.buf.format({ async = true })
                     end,
                  })
               end
            end,
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
            virtual_text = true,
            underline = true,
            update_in_insert = false,
            severity_sort = true,
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
                  [vim.diagnostic.severity.INFO] = "",
                  [vim.diagnostic.severity.HINT] = ""
               },
               numhl = {
                  [vim.diagnostic.severity.ERROR] = "ErrorMsg",
                  [vim.diagnostic.severity.WARN] = "WarningMsg",
               }
            }
         })
      end
   }
end
