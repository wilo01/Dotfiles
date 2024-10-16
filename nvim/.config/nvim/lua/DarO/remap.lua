local utils = require("DarO.utils")

vim.g.mapleader = " "
vim.keymap.set("n", "<leader>v", vim.cmd.Ex) -- Opens netrw

vim.keymap.set("v", "J", ":m '>+1<CR>gv=gv") -- Select visual and move DOWN
vim.keymap.set("v", "K", ":m '<-2<CR>gv=gv") -- Select visual and move DOWN

vim.keymap.set("i", "jk", "<Esc>")
vim.keymap.set("i", "hl", "<Esc>")
vim.keymap.set("v", "hl", "<Esc>")

-- vim.keymap.set("n", "J", "}")
-- vim.keymap.set("n", "K", "{")
-- vim.keymap.set("n", "J", "mzJ`z")
vim.keymap.set("n", "<C-d>", "<C-d>zz") -- Code jumping
vim.keymap.set("n", "<C-u>", "<C-u>zz") -- Code jumping
-- vim.keymap.set("n", "J", "<C-d>zz")
-- vim.keymap.set("n", "K", "<C-u>zz")
vim.keymap.set("n", "n", "nzzzv")         -- Helps with navigation and search
vim.keymap.set("n", "N", "Nzzzv")         -- Helps with navigation and search
vim.keymap.set("n", "x", '"_x')           -- Removes character without yank
vim.keymap.set("n", "<leader>+", '<C-a>') -- Increment number
vim.keymap.set("n", "<leader>-", '<C-x>') -- Decrement number

