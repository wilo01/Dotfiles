return {
   "nvim-treesitter/nvim-treesitter",
   build = ":TSUpdate",
   event = { "BufReadPost", "BufNewFile" },
   cmd = { "TSUpdateSync", "TSUpdate", "TSInstall" },
   config = function()
      local ok, ts_config = pcall(require, "nvim-treesitter.configs")
      if not ok then
         vim.notify("Failed to load nvim-treesitter.configs: " .. ts_config, vim.log.levels.ERROR)
         return
      end

      local parser_config = require("nvim-treesitter.parsers").get_parser_configs()
      -- FIXME:Fields cannot be injected into the reference of `ParserInfo[]` for `templ`. To do so, use `---@class` for `parser_config`.
      parser_config.templ = {
         install_info = {
            url = "https://github.com/vrischmann/tree-sitter-templ.git",
            files = { "src/parser.c", "src/scanner.c" },
            branch = "master",
         },
         filetype = "templ",
      }

      local config = {
         ensure_installed = {
            "vimdoc", "javascript", "typescript", "c", "lua", "rust",
            "jsdoc", "bash", "markdown", "markdown_inline", "query",
            "vim", "html", "css", "json", "yaml", "python", "go",
            "templ"
         },

         sync_install = false,
         auto_install = true,
         modules = {},
         ignore_install = { "phpdoc" },

         highlight = {
            enable = true,
            additional_vim_regex_highlighting = false,
         },

         indent = {
            enable = true,
         },

         incremental_selection = {
            enable = true,
            keymaps = {
               init_selection = "<C-space>",
               node_incremental = "<C-space>",
               scope_incremental = false,
               node_decremental = "<bs>",
            },
         },

         textobjects = {
            select = {
               enable = true,
               lookahead = true,
               keymaps = {
                  ["af"] = "@function.outer",
                  ["if"] = "@function.inner",
                  ["ac"] = "@class.outer",
                  ["ic"] = "@class.inner",
               },
            },
            move = {
               enable = true,
               set_jumps = true,
               goto_next_start = {
                  ["]m"] = "@function.outer",
                  ["]]"] = "@class.outer",
               },
               goto_next_end = {
                  ["]M"] = "@function.outer",
                  ["]["] = "@class.outer",
               },
               goto_previous_start = {
                  ["[m"] = "@function.outer",
                  ["[["] = "@class.outer",
               },
               goto_previous_end = {
                  ["[M"] = "@function.outer",
                  ["[]"] = "@class.outer",
               },
            },
         },
      }

      ts_config.setup(config)

      vim.filetype.add({
         extension = {
            templ = "templ",
         },
      })

      vim.treesitter.language.register("templ", "templ")

      local function ensure_parser_installed(parser_name)
         local ok_parser, _ = pcall(vim.treesitter.get_parser, 0, parser_name)
         if not ok_parser then
            vim.schedule(function()
               vim.cmd("TSInstall " .. parser_name)
            end)
         end
      end

      vim.api.nvim_create_autocmd({ "BufRead", "BufNewFile" }, {
         callback = function()
            local ft = vim.bo.filetype
            if ft and ft ~= "" then
               ensure_parser_installed(ft)
            end
         end,
      })
   end,

   dependencies = {
      {
         "nvim-treesitter/nvim-treesitter-textobjects",
         event = "VeryLazy",
      },
      {
         "nvim-treesitter/nvim-treesitter-context",
         event = "VeryLazy",
         opts = {
            enable = true,
            max_lines = 0,
            min_window_height = 0,
            line_numbers = true,
            multiline_threshold = 20,
            trim_scope = 'outer',
            mode = 'cursor',
            separator = nil,
            zindex = 20,
         },
      },
   },
}
