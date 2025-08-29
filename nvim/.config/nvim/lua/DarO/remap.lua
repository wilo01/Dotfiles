vim.g.mapleader = " "

-- Text Actions
vim.keymap.set("v", "J", ":m '>+1<CR>gv=gv", { desc = "Move selected text down" })
vim.keymap.set("v", "K", ":m '<-2<CR>gv=gv", { desc = "Move selected text up" })
vim.keymap.set("n", "H", "gt0", { desc = "Move cursor to the begginig of the current line" })
vim.keymap.set("n", "L", "gt$", { desc = "Move cursor to the end of the current line" })
vim.keymap.set("n", "<A-h>", "mzJ`z", { desc = "Move text lines without moving cursor" })
vim.keymap.set('n', '<leader>/', '/<C-r>+<CR>zz', { desc = "Search with clipboard text" })
vim.keymap.set('n', '<C-i>', '"+yi[', { desc = "Yank inside square brackets" })
vim.keymap.set("n", "<leader>R", [[:%s/\<<C-r><C-w>\>/<C-r><C-w>/gI<Left><Left><Left>]],
   { desc = "Replace text occurrences of the word under cursor" })
vim.keymap.set("v", "<leader>8", "c**<C-r>\"**<Esc>", { desc = "Markdown Bold text with **TEXT**" })
vim.keymap.set("v", "<leader>{", "c{<C-r>\"}<Esc>", { desc = "Wrap text with { }" })
vim.keymap.set("v", "<leader>}", "c{<C-r>\"}<Esc>", { desc = "Wrap text with { }" })
vim.keymap.set("v", "<leader>[", "c[<C-r>\"]<Esc>", { desc = "Wrap text with [ ]" })
vim.keymap.set("v", "<leader>]", "c[<C-r>\"]<Esc>", { desc = "Wrap text with [ ]" })
vim.keymap.set("v", "<leader>(", "c(<C-r>\")<Esc>", { desc = "Wrap text with ( )" })
vim.keymap.set("v", "<leader>)", "c(<C-r>\")<Esc>", { desc = "Wrap text with ( )" })
vim.keymap.set("v", '<leader>"', 'c"<C-r>""<Esc>', { desc = 'Wrap text with " "' })
vim.keymap.set("v", "<leader>'", "c'<C-r>\"'<Esc>", { desc = "Wrap text with ' '" })
vim.keymap.set("n", "<C-_>", "gcc", { desc = "Toggle comment for current line", remap = true })
vim.keymap.set("x", "<C-_>", "gc", { desc = "Toggle comment for visual selection", remap = true })

-- Escape Mode
vim.keymap.set({ "n", "i", "v" }, "qq", "<Esc>", { desc = "Escape with qq" })
vim.keymap.set("i", "<C-c>", "<Esc>", { desc = "Escape insert mode with Ctrl+C" })
vim.keymap.set("n", "<Esc>", "<cmd>nohlsearch<CR>", { desc = "Escape with no hl search" })

-- Navigation Enhancements
vim.keymap.set("n", "<C-d>", "<C-d>zz", { desc = "Scroll down and center" })
vim.keymap.set("n", "<C-u>", "<C-u>zz", { desc = "Scroll up and center" })
vim.keymap.set("n", "n", "nzzzv", { desc = "Center cursor on next search result" })
vim.keymap.set("n", "N", "Nzzzv", { desc = "Center cursor on previous search result" })
vim.keymap.set("n", "<A-j>", "}zz", { desc = "Jump to next empty line and center" })
vim.keymap.set("n", "<A-k>", "{zz", { desc = "Jump to previous empty line and center" })

