local function setup_diagnostics()
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

local function setup_mason()
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

local function get_server_configs(capabilities)
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
      pyright = {
         root_markers = { "pyproject.toml", "setup.py", "setup.cfg", "requirements.txt", "Pipfile", ".git" },
         settings = {
            python = {
               analysis = {
                  autoImportCompletions = true,
                  typeCheckingMode = "standard",
                  diagnosticMode = "workspace",
                  useLibraryCodeForTypes = true,
                  autoSearchPaths = true,
                  diagnosticSeverityOverrides = {
                     reportMissingImports = "error",
                     reportUndefinedVariable = "error",
                     reportGeneralTypeIssues = "warning",
                     reportOptionalMemberAccess = "warning",
                     reportOptionalSubscript = "warning",
                     reportPrivateUsage = "warning",
                     reportUnusedImport = "information",
                     reportUnusedVariable = "information",
                  }
               }
            }
         },
         capabilities = capabilities,
      },
      ruff = {
         root_markers = { "pyproject.toml", "ruff.toml", ".ruff.toml", "setup.py", "setup.cfg", "requirements.txt", ".git" },
         settings = {
            init_options = {
               settings = {
                  args = {},
                  lint = {
                     enable = true,
                     run = "onType",
                  },
                  format = {
                     enable = true,
                  },
                  organizeImports = {
                     enable = true,
                  },
               }
            }
         },
         capabilities = capabilities,
      },
   }
end

local function setup_format_on_save()
   vim.api.nvim_create_autocmd("BufWritePre", {
      group = vim.api.nvim_create_augroup("LspFormat", { clear = true }),
      callback = function()
         vim.lsp.buf.format({
            async = false,
            filter = function(client)
               local filetype = vim.bo.filetype
               if filetype == "javascript" or filetype == "typescript" or
                   filetype == "javascriptreact" or filetype == "typescriptreact" then
                  return client.name == "eslint"
               end

               if filetype == "python" then
                  return client.name == "ruff"
               end
               return client.name ~= "eslint" and client.name ~= "ruff-lsp"
            end
         })
      end,
   })
end

return {
   {
      "williamboman/mason.nvim",
      config = setup_mason
   },
   {
      "williamboman/mason-lspconfig.nvim",
      dependencies = { "williamboman/mason.nvim" },
      config = function()
         require("mason-lspconfig").setup({
            ensure_installed = {
               "ts_ls", "eslint", "lua_ls", "bashls", "gopls", "dockerls", "yamlls", "zls", "pyright", "ruff"
            },
            automatic_installation = true,
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
      "neovim/nvim-lspconfig",
      dependencies = { "williamboman/mason.nvim", "hrsh7th/cmp-nvim-lsp" },
      config = function()
         local ok_cmp, cmp_lsp = pcall(require, 'cmp_nvim_lsp')
         if not ok_cmp then
            vim.notify("Failed to load cmp_nvim_lsp", vim.log.levels.ERROR)
            return
         end

         local capabilities = cmp_lsp.default_capabilities()

         vim.lsp.config['*'] = { capabilities = capabilities }

         local server_configs = get_server_configs(capabilities)
         for server, config in pairs(server_configs) do
            vim.lsp.config[server] = config
         end

         local servers = {
            'gopls', 'lua_ls', 'eslint', 'ts_ls', 'dockerls',
            'yamlls', 'zls', 'bashls', 'pyright', 'ruff'
         }
         local ok_enable, err = pcall(vim.lsp.enable, servers)
         if not ok_enable then
            vim.notify("Failed to enable LSP servers: " .. tostring(err), vim.log.levels.WARN)
         end

         vim.lsp.set_log_level("WARN")
         setup_diagnostics()
         setup_format_on_save()

         vim.g.zig_fmt_parse_errors = 0
         vim.g.zig_fmt_autosave = 0
      end
   }
}