-- greatest remap ever
vim.keymap.set("x", "<leader>p", [["_dP]]) -- Replace by yanked text

-- next greatest remap ever : asbjornHaland
vim.keymap.set({ "n", "v" }, "<leader>y", [["+y]]) -- Yank to clipboard
vim.keymap.set("n", "<leader>Y", [["+Y]])          -- Yank to clipboard
vim.keymap.set("v", "<C-c>", [["+Y]])              -- Yank to clipboard with Ctrl + C

vim.keymap.set({ "n", "v" }, "<leader>d", [["_d]]) -- Yank to clipboard

-- This is going to get me cancelled
vim.keymap.set("i", "<C-c>", "<Esc>")

vim.keymap.set("n", "Q", "<nop>")
vim.keymap.set("n", "<C-f>", "<cmd>silent !tmux neww tmux-sessionizer<CR>") -- Switch projects
-- Format with LSP and save
vim.keymap.set("n", "<leader>f", function()
   vim.lsp.buf.format()
   vim.cmd("write")
end, bufopts)
vim.keymap.set("v", "<leader>f", function()
   vim.lsp.buf.format()
   vim.cmd("write")
end, bufopts)

vim.keymap.set("n", "<C-k>", "<cmd>cnext<CR>zz")                                         -- Quick fix navigation
vim.keymap.set("n", "<C-j>", "<cmd>cprev<CR>zz")                                         -- Quick fix navigation
vim.keymap.set("n", "<leader>k", "<cmd>lnext<CR>zz")                                     -- Quick fix navigation
vim.keymap.set("n", "<leader>j", "<cmd>lprev<CR>zz")                                     -- Quick fix navigation

vim.keymap.set("n", "<leader>R", [[:%s/\<<C-r><C-w>\>/<C-r><C-w>/gI<Left><Left><Left>]]) -- Replace / rename all selected text
vim.keymap.set("v", "<leader>R", [[:%s/\<<C-r><C-w>\>/<C-r><C-w>/gI<Left><Left><Left>]]) -- Replace / rename all selected text
vim.keymap.set("n", "<leader>x", "<cmd>!chmod +x %<CR>", { silent = true })              -- Make current file executable

vim.keymap.set("n", "<leader>vpp", "<cmd>e ~/.dotfiles/nvim/.config/nvim/lua/DarO/packer.lua<CR>");
vim.keymap.set("n", "<leader>mr", "<cmd>CellularAutomaton make_it_rain<CR>");

-- Markdown Preview
vim.keymap.set("n", "<leader>m", "<CMD>MarkdownPreview<CR>")               -- Markdown preview
vim.keymap.set("n", "<leader>mn", "<CMD>MarkdownPreviewStop<CR>")          -- Markdown preview stop

vim.keymap.set("n", "<leader>va", "<CMD>Gitsigns preview_hunk_inline<CR>") -- Git preview
vim.keymap.set("n", "<leader>vs", "<CMD>Gitsigns diffthis<CR>")            -- Git diff
vim.keymap.set("n", "<leader>vt", "<CMD>Gitsigns toggle_deleted<CR>")
vim.keymap.set("n", "<leader>vb", "<CMD>Gitsigns blame_line<CR>")
vim.keymap.set("n", "J", "<CMD>Gitsigns next_hunk<CR>")
vim.keymap.set("n", "K", "<CMD>Gitsigns prev_hunk<CR>")

-- Telescope
vim.keymap.set("n", "<leader><leader>", ":Telescope smart_open<cr>", { desc = "Smart open" })
vim.keymap.set("n", "<leader>fh", ":Telescope help_tags<cr>", { desc = "Find help tags" })
-- stylua: ignore start
vim.keymap.set("n", "<leader>df", function() utils.telescope_diff_file() end, { desc = "Diff file with current buffer" })
vim.keymap.set("n", "<leader>dg", function() utils.telescope_diff_from_history() end, { desc = "Diff from git history" })

-- Usefull aditional remap for going back from go to definition
vim.keymap.set("n", "gb", "<C-o>")

-- Keybinds on selection
vim.keymap.set("v", "<leader>cl", function()
   vim.cmd('normal! "+y')
   local selected_text = vim.fn.getreg('+')
   local snippet = {
      "console.warn({",
      string.format("\t'%s': %s", selected_text, selected_text),
      "});"
   }
   vim.api.nvim_put(snippet, 'l', true, true)
end)
-- stylua: ignore end

-- New un used yet
-- vim.keymap.set("n", "<leader>vwm", function()
--     require("vim-with-me").StartVimWithMe()
-- end)
-- vim.keymap.set("n", "<leader>svwm", function()
--     require("vim-with-me").StopVimWithMe()
-- end)

-- vim.keymap.set("x", "<leader>p", [["_dP]])
-- vim.keymap.set(
--     "n",
--     "<leader>ee",
--     "oif err != nil {<CR>}<Esc>Oreturn err<Esc>"
-- )

-- vim.keymap.set("n", "<leader>vpp", "<cmd>e ~/.dotfiles/nvim/.config/nvim/lua/theprimeagen/packer.lua<CR>");
-- vim.keymap.set("n", "<leader>mr", "<cmd>CellularAutomaton make_it_rain<CR>");

--[[ Other keymaps and navigation

# SEARCH & NAVIGATION #
## harpoon keymaps
<leader>a                           -- Add file to harpoon menu
<C-e>                               -- Toggle quick menu
<C-h>                               -- Navigate to file left
<C-l>                               -- Navigate to file right
<leader>pf                          -- Find files
<C-p>                               -- Find git files
<leader>s                           -- Find by grep word
<leader>v                           -- File navigation with netrw
<leader>pf                          -- File navigation with fzf search
%                                   -- Jump beetween prentesies NVIM

# TEXT MANIPULATION
<leader>x                           -- Make file executable
<leader>p                           -- Replace by yanked text
<leader>y                           -- Ynak to clipboard
<leader>Y                           -- Ynak to clipboard
<leader>d                           -- Ynak to clipboard

## tmux keymaps
<leader>s                           -- Jump between tmux sesnions
<C-f>f                              -- Search with tmux-sessionizer

## Git mappings
<leader>gs                          -- Git actions

## LSP mappings
<C-p>                               -- Select previous LSP item
<C-n>                               -- Select next LSP item
<C-y>                               -- Confirm LSP
<C-Space>                           -- Mapping complete

lsp.on_attach(function(client, bufnr)
    local opts = {buffer = bufnr, remap = false}
    vim.keymap.set("n", "gd", function() vim.lsp.buf.definition() end, opts)
    vim.keymap.set("n", "<leader>k", function() vim.lsp.buf.hover() end, opts)
    vim.keymap.set("n", "<leader>vws", function() vim.lsp.buf.workspace_symbol() end, opts)
    vim.keymap.set('n', "<leader>vd", function() vim.diagnostic.open_float() end, opts)      -- ESlint or other lsp error / warnings
    vim.keymap.set("n", "[d", function() vim.lsp.buf.goto_next() end, opts)
    vim.keymap.set("n", "]d", function() vim.lsp.buf.goto_prev() end, opts)
    vim.keymap.set("n", "<leader>vca", function() vim.lsp.buf.code_action() end, opts)
    vim.keymap.set("n", "<leader>vrr", function() vim.lsp.buf.references() end, opts)
    vim.keymap.set("n", "<leadervca>vrn", function() vim.lsp.buf.rename() end, opts)
    vim.keymap.set("n", "<C-i>", function() vim.lsp.buf.signature_help() end, opts)
end)

END ]] --