-- Editing Utilities
vim.keymap.set("n", "x", '"_x', { desc = "Delete character without yanking" })
vim.keymap.set("n", "<leader>+", '<C-a>', { desc = "Increment number" })
vim.keymap.set("n", "<leader>=", "<C-a>", { desc = "Increment number" })
vim.keymap.set("x", "<leader>=", "g<C-a>", { desc = "Increment numbers across selection" })
vim.keymap.set("n", "<leader>-", '<C-x>', { desc = "Decrement number" })
vim.keymap.set("x", "<leader>-", 'g<C-x>', { desc = "Decrement numbers across selection" })
vim.keymap.set({ "n", "v" }, "yc", "yy<cmd>normal gcc<CR>p", { desc = "Duplicate a line and comment out the first line" })
vim.keymap.set("n", "<leader>oc", function()
   local filenameAndLine = vim.fn.expand("%:t") .. ":" .. vim.fn.line(".")
   local escaped = filenameAndLine:gsub([[\]], [[\\]])
       :gsub([["]], [[\\"]])
       :gsub([[']], [[\']])
   local script = [[
    tell application "Google Chrome"
      activate
      tell application "System Events"
        keystroke "i" using {command down, option down}
        delay 0.5
        keystroke "p" using command down
        delay 1
        keystroke "<<filenameAndLine>>"
      end tell
    end tell
  ]]

   script = script:gsub("<<filenameAndLine>>", escaped)
   vim.print("Running script: " .. script)
   vim.system({
      "osascript",
      "-e",
      script,
   })
end, { desc = "Open chrome dev tools and run \"open file\" with current file and line" })
vim.keymap.set('n', '<leader>,', function()
   local word = vim.fn.expand('<cword>')
   if word == 'true' then
      vim.cmd('normal! ciwfalse')
   elseif word == 'false' then
      vim.cmd('normal! ciwtrue')
   end
end, { desc = "Toggle true/false" })
vim.keymap.set("n", "<leader>c/", function()
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
end, { desc = "Remove comment from current line" })

vim.keymap.set("n", "<leader>*", function()
   local gitsigns = require('gitsigns')
   local ft = vim.bo.filetype
   local comment_patterns = {
      javascript = "//",
      lua = "%-%-",
      python = "#",
      c = "//",
      sh = "#",
      bash = "#",
      zsh = "#",
      vim = '"',
      sql = "%-%- ",
   }

   local pattern = comment_patterns[ft]

   if not pattern then
      vim.notify("No comment pattern defined for filetype: " .. ft, vim.log.levels.WARN)
      return
   end

   local bufnr = vim.api.nvim_get_current_buf()
   local hunks = gitsigns.get_hunks(bufnr)

   if not hunks or #hunks == 0 then
      vim.notify("No git changes found in current buffer", vim.log.levels.INFO)
      return
   end

   local removed_count = 0
   local deleted_lines = 0
   local lines_to_delete = {}

   local function find_trailing_comment(line, comment_pattern)
      local in_string = false
      local string_char = nil
      local escaped = false

      for i = 1, #line do
         local char = line:sub(i, i)
         local next_chars = line:sub(i, math.min(i + #comment_pattern - 1, #line))

         if escaped then
            escaped = false
         elseif char == "\\" and in_string then
            escaped = true
         elseif not in_string and (char == '"' or char == "'") then
            in_string = true
            string_char = char
         elseif in_string and char == string_char and not escaped then
            in_string = false
            string_char = nil
         elseif not in_string and next_chars == comment_pattern then
            local before = line:sub(1, i - 1)
            if before:match("[%s,;%)%}%]%>]$") or before:match("^%s*$") then
               return i
            end
         end
      end

      return nil
   end

   for _, hunk in ipairs(hunks) do
      if hunk.added and hunk.added.start and hunk.added.count > 0 then
         local start_line = hunk.added.start - 1 -- Convert to 0-based
         local end_line = start_line + hunk.added.count - 1

         for line_idx = start_line, end_line do
            local lines = vim.api.nvim_buf_get_lines(bufnr, line_idx, line_idx + 1, false)
            if lines[1] then
               local original_line = lines[1]

               if original_line:match("^%s*" .. pattern) then
                  table.insert(lines_to_delete, line_idx)
                  removed_count = removed_count + 1
               else
                  local comment_pos = find_trailing_comment(original_line, pattern:gsub("%%", ""))
                  if comment_pos then
                     local updated_line = original_line:sub(1, comment_pos - 1):gsub("%s+$", "")
                     vim.api.nvim_buf_set_lines(bufnr, line_idx, line_idx + 1, false, { updated_line })
                     removed_count = removed_count + 1
                  end
               end
            end
         end
      end
   end

   table.sort(lines_to_delete, function(a, b) return a > b end)
   for _, line_idx in ipairs(lines_to_delete) do
      vim.api.nvim_buf_set_lines(bufnr, line_idx, line_idx + 1, false, {})
      deleted_lines = deleted_lines + 1
   end

   local msg = string.format("Removed comments from %d changed lines", removed_count)
   if deleted_lines > 0 then
      msg = msg .. string.format(" (deleted %d empty lines)", deleted_lines)
   end
   vim.notify(msg, vim.log.levels.INFO)
end, { desc = "Remove comments from git-changed lines (removes Standalone and Trailing comments)" })

-- Clipboard Operations
vim.keymap.set("x", "p", "\"_dP", { desc = "Replace with yanked text, and keep yanked" })
vim.keymap.set("x", "<leader>p", "p", { desc = "Replace with yanked text, and new yanked text" })
vim.keymap.set({ "n", "v" }, "<leader>y", "\"+y", { desc = "Yank to clipboard" })
vim.keymap.set("n", "<leader>Y", "\"+Y", { desc = "Yank line to clipboard" })
vim.keymap.set("v", "<C-c>", "\"+y", { desc = "Yank selection to clipboard with Ctrl+C" })
vim.keymap.set({ "n", "v" }, "<leader>z", "\"_d", { desc = "Delete without yanking" })
vim.keymap.set({ "n", "v" }, "<leader>zy", "\"+d", { desc = "Delete & Yank to clipboard" })
-- vim.keymap.set('n', '<leader><leader>', ':%y+<CR>', { desc = "Yank entire file to clipboard without moving cursor" })
vim.keymap.set('n', '<leader><leader>', 'ggVG"+y', { desc = "Select all and yank to clipboard" })
-- vim.keymap.set("n", "<leader><leader>", function()
--    vim.cmd("silent !tmux split-window -dh")
-- end, { desc = "Open remap file in new tmux split by sending keys" })

-- Disabling Default Mappings
vim.keymap.set("n", "Q", "<nop>", { desc = "Disable 'Q'" })

-- Tmux Integration (Needs tmux-sessionizer export in shell rc)
vim.keymap.set("n", "<C-f>", "<cmd>silent !tmux neww tmux-sessionizer<CR>",
   { desc = "Switch projects using tmux-sessionizer" })

-- LSP Formatting
vim.keymap.set({ "n", "v" }, "<leader>f", function()
   vim.lsp.buf.format()
   vim.cmd("write")
end, { desc = "Format and save with LSP" })
vim.keymap.set("n", "<leader>d", function()
   vim.diagnostic.open_float(nil, { focusable = false, source = "if_many" })
end, { desc = "Show diagnostic errors and warnings in a floating window" })

-- Quickfix and Location List Navigation
vim.keymap.set("n", "<C-j>", "<cmd>cnext<CR>zz", { desc = "Next quickfix item" })
vim.keymap.set("n", "<C-k>", "<cmd>cprev<CR>zz", { desc = "Previous quickfix item" })
vim.keymap.set("n", "<leader>k", "<cmd>lnext<CR>zz", { desc = "Next location list item" })
vim.keymap.set("n", "<leader>j", "<cmd>lprev<CR>zz", { desc = "Previous location list item" })

-- Quickfix Preview: Use <C-CR> in quickfix window to preview without switching
vim.api.nvim_create_autocmd("FileType", {
   pattern = "qf",
   callback = function()
      -- Map Ctrl+Enter to preview current quickfix item without leaving quickfix window
      vim.keymap.set("n", "<C-CR>", function()
         -- Get current quickfix entry
         local qf_idx = vim.fn.line('.')
         local qf_list = vim.fn.getqflist()
         
         if qf_idx > 0 and qf_idx <= #qf_list then
            local entry = qf_list[qf_idx]
            
            -- Find the main window (not quickfix)
            local main_win = nil
            for _, win in ipairs(vim.api.nvim_list_wins()) do
               local buf = vim.api.nvim_win_get_buf(win)
               local ft = vim.api.nvim_buf_get_option(buf, 'filetype')
               if ft ~= 'qf' then
                  main_win = win
                  break
               end
            end
            
            if main_win and entry.bufnr > 0 then
               -- Switch to main window temporarily
               vim.api.nvim_set_current_win(main_win)
               -- Load the buffer
               vim.api.nvim_win_set_buf(main_win, entry.bufnr)
               -- Jump to the line
               vim.api.nvim_win_set_cursor(main_win, {entry.lnum, entry.col - 1})
               -- Center the screen
               vim.cmd('normal! zz')
               -- Switch back to quickfix window
               vim.cmd('wincmd p')
            end
         end
      end, { buffer = true, desc = "Preview quickfix item without switching (stay in quickfix)" })
   end,
   desc = "Setup quickfix preview keymaps"
})

-- File Operations / File Navigation
vim.keymap.set("n", "<leader>v", vim.cmd.Ex, { desc = "Open Netrw" })
vim.keymap.set("n", "gb", "<C-o>zz", { desc = "Go back, and center" })
vim.keymap.set("n", "<leader>x", "<cmd>!chmod +x %<CR>", { desc = "Make file executable", silent = true })
vim.keymap.set("n", "<leader>mr", "<cmd>CellularAutomaton make_it_rain<CR>", { desc = "Make it rain animation" })
vim.keymap.set('n', '<leader>fp', function()
   local filepath = vim.fn.expand('%:p')
   local home_dir = vim.fn.getenv('HOME')
   local file_path = filepath:gsub('^' .. home_dir, '~')

   vim.fn.setreg('+', file_path)
   vim.notify('Copied path: ' .. file_path)
end, { desc = 'Copy current file path to clipboard (pwd)' })
vim.keymap.set('n', '<leader>rp', function()
   local relative_path = vim.fn.expand('%:~:.')
   vim.fn.setreg('+', relative_path)
   vim.notify('Copied path: ' .. relative_path)
end, { desc = 'Copy current file path to clipboard (relative)' })
vim.keymap.set("n", "<leader>ov", function()
   local filepath = vim.fn.expand('%:p')
   vim.system({ 'code', filepath })
end, { desc = "Open current file in VSCode" })

-- Markdown Preview
vim.keymap.set("n", "<leader>m", "<CMD>MarkdownPreview<CR>", { desc = "Start Markdown preview" })
vim.keymap.set("n", "<leader>mn", "<CMD>MarkdownPreviewStop<CR>", { desc = "Stop Markdown preview" })

-- Gitsigns Integration
vim.keymap.set("n", "<leader>va", "<CMD>Gitsigns preview_hunk_inline<CR>", { desc = "Gitsigns preview Git hunk" })
vim.keymap.set("n", "<leader>vs", "<CMD>Gitsigns diffthis<CR>", { desc = "Gitsigns Diff current buffer" })
vim.keymap.set("n", "<leader>bl", "<CMD>Gitsigns blame<CR>", { desc = "Gitsigns Blame current file" }) -- [ ] TODO: Add toggle blame
vim.keymap.set("n", "<leader>vt", "<CMD>Gitsigns toggle_deleted<CR>", { desc = "Gitsigns Toggle deleted lines" })
vim.keymap.set("n", "<leader>vb", "<CMD>Gitsigns blame_line<CR>", { desc = "Gitsigns Blame current line" })
vim.keymap.set("n", "<leader>rg", "<CMD>Gitsigns reset_hunk<CR>", { desc = "Gitsigns Reset Hunk (Reset git, diff)" })
vim.keymap.set("n", "<leader>sh", "<CMD>Gitsigns stage_hunk<CR>", { desc = "Gitsigns Stage Hunk" })
vim.keymap.set("n", "<leader>sf", "<CMD>Gitsigns stage_buffer<CR>", { desc = "Gitsigns Stage entire File/Buffer" })
vim.keymap.set("n", "<leader>uf", "<CMD>Gitsigns reset_buffer_index<CR>",
   { desc = "Gitsigns Unstage entire File/Buffer" }) -- [ ] TODO: Toggle Unstage entire File/Buffer
vim.keymap.set("n", "J", "<CMD>Gitsigns next_hunk<CR>zz", { desc = "Gitsigns go to next Git hunk and jump to center" })
vim.keymap.set("n", "K", "<CMD>Gitsigns prev_hunk<CR>zz",
   { desc = "Gitsigns go to previous Git hunk and jump to center" })

-- Insert Console Snippets
vim.keymap.set("v", "<leader>cl", function()
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
end, { desc = "Insert object console.warn snippet with selection (log, debugger)" })

vim.keymap.set("v", "<leader>cn", function()
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
end, { desc = "Insert without object console.warn snippet with selection (log, debugger)" })

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
      "ca_log_pak.log_warning('Dwdw', '" ..
      clipboard_content .. ": ' || " .. clipboard_content .. ", '" .. current_file .. "');"
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

-- Spellinguseful actions with description
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

-- GitHub, Gitlab Integration
vim.keymap.set("n", "<leader>gr", function()
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
end, { desc = "Git open current file in TDS GitHub / GitLab at cursor" })

local function open_git_online()
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
end

vim.keymap.set("n", "<leader>og", open_git_online, { desc = "Open current file in GitHub or GitLab at cursor" })

-- CSV editing format (Auto close on save in -> autocmds.lua)
vim.g.is_csv_prettified = false
vim.keymap.set("n", "<leader>t", function()
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
end, { desc = "Toggle CSV formatting for csv edit", noremap = true, silent = true })

-- Quickfix Navigation
vim.keymap.set("n", "<C-w>p", "<C-w>p", { desc = "Toggle between quickfix and file (previous window)" })
vim.keymap.set("n", "<C-w>w", "<C-w>w", { desc = "Cycle through all windows with quickfix" })
vim.keymap.set("n", "<C-w>j", "<C-w>j", { desc = "Jump down to window quickfix" })
vim.keymap.set("n", "<C-w>k", "<C-w>k", { desc = "Jump up to window from quickfix" })
