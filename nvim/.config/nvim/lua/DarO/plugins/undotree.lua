local defer = require("DarO.defer")

defer.on_cmd({
   repo = "mbbill/undotree",
   cmds = { "UndotreeToggle", "UndotreeShow", "UndotreeHide", "UndotreeFocus" },
   keymaps = function(ensure)
      vim.keymap.set("n", "<leader>u", function()
         if not ensure() then return end
         vim.cmd.UndotreeToggle()
      end, { desc = "Toggle UndoTree" })
   end,
})
