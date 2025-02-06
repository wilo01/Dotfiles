return {
   "L3MON4D3/LuaSnip",
   version = "v2.*",
   build = "make install_jsregexp",

   dependencies = { "rafamadriz/friendly-snippets" },

   config = function()
      local ls = require("luasnip")
      local snip = ls.snippet
      local func = ls.function_node

      ls.filetype_extend("javascript", { "jsdoc" })

      ls.config.set_config {
         history = true,
         ext_base_prio = 200,
         ext_prio_increase = 1,
         enable_autosnippets = true,
         store_selection_keys = "<TAB>",
      }

      local date = function()
         return { os.date "%d.%m.%Y" }
      end

      local filename = function()
         return { vim.fn.expand "%:p" }
      end

      -- Make sure to not pass an invalid command, as io.popen() may write over nvim-text.
      local function bash(_, _, command)
         local file = io.popen(command, "r")
         local res = {}
         for line in file:lines() do
            table.insert(res, line)
         end
         return res
      end

      --- TODO: What is expand?
      vim.keymap.set({ "i" }, "<C-s>e", function()
         ls.expand()
      end, { desc = "Expand snippet", silent = true })

      vim.keymap.set({ "i", "s" }, "<C-s>;", function()
         ls.jump(1)
      end, { desc = "Jump forward in snippet", silent = true })

      vim.keymap.set({ "i", "s" }, "<C-s>,", function()
         ls.jump(-1)
      end, { desc = "Jump backward in snippet", silent = true })

      vim.keymap.set({ "i", "s" }, "<C-E>", function()
         if ls.choice_active() then
            ls.change_choice(1)
         end
      end, { desc = "Cycle through snippet choices", silent = true })

      ls.add_snippets(nil, {
         all = {
            ls.parser.parse_snippet("log_dbms", "DBMS_OUTPUT.PUT_LINE('Dwdw ${1}: ' || ${1});${2}"),
            ls.parser.parse_snippet("log_pak", "ca_log_pak.log_warning('Dwdw', '${1}: ' || ${1});${2}"),
            snip({
               trig = "date",
               namr = "Date",
               dscr = "Date in the form of DD.MM.YYYY",
            }, {
               func(date, {}),
            }),
            snip({
               trig = "pwd",
               namr = "PWD",
               dscr = "Path to current working directory",
            }, {
               func(bash, {}, { user_args = { "pwd" } }),
            }),
            snip({
               trig = "filename",
               namr = "Filename",
               dscr = "Absolute path to file",
            }, {
               func(filename, {}),
            })
         },
         javascript = {
            ls.parser.parse_snippet("clwo", "console.warn({\n\t'${1}': ${1}\n});${2}"),
            ls.parser.parse_snippet("clw", "console.warn('${1}', ${1})"),
            ls.parser.parse_snippet("clg", "console.log('${1}');${2}"),
            ls.parser.parse_snippet("clo", "console.log('${1}Obj', ${2}Obj);${3}"),
            ls.parser.parse_snippet("ccl", "console.clear();${1}"),
            ls.parser.parse_snippet("cer", "console.error('${1}');${2}"),
            ls.parser.parse_snippet("ctr", "console.trace();${1}"),
            ls.parser.parse_snippet("clt", "console.table('${1}');${2}"),
            ls.parser.parse_snippet("cin", "console.info('${1}');${2}"),
            ls.parser.parse_snippet("cco", "console.count('${1}');${2}"),
            ls.parser.parse_snippet("db", "debugger;${1}"),
            ls.parser.parse_snippet("deb", "debugger;${1}"),
            ls.parser.parse_snippet("if", "if (${1}) {}${2}"),
            ls.parser.parse_snippet("log_warn", "console.warn('${1}', ${1})"),
            ls.parser.parse_snippet("log_warn_obj", "console.warn({\n\t'${1}': ${1}\n});${2}"),
            ls.parser.parse_snippet("tryc", "try {\n\t${1}\n} catch (error) {\n\tconsole.error('An error occurred:', error);\n}")
         },
         sh = {
            ls.parser.parse_snippet("shebang", "#!/bin/sh\n${0}"),
         },
         python = {
            ls.parser.parse_snippet("shebang", "#!/usr/bin/env python3\n\n${0}"),
         },
         lua = {
            ls.parser.parse_snippet("shebang", "#!/usr/bin/lua\n\n${0}"),
            ls.parser.parse_snippet("req", "require('${1:Module-name}')\n${0}"),
            ls.parser.parse_snippet("func", "function(${1:Arguments})\n\t${2}\nend\n${0}"),
            ls.parser.parse_snippet("forp", "for ${1:k}, ${2:v} in pairs(${3:table}) do\n\t${4}\nend\n${0}"),
            ls.parser.parse_snippet("fori", "for ${1:k}, ${2:v} in ipairs(${3:table}) do\n\t${4}\nend\n${0}"),
            ls.parser.parse_snippet("if", "if ${1} then\n\t${2}\nend\n${0}"),
            ls.parser.parse_snippet("M", "local M = {}\n${0}\nreturn M"),
         },
         markdown = {
            ls.parser.parse_snippet("img", "![${1:alt text}](${2:image_url})\n${0}"),
            ls.parser.parse_snippet("image", "![${1:alt text}](${2:image_url})\n${0}"),
            ls.parser.parse_snippet("link", "[${1:link text}](${2:url})\n${0}"),
            ls.parser.parse_snippet("url", "[${1:link text}](${2:url})\n${0}"),
            ls.parser.parse_snippet("codewrap", "```${1:Language}\n${2}\n```\n${0}"),
            ls.parser.parse_snippet("code", "```${1:Language}\n${2:Content}\n```\n${0}"),
         }
      })
   end,
}
