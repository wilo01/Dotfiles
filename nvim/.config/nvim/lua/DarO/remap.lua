local utils = require("DarO.utils")

vim.g.mapleader = " "

-- File Navigation
vim.keymap.set("n", "<leader>v", vim.cmd.Ex, { desc = "Open Netrw" })

-- Visual Mode Text Movement
vim.keymap.set("v", "J", ":m '>+1<CR>gv=gv", { desc = "Move selected text down" })
vim.keymap.set("v", "K", ":m '<-2<CR>gv=gv", { desc = "Move selected text up" })

-- Insert and Visual Mode Escape
vim.keymap.set("i", "jk", "<Esc>", { desc = "Escape insert mode with 'jk'" })
vim.keymap.set("i", "hl", "<Esc>", { desc = "Escape insert mode with 'hl'" })
vim.keymap.set("v", "hl", "<Esc>", { desc = "Escape visual mode with 'hl'" })

-- Navigation Enhancements
vim.keymap.set("n", "<C-d>", "<C-d>zz", { desc = "Scroll down and center" })
vim.keymap.set("n", "<C-u>", "<C-u>zz", { desc = "Scroll up and center" })
vim.keymap.set("n", "n", "nzzzv", { desc = "Center cursor on next search result" })
vim.keymap.set("n", "N", "Nzzzv", { desc = "Center cursor on previous search result" })

-- Editing Utilities
vim.keymap.set("n", "x", '"_x', { desc = "Delete character without yanking" })
vim.keymap.set("n", "<leader>+", '<C-a>', { desc = "Increment number" })
vim.keymap.set("n", "<leader>-", '<C-x>', { desc = "Decrement number" })

