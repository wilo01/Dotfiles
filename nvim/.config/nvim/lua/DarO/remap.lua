vim.g.mapleader = " "

-- Text Actions
vim.keymap.set("v", "J", ":m '>+1<CR>gv=gv", { desc = "Move selected text down" })
vim.keymap.set("v", "K", ":m '<-2<CR>gv=gv", { desc = "Move selected text up" })
vim.keymap.set("n", "H", "mzJ`z", { desc = "Move text lines without moving cursor" })
vim.keymap.set('n', '<leader>/', '/<C-r>+<CR>', { desc = "Search with clipboard text" })
vim.keymap.set("n", "<leader>R", [[:%s/\<<C-r><C-w>\>/<C-r><C-w>/gI<Left><Left><Left>]],
   { desc = "Replace text occurrences of the word under cursor" })

-- Escape Mode
vim.keymap.set({ "n", "i", "v" }, "qq", "<Esc>", { desc = "Escape with qq" })
vim.keymap.set("i", "<C-c>", "<Esc>", { desc = "Escape insert mode with Ctrl+C" })
vim.keymap.set("n", "<Esc>", "<cmd>nohlsearch<CR>", { desc = "Escape with no hl search" })

-- Navigation Enhancements
vim.keymap.set("n", "<C-d>", "<C-d>zz", { desc = "Scroll down and center" })
vim.keymap.set("n", "<C-u>", "<C-u>zz", { desc = "Scroll up and center" })
vim.keymap.set("n", "n", "nzzzv", { desc = "Center cursor on next search result" })
vim.keymap.set("n", "N", "Nzzzv", { desc = "Center cursor on previous search result" })

-- Editing Utilities
vim.keymap.set("n", "x", '"_x', { desc = "Delete character without yanking" })
vim.keymap.set("n", "<leader>+", '<C-a>', { desc = "Increment number" })
vim.keymap.set("n", "<leader>=", "<C-a>", { desc = "Increment number" })
vim.keymap.set("x", "<leader>=", "g<C-a>", { desc = "Increment numbers across selection" })
vim.keymap.set("n", "<leader>-", '<C-x>', { desc = "Decrement number" })
vim.keymap.set("x", "<leader>-", 'g<C-x>', { desc = "Decrement numbers across selection" })
vim.keymap.set({ "n", "v" }, "yc", "yy<cmd>normal gcc<CR>p", { desc = "Duplicate a line and comment out the first line" })

-- Clipboard Operations
vim.keymap.set("x", "p", "\"_dP", { desc = "Replace with yanked text, and keep yanked" })
vim.keymap.set("x", "<leader>p", "p", { desc = "Replace with yanked text, and new yanked text" })
vim.keymap.set({ "n", "v" }, "<leader>y", "\"+y", { desc = "Yank to clipboard" })
vim.keymap.set("n", "<leader>Y", "\"+Y", { desc = "Yank line to clipboard" })
vim.keymap.set("v", "<C-c>", "\"+y", { desc = "Yank selection to clipboard with Ctrl+C" })
vim.keymap.set({ "n", "v" }, "<leader>d", "\"_d", { desc = "Delete without yanking" })
vim.keymap.set({ "n", "v" }, "<leader>x", "\"+d", { desc = "Delete & Yank to clipboard" })

-- Disabling Default Mappings
vim.keymap.set("n", "Q", "<nop>", { desc = "Disable 'Q'" })

-- Tmux Integration (Needs tmux-sessionizer export in shell rc)
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

-- File Operations / File Navigation
vim.keymap.set("n", "<leader>v", vim.cmd.Ex, { desc = "Open Netrw" })
vim.keymap.set("n", "gb", "<C-o>zz", { desc = "Go back, and center" })
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
vim.keymap.set("n", "<leader>va", "<CMD>Gitsigns preview_hunk_inline<CR>", { desc = "Gitsigns preview Git hunk" })
vim.keymap.set("n", "<leader>vs", "<CMD>Gitsigns diffthis<CR>", { desc = "Gitsigns Diff current buffer" })
vim.keymap.set("n", "<leader>bl", "<CMD>Gitsigns blame<CR>", { desc = "Gitsigns Blame current file" })
vim.keymap.set("n", "<leader>vt", "<CMD>Gitsigns toggle_deleted<CR>", { desc = "Gitsigns Toggle deleted lines" })
vim.keymap.set("n", "<leader>vb", "<CMD>Gitsigns blame_line<CR>", { desc = "Gitsigns Blame current line" })
vim.keymap.set("n", "J", "<CMD>Gitsigns next_hunk<CR>zz", { desc = "Gitsigns go to next Git hunk and jump to center" })
vim.keymap.set("n", "K", "<CMD>Gitsigns prev_hunk<CR>zz",
   { desc = "Gitsigns go to previous Git hunk and jump to center" })

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
end, { desc = "Insert object console.warn snippet with selection (log, debugger)" })

vim.keymap.set("v", "<leader>ck", function()
   vim.cmd('normal! "+y')
   local selected_text = vim.fn.getreg('+')
   local snippet = {
      "if (" .. selected_text .. ") {",
      "}"
   }
   vim.api.nvim_put(snippet, 'l', true, true)
end, { desc = "Insert if statement snippet with clipboard content (log, debugger)" })

