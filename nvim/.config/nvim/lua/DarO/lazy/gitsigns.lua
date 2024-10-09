return {
   "lewis6991/gitsigns.nvim",
   config = function()
      local status, gitsigns = pcall(require, "gitsigns")

      if not status then
         return
      end

      gitsigns.setup({
         signs = {
            add = { text = "│" },
            change = { text = "│" },
            delete = { text = "_" },
            topdelete = { text = "‾" },
            changedelete = { text = "~" },
            untracked = { text = "┆" },
         },
         signcolumn = true, -- Toggle with `:Gitsigns toggle_signs`
         numhl = false,     -- Toggle with `:Gitsigns toggle_numhl`
         linehl = false,    -- Toggle with `:Gitsigns toggle_linehl`
         word_diff = false, -- Toggle with `:Gitsigns toggle_word_diff`
         watch_gitdir = {
            interval = 1000,
            follow_files = true,
         },
         attach_to_untracked = true,
         current_line_blame = true, -- Toggle with `:Gitsigns toggle_current_line_blame`
         current_line_blame_opts = {
            virt_text = true,
            virt_text_pos = "eol", -- 'eol' | 'overlay' | 'right_align'
            delay = 100,
            ignore_whitespace = false,
         },
         current_line_blame_formatter = "<author>, <author_time:%Y-%m-%d> - <summary>",
         sign_priority = 6,
         update_debounce = 100,
         status_formatter = nil,  -- Use default
         max_file_length = 40000, -- Disable if file is longer than this (in lines)
         preview_config = {
            -- Options passed to nvim_open_win
            border = "single",
            style = "minimal",
            relative = "cursor",
            row = 0,
            col = 1,
         },
      })

      -- Autocmd for removing trailing whitespaces on changed lines
      local augroup = vim.api.nvim_create_augroup
      local TheDaroGroup = augroup('TheDarO', {})
      local autocmd = vim.api.nvim_create_autocmd

      -- Autocmd for removing trailing whitespaces on changed lines
      autocmd({ "BufWritePre" }, {
         group = TheDaroGroup,
         pattern = "*",
         callback = function()
            -- Get the current buffer
            local bufnr = vim.api.nvim_get_current_buf()
            -- Get the hunks (changed lines)
            local hunk_lines = gitsigns.get_hunks(bufnr)

            -- Iterate over each hunk and remove trailing whitespaces in that range
            if hunk_lines then
               for _, hunk in ipairs(hunk_lines) do
                  local start_line = hunk.added.start
                  local end_line = hunk.added.start + hunk.added.count - 1
                  -- Only run the command if the range is valid
                  if start_line <= end_line then
                     -- Remove trailing whitespaces only in the modified range
                     vim.api.nvim_buf_call(bufnr, function()
                        vim.cmd(string.format('%d,%ds/\\s\\+$//e', start_line, end_line))
                     end)
                  end
               end
            end
         end,
      })
   end
}
