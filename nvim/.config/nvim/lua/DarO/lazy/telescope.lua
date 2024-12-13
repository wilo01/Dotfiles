return {
   "nvim-telescope/telescope.nvim",
   tag = "0.1.5",
   dependencies = {
      "nvim-lua/plenary.nvim",
      "nvim-telescope/telescope-file-browser.nvim",
      dependencies = { "nvim-telescope/telescope.nvim", "nvim-lua/plenary.nvim" }
   },
   config = function()
      local utils = require("DarO.utils")
      local telescope = require('telescope')
      local builtin = require('telescope.builtin')
      telescope.setup({
         defaults = {
            file_ignore_patterns = { "node_modules" },
            preview = {
               timeout = 2000,
            },
         },
      })
      telescope.load_extension('file_browser')

      local grep_opts = {
         auto_depth = true,
         follow_symlinks = true,
         grep_open_files = false -- if true, restrict search to open files only
      }

      vim.keymap.set('n', '<leader>pf', function()
         builtin.find_files()
      end, { desc = "Telescope Find files" })

      vim.keymap.set('n', '<C-p>', function()
         builtin.git_files()
      end, { desc = "Telescope Find Git files" })

      vim.keymap.set('n', '<leader>l', function()
         builtin.live_grep(grep_opts)
      end, { desc = "Telescope Live Grep" })

      vim.keymap.set('n', '<leader>ws', function()
         builtin.grep_string(vim.tbl_extend("force", grep_opts, {
            search = vim.fn.expand("<cword>")
         }))
      end, { desc = "Telescope Grep current word" })

      vim.keymap.set('n', '<leader>S', function()
         builtin.grep_string(vim.tbl_extend("force", grep_opts, {
            search = vim.fn.expand("<cWORD>")
         }))
      end, { desc = "Telescope Grep current WORD" })

      vim.keymap.set('n', '<leader>s', function()
         builtin.grep_string(vim.tbl_extend("force", grep_opts, {
            search = vim.fn.input("Grep > ")
         }))
      end, { desc = "Telescope Grep for input" })

      vim.keymap.set('n', '<leader>c', function()
         builtin.grep_string(vim.tbl_extend("force", grep_opts, {
            search = vim.fn.input("Grep > ", vim.fn.getreg('+'))
         }))
      end, { desc = "Telescope Grep for clipboard content" })

      vim.keymap.set('v', '<leader>c', function()
         vim.cmd('normal! "+y')
         builtin.grep_string(vim.tbl_extend("force", grep_opts, {
            search = vim.fn.input("Grep > ", vim.fn.getreg('+'))
         }))
      end, { desc = "Telescope Grep for visually selected text" })

      vim.keymap.set('n', '<leader>r', function()
         builtin.registers()
      end, { desc = "Telescope Search Registers" })

      vim.keymap.set('n', '<leader>b', function()
         builtin.buffers({
            sort_mru = true,
            sort_lastused = true,
            initial_mode = "insert",
            theme = "ivy"
         })
      end, { desc = "Telescope list buffers (buffer jumper, file jumps, last used files)" })

      vim.keymap.set('n', '<leader>ht', function()
         builtin.help_tags()
      end, { desc = "Telescope find help tags" })

      vim.keymap.set('n', '<leader>pb', function()
         telescope.extensions.file_browser.file_browser({
            layout_config = { width = 0.8, height = 0.89 },
            initial_mode = "normal",
            hidden = true,
            no_ignore = true
         })
      end, { desc = "Telescope open file browser" })

      vim.keymap.set("n", "<leader>km", function()
         builtin.keymaps()
      end, { desc = "Telescope show keymaps" })

      vim.keymap.set("n", "<leader>df", function()
         utils.telescope_diff_file()
      end, { desc = "Telescope diff file with current buffer" })

      vim.keymap.set("n", "<leader>dg", function()
         utils.telescope_diff_from_history()
      end, { desc = "Telescope diff from Git history" })
   end
}
