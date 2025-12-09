return {
   "stevearc/conform.nvim",
   event = { "BufWritePre" },
   cmd = { "ConformInfo" },
   keys = {
      { "<leader>cf", function() require("conform").format({ async = true }) end, desc = "Format buffer" },
   },
   opts = {
      formatters_by_ft = {
         dotenv = { "dotenv_json" },
      },
      formatters = {
         dotenv_json = {
            command = vim.fn.stdpath("config") .. "/scripts/format-dotenv.sh",
            stdin = true,
         },
      },
   },
}
