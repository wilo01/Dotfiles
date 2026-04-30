local ok_utils, utils = pcall(require, "DarO.utils")
if not ok_utils then
   vim.notify("Failed to load DarO.utils: " .. tostring(utils), vim.log.levels.ERROR)
   return
end
local autocmd = vim.api.nvim_create_autocmd
local augroup = vim.api.nvim_create_augroup
local general = augroup("General Settings", { clear = true })

local CSV_STATE_COMPACT = 0
local CSV_STATE_NARROW = 1
local CSV_STATE_WIDE = 2
local CSV_COL_CAP = 24 -- Narrow-state cap; longer cells show first 23 chars + "…"
local csv_buffer_data = {}
local csv_ns = vim.api.nvim_create_namespace("DarO_csv_align")

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
      csv_buffer_data[bufnr] = { state = CSV_STATE_COMPACT, last_cursor = nil }
   end
   return csv_buffer_data[bufnr]
end

local function csv_show_alignment(buf, cap)
   local lines = vim.api.nvim_buf_get_lines(buf, 0, -1, false)
   local visual_widths = {}

   for _, line in ipairs(lines) do
      if line:find(",", 1, true) then
         local cols = csv_parse_line(line)
         for i, col in ipairs(cols) do
            local vw = cap and math.min(#col, cap) or #col
            visual_widths[i] = math.max(visual_widths[i] or 0, vw)
         end
      end
   end

   vim.api.nvim_buf_clear_namespace(buf, csv_ns, 0, -1)

   for lnum, line in ipairs(lines) do
      if line:find(",", 1, true) then
         local cols = csv_parse_line(line)
         local byte_pos = 0
         for i, col in ipairs(cols) do
            if cap and #col > cap then
               vim.api.nvim_buf_set_extmark(buf, csv_ns, lnum - 1,
                  byte_pos + cap - 1, {
                     end_col = byte_pos + #col,
                     conceal = "…",
                  })
            end

            if i < #cols then
               local visual_len = cap and math.min(#col, cap) or #col
               local pad = (visual_widths[i] or 0) - visual_len
               if pad > 0 then
                  vim.api.nvim_buf_set_extmark(buf, csv_ns, lnum - 1,
                     byte_pos + #col, {
                        virt_text = { { string.rep(" ", pad), "NonText" } },
                        virt_text_pos = "inline",
                     })
               end
            end

            byte_pos = byte_pos + #col + 1
         end
      end
   end
end

local function csv_clear_alignment(buf)
   vim.api.nvim_buf_clear_namespace(buf, csv_ns, 0, -1)
end

local function csv_cell_at_cursor(buf)
   local row, col = unpack(vim.api.nvim_win_get_cursor(0))
   local line = vim.api.nvim_buf_get_lines(buf, row - 1, row, false)[1]
   if not line or not line:find(",", 1, true) then return nil end

   local cells = csv_parse_line(line)
   local byte_pos = 0
   for _, cell in ipairs(cells) do
      if col >= byte_pos and col <= byte_pos + #cell then
         return cell
      end
      byte_pos = byte_pos + #cell + 1
   end
   return cells[#cells]
end

local function csv_show_cell_popup(buf)
   local content = csv_cell_at_cursor(buf)
   if not content or content == "" then
      vim.notify("CSV: empty or no cell under cursor", vim.log.levels.INFO)
      return
   end

   local lines = vim.split(content, "\n", { plain = true })
   local pbuf = vim.api.nvim_create_buf(false, true)
   vim.api.nvim_buf_set_lines(pbuf, 0, -1, false, lines)
   vim.bo[pbuf].bufhidden = "wipe"

   local max_w = 0
   for _, l in ipairs(lines) do
      max_w = math.max(max_w, vim.fn.strdisplaywidth(l))
   end
   local width = math.max(10, math.min(max_w, vim.o.columns - 6))
   local height = math.min(#lines, math.floor(vim.o.lines * 0.4))

   local win = vim.api.nvim_open_win(pbuf, false, {
      relative = "cursor",
      row = 1,
      col = 0,
      width = width,
      height = height,
      border = "rounded",
      style = "minimal",
      focusable = false,
   })

   vim.api.nvim_create_autocmd({ "CursorMoved", "InsertEnter", "BufLeave" }, {
      once = true,
      callback = function()
         if vim.api.nvim_win_is_valid(win) then
            vim.api.nvim_win_close(win, true)
         end
      end,
   })
end

--- Snap cursor past concealed cell tails so h/l/w/e/b feel natural in narrow/wide.
local function csv_handle_cursor_moved(buf)
   local data = csv_buffer_data[buf]
   if not data or data.state == CSV_STATE_COMPACT then return end
   if vim.api.nvim_get_mode().mode ~= "n" then return end

   local row, col = unpack(vim.api.nvim_win_get_cursor(0))
   local prev = data.last_cursor

   local marks = vim.api.nvim_buf_get_extmarks(buf, csv_ns,
      { row - 1, 0 }, { row - 1, -1 }, { details = true })

   for _, m in ipairs(marks) do
      local mcol, det = m[3], m[4]
      if det and det.conceal and col > mcol and col < det.end_col then
         local target
         if prev and prev.row == row and prev.col >= det.end_col then
            target = mcol
         else
            target = det.end_col
         end
         vim.api.nvim_win_set_cursor(0, { row, target })
         data.last_cursor = { row = row, col = target }
         return
      end
   end

   data.last_cursor = { row = row, col = col }
end

--- Jump cursor to the start of the next/previous CSV cell.
local function csv_jump_cell(buf, dir)
   local row, col = unpack(vim.api.nvim_win_get_cursor(0))
   local line = vim.api.nvim_buf_get_lines(buf, row - 1, row, false)[1]
   if not line or not line:find(",", 1, true) then return end

   local cells = csv_parse_line(line)
   local boundaries = { 0 }
   local pos = 0
   for i = 1, #cells - 1 do
      pos = pos + #cells[i] + 1
      table.insert(boundaries, pos)
   end

   local target
   if dir == "next" then
      for _, b in ipairs(boundaries) do
         if b > col then target = b break end
      end
      target = target or boundaries[#boundaries]
   else
      for i = #boundaries, 1, -1 do
         if boundaries[i] < col then target = boundaries[i] break end
      end
      target = target or boundaries[1]
   end

   vim.api.nvim_win_set_cursor(0, { row, target })
end

autocmd("FileType", {
   pattern = "csv",
   callback = function()
      local bufnr = vim.api.nvim_get_current_buf()

      vim.defer_fn(function()
         vim.notify("CSV: <leader>t toggles view · <Tab>/<S-Tab> jump cells · K previews cell", vim.log.levels.INFO)
      end, 100)

      local motion_group = augroup("DarO_csv_motion_" .. bufnr, { clear = true })
      autocmd("CursorMoved", {
         group = motion_group,
         buffer = bufnr,
         callback = function() csv_handle_cursor_moved(bufnr) end,
         desc = "CSV: snap cursor past concealed cell tails",
      })

      vim.keymap.set("n", "<Tab>", function() csv_jump_cell(bufnr, "next") end,
         { buffer = bufnr, desc = "CSV: jump to next cell", noremap = true, silent = true })
      vim.keymap.set("n", "<S-Tab>", function() csv_jump_cell(bufnr, "prev") end,
         { buffer = bufnr, desc = "CSV: jump to previous cell", noremap = true, silent = true })

      vim.keymap.set("n", "<leader>t", function()
         local buf = vim.api.nvim_get_current_buf()
         local data = csv_get_state(buf)

         if data.state == CSV_STATE_COMPACT then
            local narrow_cap = vim.g.csv_col_cap_narrow or CSV_COL_CAP
            data.prev_conceallevel = vim.wo.conceallevel
            data.prev_concealcursor = vim.wo.concealcursor
            vim.wo.conceallevel = 2
            vim.wo.concealcursor = "nc"
            csv_show_alignment(buf, narrow_cap)
            data.state = CSV_STATE_NARROW
            vim.notify("[CSV: Narrow] Aligned, " .. narrow_cap .. "-byte cap (…)", vim.log.levels.INFO)
         elseif data.state == CSV_STATE_NARROW then
            local wide_cap = vim.g.csv_col_cap_wide -- nil = uncapped/full
            csv_show_alignment(buf, wide_cap)
            data.state = CSV_STATE_WIDE
            vim.notify("[CSV: Wide] " .. (wide_cap and (wide_cap .. "-byte cap (…)") or "full content"), vim.log.levels.INFO)
         else
            csv_clear_alignment(buf)
            vim.wo.conceallevel = data.prev_conceallevel or 0
            vim.wo.concealcursor = data.prev_concealcursor or ""
            data.state = CSV_STATE_COMPACT
            vim.notify("[CSV: Compact] Raw CSV", vim.log.levels.INFO)
         end
      end, { buffer = true, desc = "Toggle CSV: Compact → Narrow → Wide", noremap = true, silent = true })

      vim.keymap.set("n", "K", function()
         csv_show_cell_popup(vim.api.nvim_get_current_buf())
      end, { buffer = true, desc = "CSV: Preview full cell under cursor", noremap = true, silent = true })
   end,
   desc = "Setup CSV 3-state toggle",
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
      local data = csv_buffer_data[bufnr]

      vim.api.nvim_buf_clear_namespace(bufnr, csv_ns, 0, -1)

      if data and data.state == CSV_STATE_NARROW then
         csv_show_alignment(bufnr, vim.g.csv_col_cap_narrow or CSV_COL_CAP)
      elseif data and data.state == CSV_STATE_WIDE then
         csv_show_alignment(bufnr, vim.g.csv_col_cap_wide)
      else
         csv_buffer_data[bufnr] = { state = CSV_STATE_COMPACT }
      end
   end,
   desc = "CSV: Re-apply alignment after file reload",
})

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
