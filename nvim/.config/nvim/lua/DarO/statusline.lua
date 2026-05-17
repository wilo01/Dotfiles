local M = {}

local git_cache = {
   branch = '',
   ahead = 0,
   behind = 0,
   staged = 0,
   modified = 0,
   untracked = 0,
   stashed = 0,
   last_update = 0
}

local os_icon_cache = nil

local CACHE_DURATION = 2000

local icons = {
   folder = '',
   git = ' ',
   os = {
      ubuntu = '',
      fedora = '',
      arch = '',
      debian = '',
      centos = '',
      manjaro = '',
      opensuse = '',
      linux = '',
   },
   git_indicators = {
      ahead = '⇡',
      behind = '⇣',
      staged = '+',
      modified = '!',
      untracked = '?',
      stashed = '*'
   }
}

local function setup_highlights()
   vim.api.nvim_set_hl(0, 'StatusLineDir', { ctermfg = 39, fg = '#00afff', bold = true })
   vim.api.nvim_set_hl(0, 'StatusLineFile', { ctermfg = 220, fg = '#ffd700', bold = true })
   vim.api.nvim_set_hl(0, 'StatusLineGitClean', { ctermfg = 76, fg = '#5fd700', bold = true })
   vim.api.nvim_set_hl(0, 'StatusLineGitDirty', { ctermfg = 178, fg = '#d7af00', bold = true })
   vim.api.nvim_set_hl(0, 'StatusLineGitUntracked', { ctermfg = 39, fg = '#00afff', bold = true })
   vim.api.nvim_set_hl(0, 'StatusLineSeparator', { ctermfg = 255, fg = '#ffffff' })
   vim.api.nvim_set_hl(0, 'StatusLineInfo', { ctermfg = 255, fg = '#ffffff' })
   vim.api.nvim_set_hl(0, 'StatusLineOS', { ctermfg = 255, fg = '#ffffff', bold = true })
end

local function get_os_icon()
   if os_icon_cache then
      return os_icon_cache
   end

   local function check_os_file(file)
      local f = io.open(file, 'r')
      if not f then return nil end
      local content = f:read('*a')
      f:close()
      return content:lower()
   end

   local os_info = check_os_file('/etc/os-release') or check_os_file('/etc/lsb-release') or ''

   if os_info:match('ubuntu') then
      os_icon_cache = icons.os.ubuntu
   elseif os_info:match('fedora') then
      os_icon_cache = icons.os.fedora
   elseif os_info:match('arch') then
      os_icon_cache = icons.os.arch
   elseif os_info:match('debian') then
      os_icon_cache = icons.os.debian
   elseif os_info:match('centos') or os_info:match('rhel') then
      os_icon_cache = icons.os.centos
   elseif os_info:match('manjaro') then
      os_icon_cache = icons.os.manjaro
   elseif os_info:match('opensuse') or os_info:match('suse') then
      os_icon_cache = icons.os.opensuse
   else
      os_icon_cache = icons.os.linux
   end

   return os_icon_cache
end


-- Safe string substring with bounds checking
local function safe_sub(str, start_pos, end_pos)
   if not str or str == '' then return '' end

   local len = #str
   if start_pos < 1 then start_pos = 1 end
   if start_pos > len then return '' end

   if end_pos then
      if end_pos < start_pos then return '' end
      if end_pos > len then end_pos = len end
      return str:sub(start_pos, end_pos)
   else
      if start_pos > len then return '' end
      return str:sub(start_pos)
   end
end

