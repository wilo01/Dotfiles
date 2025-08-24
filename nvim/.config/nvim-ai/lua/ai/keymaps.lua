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
map("x", "p", '"_dP', desc("Replace with yanked text, and keep yanked"))
map("x", "<leader>p", "p", desc("Replace with yanked text, and new yanked text"))
map({ "n", "v" }, "<leader>y", '"+y', desc("Yank to system clipboard"))
map("n", "<leader>Y", '"+Y', desc("Yank line to system clipboard"))
map("v", "<C-c>", '"+y', desc("Yank selection to clipboard with Ctrl+C"))
map({ "n", "v" }, "<leader>z", '"_d', desc("Delete without yanking"))
map({ "n", "v" }, "<leader>zy", '"+d', desc("Delete & Yank to clipboard"))
map("n", "<leader><leader>", 'ggVG"+y', desc("Select all and yank to clipboard"))
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
map("v", "<leader>}", [[c{<C-r>"}<Esc>]], desc("Wrap with {}"))
map("v", "<leader>[", [[c[<C-r>"]<Esc>]], desc("Wrap with []"))
map("v", "<leader>]", [[c[<C-r>"]<Esc>]], desc("Wrap with []"))
map("v", "<leader>(", [[c(<C-r>")<Esc>]], desc("Wrap with ()"))
map("v", "<leader>)", [[c(<C-r>")<Esc>]], desc("Wrap with ()"))
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
map("n", "<leader>v", vim.cmd.Ex, desc("Open Netrw"))
map("n", "gb", "<C-o>zz", desc("Go back, and center"))
map("n", "<leader>u", vim.cmd.UndotreeToggle, desc("Toggle undo tree"))

-- Diagnostic keymaps (built-in)
map("n", "[d", vim.diagnostic.goto_prev, desc("Previous diagnostic"))
map("n", "]d", vim.diagnostic.goto_next, desc("Next diagnostic"))
map("n", "gl", vim.diagnostic.open_float, desc("Show diagnostic"))
map("n", "<leader>cd", vim.diagnostic.setloclist, desc("Diagnostics to location list"))

-- Quickfix and location list
map("n", "<C-j>", "<cmd>cnext<CR>zz", desc("Next quickfix item"))
map("n", "<C-k>", "<cmd>cprev<CR>zz", desc("Previous quickfix item"))
map("n", "<leader>k", "<cmd>lnext<CR>zz", desc("Next location list item"))
map("n", "<leader>j", "<cmd>lprev<CR>zz", desc("Previous location list item"))
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
map({ "n", "v" }, "<leader>f", function()
   vim.lsp.buf.format()
   vim.cmd("write")
end, desc("Format and save with LSP"))

map("n", "<leader>d", function()
   vim.diagnostic.open_float(nil, { focusable = false, source = "if_many" })
end, desc("Show diagnostic errors and warnings in a floating window"))

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

-- File Operations / File Navigation
map("n", "<leader>x", "<cmd>!chmod +x %<CR>", desc("Make file executable"))
map("n", "<leader>mr", "<cmd>CellularAutomaton make_it_rain<CR>", desc("Make it rain animation"))
map('n', '<leader>fp', function()
   local filepath = vim.fn.expand('%:p')
   local home_dir = vim.fn.getenv('HOME')
   local file_path = filepath:gsub('^' .. home_dir, '~')
   vim.fn.setreg('+', file_path)
   vim.notify('Copied path: ' .. file_path)
end, desc('Copy current file path to clipboard (pwd)'))
map('n', '<leader>rp', function()
   local relative_path = vim.fn.expand('%:~:.')
   vim.fn.setreg('+', relative_path)
   vim.notify('Copied path: ' .. relative_path)
end, desc('Copy current file path to clipboard (relative)'))
map("n", "<leader>ov", function()
   local filepath = vim.fn.expand('%:p')
   vim.system({ 'code', filepath })
end, desc("Open current file in VSCode"))

-- Markdown Preview
map("n", "<leader>m", "<CMD>MarkdownPreview<CR>", desc("Start Markdown preview"))
map("n", "<leader>mn", "<CMD>MarkdownPreviewStop<CR>", desc("Stop Markdown preview"))

-- Tmux Integration (Needs tmux-sessionizer export in shell rc)
map("n", "<C-f>", "<cmd>silent !tmux neww tmux-sessionizer<CR>", desc("Switch projects using tmux-sessionizer"))

-- Disabling Default Mappings
map("n", "Q", "<nop>", desc("Disable 'Q'"))

-- Yank inside square brackets
map('n', '<C-i>', '"+yi[', desc("Yank inside square brackets"))

-- Alternative for join lines (since J is used for Git navigation)
map("n", "<leader>J", "J", desc("Join lines (standard vim J)"))

-- Gitsigns Integration (will be setup in gitsigns config but added here for reference)
map("n", "<leader>va", "<CMD>Gitsigns preview_hunk_inline<CR>", desc("Gitsigns preview Git hunk"))
map("n", "<leader>vs", "<CMD>Gitsigns diffthis<CR>", desc("Gitsigns Diff current buffer"))
map("n", "<leader>bl", "<CMD>Gitsigns blame<CR>", desc("Gitsigns Blame current file"))
map("n", "<leader>vt", "<CMD>Gitsigns toggle_deleted<CR>", desc("Gitsigns Toggle deleted lines"))
map("n", "<leader>vb", "<CMD>Gitsigns blame_line<CR>", desc("Gitsigns Blame current line"))
map("n", "<leader>rg", "<CMD>Gitsigns reset_hunk<CR>", desc("Gitsigns Reset Hunk (Reset git, diff)"))
map("n", "<leader>sh", "<CMD>Gitsigns stage_hunk<CR>", desc("Gitsigns Stage Hunk"))
map("n", "<leader>sf", "<CMD>Gitsigns stage_buffer<CR>", desc("Gitsigns Stage entire File/Buffer"))
map("n", "<leader>uf", "<CMD>Gitsigns reset_buffer_index<CR>", desc("Gitsigns Unstage entire File/Buffer"))
map("n", "J", "<CMD>Gitsigns next_hunk<CR>zz", desc("Gitsigns go to next Git hunk and jump to center"))
map("n", "K", "<CMD>Gitsigns prev_hunk<CR>zz", desc("Gitsigns go to previous Git hunk and jump to center"))

-- Insert Console Snippets
map("v", "<leader>cl", function()
   vim.cmd('normal! "+y')
   local selected_text = vim.fn.getreg('+')
   local snippet = {
      "console.warn('', {",
      string.format("\t'%s': %s,", selected_text, selected_text),
      "});"
   }
   vim.api.nvim_put(snippet, 'l', true, true)
   vim.lsp.buf.format()
   vim.cmd("write")
end, desc("Insert object console.warn snippet with selection"))

map("v", "<leader>cn", function()
   vim.cmd('normal! "+y')
   local selected_text = vim.fn.getreg('+')
   local snippet = {
      "console.warn(",
      string.format("\t'%s'", selected_text),
      ");"
   }
   vim.api.nvim_put(snippet, 'l', true, true)
   vim.lsp.buf.format()
   vim.cmd("write")
end, desc("Insert without object console.warn snippet with selection"))

map("v", "<leader>ck", function()
   vim.cmd('normal! "+y')
   local selected_text = vim.fn.getreg('+')
   local snippet = {
      "if (" .. selected_text .. ") {",
      "}"
   }
   vim.api.nvim_put(snippet, 'l', true, true)
end, desc("Insert if statement snippet with clipboard content"))

map("v", "<leader>cj", function()
   vim.cmd('normal! "+y')
   local clipboard_content = vim.fn.getreg('+')
   local current_file = vim.fn.expand('%:t:r')
   local snippet = {
      "ca_log_pak.log_warning('Dwdw', '" ..
      clipboard_content .. ": ' || " .. clipboard_content .. ", '" .. current_file .. "');"
   }
   vim.api.nvim_put(snippet, 'l', true, true)
end, desc("Insert ca_log_pak.log_warning snippet with clipboard content"))

map("v", "<leader>ch", function()
   vim.cmd('normal! "+y')
   local clipboard_content = vim.fn.getreg('+')
   local snippet = {
      "DBMS_OUTPUT.PUT_LINE('Dwdw " .. clipboard_content .. ": ' || " .. clipboard_content .. ");"
   }
   vim.api.nvim_put(snippet, 'l', true, true)
end, desc("Insert DBMS_OUTPUT snippet with clipboard content"))

-- Comment removal utilities
map("n", "<leader>c/", function()
   local ft = vim.bo.filetype
   local comment_patterns = {
      javascript = "// ",
      lua = "%-%- ",
      python = "# ",
      c = "// ",
      sql = "%-%- ",
   }
   
   local pattern = comment_patterns[ft]
   
   if not pattern then
      vim.notify("No comment pattern defined for filetype: " .. ft, vim.log.levels.WARN)
      return
   end
   
   local line_number = vim.fn.line(".") - 1
   local current_line = vim.api.nvim_buf_get_lines(0, line_number, line_number + 1, false)[1]
   
   if not current_line then
      vim.notify("Could not fetch the current line!", vim.log.levels.ERROR)
      return
   end
   
   local updated_line = current_line:gsub(pattern .. ".*", "")
   vim.api.nvim_buf_set_lines(0, line_number, line_number + 1, false, { updated_line })
   vim.notify("Comment removed from the current line!", vim.log.levels.INFO)
end, desc("Remove comment from current line"))

-- GitHub/GitLab Integration
map("n", "<leader>gr", function()
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
end, desc("Git open current file in TDS GitHub / GitLab at cursor"))

map("n", "<leader>og", function()
   local remote_url = vim.fn.system("git config --get remote.origin.url"):gsub("\n", "")
   local repo_path
   local base_url
   
   local git_hosts = {
      github = {
         pattern = "github",
         base_url = "https://github.com",
         ssh_pattern = "git@[^:]*github[^:]*:(.+)%.git",
         https_pattern = "https://[^/]*github[^/]*/(.+)%.git"
      },
      gitlab = {
         pattern = "gitlab",
         base_url = "https://gitlab.com",
         ssh_pattern = "git@[^:]*gitlab[^:]*:(.+)%.git",
         https_pattern = "https://[^/]*gitlab[^/]*/(.+)%.git"
      }
   }
   
   local detected_host = nil
   for host_name, host_config in pairs(git_hosts) do
      if remote_url:find(host_config.pattern) then
         detected_host = host_config
         break
      end
   end
   
   if not detected_host then
      print("Error: Unsupported remote host!")
      return
   end
   
   repo_path = remote_url:match(detected_host.ssh_pattern) or remote_url:match(detected_host.https_pattern)
   base_url = detected_host.base_url
   
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
end, desc("Open current file in GitHub or GitLab at cursor"))

-- CSV editing format (Auto close on save in -> autocmds.lua)
vim.g.is_csv_prettified = false
map("n", "<leader>t", function()
   local lines = vim.api.nvim_buf_get_lines(0, 0, -1, false)
   
   if vim.g.is_csv_prettified then
      local cleaned_lines = {}
      for _, line in ipairs(lines) do
         local cleaned_line = line:gsub("%s*,%s*", ","):gsub("%s+$", "")
         table.insert(cleaned_lines, cleaned_line)
      end
      vim.api.nvim_buf_set_lines(0, 0, -1, false, cleaned_lines)
      print("CSV prettification disabled.")
   else
      -- Configuration constants
      local MAX_COLUMN_WIDTH = 100 -- Maximum width for any column
      local ELLIPSIS = "..."
      local MAX_FORMAT_WIDTH = 144 -- Lua string.format limitation
      
      local max_lengths = {}
      
      -- Calculate maximum lengths for each column with limits
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
            
            -- Ensure we don't exceed format limits
            if max_len > MAX_FORMAT_WIDTH then
               max_len = MAX_FORMAT_WIDTH
            end
            
            -- Truncate long fields with ellipsis
            local formatted_col = col
            if #col > MAX_COLUMN_WIDTH then
               formatted_col = col:sub(1, MAX_COLUMN_WIDTH - #ELLIPSIS) .. ELLIPSIS
            end
            
            -- Safe string formatting with error handling
            local success, result = pcall(string.format, "%-" .. max_len .. "s", formatted_col)
            if success then
               cols[i] = result
            else
               -- Fallback: just pad manually if string.format fails
               cols[i] = formatted_col .. string.rep(" ", math.max(0, max_len - #formatted_col))
               vim.notify("Warning: String format failed for column " .. i .. ", using fallback padding",
                  vim.log.levels.WARN)
            end
         end
         table.insert(prettified_lines, table.concat(cols, " , "))
      end
      
      vim.api.nvim_buf_set_lines(0, 0, -1, false, prettified_lines)
      print("CSV prettification enabled.")
   end
   
   vim.g.is_csv_prettified = not vim.g.is_csv_prettified
end, desc("Toggle CSV formatting for csv edit"))

-- Spelling keymaps
map("n", "z=", "z=", desc("Spelling suggestions"))
map("n", "zg", "zg", desc("Spelling add word to spellfile as good word"))
map("n", "zG", "zG", desc("Spelling add word to internal word list as good word"))
map("n", "zw", "zw", desc("Spelling mark word as wrong (bad) in spellfile"))
map("n", "zW", "zW", desc("Spelling mark word as wrong (bad) in internal word list"))
map("n", "zug", "zug", desc("Spelling remove word from spellfile as good/bad"))
map("n", "zuG", "zuG", desc("Spelling remove word from internal word list as good/bad"))
map("n", "]s", "]s", desc("Spelling Next misspelled word"))
map("n", "[s", "[s", desc("Spelling Previous misspelled word"))
map("n", "]S", "]S", desc("Spelling Next bad word only"))
map("n", "[S", "[S", desc("Spelling Previous bad word only"))
map("n", "]r", "]r", desc("Spelling Next rare word"))
map("n", "[r", "[r", desc("Spelling Previous rare word"))