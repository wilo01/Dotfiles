-- Common configurations that work across all Neovim versions
local common = {}

-- Consolidated CMP setup function
function common.setup_cmp()
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

-- Consolidated diagnostic setup function
function common.setup_diagnostics()
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

-- Mason setup function
function common.setup_mason()
   require("mason").setup({
      ui = {
         check_outdated_packages_on_open = true,
         border = "rounded",
         icons = {
            package_installed = "✓",
            package_pending = "➜",
            package_uninstalled = "✗"
         }
      },
      log_level = vim.log.levels.INFO,
      max_concurrent_installers = 4,
      pip = {
         upgrade_pip = true,
      },
   })
end

-- Common server configurations
function common.get_server_configs(capabilities)
   return {
      gopls = {
         root_markers = { "go.work", "go.mod", ".git" },
         settings = {
            gopls = {
               analyses = { unusedparams = true },
               staticcheck = true,
            },
         },
         capabilities = capabilities,
      },
      lua_ls = {
         root_markers = { ".luarc.json", ".luarc.jsonc", ".git" },
         settings = {
            Lua = {
               runtime = { version = "LuaJIT" },
               diagnostics = {
                  globals = { "bit", "vim", "it", "describe", "before_each", "after_each" },
               },
               workspace = {
                  checkThirdParty = false,
               },
               telemetry = {
                  enable = false,
               },
            }
         },
         capabilities = capabilities,
      },
      eslint = {
         root_markers = { ".eslintrc.js", ".eslintrc.json", "eslint.config.js", "package.json", ".git" },
         settings = {
            format = { enable = true },
            codeActionOnSave = {
               enable = true,
               mode = "all"
            },
         },
         capabilities = capabilities,
      },
      ts_ls = {
         root_markers = { "package.json", "tsconfig.json", "jsconfig.json", ".git" },
         settings = {
            typescript = {
               format = { indentSize = 3, tabSize = 3 },
            },
            javascript = {
               format = { indentSize = 3, tabSize = 3 },
            },
         },
         capabilities = capabilities,
      },
      dockerls = {
         root_markers = { "Dockerfile", ".git" },
         capabilities = capabilities,
      },
      yamlls = {
         settings = {
            yaml = {
               schemas = {
                  ["https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json"] =
                  "/docker-compose.yml"
               }
            }
         },
         capabilities = capabilities,
      },
      zls = {
         root_markers = { "build.zig", "build.zig.zon", ".git" },
         settings = {
            zls = {
               enable_inlay_hints = true,
               enable_snippets = true,
               warn_style = true,
            },
         },
         capabilities = capabilities,
      },
      bashls = {
         root_markers = { ".git" },
         capabilities = capabilities,
      },
   }
end

-- Format on save setup
function common.setup_format_on_save()
   vim.api.nvim_create_autocmd("BufWritePre", {
      group = vim.api.nvim_create_augroup("LspFormat", { clear = true }),
      callback = function()
         vim.lsp.buf.format({
            async = false,
            filter = function(client)
               -- Prefer ESLint for JS/TS files, others for everything else
               local filetype = vim.bo.filetype
               if filetype == "javascript" or filetype == "typescript" or
                   filetype == "javascriptreact" or filetype == "typescriptreact" then
                  return client.name == "eslint"
               end
               return client.name ~= "eslint"
            end
         })
      end,
   })
end

