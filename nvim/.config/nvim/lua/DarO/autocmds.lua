local utils = require("DarO.utils")
local autocmd = vim.api.nvim_create_autocmd
local augroup = vim.api.nvim_create_augroup
local DarOGroup = augroup('DarO', {})
local yank_group = augroup('HighlightYank', {})

-- General Settings
local general = augroup("General Settings", { clear = true })

autocmd("BufEnter", {
   callback = function()
      vim.opt.formatoptions:remove({ "c", "r", "o" })
   end,
   group = general,
   desc = "Autocmds Disable New Line Comment",
})

autocmd("BufEnter", {
   callback = function(opts)
      if vim.bo[opts.buf].filetype == "bicep" then
         vim.bo.commentstring = "// %s"
      end
   end,
   group = general,
   desc = "Autocmds Set Bicep Comment String",
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
      utils.open_help(data.buf)
   end,
   group = general,
   desc = "Autocmds Redirect help to floating window",
})

function R(name)
   require("plenary.reload").reload_module(name)
end

vim.filetype.add({
   extension = {
      templ = 'templ',
   }
})

autocmd('TextYankPost', {
   group = yank_group,
   pattern = '*',
   callback = function()
      vim.highlight.on_yank({
         higroup = 'IncSearch',
         timeout = 40,
      })
   end,
})

autocmd({ "BufWritePre" }, {
   group = DarOGroup,
   pattern = { "*.md", "*.lua", "*.js", "*.jsx", "*.ts", "*.rs", "*.go", "*.py" },
   command = [[%s/\s\+$//e]],
})

autocmd('LspAttach', {
   group = DarOGroup,
   callback = function(e)
      local opts = { buffer = e.buf }
      vim.keymap.set("n", "gd", function()
         vim.lsp.buf.definition()
      end, { desc = "Autocmds Go to definition", unpack(opts) })

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

      vim.keymap.set("n", "<leader>vrr", function()
         vim.lsp.buf.references()
      end, { desc = "Autocmds Show references", unpack(opts) })

      vim.keymap.set("n", "<leader>vrn", function()
         vim.lsp.buf.rename()
      end, { desc = "Autocmds Rename symbol", unpack(opts) })

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

-- Auto-close CSV edit formatting before saving
autocmd("BufWritePre", {
   pattern = "*.csv",
   callback = function()
      if not vim.g.csv_prettify_ind then
         print("CSV prettify functionality is disabled.")
         return
      end

      if vim.g.is_csv_prettified then
         local lines = vim.api.nvim_buf_get_lines(0, 0, -1, false)
         local cleaned_lines = {}
         for _, line in ipairs(lines) do
            local cleaned_line = line:gsub("%s*,%s*", ","):gsub("%s+$", "")
            table.insert(cleaned_lines, cleaned_line)
         end
         vim.api.nvim_buf_set_lines(0, 0, -1, false, cleaned_lines)
         print("CSV compacted before saving.")
         vim.g.is_csv_prettified = false
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
