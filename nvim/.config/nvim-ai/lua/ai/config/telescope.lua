-- Telescope configuration
-- Optimized for performance with smart defaults

local telescope = require("telescope")
local actions = require("telescope.actions")
local builtin = require("telescope.builtin")

telescope.setup({
   defaults = {
      prompt_prefix = " ",
      selection_caret = " ",
      entry_prefix = "  ",
      multi_icon = " ",

      sorting_strategy = "ascending",
      layout_strategy = "flex",
      layout_config = {
         horizontal = {
            prompt_position = "top",
            preview_width = 0.55,
         },
         vertical = {
            mirror = false,
         },
         flex = {
            flip_columns = 120,
         },
         width = 0.87,
         height = 0.80,
         preview_cutoff = 120,
      },

      mappings = {
         i = {
            ["<C-n>"] = actions.cycle_history_next,
            ["<C-p>"] = actions.cycle_history_prev,
            ["<C-j>"] = actions.move_selection_next,
            ["<C-k>"] = actions.move_selection_previous,
            ["<C-c>"] = actions.close,
            ["<CR>"] = actions.select_default,
            ["<C-x>"] = actions.select_horizontal,
            ["<C-v>"] = actions.select_vertical,
            ["<C-t>"] = actions.select_tab,
            ["<C-u>"] = actions.preview_scrolling_up,
            ["<C-d>"] = actions.preview_scrolling_down,
            ["<PageUp>"] = actions.results_scrolling_up,
            ["<PageDown>"] = actions.results_scrolling_down,
            ["<Tab>"] = actions.toggle_selection + actions.move_selection_worse,
            ["<S-Tab>"] = actions.toggle_selection + actions.move_selection_better,
            ["<C-q>"] = actions.send_to_qflist + actions.open_qflist,
            ["<M-q>"] = actions.send_selected_to_qflist + actions.open_qflist,
            ["<C-l>"] = actions.complete_tag,
            ["<C-/>"] = actions.which_key,
         },
         n = {
            ["<esc>"] = actions.close,
            ["<CR>"] = actions.select_default,
            ["<C-x>"] = actions.select_horizontal,
            ["<C-v>"] = actions.select_vertical,
            ["<C-t>"] = actions.select_tab,
            ["<Tab>"] = actions.toggle_selection + actions.move_selection_worse,
            ["<S-Tab>"] = actions.toggle_selection + actions.move_selection_better,
            ["<C-q>"] = actions.send_to_qflist + actions.open_qflist,
            ["<M-q>"] = actions.send_selected_to_qflist + actions.open_qflist,
            ["j"] = actions.move_selection_next,
            ["k"] = actions.move_selection_previous,
            ["H"] = actions.move_to_top,
            ["M"] = actions.move_to_middle,
            ["L"] = actions.move_to_bottom,
            ["<Down>"] = actions.move_selection_next,
            ["<Up>"] = actions.move_selection_previous,
            ["gg"] = actions.move_to_top,
            ["G"] = actions.move_to_bottom,
            ["<C-u>"] = actions.preview_scrolling_up,
            ["<C-d>"] = actions.preview_scrolling_down,
            ["<PageUp>"] = actions.results_scrolling_up,
            ["<PageDown>"] = actions.results_scrolling_down,
            ["?"] = actions.which_key,
         },
      },

      vimgrep_arguments = {
         "rg",
         "--color=never",
         "--no-heading",
         "--with-filename",
         "--line-number",
         "--column",
         "--smart-case",
         "--trim", -- Remove indentation
      },

      file_ignore_patterns = {
         "%.git/",
         "%.cache",
         "%.min%.js",
         "%.min%.css",
         "node_modules/",
         "vendor/",
         "dist/",
         "build/",
         "__pycache__/",
         "%.pyc",
         "%.pyo",
         "%.pyd",
         "%.so",
         "%.dylib",
         "%.dll",
         "%.class",
         "%.exe",
         "%.o",
         "%.a",
         "%.lib",
         "%.png",
         "%.jpg",
         "%.jpeg",
         "%.gif",
         "%.svg",
         "%.ico",
         "%.pdf",
         "%.zip",
         "%.tar",
         "%.gz",
         "%.7z",
         "%.rar",
      },

      path_display = { "truncate" },
      winblend = 0,
      border = {},
      borderchars = { "─", "│", "─", "│", "╭", "╮", "╯", "╰" },
      color_devicons = true,
      use_less = true,
      set_env = { ["COLORTERM"] = "truecolor" },

      file_previewer = require("telescope.previewers").vim_buffer_cat.new,
      grep_previewer = require("telescope.previewers").vim_buffer_vimgrep.new,
      qflist_previewer = require("telescope.previewers").vim_buffer_qflist.new,

      -- Performance optimizations
      file_sorter = require("telescope.sorters").get_fuzzy_file,
      generic_sorter = require("telescope.sorters").get_generic_fuzzy_sorter,
      prefilter_sorter = require("telescope.sorters").prefilter,
   },

   pickers = {
      find_files = {
         find_command = { "fd", "--type", "f", "--strip-cwd-prefix", "--hidden", "--exclude", ".git" },
         hidden = true,
      },
      live_grep = {
         additional_args = function()
            return { "--hidden" }
         end,
      },
      buffers = {
         show_all_buffers = true,
         sort_lastused = true,
         theme = "dropdown",
         previewer = false,
         mappings = {
            i = {
               ["<c-d>"] = actions.delete_buffer,
            },
            n = {
               ["dd"] = actions.delete_buffer,
            },
         },
      },
      git_files = {
         show_untracked = true,
      },
      diagnostics = {
         theme = "ivy",
         initial_mode = "normal",
         layout_config = {
            preview_cutoff = 9999,
         },
      },
   },

   extensions = {
      fzf = {
         fuzzy = true,
         override_generic_sorter = true,
         override_file_sorter = true,
         case_mode = "smart_case",
      },
   },
})

