local ok_utils, utils = pcall(require, "DarO.utils")
if not ok_utils then
   vim.notify("Failed to load DarO.utils: " .. tostring(utils), vim.log.levels.ERROR)
   return
end
local autocmd = vim.api.nvim_create_autocmd
local augroup = vim.api.nvim_create_augroup
local general = augroup("General Settings", { clear = true })

local CSV_STATE_COMPACT = 0
local CSV_STATE_FULL = 1
local csv_buffer_data = {}

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

--- Parse CSV line respecting quoted fields
--- @param line string CSV line to parse
--- @return table columns
local function csv_parse_line(line)
   local cols = {}
   local current = ""
   local in_quotes = false
   local i = 1

   while i <= #line do
      local char = line:sub(i, i)

      if char == '"' then
         in_quotes = not in_quotes
         current = current .. char
      elseif char == "," and not in_quotes then
         table.insert(cols, current)
         current = ""
      else
         current = current .. char
      end
      i = i + 1
   end

   table.insert(cols, current)
   return cols
end

local function csv_get_state(bufnr)
   if not csv_buffer_data[bufnr] then
      csv_buffer_data[bufnr] = { state = CSV_STATE_COMPACT }
   end
   return csv_buffer_data[bufnr]
end

local function csv_prettify(lines)
   local max_lengths = {}

   -- First pass: find the actual maximum length for each column
   for _, line in ipairs(lines) do
      local cols = csv_parse_line(line)
      for i, col in ipairs(cols) do
         max_lengths[i] = math.max(max_lengths[i] or 0, #col)
      end
   end

   -- Second pass: pad all columns to their calculated max width
   local prettified = {}
   for _, line in ipairs(lines) do
      local cols = csv_parse_line(line)
      for i, col in ipairs(cols) do
         local target_width = max_lengths[i] or 0
         local padding = target_width - #col
         cols[i] = col .. string.rep(" ", padding)
      end
      table.insert(prettified, table.concat(cols, " , "))
   end
   return prettified
end

local function csv_compact(lines)
   local compacted = {}
   for _, line in ipairs(lines) do
      local cleaned = line:gsub("%s*,%s*", ",")
      cleaned = cleaned:gsub("%s+$", "")
      compacted[#compacted + 1] = cleaned
   end
   return compacted
end

autocmd("FileType", {
   pattern = "csv",
   callback = function()
      vim.defer_fn(function()
         vim.notify("CSV: <leader>t toggles Compact ↔ Full", vim.log.levels.INFO)
      end, 100)

      vim.keymap.set("n", "<leader>t", function()
         local buf = vim.api.nvim_get_current_buf()
         local data = csv_get_state(buf)

         if data.needs_restore and data.prettified_for_restore then
            vim.api.nvim_buf_set_lines(buf, 0, -1, false, data.prettified_for_restore)
            data.prettified_for_restore = nil
            data.needs_restore = nil
            vim.notify("[CSV: Full] Recovered from failed save", vim.log.levels.WARN)
            return
         end

         local lines = vim.api.nvim_buf_get_lines(buf, 0, -1, false)

         if data.state == CSV_STATE_COMPACT then
            local prettified = csv_prettify(lines)
            vim.api.nvim_buf_set_lines(buf, 0, -1, false, prettified)
            data.state = CSV_STATE_FULL
            vim.notify("[CSV: Full] Pretty view - editable", vim.log.levels.INFO)
         else
            local compacted = csv_compact(lines)
            vim.api.nvim_buf_set_lines(buf, 0, -1, false, compacted)
            data.state = CSV_STATE_COMPACT
            vim.notify("[CSV: Compact] Raw CSV", vim.log.levels.INFO)
         end
      end, { buffer = true, desc = "Toggle CSV: Compact ↔ Full", noremap = true, silent = true })
   end,
   desc = "Setup CSV 2-state toggle",
})

autocmd("BufWritePre", {
   pattern = "*.csv",
   callback = function()
      if not vim.g.csv_prettify_ind then return end

      local bufnr = vim.api.nvim_get_current_buf()
      local data = csv_buffer_data[bufnr]
      if not data or data.state == CSV_STATE_COMPACT then return end

      local lines = vim.api.nvim_buf_get_lines(bufnr, 0, -1, false)
      data.prettified_for_restore = lines
      data.needs_restore = true

      local compacted = csv_compact(lines)
      vim.api.nvim_buf_set_lines(bufnr, 0, -1, false, compacted)
   end,
   desc = "CSV: Compact before save",
})

autocmd("BufWritePost", {
   pattern = "*.csv",
   callback = function()
      if not vim.g.csv_prettify_ind then return end

      local bufnr = vim.api.nvim_get_current_buf()
      local data = csv_buffer_data[bufnr]
      if not data or not data.needs_restore then return end

      vim.api.nvim_buf_set_lines(bufnr, 0, -1, false, data.prettified_for_restore)
      vim.bo[bufnr].modified = false
      data.prettified_for_restore = nil
      data.needs_restore = nil
      vim.notify("Saved compacted CSV", vim.log.levels.INFO)
   end,
   desc = "CSV: Restore Full view after save",
})

autocmd({ "BufUnload", "BufWipeout" }, {
   pattern = "*.csv",
   callback = function()
      local bufnr = vim.api.nvim_get_current_buf()
      csv_buffer_data[bufnr] = nil
   end,
   desc = "CSV: Cleanup state on buffer close",
})

autocmd("BufReadPost", {
   pattern = "*.csv",
   callback = function()
      local bufnr = vim.api.nvim_get_current_buf()
      csv_buffer_data[bufnr] = { state = CSV_STATE_COMPACT }
   end,
   desc = "CSV: Reset state on file reload",
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