vim.keymap.set("v", "<leader>cj", function()
   vim.cmd('normal! "+y')
   local clipboard_content = vim.fn.getreg('+')
   local current_file = vim.fn.expand('%:t:r')
   local snippet = {
      "ca_log_pak.log_warning('Dwdw', '" .. clipboard_content .. ": ' || " .. clipboard_content .. ", '" .. current_file .. "');"
   }
   vim.api.nvim_put(snippet, 'l', true, true)
end, { desc = "Insert ca_log_pak.log_warning snippet with clipboard content (log, debugger)" })

vim.keymap.set("v", "<leader>ch", function()
   vim.cmd('normal! "+y')
   local clipboard_content = vim.fn.getreg('+')
   local snippet = {
      "DBMS_OUTPUT.PUT_LINE('Dwdw " .. clipboard_content .. ": ' || " .. clipboard_content .. ");"
   }
   vim.api.nvim_put(snippet, 'l', true, true)
end, { desc = "Insert DBMS_OUTPUT snippet with clipboard content (log, debugger)" })

-- Other useful actions with description
vim.keymap.set("n", "z=", "z=", { desc = "Spelling suggestions" })
vim.keymap.set("n", "zg", "zg", { desc = "Spelling add word to spellfile as good word" })
vim.keymap.set("n", "zG", "zG", { desc = "Spelling add word to internal word list as good word" })
vim.keymap.set("n", "zw", "zw", { desc = "Spelling mark word as wrong (bad) in spellfile" })
vim.keymap.set("n", "zW", "zW", { desc = "Spelling mark word as wrong (bad) in internal word list" })
vim.keymap.set("n", "zug", "zug", { desc = "Spelling remove word from spellfile as good/bad" })
vim.keymap.set("n", "zuG", "zuG", { desc = "Spelling remove word from internal word list as good/bad" })
vim.keymap.set("n", "]s", "]s", { desc = "Spelling Next misspelled word" })
vim.keymap.set("n", "[s", "[s", { desc = "Spelling Previous misspelled word" })
vim.keymap.set("n", "]S", "]S", { desc = "Spelling Next bad word only" })
vim.keymap.set("n", "[S", "[S", { desc = "Spelling Previous bad word only" })
vim.keymap.set("n", "]r", "]r", { desc = "Spelling Next rare word" })
vim.keymap.set("n", "[r", "[r", { desc = "Spelling Previous rare word" })

-- Gitlab Integration
local function open_in_tds_gitlab()
   local gitlab_url = "https://gitlab.tds.ie"
   local remote_url = vim.fn.system("git config --get remote.origin.url"):gsub("\n", "")
   local repo_path = remote_url:match("git@[^:]+:(.+)%.git") or remote_url:match("https://[^/]+/(.+)%.git")
   local branch = vim.fn.system("git branch --show-current"):gsub("\n", "")
   local file_path = vim.fn.expand("%:p")
   local git_root = vim.fn.system("git rev-parse --show-toplevel"):gsub("\n", "")

   if not repo_path or git_root == "" then
      print("Error: Not a Git repository or remote not configured!")
      return
   end

   local relative_path = file_path:sub(#git_root + 2)
   local cursor_line = vim.fn.line(".")
   local url = string.format("%s/%s/-/blob/%s/%s#L%s", gitlab_url, repo_path, branch, relative_path, cursor_line)

   vim.fn.system(string.format("xdg-open '%s'", url))
   print("Opening: " .. url)
end

vim.keymap.set("n", "<leader>ob", open_in_tds_gitlab, { desc = "Open current file in TDS GitLab at cursor" })

local function open_git_online()
   local remote_url = vim.fn.system("git config --get remote.origin.url"):gsub("\n", "")
   local repo_path
   local base_url

   if remote_url:match("github.com") then
      repo_path = remote_url:match("git@github.com:(.+)%.git") or remote_url:match("https://github.com/(.+)%.git")
      base_url = "https://github.com"
   elseif remote_url:match("gitlab.com") then
      repo_path = remote_url:match("git@gitlab.com:(.+)%.git") or remote_url:match("https://gitlab.com/(.+)%.git")
      base_url = "https://gitlab.com"
   else
      print("Error: Unsupported remote host!")
      return
   end

   local branch = vim.fn.system("git branch --show-current"):gsub("\n", "")
   local file_path = vim.fn.expand("%:p")
   local git_root = vim.fn.system("git rev-parse --show-toplevel"):gsub("\n", "")

   if not repo_path or git_root == "" then
      print("Error: Not a Git repository or remote not configured!")
      return
   end

   local relative_path = file_path:sub(#git_root + 2)
   local cursor_line = vim.fn.line(".")

   local url
   if base_url == "https://github.com" then
      url = string.format("%s/%s/blob/%s/%s#L%s", base_url, repo_path, branch, relative_path, cursor_line)
   elseif base_url == "https://gitlab.com" then
      url = string.format("%s/%s/-/blob/%s/%s#L%s", base_url, repo_path, branch, relative_path, cursor_line)
   end

   vim.fn.system(string.format("xdg-open '%s'", url))
   print("Opening: " .. url)
end

vim.keymap.set("n", "<leader>og", open_git_online, { desc = "Open current file in GitHub or GitLab at cursor" })
