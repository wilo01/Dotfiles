return {
   "stevearc/conform.nvim",
   event = { "BufReadPre" },
   cmd = { "ConformInfo" },
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
      -- Use conform as fallback when LSP doesn't have a formatter
      format_on_save = function(bufnr)
         -- Only auto-format dotenv files
         if vim.bo[bufnr].filetype == "dotenv" then
            return { timeout_ms = 5000, lsp_fallback = false }
         end
         return false
      end,
   },
}