local function shorten_path(filepath)
   if filepath == '' or filepath == '[No Name]' then
      return filepath
   end

   local home = vim.fn.expand('~')
   local full_path = vim.fn.expand(filepath)

   -- Safe path replacement with boundary validation
   if #full_path >= #home and safe_sub(full_path, 1, #home) == home then
      full_path = '~' .. safe_sub(full_path, #home + 1)
   end

   local parts = {}
   for part in full_path:gmatch('[^/]+') do
      table.insert(parts, part)
   end

   if #parts <= 3 then
      return full_path
   end

   local shortened = {}

   table.insert(shortened, parts[1])

   for i = 2, #parts - 2 do
      local part = parts[i] or ''
      table.insert(shortened, safe_sub(part, 1, 1))
   end

   if #parts > 1 then
      table.insert(shortened, parts[#parts - 1])
      table.insert(shortened, parts[#parts])
   end

   return table.concat(shortened, '/')
end

local function get_git_branch()
   local handle = io.popen('git branch --show-current 2>/dev/null')
   if not handle then return '' end

   local branch = handle:read("*a")
   handle:close()

   if not branch then return '' end
   branch = branch:gsub('\n', '')
   return branch ~= '' and branch or ''
end

local function get_git_ahead_behind()
   local handle = io.popen('git rev-list --count --left-right @{upstream}...HEAD 2>/dev/null')
   if not handle then return 0, 0 end

   local result = handle:read("*a")
   handle:close()

   if not result or result == '' then return 0, 0 end
   result = result:gsub('\n', '')

   if result == '' then return 0, 0 end

   local behind, ahead = result:match('(%d+)%s+(%d+)')
   return tonumber(ahead) or 0, tonumber(behind) or 0
end

local function get_git_status_counts()
   local handle = io.popen('git status --porcelain 2>/dev/null')
   if not handle then return 0, 0, 0 end

   local output = handle:read("*a")
   handle:close()

   if output == '' then return 0, 0, 0 end

   local staged = 0
   local modified = 0
   local untracked = 0

   for line in output:gmatch('[^\r\n]+') do
      -- Safe extraction of git status characters with bounds checking
      if line and #line >= 2 then
         local index_status = safe_sub(line, 1, 1)
         local work_status = safe_sub(line, 2, 2)

         if index_status:match('[MADRC]') then
            staged = staged + 1
         end

         if work_status:match('[MADRC]') then
            modified = modified + 1
         end

         if index_status == '?' and work_status == '?' then
            untracked = untracked + 1
         end
      end
   end

   return staged, modified, untracked
end

local function get_git_stash_count()
   local handle = io.popen('git stash list 2>/dev/null | wc -l')
   if not handle then return 0 end

   local count = handle:read("*a")
   handle:close()

   if not count then return 0 end
   count = count:gsub('\n', '')

   return tonumber(count) or 0
end

local function update_git_cache()
   local current_time = vim.uv.now()

   if current_time - git_cache.last_update < CACHE_DURATION then
      return
   end

   local handle = io.popen('git rev-parse --is-inside-work-tree 2>/dev/null')
   if not handle then
      -- Reset git cache on command failure
      git_cache = {
         branch = '',
         ahead = 0,
         behind = 0,
         staged = 0,
         modified = 0,
         untracked = 0,
         stashed = 0,
         last_update = current_time
      }
      return
   end

   local is_git_repo = handle:read("*a")
   handle:close()

   if not is_git_repo then
      git_cache = {
         branch = '',
         ahead = 0,
         behind = 0,
         staged = 0,
         modified = 0,
         untracked = 0,
         stashed = 0,
         last_update = current_time
      }
      return
   end

   is_git_repo = is_git_repo:gsub('\n', '')

   if is_git_repo ~= 'true' then
      git_cache = {
         branch = '',
         ahead = 0,
         behind = 0,
         staged = 0,
         modified = 0,
         untracked = 0,
         stashed = 0,
         last_update = current_time
      }
      return
   end

   -- Safely update git cache with error handling
   pcall(function()
      git_cache.branch = get_git_branch()
      git_cache.ahead, git_cache.behind = get_git_ahead_behind()
      git_cache.staged, git_cache.modified, git_cache.untracked = get_git_status_counts()
      git_cache.stashed = get_git_stash_count()
   end)
   git_cache.last_update = current_time
end

local function format_git_status()
   update_git_cache()

   if git_cache.branch == '' then
      return '', false
   end

   local parts = {}

   local branch_color = '%#StatusLineGitClean#'

   local branch_part = branch_color .. icons.git .. ' ' .. git_cache.branch .. '%*'

   local branch_is_dirty = (git_cache.modified > 0 or git_cache.untracked > 0)

   if git_cache.behind > 0 then
      table.insert(parts, '%#StatusLineGitClean#' .. icons.git_indicators.behind .. git_cache.behind .. '%*')
   end

   if git_cache.ahead > 0 then
      table.insert(parts, '%#StatusLineGitClean#' .. icons.git_indicators.ahead .. git_cache.ahead .. '%*')
   end

   if git_cache.stashed > 0 then
      table.insert(parts, '%#StatusLineGitClean#' .. icons.git_indicators.stashed .. git_cache.stashed .. '%*')
   end

   if git_cache.staged > 0 then
      table.insert(parts, '%#StatusLineGitDirty#' .. icons.git_indicators.staged .. git_cache.staged .. '%*')
   end

   if git_cache.modified > 0 then
      table.insert(parts, '%#StatusLineGitDirty#' .. icons.git_indicators.modified .. git_cache.modified .. '%*')
   end

   if git_cache.untracked > 0 then
      table.insert(parts, '%#StatusLineGitUntracked#' .. icons.git_indicators.untracked .. git_cache.untracked .. '%*')
   end

   local git_status = branch_part

   if #parts > 0 then
      git_status = git_status .. ' ' .. table.concat(parts, ' ')
   end

   return git_status, branch_is_dirty
end

local function get_file_path()
   local filepath = vim.fn.expand('%')
   if filepath == '' then
      return '[No Name]'
   end

   local cwd = vim.fn.getcwd()
   local full_path = vim.fn.expand('%:p')
   local home = vim.fn.expand('~')

   -- Safe path processing with boundary checks
   if #cwd >= #home and safe_sub(cwd, 1, #home) == home then
      cwd = '~' .. safe_sub(cwd, #home + 1)
   end
   if #full_path >= #home and safe_sub(full_path, 1, #home) == home then
      full_path = '~' .. safe_sub(full_path, #home + 1)
   end

   local display_path

   if #full_path >= #cwd and safe_sub(full_path, 1, #cwd) == cwd then
      display_path = safe_sub(full_path, #cwd + 2)
      if display_path == '' then
         display_path = vim.fn.expand('%:t')
      end
   else
      display_path = shorten_path(filepath)
   end

   local modified = vim.bo.modified and ' [+]' or ''
   local readonly = vim.bo.readonly and ' [RO]' or ''

   return display_path .. modified .. readonly
end

local function get_cursor_position()
   local line = vim.fn.line('.')
   local col = vim.fn.col('.')
   local total_lines = vim.fn.line('$')
   return string.format('%d:%d/%d', line, col, total_lines)
end

local function get_filetype()
   local ft = vim.bo.filetype
   return ft ~= '' and ft or 'no ft'
end

local function get_file_encoding()
   local encoding = vim.bo.fileencoding
   if encoding == '' then
      encoding = vim.o.encoding
   end
   return encoding
end

function M.statusline()
   -- Wrap statusline generation in pcall to prevent crashes
   local success, result = pcall(function()
      local git_status = format_git_status()
      local cursor_pos = get_cursor_position()
      local filetype = get_filetype()
      local encoding = get_file_encoding()
      local file_path = get_file_path()
      local os_icon = get_os_icon()

      local left_parts = {}

      -- Safe string concatenation with validation
      if os_icon and os_icon ~= '' then
         table.insert(left_parts, '%#StatusLineOS#' .. os_icon .. '%*')
      end

      if file_path and file_path ~= '' then
         table.insert(left_parts, '%#StatusLineDir#' .. icons.folder .. ' ' .. file_path .. '%*')
      end

      if git_status and git_status ~= '' then
         table.insert(left_parts, '%#StatusLineSeparator#on%* ' .. git_status)
      end

      local left = table.concat(left_parts, ' ')

      local right_parts = {}
      if filetype and filetype ~= '' then
         table.insert(right_parts, '%#StatusLineInfo#' .. filetype .. '%*')
      end
      table.insert(right_parts, '%#StatusLineSeparator#|%*')
      if encoding and encoding ~= '' then
         table.insert(right_parts, '%#StatusLineInfo#' .. encoding .. '%*')
      end
      table.insert(right_parts, '%#StatusLineSeparator#|%*')
      if cursor_pos and cursor_pos ~= '' then
         table.insert(right_parts, '%#StatusLineInfo#' .. cursor_pos .. '%*')
      end

      local right = table.concat(right_parts, ' ')
      return left .. '%=' .. right
   end)

   if success and result then
      return result
   else
      -- Fallback statusline on error
      return ' [Error in statusline] %=%l:%c/%L '
   end
end

function M.setup()
   setup_highlights()

   vim.o.statusline = '%!v:lua.require("DarO.statusline").statusline()'

   vim.api.nvim_create_autocmd({ 'BufEnter', 'BufWrite', 'FocusGained', 'VimResume' }, {
      group = vim.api.nvim_create_augroup('CustomStatuslineRefresh', { clear = true }),
      callback = function()
         pcall(function()
            git_cache.last_update = 0
            vim.cmd('redrawstatus')
         end)
      end
   })

   vim.api.nvim_create_autocmd('ColorScheme', {
      group = vim.api.nvim_create_augroup('StatuslineHighlights', { clear = true }),
      callback = function()
         pcall(setup_highlights)
      end
   })

   -- Safer timer with error handling
   local timer = vim.uv.new_timer()
   if timer then
      timer:start(5000, 5000, vim.schedule_wrap(function()
         pcall(function()
            if vim.api.nvim_get_mode().mode == 'n' then
               git_cache.last_update = 0
               vim.cmd('redrawstatus')
            end
         end)
      end))
   end
end

return M
