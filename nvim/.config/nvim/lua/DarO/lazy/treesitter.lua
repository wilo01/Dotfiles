return {
   {
      "nvim-treesitter/nvim-treesitter",
      event = "VeryLazy",
      build = ":TSUpdate",
      cmd = { "TSUpdateSync", "TSUpdate", "TSInstall" },
      config = function()
         local configs = require("nvim-treesitter.configs")

         configs.setup({
            ensure_installed = {
               "vimdoc", "javascript", "typescript", "c", "lua", "rust",
               "jsdoc", "bash", "markdown", "markdown_inline", "query",
               "vim", "html", "css", "json", "yaml", "python", "go",
               "templ", "sql", "xml", "toml", "tsx", "dockerfile", "vue",
               "svelte", "php", "java", "regex", "elixir", "heex", "eex"
            },
            auto_install = true,
            sync_install = false,
            highlight = { enable = true },
            indent = { enable = true },
            modules = {},
            ignore_install = {},
         })
      end,
   },
}
