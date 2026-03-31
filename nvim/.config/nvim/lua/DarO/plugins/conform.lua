require("conform").setup({
   formatters_by_ft = {
      dotenv = { "dotenv_json" },
   },
   formatters = {
      dotenv_json = {
         command = vim.fn.stdpath("config") .. "/scripts/format-dotenv.sh",
         stdin = true,
      },
   },
   format_on_save = function(bufnr)
      if vim.bo[bufnr].filetype == "dotenv" then
         return { timeout_ms = 5000, lsp_fallback = false }
      end
      return false
   end,
})
