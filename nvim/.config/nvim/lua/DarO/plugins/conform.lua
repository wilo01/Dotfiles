require("conform").setup({
   formatters_by_ft = {
      dotenv = { "dotenv_json" },
      json = { "jq" },
      jsonc = { "jq" },
   },
   formatters = {
      dotenv_json = {
         command = vim.fn.stdpath("config") .. "/scripts/format-dotenv.sh",
         stdin = true,
      },
   },
   format_on_save = function(bufnr)
      local ft = vim.bo[bufnr].filetype
      if ft == "dotenv" or ft == "json" or ft == "jsonc" then
         return { timeout_ms = 5000, lsp_fallback = false }
      end
      return false
   end,
})
