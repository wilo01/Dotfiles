local ok_utils, utils = pcall(require, "DarO.utils")
if not ok_utils then
   vim.notify("Failed to load DarO.utils: " .. tostring(utils), vim.log.levels.ERROR)
   return
end
local autocmd = vim.api.nvim_create_autocmd
local augroup = vim.api.nvim_create_augroup
local general = augroup("General Settings", { clear = true })

autocmd("BufEnter", {
   callback = function()
      vim.opt.formatoptions:remove({ "c", "r", "o" })
   end,
   group = general,
   desc = "Autocmds Disable New Line Comment",
})

autocmd("FileType", {
   pattern = "*",
   callback = function(opts)
      local ft = vim.bo[opts.buf].filetype
      local comment_patterns = require("DarO.comment-patterns")
      local commentstring = comment_patterns.get_commentstring(ft)

      if comment_patterns.commentstrings[ft] then
         vim.bo[opts.buf].commentstring = commentstring
      end
   end,
   group = general,
   desc = "Set appropriate commentstring for each filetype",
})

autocmd("BufEnter", {
   pattern = { "*.md", "*.txt" },
   callback = function()
      vim.opt_local.spell = true
   end,
   group = general,
   desc = "Autocmds Enable spell checking on specific filetypes",
})

autocmd("BufWinEnter", {
   callback = function(data)
      if utils and utils.open_help then
         pcall(utils.open_help, data.buf)
      end
   end,
   group = general,
   desc = "Autocmds Redirect help to floating window",
})

function R(name)
   -- Native module reloading without plenary
   package.loaded[name] = nil
   return require(name)
end

vim.filetype.add({
   extension = {
      templ = 'templ',
   },
   filename = {
      ['.env'] = 'dotenv',
   },
   pattern = {
      ['.*%.env[^/]*'] = 'dotenv', -- Matches any file starting with .env
   }
})

autocmd('TextYankPost', {
   group = augroup('HighlightYank', {}),
   pattern = '*',
   callback = function()
      vim.highlight.on_yank({
         higroup = 'IncSearch',
         timeout = 40,
      })
   end,
})

