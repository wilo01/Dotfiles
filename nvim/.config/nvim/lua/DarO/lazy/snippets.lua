return {
   {
      "hrsh7th/nvim-cmp",
      event = "InsertEnter",
      dependencies = {
         { "hrsh7th/cmp-buffer" },
         { "hrsh7th/cmp-path" },
         { "hrsh7th/cmp-nvim-lsp" },
      },
      config = function()
         local dynamic_cache = {}
         local CACHE_TTL = 5000

         local function cached_dynamic(fn, cache_key)
            return function()
               local now = vim.uv.now()
               local cached = dynamic_cache[cache_key]

               if cached and (now - cached.time) < CACHE_TTL then
                  return cached.value
               end

               local ok, result = pcall(fn)
               local value = ok and result or ""
               dynamic_cache[cache_key] = { value = value, time = now }
               return value
            end
         end

         local get_date = cached_dynamic(function()
            return os.date("%d.%m.%Y")
         end, "date")

         local get_timestamp = cached_dynamic(function()
            return os.date("%d.%m.%Y %H:%M:%S")
         end, "timestamp")

         local get_filename = cached_dynamic(function()
            return vim.fn.expand("%:p")
         end, "filename")

         local get_pwd = cached_dynamic(function()
            local cwd = vim.fn.getcwd()
            return cwd:gsub("([\\$`])", "\\%1")
         end, "pwd")

         local registry = {
            all = {},
            javascript = {},
            typescript = {},
            javascriptreact = {},
            typescriptreact = {},
            sh = {},
            python = {},
            lua = {},
            markdown = {},
            sql = {},
            plsql = {},
         }

         local function snip(filetype, trigger, body, description)
            if not registry[filetype] then
               registry[filetype] = {}
            end
            registry[filetype][trigger] = {
               trigger = trigger,
               body = body,
               description = description or trigger,
            }
         end

         local function setup_snippets()
            snip("all", "todo", "[ ] TODO: $0")
            snip("all", "todo-", "- [ ] TODO: $0")
            snip("all", "todo--", "-- [ ] TODO: $0")
            snip("all", "datetime", get_timestamp, "Insert current date and time")
            snip("all", "timestamp", get_timestamp, "Insert current timestamp")
            snip("all", "date", get_date, "Insert current date")
            snip("all", "pwd", get_pwd, "Insert current working directory")
            snip("all", "filename", get_filename, "Insert current filename with path")

            snip("sql", "log_dbms", "DBMS_OUTPUT.PUT_LINE('Dwdw ${1}: ' || ${1});$0")
            snip("sql", "log_pak", "ca_log_pak.log_warning('Dwdw', '${1}: ' || ${1});$0")
            snip("plsql", "log_dbms", "DBMS_OUTPUT.PUT_LINE('Dwdw ${1}: ' || ${1});$0")
            snip("plsql", "log_pak", "ca_log_pak.log_warning('Dwdw', '${1}: ' || ${1});$0")

            local js_filetypes = { "javascript", "typescript", "javascriptreact", "typescriptreact" }
            for _, ft in ipairs(js_filetypes) do
               snip(ft, "clwo", "console.warn('', {\n\t'${1}': ${1}\n});$0")
               snip(ft, "clw", "console.warn('${1}', ${1})")
               snip(ft, "clg", "console.log('${1}');$0")
               snip(ft, "clo", "console.log('${1}Obj', ${2}Obj);$0")
               snip(ft, "ccl", "console.clear();$0")
               snip(ft, "cer", "console.error('${1}');$0")
               snip(ft, "ctr", "console.trace();$0")
               snip(ft, "clt", "console.table('${1}');$0")
               snip(ft, "cin", "console.info('${1}');$0")
               snip(ft, "cco", "console.count('${1}');$0")
               snip(ft, "db", "debugger;$0")
               snip(ft, "deb", "debugger;$0")
               snip(ft, "if", "if (${1}) {\n\t$0\n}")
               snip(ft, "log_warn", "console.warn('${1}', ${1})")
               snip(ft, "log_warn_obj", "console.warn({\n\t'${1}': ${1}\n});$0")
               snip(ft, "tryc", "try {\n\t${1}\n} catch (error) {\n\tconsole.error('An error occurred:', error);\n}")
            end

            snip("sh", "shebang", "#!/bin/sh\n$0")
            snip("python", "shebang", "#!/usr/bin/env python3\n\n$0")

            snip("lua", "shebang", "#!/usr/bin/lua\n\n$0")
            snip("lua", "req", "require('${1:Module-name}')\n$0")
            snip("lua", "func", "function(${1:Arguments})\n\t${2}\nend\n$0")
            snip("lua", "forp", "for ${1:k}, ${2:v} in pairs(${3:table}) do\n\t${4}\nend\n$0")
            snip("lua", "fori", "for ${1:k}, ${2:v} in ipairs(${3:table}) do\n\t${4}\nend\n$0")
            snip("lua", "if", "if ${1} then\n\t${2}\nend\n$0")
            snip("lua", "M", "local M = {}\n$0\nreturn M")

            snip("markdown", "img", "![${2:alt text}](${1:})$0")
            snip("markdown", "image", "![${2:alt text}](${1:})$0")
            snip("markdown", "link", "[${2:link text}](${1:})$0")
            snip("markdown", "url", "[${2:link text}](${1:})$0")
            snip("markdown", "codewrap", "```${1:Language}\n${2}\n```\n$0")
            snip("markdown", "code", "```${2:Language}\n${1:}\n```\n$0")
         end

         local filetype_cache = {}

         local function get_buf_snips()
            local ft = vim.bo.filetype

            if not filetype_cache[ft] then
               local snips = {}

               if registry.all then
                  for _, snippet in pairs(registry.all) do
                     table.insert(snips, snippet)
                  end
               end

               if ft and registry[ft] then
                  for _, snippet in pairs(registry[ft]) do
                     table.insert(snips, snippet)
                  end
               end

               filetype_cache[ft] = snips
            end

            return filetype_cache[ft]
         end

         local function register_cmp_source()
            local cmp_source = {}
            local cache = setmetatable({}, { __mode = 'k' })

            function cmp_source.complete(_, _, callback)
               local bufnr = vim.api.nvim_get_current_buf()
               if not cache[bufnr] then
                  local completion_items = vim.tbl_map(function(s)
                     local body = s.body
                     if type(body) == "function" then
                        body = body()
                     end

                     local item = {
                        word = s.trigger,
                        label = s.trigger,
                        kind = vim.lsp.protocol.CompletionItemKind.Snippet,
                        insertText = body,
                        insertTextFormat = vim.lsp.protocol.InsertTextFormat.Snippet,
                     }
                     return item
                  end, get_buf_snips())

                  cache[bufnr] = completion_items
               end

               callback(cache[bufnr])
            end

            local augroup = vim.api.nvim_create_augroup('SnippetCache', { clear = true })
            vim.api.nvim_create_autocmd({ "BufEnter", "FileType", "BufDelete" }, {
               group = augroup,
               callback = function(args)
                  cache[args.buf] = nil

                  if args.event == "FileType" then
                     filetype_cache[vim.bo[args.buf].filetype] = nil
                  end
               end,
            })

            require("cmp").register_source("native_snippets", cmp_source)
         end

         vim.opt.completeopt = { "menu", "menuone", "noselect" }

         local cmp = require("cmp")
         local select_opts = { behavior = cmp.SelectBehavior.Insert }

         setup_snippets()

         register_cmp_source()

         cmp.setup({
            snippet = {
               expand = function(args)
                  vim.snippet.expand(args.body)
               end,
            },
            sources = {
               { name = "path" },
               { name = "nvim_lsp",        keyword_length = 3 },
               { name = "buffer",          keyword_length = 3 },
               { name = "native_snippets", keyword_length = 1 },
            },
            window = {
               documentation = cmp.config.window.bordered(),
            },
            formatting = {
               fields = { "menu", "abbr", "kind" },
               format = function(entry, item)
                  local menu_icon = {
                     nvim_lsp = "",
                     native_snippets = "",
                     buffer = "",
                     path = "",
                  }
                  item.menu = menu_icon[entry.source.name]
                  return item
               end,
            },
            mapping = {
               ["<Up>"] = cmp.mapping.select_prev_item(select_opts),
               ["<Down>"] = cmp.mapping.select_next_item(select_opts),
               ["<C-k>"] = cmp.mapping.select_prev_item(select_opts),
               ["<C-j>"] = cmp.mapping.select_next_item(select_opts),
               ["<C-u>"] = cmp.mapping.scroll_docs(-4),
               ["<C-f>"] = cmp.mapping.scroll_docs(4),
               ["<C-e>"] = cmp.mapping.abort(),
               ["<CR>"] = cmp.mapping.confirm({ select = false }),

               ["<C-d>"] = cmp.mapping(function(fallback)
                  if vim.snippet.active({ direction = 1 }) then
                     vim.snippet.jump(1)
                  else
                     fallback()
                  end
               end, { "i", "s" }),

               ["<C-b>"] = cmp.mapping(function(fallback)
                  if vim.snippet.active({ direction = -1 }) then
                     vim.snippet.jump(-1)
                  else
                     fallback()
                  end
               end, { "i", "s" }),

               ["<Tab>"] = cmp.mapping(function(fallback)
                  local col = vim.fn.col(".") - 1
                  if cmp.visible() then
                     cmp.select_next_item(select_opts)
                  elseif vim.snippet.active({ direction = 1 }) then
                     vim.snippet.jump(1)
                  elseif col == 0 or vim.fn.getline("."):sub(col, col):match("%s") then
                     fallback()
                  else
                     cmp.complete()
                  end
               end, { "i", "s" }),

               ["<S-Tab>"] = cmp.mapping(function(fallback)
                  if cmp.visible() then
                     cmp.select_prev_item(select_opts)
                  elseif vim.snippet.active({ direction = -1 }) then
                     vim.snippet.jump(-1)
                  else
                     fallback()
                  end
               end, { "i", "s" }),
            },
         })

         vim.keymap.set({ "i" }, "<C-s>e", function()
            local current_word = vim.fn.expand('<cword>')
            local filetype = vim.bo.filetype
            local snippet = nil

            if registry[filetype] and registry[filetype][current_word] then
               snippet = registry[filetype][current_word]
            elseif registry.all[current_word] then
               snippet = registry.all[current_word]
            end

            if snippet then
               vim.cmd('normal! diw')
               local body = snippet.body
               if type(body) == "function" then
                  body = body()
               end
               vim.snippet.expand(body)
            else
               vim.notify("No snippet found for: " .. current_word, vim.log.levels.INFO)
            end
         end, { desc = "Expand snippet under cursor", silent = true })

         vim.keymap.set({ "i", "s" }, "<C-s>;", function()
            if vim.snippet.active({ direction = 1 }) then
               vim.snippet.jump(1)
            end
         end, { desc = "Jump forward in snippet", silent = true })

         vim.keymap.set({ "i", "s" }, "<C-s>,", function()
            if vim.snippet.active({ direction = -1 }) then
               vim.snippet.jump(-1)
            end
         end, { desc = "Jump backward in snippet", silent = true })

         vim.keymap.set({ "i", "s" }, "<C-l>", function()
            if vim.snippet.active() then
               vim.snippet.stop()
            end
         end, { desc = "Stop snippet", silent = true })
      end,
   },
}