-- Modern Neovim 0.11+ configuration
local function setup_modern_lsp()
   return {
      {
         "williamboman/mason.nvim",
         config = common.setup_mason
      },
      {
         "williamboman/mason-lspconfig.nvim",
         dependencies = { "williamboman/mason.nvim" },
         config = function()
            require("mason-lspconfig").setup({
               ensure_installed = {
                  "ts_ls", "eslint", "lua_ls", "bashls", "gopls", "dockerls", "yamlls", "zls"
               },
               automatic_installation = true,
            })
         end
      },
      {
         "hrsh7th/nvim-cmp",
         dependencies = {
            "hrsh7th/cmp-nvim-lsp", "hrsh7th/cmp-buffer", "hrsh7th/cmp-path",
            "L3MON4D3/LuaSnip", "saadparwaiz1/cmp_luasnip",
         },
         config = common.setup_cmp
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
         "neovim/nvim-lspconfig",
         dependencies = { "williamboman/mason.nvim", "hrsh7th/nvim-cmp" },
         config = function()
            local ok_cmp, cmp_lsp = pcall(require, 'cmp_nvim_lsp')
            if not ok_cmp then
               vim.notify("Failed to load cmp_nvim_lsp", vim.log.levels.ERROR)
               return
            end

            local capabilities = cmp_lsp.default_capabilities()

            -- Modern vim.lsp.config setup
            vim.lsp.config['*'] = { capabilities = capabilities }

            local server_configs = common.get_server_configs(capabilities)
            for server, config in pairs(server_configs) do
               vim.lsp.config[server] = config
            end

            -- Safely enable servers (only available in 0.11+)
            if vim.lsp.enable then
               local servers = { 'gopls', 'lua_ls', 'eslint', 'ts_ls', 'dockerls', 'yamlls', 'zls', 'bashls' }
               local ok_enable, err = pcall(vim.lsp.enable, servers)
               if not ok_enable then
                  vim.notify("Failed to enable LSP servers: " .. tostring(err), vim.log.levels.WARN)
               end
            end

            vim.lsp.set_log_level("WARN")
            common.setup_diagnostics()
            common.setup_format_on_save()

            vim.g.zig_fmt_parse_errors = 0
            vim.g.zig_fmt_autosave = 0
         end
      }
   }
end

-- Legacy Neovim <0.11 configuration
local function setup_legacy_lsp()
   return {
      "neovim/nvim-lspconfig",
      dependencies = {
         "williamboman/mason.nvim",
         "williamboman/mason-lspconfig.nvim",
         "hrsh7th/cmp-nvim-lsp", "hrsh7th/cmp-buffer", "hrsh7th/cmp-path", "hrsh7th/cmp-cmdline",
         "hrsh7th/nvim-cmp", "L3MON4D3/LuaSnip", "saadparwaiz1/cmp_luasnip",
         "j-hui/fidget.nvim", "jose-elias-alvarez/null-ls.nvim", "jay-babu/mason-null-ls.nvim",
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
         local cmp_lsp = require("cmp_nvim_lsp")
         local null_ls = require("null-ls")

         local capabilities = vim.tbl_deep_extend(
            "force", {}, vim.lsp.protocol.make_client_capabilities(), cmp_lsp.default_capabilities()
         )

         common.setup_mason()

         require("mason-null-ls").setup({
            automatic_installation = true,
            ensure_installed = { "gofmt", "prettierd" },
         })

         require("mason-lspconfig").setup({
            automatic_installation = true,
            ensure_installed = { "eslint", "lua_ls", "gopls", "dockerls", "yamlls" },
            handlers = {
               function(server_name)
                  local server_configs = common.get_server_configs(capabilities)
                  local config = server_configs[server_name] or { capabilities = capabilities }
                  require("lspconfig")[server_name].setup(config)
               end,
            }
         })

         -- Manual setup for servers not handled by mason-lspconfig
         local lspconfig = require("lspconfig")
         lspconfig.zls.setup({
            root_dir = lspconfig.util.root_pattern(".git", "build.zig", "zls.json"),
            capabilities = capabilities,
            settings = {
               zls = {
                  enable_inlay_hints = true,
                  enable_snippets = true,
                  warn_style = true,
               },
            },
         })

         null_ls.setup({
            sources = {
               null_ls.builtins.formatting.gofmt,
               null_ls.builtins.formatting.prettierd.with({
                  filetypes = { "json", "yaml", "typescript", "html", "markdown" },
                  extra_args = { "--ignore-path", "/dev/null", "--ignore-patterns", "%[(.-)%]" },
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

         common.setup_cmp()
         common.setup_diagnostics()

         vim.g.zig_fmt_parse_errors = 0
         vim.g.zig_fmt_autosave = 0
      end
   }
end

-- Return appropriate configuration based on Neovim version
if vim.fn.has('nvim-0.11') == 1 then
   return setup_modern_lsp()
else
   return setup_legacy_lsp()
end

