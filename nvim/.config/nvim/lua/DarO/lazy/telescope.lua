return {
   "nvim-telescope/telescope.nvim",
   tag = "0.1.5",
   dependencies = {
      "nvim-lua/plenary.nvim",
      "nvim-telescope/telescope-file-browser.nvim",
      dependencies = { "nvim-telescope/telescope.nvim", "nvim-lua/plenary.nvim" }
   },
   file_ignore_patterns = {
      "node_modules"
   },
   config = function()
      local telescope = require('telescope')
      telescope.setup({})
      telescope.load_extension('file_browser')

      local builtin = require('telescope.builtin')
      local grep_opts = {
         auto_depth = true,
         follow_symlinks = true,
         grep_open_files = true
      }

      vim.keymap.set('n', '<leader>pf', builtin.find_files, {})
      vim.keymap.set('n', '<C-p>', builtin.git_files, {})

      vim.keymap.set('n', '<leader>ws', function()
         builtin.grep_string(vim.tbl_extend("force", grep_opts, {
            search = vim.fn.expand("<cword>")
         }))
      end)

      vim.keymap.set('n', '<leader>Ws', function()
         builtin.grep_string(vim.tbl_extend("force", grep_opts, {
            search = vim.fn.expand("<cWORD>")
         }))
      end)

      vim.keymap.set('n', '<leader>s', function()
         builtin.grep_string(vim.tbl_extend("force", grep_opts, {
            search = vim.fn.input("Grep > ")
         }))
      end)

      vim.keymap.set('n', '<leader>c', function()
         builtin.grep_string(vim.tbl_extend("force", grep_opts, {
            search = vim.fn.input("Grep > ", vim.fn.getreg('+'))
         }))
      end)

      vim.keymap.set('v', '<leader>c', function()
         vim.cmd('normal! "+y')
         builtin.grep_string(vim.tbl_extend("force", grep_opts, {
            search = vim.fn.input("Grep > ", vim.fn.getreg('+'))
         }))
      end)

      vim.keymap.set('n', '<leader>r', function()
         builtin.buffers({
            sort_mru = true,
            sort_lastused = true,
            initial_mode = "insert",
            theme = "ivy"
         })
      end)

      vim.keymap.set('n', '<leader>ht', builtin.help_tags, {})

      vim.keymap.set('n', '<leader>pb', function()
         telescope.extensions.file_browser.file_browser({
            layout_config = { width = 0.8, height = 0.89 },
            initial_mode = "normal",
            hidden = true,
            no_ignore = true
         })
      end)
   end
}
