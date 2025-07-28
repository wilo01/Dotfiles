if vim.fn.has 'nvim-0.11' == 1 then
   return {
      {
         "williamboman/mason.nvim",
         config = function()
            require("mason").setup()
         end
      },
      {
         "williamboman/mason-lspconfig.nvim",
         dependencies = { "williamboman/mason.nvim" },
         config = function()
            require("mason-lspconfig").setup({
               ensure_installed = {
                  "ts_ls",
                  "eslint",
                  "lua_ls",
                  "bashls",
                  "gopls",
                  "dockerls",
                  "yamlls",
                  "zls"
               },
               automatic_installation = true,
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
      {
         "neovim/nvim-lspconfig", -- Add this to ensure LSP config is triggered
         dependencies = { "williamboman/mason.nvim", "hrsh7th/nvim-cmp" },
         config = function()
            local capabilities = require('cmp_nvim_lsp').default_capabilities()

            vim.lsp.config['*'] = {
               capabilities = capabilities,
            }

            vim.lsp.config.gopls = {
               root_markers = { "go.work", "go.mod", ".git" },
               settings = {
                  gopls = {
                     analyses = { unusedparams = true },
                     staticcheck = true,
                  },
               },
            }

            vim.lsp.config.lua_ls = {
               root_markers = { ".luarc.json", ".luarc.jsonc", ".git" },
               settings = {
                  Lua = {
                     runtime = { version = "Lua 5.1" },
                     diagnostics = {
                        globals = { "bit", "vim", "it", "describe", "before_each", "after_each" },
                     }
                  }
               }
            }

            vim.lsp.config.eslint = {
               root_markers = { ".eslintrc.js", ".eslintrc.json", "eslint.config.js", "package.json", ".git" },
               settings = {
                  format = { enable = true },
                  codeActionOnSave = {
                     enable = true,
                     mode = "all"
                  },
               },
            }

            vim.lsp.config.dockerls = {
               root_markers = { "Dockerfile", ".git" },
            }

            vim.lsp.config.yamlls = {
               settings = {
                  yaml = {
                     schemas = {
                        ["https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json"] =
                        "/docker-compose.yml"
                     }
                  }
               }
            }

            vim.lsp.config.zls = {
               root_markers = { "build.zig", "build.zig.zon", ".git" },
               settings = {
                  zls = {
                     enable_inlay_hints = true,
                     enable_snippets = true,
                     warn_style = true,
                  },
               },
            }

            vim.lsp.config.bashls = {
               root_markers = { ".git" },
            }

            vim.lsp.config.ts_ls = {
               root_markers = { "package.json", "tsconfig.json", "jsconfig.json", ".git" },
               settings = {
                  typescript = {
                     format = { indentSize = 3, tabSize = 3 },
                  },
                  javascript = {
                     format = { indentSize = 3, tabSize = 3 },
                  },
               },
            }

            local servers = { 'gopls', 'lua_ls', 'eslint', 'ts_ls', 'dockerls', 'yamlls', 'zls', 'bashls' }
            vim.lsp.enable(servers)

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

            -- Format on save using built-in LSP formatting with client filtering
            vim.api.nvim_create_autocmd("BufWritePre", {
               group = vim.api.nvim_create_augroup("LspFormat", { clear = true }),
               callback = function()
                  vim.lsp.buf.format({
                     async = false,
                     filter = function(client)
                        -- Only format with servers that support formatting
                        -- ESLint doesn't provide formatting, only linting
                        return client.name ~= "eslint"
                     end
                  })
               end,
            })

            -- Manual format keybinding
            vim.keymap.set("n", "<leader>f", function()
               vim.lsp.buf.format({
                  async = false,
                  filter = function(client)
                     return client.name ~= "eslint"
                  end
               })
            end, { desc = "Format buffer with LSP" })
         end
      }
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
               --                vim.api.nvim_create_autocmd("BufWritePre", {
               --                buffer = bufnr,
               --                callback = function()
               --                      vim.lsp.buf.format({ async = true })
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
