return {
   "nvim-telescope/telescope.nvim",
   tag = "0.1.8",
   dependencies = {
      "nvim-lua/plenary.nvim",
      "nvim-telescope/telescope-file-browser.nvim",
      { "nvim-telescope/telescope-fzf-native.nvim", build = 'make' },
      dependencies = { "nvim-telescope/telescope.nvim", "nvim-lua/plenary.nvim" }
   },
   config = function()
      local utils = require("DarO.utils")
      local telescope = require('telescope')
      local builtin = require('telescope.builtin')
      local previewers = require('telescope.previewers')

      -- Live multigrep section start
      local pickers = require("telescope.pickers")
      local finders = require("telescope.finders")
      local make_entry = require("telescope.make_entry")
      local conf = require("telescope.config").values

      local function live_multigrep(opts)
         opts = opts or {}
         opts.cwd = opts.cwd or vim.uv.cwd() -- [ ] TODO: Undefined field `cwd`.

         local finder = finders.new_async_job {
            command_generator = function(prompt)
               if not prompt or prompt == "  " then
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

               ---@diagnostic disable-next-line: deprecated
               return vim.tbl_flatten {
                  args,
                  { "--color=never", "--no-heading", "--with-filename", "--line-number", "--column", "--smart-case" },
               }
            end,
            entry_maker = make_entry.gen_from_vimgrep(opts),
            cwd = opts.cwd,
         }

         pickers.new(opts, {
            debounce = 100,
            prompt_title = "Multi Grep",
            finder = finder,
            previewer = conf.grep_previewer(opts),
            sorter = require("telescope.sorters").empty(),
         }):find()
      end
      -- Live multigrep section end

      telescope.setup({
         defaults = {
            -- Example: { "^scratch/" } -- ignore all files in scratch directory
            -- Example: { "%.npz" } -- ignore all npz files
            file_ignore_patterns = {
               "node_modules",
               "ext%-all%.js$",
               "ext%-all%-debug%.js$",
               "ext%-all%-rtl%.js$",
               "ext%-all%-rtl%-debug%.js$",
               "source/ui/api/bundle%.js",
               -- "source/ui%-student%-portal",
               "source/ui%-muster",
               -- "source/server/database/sql/student/",
               "source/server/rtmuster",
               "test/Cypress/cypress/e2e/safe/"
            },
            layout_strategy = 'vertical',
            layout_config = { width = 0.9, height = 0.89 },
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

      local delta = previewers.new_termopen_previewer({
         get_command = function(entry)
            local status = entry.status or ""

            local function get_git_command()
               if status:match("^[AMD]") then
                  return { 'git', 'diff', '--cached', '--color=always', '--', entry.value }
               end

               if status:match("^.[MD]") then
                  return { 'git', 'diff', '--color=always', '--', entry.value }
               end

               if status == '??' then
                  return { 'git', 'diff', '--no-index', '--color=always', '/dev/null', entry.value }
               end

               return { 'git', 'diff', '--color=always', 'HEAD', '--', entry.value }
            end

            local cmd = get_git_command()

            if status == '??' and vim.fn.executable('delta') ~= 1 then
               if vim.fn.executable('bat') == 1 then
                  return { 'bat', '--style=numbers', '--color=always', entry.value }
               end
               return { 'cat', entry.value }
            end

            if vim.fn.executable('delta') == 1 then
               return { 'sh', '-c', table.concat(cmd, ' ') .. ' | delta' }
            end

            return cmd
         end
      })

      vim.keymap.set('n', '<leader>l', function()
         live_multigrep()
      end, { desc = "Telescope Live Multi Grep" })

      vim.keymap.set('n', '<leader>lg', function()
         builtin.live_grep(grep_opts)
      end, { desc = "Telescope Live Grep" })

      vim.keymap.set('n', '<leader>pf', function()
         builtin.find_files()
      end, { desc = "Telescope Find files" })

      vim.keymap.set('n', '<C-p>', function()
         local search_dir

         if utils.is_git_repo() then
            search_dir = vim.fn.systemlist("git rev-parse --show-toplevel")[1]
         else
            search_dir = vim.fn.expand('%:p:h')
         end

         builtin.find_files({
            cwd = search_dir,
            hidden = true,
            no_ignore = true,
            follow = true,
         })
      end, { desc = "Telescope Find all files (smart path detection)" })

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
            layout_config = { width = 0.9, height = 0.89 },
            initial_mode = "normal",
            hidden = true,
            no_ignore = true
         })
      end, { desc = "Telescope open file browser" })

      vim.keymap.set("n", "<leader>km", function()
         builtin.keymaps()
      end, { desc = "Telescope show keymaps" })

      vim.keymap.set("n", "<leader>df", function()
         if not utils.is_git_repo() then
            vim.notify("Not a Git repository, cannot show diff file with current buffer", vim.log.levels.WARN)
            return
         end

         utils.telescope_diff_file()
      end, { desc = "Telescope diff file with current buffer" })

      vim.keymap.set("n", "<leader>dg", function()
         if not utils.is_git_repo() then
            vim.notify("Not a Git repository, cannot show diff history", vim.log.levels.WARN)
            return
         end

         utils.telescope_diff_from_history()
      end, { desc = "Telescope diff from Git history" })

      vim.keymap.set('n', '<leader>ge', function()
         if not utils.is_git_repo() then
            vim.notify("Not a Git repository, cannot show edited files", vim.log.levels.WARN)
            return
         end

         builtin.git_status({
            previewer = delta,
            sort_lastused = true,
         })
      end, { desc = "Telescope list Git edited files (Git changed files)" })
   end
}
