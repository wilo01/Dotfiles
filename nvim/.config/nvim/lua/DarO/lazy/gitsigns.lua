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
            local bufnr = vim.api.nvim_get_current_buf()
            local hunk_lines = gitsigns.get_hunks(bufnr)

            if hunk_lines and #hunk_lines > 0 then
               for _, hunk in ipairs(hunk_lines) do
                  local start_line = hunk.added and hunk.added.start or nil
                  local count = hunk.added and hunk.added.count or 0
                  if start_line and count > 0 then
                     local end_line = start_line + count - 1
                     if start_line <= end_line then
                        vim.api.nvim_buf_call(bufnr, function()
                           vim.cmd(string.format("%d,%ds/\\s\\+$//e", start_line, end_line))
                        end)
                     end
                  end
               end
            end
         end,
      })
   end
}