-- Load extensions
telescope.load_extension("fzf")

-- Custom functions
local M = {}

-- Live grep with glob pattern support
M.live_multigrep = function(opts)
   opts = opts or {}
   opts.cwd = opts.cwd or vim.loop.cwd()

   local finder = require("telescope.finders").new_async_job({
      command_generator = function(prompt)
         if not prompt or prompt == "" then
            return nil
         end

         local pieces = vim.split(prompt, "  ")
         local args = { "rg" }

         if pieces[1] then
            table.insert(args, "-e")
            table.insert(args, pieces[1])
         end

         if pieces[2] then
            table.insert(args, "-g")
            table.insert(args, pieces[2])
         end

         return vim.tbl_flatten({
            args,
            { "--color=never", "--no-heading", "--with-filename", "--line-number", "--column", "--smart-case" },
         })
      end,
      entry_maker = require("telescope.make_entry").gen_from_vimgrep(opts),
   })

   require("telescope.pickers").new(opts, {
      prompt_title = "Live Grep (pattern  glob)",
      finder = finder,
      previewer = require("telescope.config").values.grep_previewer(opts),
      sorter = require("telescope.config").values.generic_sorter(opts),
   }):find()
end

-- Grep string under cursor
M.grep_string_visual = function()
   local visual_selection = function()
      local save_reg = vim.fn.getreg("v")
      vim.cmd('noau normal! "vy')
      local text = vim.fn.getreg("v")
      vim.fn.setreg("v", save_reg)
      text = string.gsub(text, "\n", "")
      if #text > 0 then
         return text
      else
         return ""
      end
   end
   builtin.grep_string({ search = visual_selection() })
end

-- Additional keymaps for Telescope
vim.keymap.set("n", "<leader>fm", M.live_multigrep, { desc = "Live multigrep" })
vim.keymap.set("v", "<leader>fg", M.grep_string_visual, { desc = "Grep visual selection" })

return M