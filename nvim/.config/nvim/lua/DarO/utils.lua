local M = {}

--- Get highlight properties for a given highlight name
--- @param name string The highlight group name
--- @param fallback? table The fallback highlight properties
--- @return table properties # the highlight group properties
function M.get_hlgroup(name, fallback)
   if vim.fn.hlexists(name) == 1 then
      local group = vim.api.nvim_get_hl(0, { name = name })

      local hl = {
         fg = group.fg == nil and "NONE" or M.parse_hex(group.fg),
         bg = group.bg == nil and "NONE" or M.parse_hex(group.bg),
      }

      return hl
   end
   return fallback or {}
end

--- Remove a buffer by its number without affecting window layout
--- @param buf? number The buffer number to delete
function M.delete_buffer(buf)
   if buf == nil or buf == 0 then
      buf = vim.api.nvim_get_current_buf()
   end

   vim.api.nvim_command("bwipeout " .. buf)
end

--- Switch to the previous buffer
function M.switch_to_previous_buffer()
   local ok, _ = pcall(function()
      vim.cmd("buffer #")
   end)
   if not ok then
      vim.notify("No other buffer to switch to!", 3, { title = "Warning" })
   end
end

--- Get the number of open buffers
--- @return number
function M.get_buffer_count()
   local count = 0
   for _, buf in ipairs(vim.api.nvim_list_bufs()) do
      if vim.fn.bufname(buf) ~= "" then
         count = count + 1
      end
   end
   return count
end

--- Parse a given integer color to a hex value.
--- @param int_color number
function M.parse_hex(int_color)
   return string.format("#%x", int_color)
end

--- Create a centered floating window of a given width and height, relative to the size of the screen.
--- @param width number width of the window where 1 is 100% of the screen
--- @param height number height of the window - between 0 and 1
--- @param buf number The buffer number
--- @return number The window number
function M.open_centered_float(width, height, buf)
   buf = buf or vim.api.nvim_create_buf(false, true)
   local win_width = math.floor(vim.o.columns * width)
   local win_height = math.floor(vim.o.lines * height)
   local offset_y = math.floor((vim.o.lines - win_height) / 2)
   local offset_x = math.floor((vim.o.columns - win_width) / 2)

   local win = vim.api.nvim_open_win(buf, true, {
      relative = "editor",
      width = win_width,
      height = win_height,
      row = offset_y,
      col = offset_x,
      style = "minimal",
      border = "single",
   })

   return win
end

--- Open the help window in a floating window
--- @param buf number The buffer number
function M.open_help(buf)
   if buf ~= nil and vim.bo[buf].filetype == "help" then
      local help_win = vim.api.nvim_get_current_win()
      local new_win = M.open_centered_float(0.6, 0.7, buf)

      vim.api.nvim_buf_set_keymap(buf, "n", "q", ":q!<CR>", {
         nowait = true,
         noremap = true,
         silent = true,
      })

      vim.wo[help_win].scroll = vim.wo[new_win].scroll
      vim.api.nvim_win_close(help_win, true)
   end
end

--- Run a shell command and return the output
--- @param cmd table The command to run in the format { "command", "arg1", "arg2", ... }
--- @param cwd? string The current working directory
--- @return table stdout, number? return_code, table? stderr
function M.get_cmd_output(cmd, cwd)
   if type(cmd) ~= "table" then
      vim.notify("Command must be a table", vim.log.levels.ERROR, { title = "Error" })
      return {}, -1, {}
   end

   local result = vim.system(cmd, {
      cwd = cwd,
      text = true,
      timeout = 30000
   }):wait()

   if result.signal == 2 then
      vim.notify("Command interrupted by user", vim.log.levels.WARN)
   elseif result.code and result.code ~= 0 then
      local cmd_str = table.concat(cmd, " ")
      if #cmd_str > 50 then
         cmd_str = cmd_str:sub(1, 47) .. "..."
      end
      vim.notify(string.format("Command failed: %s (exit code: %d)", cmd_str, result.code), vim.log.levels.ERROR)
   end

   local stdout = {}
   if result.stdout then
      for line in (result.stdout .. "\n"):gmatch("([^\r\n]*)\r?\n") do
         if line ~= "" then
            table.insert(stdout, line)
         end
      end
   end

   local stderr = {}
   if result.stderr then
      for line in (result.stderr .. "\n"):gmatch("([^\r\n]*)\r?\n") do
         if line ~= "" then
            table.insert(stderr, line)
         end
      end
   end

   return stdout, result.code or -1, stderr
end

--- Write a table of lines to a file
--- @param file string Path to the file
--- @param lines table Table of lines to write to the file
function M.write_to_file(file, lines)
   if not lines or #lines == 0 then
      return
   end
   local buf = io.open(file, "w")
   for _, line in ipairs(lines) do
      if buf ~= nil then
         buf:write(line .. "\n")
      end
   end

   if buf ~= nil then
      buf:close()
   end
