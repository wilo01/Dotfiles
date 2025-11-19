local LSP_SERVERS = {
   'gopls', 'lua_ls', 'eslint', 'ts_ls',
   'dockerls', 'yamlls', 'zls', 'bashls',
   'pyright', 'ruff'
}

local CONFIG = {
   INDENT_SIZE = 3,
   FORMAT_TIMEOUT_MS = 5000,
   EXCLUDED_FORMAT_FILETYPES = { "markdown", "text", "gitcommit" },
}

return {
   {
      "williamboman/mason.nvim",
      config = function()
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
   },
   {
      "williamboman/mason-lspconfig.nvim",
      dependencies = { "williamboman/mason.nvim" },
      config = function()
         require("mason-lspconfig").setup({
            ensure_installed = LSP_SERVERS,
            automatic_installation = true,
         })
      end
   },
   {
      "WhoIsSethDaniel/mason-tool-installer.nvim",
      dependencies = { "williamboman/mason.nvim" },
      config = function()
         require("mason-tool-installer").setup({
            ensure_installed = {
               "golangci-lint",
               "gofumpt",
               "goimports",
            },
            auto_update = false,
            run_on_start = true,
            start_delay = 3000,
         })
      end
   },
   {
      "hrsh7th/nvim-cmp",
      dependencies = {
         "hrsh7th/cmp-nvim-lsp", "hrsh7th/cmp-buffer", "hrsh7th/cmp-path",
         "L3MON4D3/LuaSnip", "saadparwaiz1/cmp_luasnip",
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
            { path = "luvit-meta/library", words = { "vim%.uv" } },
            { path = "lazy.nvim",          words = { "Lazy" } },
         },
         integrations = {
            lspconfig = true,
            cmp = true,
         },
      }
   },
   {
      "neovim/nvim-lspconfig",
      dependencies = { "williamboman/mason.nvim", "hrsh7th/nvim-cmp" },
      config = function()
         local capabilities = vim.lsp.protocol.make_client_capabilities()
         local ok, cmp_lsp = pcall(require, 'cmp_nvim_lsp')
         if ok then
            capabilities = cmp_lsp.default_capabilities()
         end

         vim.lsp.config['*'] = { capabilities = capabilities }
         vim.lsp.config.gopls = {
            root_markers = { "go.work", "go.mod", ".git" },
            settings = {
               gopls = {
                  analyses = {
                     unusedparams = true,
                     unusedwrite = true,
                     unusedresult = true,
                     shadow = true,
                     nilness = true,
                     useany = true,
                     ST1003 = true,
                     ST1016 = true,
                     SA4017 = true,
                     SA5001 = true,
                  },
                  staticcheck = true,
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
         }
         vim.lsp.config.lua_ls = {
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
         }
         vim.lsp.config.eslint = {
            root_markers = { ".eslintrc.js", ".eslintrc.json", "eslint.config.js", "package.json", ".git" },
            settings = {
               format = { enable = true },
            },
         }

         vim.lsp.config.ts_ls = {
            root_markers = { "package.json", "tsconfig.json", "jsconfig.json", ".git" },
            settings = {
               typescript = {
                  format = { indentSize = CONFIG.INDENT_SIZE, tabSize = CONFIG.INDENT_SIZE },
                  inlayHints = {
                     includeInlayParameterNameHints = "all",
                     includeInlayParameterNameHintsWhenArgumentMatchesName = true,
                     includeInlayFunctionParameterTypeHints = true,
                     includeInlayVariableTypeHints = true,
                     includeInlayPropertyDeclarationTypeHints = true,
                     includeInlayFunctionLikeReturnTypeHints = true,
                     includeInlayEnumMemberValueHints = true,
                  },
               },
               javascript = {
                  format = { indentSize = CONFIG.INDENT_SIZE, tabSize = CONFIG.INDENT_SIZE },
                  inlayHints = {
                     includeInlayParameterNameHints = "all",
                     includeInlayParameterNameHintsWhenArgumentMatchesName = true,
                     includeInlayFunctionParameterTypeHints = true,
                     includeInlayVariableTypeHints = true,
                     includeInlayPropertyDeclarationTypeHints = true,
                     includeInlayFunctionLikeReturnTypeHints = true,
                     includeInlayEnumMemberValueHints = true,
                  },
               },
            },
         }

         vim.lsp.config.dockerls = {
            root_markers = { "Dockerfile", ".git" },
         }
         vim.lsp.config.yamlls = {
            root_markers = { ".git", "package.json", "docker-compose.yml" },
            settings = {
               yaml = {
                  schemas = {
                     ["https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json"] =
                     "/docker-compose.yml"
                  }
               }
            },
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
            root_markers = { ".bashrc", ".bash_profile", "scripts/", ".git" },
            filetypes = { "sh", "bash" },
         }

         vim.lsp.config.pyright = {
            root_markers = { "pyproject.toml", "setup.py", "setup.cfg", "requirements.txt", ".git" },
            settings = {
               python = {
                  analysis = {
                     typeCheckingMode = "basic",
                     autoImportCompletions = true,
                  }
               }
            },
         }
         vim.lsp.config.ruff = {
            root_markers = { "pyproject.toml", "ruff.toml", ".ruff.toml", ".git" },
         }
         local log_level = vim.env.NVIM_LSP_LOG_LEVEL or "WARN"
         vim.lsp.set_log_level(log_level)

         vim.diagnostic.config({
            virtual_text = {
               spacing = 3,
               prefix = "●",
               severity = {
                  min = vim.diagnostic.severity.HINT,
               },
            },
            virtual_lines = false,
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
         vim.api.nvim_create_autocmd("LspAttach", {
            group = vim.api.nvim_create_augroup("UserLspConfig", { clear = true }),
            callback = function(args)
               local bufnr = args.buf
               local client = vim.lsp.get_client_by_id(args.data.client_id)

               if not client then return end

               if client:supports_method("textDocument/inlayHint") then
                  vim.lsp.inlay_hint.enable(true, { bufnr = bufnr })
               end

               if client:supports_method("textDocument/completion") then
                  vim.lsp.completion.enable(true, client.id, bufnr, { autotrigger = true })
               end

               if client:supports_method("textDocument/semanticTokens") then
                  vim.lsp.semantic_tokens.start(bufnr, client.id)
               end

               if client:supports_method("textDocument/codeLens") then
                  vim.api.nvim_create_autocmd({ "BufEnter", "CursorHold", "InsertLeave" }, {
                     group = vim.api.nvim_create_augroup("LspCodelens", { clear = false }),
                     buffer = bufnr,
                     callback = function()
                        vim.lsp.codelens.refresh({ bufnr = bufnr })
                     end,
                  })
               end

               if client.name == "ruff" then
                  client.server_capabilities.hoverProvider = false
               end
            end
         })

         vim.api.nvim_create_autocmd("BufWritePre", {
            group = vim.api.nvim_create_augroup("LspFormat", { clear = true }),
            callback = function(args)
               local bufnr = args.buf
               local filetype = vim.bo[bufnr].filetype

               if vim.tbl_contains(CONFIG.EXCLUDED_FORMAT_FILETYPES, filetype) then
                  return
               end

               local clients = vim.lsp.get_clients({ bufnr = bufnr })
               if #clients == 0 then
                  return
               end

               local ok, gitsigns = pcall(require, "gitsigns")
               if not ok then
                  return
               end

               local hunks = gitsigns.get_hunks(bufnr)
               if not hunks or #hunks == 0 then
                  return
               end

               local supports_range_format = false
               for _, client in ipairs(clients) do
                  if client:supports_method("textDocument/rangeFormatting", bufnr) then
                     supports_range_format = true
                     break
                  end
               end
               if supports_range_format then
                  for _, hunk in ipairs(hunks) do
                     if hunk.added and hunk.added.count > 0 then
                        local start_line = hunk.added.start
                        local end_line = start_line + hunk.added.count - 1

                        local end_col = #vim.api.nvim_buf_get_lines(bufnr, end_line - 1, end_line, false)[1]
                        vim.lsp.buf.format({
                           bufnr = bufnr,
                           async = false,
                           range = {
                              ["start"] = { start_line, 0 },
                              ["end"] = { end_line, end_col }
                           },
                           timeout_ms = CONFIG.FORMAT_TIMEOUT_MS,
                           filter = function(client)
                              return client:supports_method("textDocument/rangeFormatting", bufnr)
                           end
                        })
                     end
                  end
               else
                  vim.lsp.buf.format({
                     bufnr = bufnr,
                     async = false,
                     timeout_ms = CONFIG.FORMAT_TIMEOUT_MS,
                     filter = function(client)
                        return client:supports_method("textDocument/formatting", bufnr)
                            and client.name ~= "ruff"
                     end
                  })
               end
            end
         })

         if vim.lsp.enable then
            local ok_enable, err = pcall(vim.lsp.enable, LSP_SERVERS)
            if not ok_enable then
               vim.notify("Failed to enable LSP servers: " .. tostring(err), vim.log.levels.WARN)
            end
         end

         vim.g.zig_fmt_parse_errors = 0
         vim.g.zig_fmt_autosave = 0
      end
   },
   {
      "mfussenegger/nvim-lint",
      event = { "BufReadPre", "BufNewFile" },
      dependencies = { "williamboman/mason.nvim" },
      config = function()
         local lint = require('lint')
         lint.linters_by_ft = {
            go = { 'golangcilint' },
         }
         local mason_path = vim.fn.stdpath("data") .. "/mason/bin/golangci-lint"
         local golangci_path = vim.fn.exepath("golangci-lint")
         if golangci_path == "" and vim.fn.executable(mason_path) == 1 then
            golangci_path = mason_path
         end

         lint.linters.golangcilint = {
            cmd = golangci_path ~= "" and golangci_path or mason_path,
            stdin = false,
            args = {
               'run',
               '--out-format', 'json',
               '--issues-exit-code=0',
            },
            stream = 'stdout',
            ignore_exitcode = true,
            parser = function(output, bufnr)
               local diagnostics = {}

               if output == nil or output == "" then
                  return diagnostics
               end
               local ok, decoded = pcall(vim.json.decode, output)
               if not ok or not decoded then
                  return diagnostics
               end

               if decoded.Issues then
                  for _, issue in ipairs(decoded.Issues) do
                     local severity = vim.diagnostic.severity.WARN
                     if issue.Severity == "error" then
                        severity = vim.diagnostic.severity.ERROR
                     elseif issue.Severity == "warning" then
                        severity = vim.diagnostic.severity.WARN
                     elseif issue.Severity == "info" then
                        severity = vim.diagnostic.severity.INFO
                     end
                     local line = (issue.Pos and issue.Pos.Line) or issue.Line or 1
                     local col = (issue.Pos and issue.Pos.Column) or issue.Column or 1
                     table.insert(diagnostics, {
                        bufnr = bufnr,
                        lnum = line - 1,
                        col = col - 1,
                        message = string.format("[%s] %s", issue.FromLinter, issue.Text or issue.Message),
                        severity = severity,
                        source = "golangci-lint",
                        code = issue.FromLinter,
                     })
                  end
               end

               return diagnostics
            end,
         }
         vim.api.nvim_create_autocmd({ "BufWritePost", "BufReadPost", "InsertLeave" }, {
            group = vim.api.nvim_create_augroup("nvim-lint", { clear = true }),
            callback = function()
               local bufnr = vim.api.nvim_get_current_buf()
               if vim.api.nvim_buf_is_valid(bufnr) and vim.fn.filereadable(vim.api.nvim_buf_get_name(bufnr)) == 1 then
                  lint.try_lint()
               end
            end,
         })
      end
   }
}
