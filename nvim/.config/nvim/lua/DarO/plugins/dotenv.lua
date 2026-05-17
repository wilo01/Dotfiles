local defer = require("DarO.defer")

defer.on_cmd({
   repo = "ellisonleao/dotenv.nvim",
   cmds = { "Dotenv", "DotenvGet" },
   setup = function()
      require("dotenv").setup({
         enable_on_load = true,
         verbose = false,
         file_name = vim.fn.expand("~/.env"),
      })
   end,
})