-- Clipboard Operations
vim.keymap.set("x", "<leader>p", [["_dP]], { desc = "Replace with yanked text" })
vim.keymap.set({ "n", "v" }, "<leader>y", [["+y]], { desc = "Yank to clipboard" })
vim.keymap.set("n", "<leader>Y", [["+Y]], { desc = "Yank line to clipboard" })
vim.keymap.set("v", "<C-c>", [["+Y]], { desc = "Yank selection to clipboard with Ctrl+C" })
vim.keymap.set({ "n", "v" }, "<leader>d", [["_d]], { desc = "Delete without yanking" })

-- Escape Key Remapping
vim.keymap.set("i", "<C-c>", "<Esc>", { desc = "Escape insert mode with Ctrl+C" })

-- Disabling Default Mappings
vim.keymap.set("n", "Q", "<nop>", { desc = "Disable 'Q'" })

-- Tmux Integration
vim.keymap.set("n", "<C-f>", "<cmd>silent !tmux neww tmux-sessionizer<CR>",
   { desc = "Switch projects using tmux-sessionizer" })

-- LSP Formatting
vim.keymap.set("n", "<leader>f", function()
   vim.lsp.buf.format()
   vim.cmd("write")
end, { desc = "Format and save with LSP" })
vim.keymap.set("v", "<leader>f", function()
   vim.lsp.buf.format()
   vim.cmd("write")
end, { desc = "Format selection and save with LSP" })

-- Quickfix and Location List Navigation
vim.keymap.set("n", "<C-k>", "<cmd>cnext<CR>zz", { desc = "Next quickfix item" })
vim.keymap.set("n", "<C-j>", "<cmd>cprev<CR>zz", { desc = "Previous quickfix item" })
vim.keymap.set("n", "<leader>k", "<cmd>lnext<CR>zz", { desc = "Next location list item" })
vim.keymap.set("n", "<leader>j", "<cmd>lprev<CR>zz", { desc = "Previous location list item" })

-- Text Replacement
vim.keymap.set("n", "<leader>R", [[:%s/\<<C-r><C-w>\>/<C-r><C-w>/gI<Left><Left><Left>]],
   { desc = "Replace all occurrences of the word under cursor" })
vim.keymap.set("v", "<leader>R", [[:%s/\<<C-r><C-w>\>/<C-r><C-w>/gI<Left><Left><Left>]],
   { desc = "Replace all occurrences in selection" })

-- File Operations
vim.keymap.set("n", "<leader>x", "<cmd>!chmod +x %<CR>", { desc = "Make file executable", silent = true })
vim.keymap.set("n", "<leader>mr", "<cmd>CellularAutomaton make_it_rain<CR>", { desc = "Make it rain animation" })
vim.keymap.set('n', '<leader>fp', function()
   local filepath = vim.fn.expand('%:p')
   local home_dir = vim.fn.getenv('HOME')
   local relative_path = filepath:gsub('^' .. home_dir, '~')

   vim.fn.setreg('+', relative_path)
   vim.notify('Copied path: ' .. relative_path)
end, { desc = 'Copy current file path to clipboard (pwd)' })

-- Markdown Preview
vim.keymap.set("n", "<leader>m", "<CMD>MarkdownPreview<CR>", { desc = "Start Markdown preview" })
vim.keymap.set("n", "<leader>mn", "<CMD>MarkdownPreviewStop<CR>", { desc = "Stop Markdown preview" })

-- Gitsigns Integration
vim.keymap.set("n", "<leader>va", "<CMD>Gitsigns preview_hunk_inline<CR>", { desc = "Preview Git hunk" })
vim.keymap.set("n", "<leader>vs", "<CMD>Gitsigns diffthis<CR>", { desc = "Diff current buffer" })
vim.keymap.set("n", "<leader>vt", "<CMD>Gitsigns toggle_deleted<CR>", { desc = "Toggle deleted lines" })
vim.keymap.set("n", "<leader>vb", "<CMD>Gitsigns blame_line<CR>", { desc = "Blame current line" })
vim.keymap.set("n", "J", "<CMD>Gitsigns next_hunk<CR>", { desc = "Go to next Git hunk" })
vim.keymap.set("n", "K", "<CMD>Gitsigns prev_hunk<CR>", { desc = "Go to previous Git hunk" })

-- Telescope Integrations
vim.keymap.set("n", "<leader><leader>", ":Telescope smart_open<CR>", { desc = "Smart open with Telescope" })
vim.keymap.set("n", "<leader>fh", ":Telescope help_tags<CR>", { desc = "Find help tags" })
vim.keymap.set("n", "<leader>df", function() utils.telescope_diff_file() end, { desc = "Diff file with current buffer" })
vim.keymap.set("n", "<leader>dg", function() utils.telescope_diff_from_history() end, { desc = "Diff from Git history" })
vim.keymap.set("n", "<leader>km", ":Telescope keymaps<CR>", { desc = "Show keymaps" })

-- Navigation
vim.keymap.set("n", "gb", "<C-o>", { desc = "Go back" })

-- Insert Console Snippets
vim.keymap.set("v", "<leader>cl", function()
   vim.cmd('normal! "+y')
   local selected_text = vim.fn.getreg('+')
   local snippet = {
      "console.warn({",
      string.format("\t'%s': %s,", selected_text, selected_text),
      "});"
   }
   vim.api.nvim_put(snippet, 'l', true, true)
end, { desc = "Insert object console.warn snippet with selection" })

vim.keymap.set("v", "<leader>ck", function()
   vim.cmd('normal! "+y')
   local selected_text = vim.fn.getreg('+')
   local snippet = {
      "if (" .. selected_text .. ") {",
      "}"
   }
   vim.api.nvim_put(snippet, 'l', true, true)
end, { desc = "Insert if statement snippet with clipboard content" })

vim.keymap.set("v", "<leader>cj", function()
   vim.cmd('normal! "+y')
   local clipboard_content = vim.fn.getreg('+')
   local snippet = {
      "ca_log_pak.log_warning('Dwdw', '" .. clipboard_content .. ": ' || " .. clipboard_content .. ");"
   }
   vim.api.nvim_put(snippet, 'l', true, true)
end, { desc = "Insert ca_log_pak.log_warning snippet with clipboard content" })

vim.keymap.set("v", "<leader>ch", function()
   vim.cmd('normal! "+y')
   local clipboard_content = vim.fn.getreg('+')
   local snippet = {
      "DBMS_OUTPUT.PUT_LINE('Dwdw " .. clipboard_content .. ": ' || " .. clipboard_content .. ");"
   }
   vim.api.nvim_put(snippet, 'l', true, true)
end, { desc = "Insert DBMS_OUTPUT snippet with clipboard content" })
