local gh = require("DarO.utils").gh
local M = {}

function M.on_cmd(spec)
   local loaded = false

   -- Retries on failure (loaded stays false) — useful for transient network errors
   local function ensure()
      if loaded then return true end
      for _, c in ipairs(spec.cmds or {}) do
         pcall(vim.api.nvim_del_user_command, c)
      end
      local ok, err = pcall(vim.pack.add, { gh(spec.repo) })
      if not ok then
         vim.notify("Failed to download " .. spec.repo .. ": " .. tostring(err), vim.log.levels.ERROR)
         return false
      end
      if spec.setup then
         local sok, serr = pcall(spec.setup)
         if not sok then
            vim.notify("Failed to configure " .. spec.repo .. ": " .. tostring(serr), vim.log.levels.ERROR)
            return false
         end
      end
      loaded = true
      return true
   end

   for _, cmd in ipairs(spec.cmds or {}) do
      vim.api.nvim_create_user_command(cmd, function(info)
         pcall(vim.api.nvim_del_user_command, cmd)
         if ensure() then
            vim.cmd({ cmd = cmd, args = { info.args }, bang = info.bang })
         end
      end, { nargs = "*", bang = true })
   end

   if spec.keymaps then
      spec.keymaps(ensure)
   end

   return ensure
end

return M