autocmd({ "BufWritePre" }, {
   group = augroup('DarO', {}),
   pattern = "*",
   callback = function()
      if vim.g.disable_autoformat then
         return
      end
      vim.cmd([[%s/\s\+$//e]])
   end,
})

autocmd('LspAttach', {
   group = augroup('LspAttach', { clear = true }),
   callback = function(event)
      local opts = { buffer = event.buf }

      -- Backwards compatible document highlighting
      local ok_client, client = pcall(vim.lsp.get_client_by_id, event.data.client_id)
      if ok_client and client and client.server_capabilities and client.server_capabilities.documentHighlightProvider then
         local highlight_augroup = augroup("lsp-highlight-" .. event.buf, { clear = true })

         autocmd("CursorHold", {
            group = highlight_augroup,
            buffer = event.buf,
            callback = function()
               pcall(vim.lsp.buf.document_highlight)
            end,
         })

         autocmd("CursorMoved", {
            group = highlight_augroup,
            buffer = event.buf,
            callback = function()
               pcall(vim.lsp.buf.clear_references)
            end,
         })
      end

      vim.keymap.set("n", "gd", function()
         vim.lsp.buf.definition()
      end, { desc = "Autocmds Go to definition", unpack(opts) })

      vim.keymap.set("n", "gD", function()
         vim.lsp.buf.declaration()
      end, { desc = "Autocmds Go to declaration", unpack(opts) })

      vim.keymap.set("n", "gT", function()
         vim.lsp.buf.type_definition()
      end, { desc = "Autocmds Go to type_definition", unpack(opts) })

      vim.keymap.set("n", "<leader>K", function()
         vim.lsp.buf.hover()
      end, { desc = "Autocmds Show hover information", unpack(opts) })

      vim.keymap.set("n", "<leader>vws", function()
         vim.lsp.buf.workspace_symbol()
      end, { desc = "Autocmds Search workspace symbols", unpack(opts) })

      vim.keymap.set("n", "<leader>vd", function()
         vim.diagnostic.open_float()
      end, { desc = "Autocmds Show diagnostics in a floating window", unpack(opts) })

      vim.keymap.set("n", "<leader>vca", function()
         vim.lsp.buf.code_action()
      end, { desc = "Autocmds Show code actions", unpack(opts) })

      vim.keymap.set("n", "gr", function()
         vim.lsp.buf.references()
      end, { desc = "Autocmds Show references", unpack(opts) })

      vim.keymap.set("n", "<leader>rn", function()
         vim.lsp.buf.rename()
      end, { desc = "Autocmds Global word rename (via LSP)", unpack(opts) })

      vim.keymap.set("i", "<C-h>", function()
         vim.lsp.buf.signature_help()
      end, { desc = "Autocmds Show signature help", unpack(opts) })

      vim.keymap.set("n", "[d", function()
         vim.diagnostic.goto_next()
      end, { desc = "Autocmds Go to next diagnostic", unpack(opts) })

      vim.keymap.set("n", "]d", function()
         vim.diagnostic.goto_prev()
      end, { desc = "Autocmds Go to previous diagnostic", unpack(opts) })
   end
})

-- CSV editing: buffer-local keymap for <leader>t
autocmd("FileType", {
   pattern = "csv",
   callback = function()
      vim.defer_fn(function()
         vim.notify("Use <leader>t to toggle CSV formatting", vim.log.levels.INFO)
      end, 100)
      vim.keymap.set("n", "<leader>t", function()
         local lines = vim.api.nvim_buf_get_lines(0, 0, -1, false)
         local bufnr = vim.api.nvim_get_current_buf()
         local is_prettified = vim.b[bufnr].is_csv_prettified or false

         if is_prettified then
            local cleaned_lines = {}
            for _, line in ipairs(lines) do
               local cleaned_line = line:gsub("%s*,%s*", ","):gsub("%s+$", "")
               table.insert(cleaned_lines, cleaned_line)
            end
            vim.api.nvim_buf_set_lines(0, 0, -1, false, cleaned_lines)
            print("CSV prettification disabled.")
         else
            local MAX_COLUMN_WIDTH = 100
            local ELLIPSIS = "..."
            local MAX_FORMAT_WIDTH = 144

            local max_lengths = {}

            for _, line in ipairs(lines) do
               local cols = vim.split(line, ",", { plain = true })
               for i, col in ipairs(cols) do
                  local col_length = math.min(#col, MAX_COLUMN_WIDTH)
                  max_lengths[i] = math.max(max_lengths[i] or 0, col_length)
               end
            end

            local prettified_lines = {}
            for _, line in ipairs(lines) do
               local cols = vim.split(line, ",", { plain = true })
               for i, col in ipairs(cols) do
                  local max_len = max_lengths[i] or 0

                  if max_len > MAX_FORMAT_WIDTH then
                     max_len = MAX_FORMAT_WIDTH
                  end

                  local formatted_col = col
                  if #col > MAX_COLUMN_WIDTH then
                     formatted_col = col:sub(1, MAX_COLUMN_WIDTH - #ELLIPSIS) .. ELLIPSIS
                  end

                  -- Skip string.format entirely - use direct padding (more reliable)
                  local padding_needed = math.max(0, max_len - #formatted_col)
                  cols[i] = formatted_col .. string.rep(" ", padding_needed)
               end
               table.insert(prettified_lines, table.concat(cols, " , "))
            end

            vim.api.nvim_buf_set_lines(0, 0, -1, false, prettified_lines)
            print("CSV prettification enabled.")
         end

         vim.b[bufnr].is_csv_prettified = not is_prettified
      end, { buffer = true, desc = "Toggle CSV formatting", noremap = true, silent = true })
   end,
   desc = "Setup CSV formatting keymap for csv files only",
})

-- Auto-close CSV edit formatting before saving
autocmd("BufWritePre", {
   pattern = "*.csv",
   callback = function()
      if not vim.g.csv_prettify_ind then
         print("CSV prettify functionality is disabled.")
         return
      end

      local bufnr = vim.api.nvim_get_current_buf()
      if vim.b[bufnr].is_csv_prettified then
         local lines = vim.api.nvim_buf_get_lines(0, 0, -1, false)
         local cleaned_lines = {}
         for _, line in ipairs(lines) do
            local cleaned_line = line:gsub("%s*,%s*", ","):gsub("%s+$", "")
            table.insert(cleaned_lines, cleaned_line)
         end
         vim.api.nvim_buf_set_lines(0, 0, -1, false, cleaned_lines)
         print("CSV compacted before saving.")
         vim.b[bufnr].is_csv_prettified = false
      else
         print("CSV already in compact format.")
      end
   end,
   desc = "Remove spaces from CSV before saving",
})

vim.api.nvim_create_user_command("CSVformatting", function()
   vim.g.csv_prettify_ind = not vim.g.csv_prettify_ind
   print("CSV prettify functionality is now " .. (vim.g.csv_prettify_ind and "enabled" or "disabled") .. ".")
end, { desc = "Toggle CSV prettify functionality globally" })

autocmd('BufReadPost', {
   callback = function(args)
      local mark = vim.api.nvim_buf_get_mark(args.buf, '"')
      local line_count = vim.api.nvim_buf_line_count(args.buf)
      if mark[1] > 0 and mark[1] <= line_count then
         vim.api.nvim_win_set_cursor(0, mark)
         vim.schedule(function()
            vim.cmd("normal! zz")
         end)
      end
   end,
   desc = "Restore cursor position on buffer read",
})
-- autocmd("FileType", {
--    pattern = "qf",
--    callback = function()
--       vim.keymap.set("n", "k", "<Up><CR><C-w>p", { buffer = true, remap = false, desc = "Navigate up quickfix" })
--       vim.keymap.set("n", "j", "<Down><CR><C-w>p", { remap = false, desc = "Navigate down quickfix" })
--    end,
-- })
