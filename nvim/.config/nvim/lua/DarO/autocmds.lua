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
   desc = "Disable New Line Comment",
})

autocmd("BufEnter", {
   callback = function(opts)
      if vim.bo[opts.buf].filetype == "bicep" then
         vim.bo.commentstring = "// %s"
      end
   end,
   group = general,
   desc = "Set Bicep Comment String",
})

autocmd("BufEnter", {
   pattern = { "*.md", "*.txt" },
   callback = function()
      vim.opt_local.spell = true
   end,
   group = general,
   desc = "Enable spell checking on specific filetypes",
})

autocmd("BufWinEnter", {
   callback = function(data)
      utils.open_help(data.buf)
   end,
   group = general,
   desc = "Redirect help to floating window",
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
      end, { desc = "Go to definition", unpack(opts) })

      vim.keymap.set("n", "<leader>K", function()
         vim.lsp.buf.hover()
      end, { desc = "Show hover information", unpack(opts) })

      vim.keymap.set("n", "<leader>vws", function()
         vim.lsp.buf.workspace_symbol()
      end, { desc = "Search workspace symbols", unpack(opts) })

      vim.keymap.set("n", "<leader>vd", function()
         vim.diagnostic.open_float()
      end, { desc = "Show diagnostics in a floating window", unpack(opts) })

      vim.keymap.set("n", "<leader>vca", function()
         vim.lsp.buf.code_action()
      end, { desc = "Show code actions", unpack(opts) })

      vim.keymap.set("n", "<leader>vrr", function()
         vim.lsp.buf.references()
      end, { desc = "Show references", unpack(opts) })

      vim.keymap.set("n", "<leader>vrn", function()
         vim.lsp.buf.rename()
      end, { desc = "Rename symbol", unpack(opts) })

      vim.keymap.set("i", "<C-h>", function()
         vim.lsp.buf.signature_help()
      end, { desc = "Show signature help", unpack(opts) })

      vim.keymap.set("n", "[d", function()
         vim.diagnostic.goto_next()
      end, { desc = "Go to next diagnostic", unpack(opts) })

      vim.keymap.set("n", "]d", function()
         vim.diagnostic.goto_prev()
      end, { desc = "Go to previous diagnostic", unpack(opts) })
   end
})
