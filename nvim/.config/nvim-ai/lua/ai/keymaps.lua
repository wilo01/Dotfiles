-- Optimized Keymaps
-- Using vim.keymap.set exclusively for consistency

local map = vim.keymap.set
local opts = { noremap = true, silent = true }

-- Helper function for description
local function desc(description)
   return vim.tbl_extend("force", opts, { desc = description })
end

-- Escape mappings
map({ "n", "i", "v" }, "qq", "<Esc>", desc("Escape with qq"))
map("i", "<C-c>", "<Esc>", desc("Escape insert mode"))
map("n", "<Esc>", "<cmd>nohlsearch<CR>", desc("Clear search highlighting"))

-- Text movement
map("v", "J", ":m '>+1<CR>gv=gv", desc("Move selected text down"))
map("v", "K", ":m '<-2<CR>gv=gv", desc("Move selected text up"))
map("n", "H", "^", desc("Move to first character of line"))
map("n", "L", "$", desc("Move to end of line"))
map("n", "<A-h>", "mzJ`z", desc("Join lines without moving cursor"))

-- Navigation improvements
map("n", "<C-d>", "<C-d>zz", desc("Scroll down and center"))
map("n", "<C-u>", "<C-u>zz", desc("Scroll up and center"))
map("n", "n", "nzzzv", desc("Next search result centered"))
map("n", "N", "Nzzzv", desc("Previous search result centered"))
map("n", "<A-j>", "}zz", desc("Next paragraph centered"))
map("n", "<A-k>", "{zz", desc("Previous paragraph centered"))
map("n", "{", "{zz", opts)
map("n", "}", "}zz", opts)

-- Better word navigation
map("n", "<C-Left>", "b", desc("Previous word"))
map("n", "<C-Right>", "w", desc("Next word"))
map("i", "<C-Left>", "<C-o>b", desc("Previous word in insert"))
map("i", "<C-Right>", "<C-o>w", desc("Next word in insert"))

-- Window management
map("n", "<C-h>", "<C-w>h", desc("Move to left window"))
map("n", "<C-j>", "<C-w>j", desc("Move to lower window"))
map("n", "<C-k>", "<C-w>k", desc("Move to upper window"))
map("n", "<C-l>", "<C-w>l", desc("Move to right window"))
map("n", "<C-Up>", "<cmd>resize +2<CR>", desc("Increase window height"))
map("n", "<C-Down>", "<cmd>resize -2<CR>", desc("Decrease window height"))
map("n", "<C-S-Left>", "<cmd>vertical resize -2<CR>", desc("Decrease window width"))
map("n", "<C-S-Right>", "<cmd>vertical resize +2<CR>", desc("Increase window width"))

-- Buffer management
map("n", "[b", "<cmd>bprevious<CR>", desc("Previous buffer"))
map("n", "]b", "<cmd>bnext<CR>", desc("Next buffer"))
map("n", "<leader>bd", "<cmd>bdelete<CR>", desc("Delete buffer"))
map("n", "<leader>bD", "<cmd>bdelete!<CR>", desc("Force delete buffer"))

-- Tab management
map("n", "[t", "<cmd>tabprevious<CR>", desc("Previous tab"))
map("n", "]t", "<cmd>tabnext<CR>", desc("Next tab"))
map("n", "<leader>tn", "<cmd>tabnew<CR>", desc("New tab"))
map("n", "<leader>tc", "<cmd>tabclose<CR>", desc("Close tab"))

-- Clipboard operations
map({ "n", "v" }, "<leader>y", '"+y', desc("Yank to system clipboard"))
map("n", "<leader>Y", '"+Y', desc("Yank line to system clipboard"))
map({ "n", "v" }, "<leader>p", '"+p', desc("Paste from system clipboard"))
map({ "n", "v" }, "<leader>P", '"+P', desc("Paste before from system clipboard"))
map({ "n", "v" }, "<leader>d", '"_d', desc("Delete without yanking"))
map("n", "x", '"_x', desc("Delete character without yanking"))

-- Search and replace
map("n", "<leader>/", "/<C-r>+<CR>zz", desc("Search with clipboard content"))
map("n", "<leader>R", [[:%s/\<<C-r><C-w>\>/<C-r><C-w>/gI<Left><Left><Left>]], desc("Replace word under cursor"))
map("v", "<leader>r", [[:s/\%V]], desc("Replace in selection"))

-- Number manipulation
map("n", "<leader>+", "<C-a>", desc("Increment number"))
map("n", "<leader>=", "<C-a>", desc("Increment number"))
map("n", "<leader>-", "<C-x>", desc("Decrement number"))
map("x", "<leader>=", "g<C-a>", desc("Increment numbers in selection"))
map("x", "<leader>-", "g<C-x>", desc("Decrement numbers in selection"))

