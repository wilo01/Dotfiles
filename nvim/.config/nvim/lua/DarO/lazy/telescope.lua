return {
   "nvim-telescope/telescope.nvim",
   tag = "0.1.5",
   dependencies = {
      "nvim-lua/plenary.nvim"
   },
   file_ignore_patterns = {
      "node_modules"
   },
   config = function()
      require('telescope').setup({})

      local builtin = require('telescope.builtin')
      vim.keymap.set('n', '<leader>pf', builtin.find_files, {})
      vim.keymap.set('n', '<C-p>', builtin.git_files, {})
      vim.keymap.set('n', '<leader>ws', function()
         local word = vim.fn.expand("<cword>")
         builtin.grep_string({ search = word })
      end)
      vim.keymap.set('n', '<leader>Ws', function()
         local word = vim.fn.expand("<cWORD>")
         builtin.grep_string({ search = word })
      end)
      vim.keymap.set('n', '<leader>s', function()
         builtin.grep_string({ search = vim.fn.input("Grep > ") })
      end)
      vim.keymap.set('n', '<leader>c', function()
         builtin.grep_string({ search = vim.fn.input("Grep > ", vim.fn.getreg('+')) });
      end)

      vim.keymap.set('n', '<leader>vh', builtin.help_tags, {})
   end
}
-- return {
--    "nvim-telescope/telescope.nvim",
--    tag = "0.1.5",
--    cmd = "Telescope",
--    dependencies = {
--       "danielfalk/smart-open.nvim",
--       "nvim-lua/plenary.nvim",
--       "nvim-tree/nvim-web-devicons",
--       {
--          "nvim-telescope/telescope-fzf-native.nvim",
--          build = function()
--             local install_path = vim.fn.stdpath("data") .. "/lazy/telescope-fzf-native.nvim"
--             vim.cmd("silent !cd " .. install_path .. " && make")
--          end,
--       },
--    },
--    file_ignore_patterns = {
--       "node_modules"
--    },
--    config = function()
--       local builtin = require('telescope.builtin')
--       local telescope = require("telescope")
--       local actions = require("telescope.actions")

--       local select_one_or_multi = function(prompt_bufnr)
--          local picker = require("telescope.actions.state").get_current_picker(prompt_bufnr)
--          local multi = picker:get_multi_selection()
--          if not vim.tbl_isempty(multi) then
--             require("telescope.actions").close(prompt_bufnr)
--             for _, j in pairs(multi) do
--                if j.path ~= nil then
--                   vim.cmd(string.format("%s %s", "edit", j.path))
--                end
--             end
--          else
--             require("telescope.actions").select_default(prompt_bufnr)
--          end
--       end

--       local home_dir = os.getenv("HOME")

--       telescope.setup({
--          extensions = {
--             resession = {
--                path_substitutions = {
--                   { find = home_dir .. "/git/", replace = "󰊢 " },
--                },
--             },

--             smart_open = {
--                match_algorithm = "fzf",
--             },
--          },
--          pickers = {
--             find_files = {
--                find_command = { "rg", "--files", "--hidden", "--glob", "!**/.git/*" },
--             },
--             buffers = {
--                theme = "dropdown",
--                previewer = false,
--                mappings = {
--                   i = {
--                      ["<c-d>"] = "delete_buffer",
--                   },
--                },
--             },
--          },
--          defaults = {
--             hidden = true,
--             prompt_prefix = "   ",
--             selection_caret = "  ",
--             entry_prefix = "  ",

--             sorting_strategy = "ascending",
--             layout_strategy = "horizontal",

--             layout_config = {
--                horizontal = {
--                   prompt_position = "top",
--                   preview_width = 0.55,
--                   results_width = 0.8,
--                },
--                vertical = {
--                   mirror = false,
--                },
--                width = 0.87,
--                height = 0.80,
--                preview_cutoff = 120,
--             },
--             path_display = {
--                filename_first = {
--                   reverse_directories = true,
--                },
--             },
--             mappings = {
--                i = {
--                   ["<esc>"] = actions.close,
--                   ["<C-k>"] = actions.move_selection_previous, -- move to prev result
--                   ["<C-j>"] = actions.move_selection_next,     -- move to next result
--                   ["<C-q>"] = actions.send_selected_to_qflist + actions.open_qflist,
--                   ["<cr>"] = select_one_or_multi,
--                },
--             },
--          },
--       })
--       vim.keymap.set('n', '<leader>pf', builtin.find_files, {})
--       vim.keymap.set('n', '<C-p>', builtin.git_files, {})
--       vim.keymap.set('n', '<leader>pws', function()
--          local word = vim.fn.expand("<cword>")
--          builtin.grep_string({ search = word })
--       end)
--       vim.keymap.set('n', '<leader>pWs', function()
--          local word = vim.fn.expand("<cWORD>")
--          builtin.grep_string({ search = word })
--       end)
--       vim.keymap.set('n', '<leader>ps', function()
--          builtin.grep_string({ search = vim.fn.input("Grep > ") })
--       end)
--       vim.keymap.set('n', '<leader>pc', function()
--          builtin.grep_string({ search = vim.fn.input("Grep > ", vim.fn.getreg('+')) });
--       end)

--       require("telescope").load_extension("fzf")
--       require("telescope").load_extension("smart_open")
--    end
-- }