end

--- Display a diff between the current buffer and a given file
--- @param file string The file to diff against the current buffer
function M.diff_file(file)
   local pos = vim.fn.getpos(".")
   local current_file = vim.fn.expand("%:p")
   vim.cmd("edit " .. file)
   vim.cmd("vert diffsplit " .. current_file)
   vim.fn.setpos(".", pos)
end

--- Display a diff between a file at a given commit and the current buffer
--- @param commit string The commit hash
--- @param file_path string The file path
function M.diff_file_from_history(commit, file_path)
   local extension = vim.fn.fnamemodify(file_path, ":e") == "" and "" or "." .. vim.fn.fnamemodify(file_path, ":e")
   local temp_file_path = os.tmpname() .. extension

   local cmd = { "git", "show", commit .. ":" .. file_path }
   local out = M.get_cmd_output(cmd)

   M.write_to_file(temp_file_path, out)
   M.diff_file(temp_file_path)
end

function M.telescope_diff_file()
   require("telescope.builtin").find_files({
      prompt_title = "Select File to Compare",
      attach_mappings = function(prompt_bufnr)
         local actions = require("telescope.actions")
         local action_state = require("telescope.actions.state")

         actions.select_default:replace(function()
            actions.close(prompt_bufnr)
            local selection = action_state.get_selected_entry()
            M.diff_file(selection.value)
         end)
         return true
      end,
   })
end

function M.telescope_diff_from_history()
   local current_file = vim.fn.fnamemodify(vim.fn.expand("%:p"), ":~:."):gsub("\\", "/")
   require("telescope.builtin").git_commits({
      git_command = { "git", "log", "--pretty=oneline", "--abbrev-commit", "--follow", "--", current_file },
      attach_mappings = function(prompt_bufnr)
         local actions = require("telescope.actions")
         local action_state = require("telescope.actions.state")

         actions.select_default:replace(function()
            actions.close(prompt_bufnr)
            local selection = action_state.get_selected_entry()
            M.diff_file_from_history(selection.value, current_file)
         end)
         return true
      end,
   })
end

function M.is_git_repo()
   local git_path    = vim.uv.cwd() .. "/.git"
   local is_git_repo = vim.uv.fs_stat(git_path)

   return is_git_repo
end

function M.print_startup_time()
   if vim.g.hide_startup_info then
      return
   end

   local elapsed_time = (vim.uv.hrtime() - vim.g.startup_time) / 1e6
   local v = vim.version()
   print(string.format("Hello DarO, Neovim v%d.%d.%d startup time: %.2f ms", v.major, v.minor, v.patch, elapsed_time))
   vim.defer_fn(function()
      vim.cmd('messages')
   end, 50)
end

--- Format only modified lines with LSP if autoformat is enabled
--- Uses lsp-format-modifications to format only git-changed lines
--- Falls back to full buffer format if range formatting is unavailable
--- Displays a warning notification if autoformat is disabled
function M.format_buffer()
   if vim.g.disable_autoformat then
      vim.notify("Unable to format: formatting is disabled (use <leader>tf to enable)", vim.log.levels.WARN)
      return
   end
   local bufnr = vim.api.nvim_get_current_buf()
   local clients = vim.lsp.get_clients({ bufnr = bufnr })

   if #clients == 0 then
      vim.notify("No LSP clients attached to buffer", vim.log.levels.WARN)
      return
   end

   local format_client = nil
   for _, client in ipairs(clients) do
      if client:supports_method("textDocument/rangeFormatting", { bufnr = bufnr }) then
         format_client = client
         break
      end
   end

   if format_client then
      local ok, format_modifications = pcall(require, "lsp-format-modifications")
      if ok then
         format_modifications.format_modifications(format_client, bufnr, {
            format_callback = function(params)
               vim.lsp.buf.format(vim.tbl_extend("force", params or {}, {
                  bufnr = bufnr,
                  async = false,
                  timeout_ms = 5000,
                  filter = function(client)
                     return not vim.g.disable_autoformat
                         and client:supports_method("textDocument/rangeFormatting", { bufnr = bufnr })
                  end
               }))
            end,
            vcs = "git",
            experimental_empty_line_handling = false,
         })
         return
      end
   end

   vim.notify("Range formatting unavailable, formatting entire buffer", vim.log.levels.INFO)
   vim.lsp.buf.format({
      bufnr = bufnr,
      async = false,
      timeout_ms = 5000,
      filter = function(client)
         return not vim.g.disable_autoformat
             and client:supports_method("textDocument/formatting", { bufnr = bufnr })
      end
   })
end

--- GitHub URL shorthand for vim.pack.add
--- @param x string Repository path (e.g. "org/repo")
--- @return string url Full GitHub URL
M.gh = function(x) return 'https://github.com/' .. x end

return M