-- Text wrapping (visual mode)
map("v", "<leader>8", [[c**<C-r>"**<Esc>]], desc("Wrap with **"))
map("v", "<leader>{", [[c{<C-r>"}<Esc>]], desc("Wrap with {}"))
map("v", "<leader>[", [[c[<C-r>"]<Esc>]], desc("Wrap with []"))
map("v", "<leader>(", [[c(<C-r>")<Esc>]], desc("Wrap with ()"))
map("v", '<leader>"', [[c"<C-r>""<Esc>]], desc('Wrap with ""'))
map("v", "<leader>'", [[c'<C-r>"'<Esc>]], desc("Wrap with ''"))

-- Quick actions
map("n", "<leader>w", "<cmd>w<CR>", desc("Save file"))
map("n", "<leader>W", "<cmd>wa<CR>", desc("Save all files"))
map("n", "<leader>q", "<cmd>q<CR>", desc("Quit"))
map("n", "<leader>Q", "<cmd>qa<CR>", desc("Quit all"))
map("n", "<leader>x", "<cmd>x<CR>", desc("Save and quit"))

-- Toggle settings
map("n", "<leader>ts", "<cmd>set spell!<CR>", desc("Toggle spell check"))
map("n", "<leader>tw", "<cmd>set wrap!<CR>", desc("Toggle line wrap"))
map("n", "<leader>tn", "<cmd>set relativenumber!<CR>", desc("Toggle relative numbers"))
map("n", "<leader>th", "<cmd>set hlsearch!<CR>", desc("Toggle search highlighting"))
map("n", "<leader>tl", "<cmd>set list!<CR>", desc("Toggle whitespace display"))

-- Built-in features
map("n", "<leader>e", "<cmd>Explore<CR>", desc("Open file explorer"))
map("n", "<leader>E", "<cmd>Lexplore<CR>", desc("Toggle file explorer"))
map("n", "<leader>u", vim.cmd.UndotreeToggle, desc("Toggle undo tree"))

-- Diagnostic keymaps (built-in)
map("n", "[d", vim.diagnostic.goto_prev, desc("Previous diagnostic"))
map("n", "]d", vim.diagnostic.goto_next, desc("Next diagnostic"))
map("n", "gl", vim.diagnostic.open_float, desc("Show diagnostic"))
map("n", "<leader>cd", vim.diagnostic.setloclist, desc("Diagnostics to location list"))

-- Quickfix and location list
map("n", "[q", "<cmd>cprevious<CR>", desc("Previous quickfix item"))
map("n", "]q", "<cmd>cnext<CR>", desc("Next quickfix item"))
map("n", "[l", "<cmd>lprevious<CR>", desc("Previous location list item"))
map("n", "]l", "<cmd>lnext<CR>", desc("Next location list item"))
map("n", "<leader>co", "<cmd>copen<CR>", desc("Open quickfix"))
map("n", "<leader>cc", "<cmd>cclose<CR>", desc("Close quickfix"))
map("n", "<leader>lo", "<cmd>lopen<CR>", desc("Open location list"))
map("n", "<leader>lc", "<cmd>lclose<CR>", desc("Close location list"))

-- Toggle true/false
map("n", "<leader>,", function()
   local word = vim.fn.expand("<cword>")
   if word == "true" then
      vim.cmd("normal! ciwfalse")
   elseif word == "false" then
      vim.cmd("normal! ciwtrue")
   end
end, desc("Toggle true/false"))

-- Duplicate line and comment
map({ "n", "v" }, "yc", "yy<cmd>normal gcc<CR>p", desc("Duplicate and comment line"))

-- Format document (will be overridden by LSP when available)
map("n", "<leader>f", function()
   vim.lsp.buf.format({ async = true })
end, desc("Format document"))

-- Command mode improvements
map("c", "<C-a>", "<Home>", { desc = "Beginning of line" })
map("c", "<C-e>", "<End>", { desc = "End of line" })
map("c", "<C-p>", "<Up>", { desc = "Previous command" })
map("c", "<C-n>", "<Down>", { desc = "Next command" })

-- Terminal mode
map("t", "<Esc><Esc>", "<C-\\><C-n>", desc("Exit terminal mode"))
map("t", "<C-h>", "<C-\\><C-n><C-w>h", desc("Move to left window from terminal"))
map("t", "<C-j>", "<C-\\><C-n><C-w>j", desc("Move to lower window from terminal"))
map("t", "<C-k>", "<C-\\><C-n><C-w>k", desc("Move to upper window from terminal"))
map("t", "<C-l>", "<C-\\><C-n><C-w>l", desc("Move to right window from terminal"))