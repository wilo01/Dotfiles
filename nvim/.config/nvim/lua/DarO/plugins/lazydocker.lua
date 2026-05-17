local defer = require("DarO.defer")

local ensure = defer.on_cmd({
   repo = "mgierada/lazydocker.nvim",
   cmds = { "Lazydocker", "LazyDocker" },
   setup = function()
      require("toggleterm").setup({})
      require("lazydocker").setup({
         border = "curved",
      })
   end,
})

vim.keymap.set("n", "<leader>ld", function()
   if not ensure() then return end
   require("lazydocker").open()
end, { desc = "Open Lazydocker floating window" })
