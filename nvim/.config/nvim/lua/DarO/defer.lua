local M = {}

function M.on_cmd(spec)
   local loaded = false

   local function ensure()
      if loaded then return end
      for _, c in ipairs(spec.cmds) do
         pcall(vim.api.nvim_del_user_command, c)
      end
      local ok, err = pcall(function()
         vim.pack.add({ gh(spec.repo) })
         if spec.setup then spec.setup() end
      end)
      if not ok then
         vim.notify("Failed to load " .. spec.repo .. ": " .. tostring(err), vim.log.levels.ERROR)
         return
      end
      loaded = true
   end

   for _, cmd in ipairs(spec.cmds) do
      vim.api.nvim_create_user_command(cmd, function(info)
         pcall(vim.api.nvim_del_user_command, cmd)
         ensure()
         vim.cmd({ cmd = cmd, args = { info.args }, bang = info.bang })
      end, { nargs = "*", bang = true })
   end

   if spec.keymaps then
      spec.keymaps(ensure)
   end

   return ensure
end

return M
