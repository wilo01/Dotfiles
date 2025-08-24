-- Treesitter configuration
-- Optimized for performance with selective language loading

require("nvim-treesitter.configs").setup({
   -- Only install parsers for languages you actually use
   ensure_installed = {
      "lua",
      "vim",
      "vimdoc",
      "query",
      "regex",
      "bash",
      "markdown",
      "markdown_inline",
      "json",
      "jsonc",
      "yaml",
      "toml",
      "html",
      "css",
      "javascript",
      "typescript",
      "tsx",
      "python",
      "rust",
      "go",
      "c",
      "cpp",
      "diff",
      "git_config",
      "git_rebase",
      "gitattributes",
      "gitcommit",
      "gitignore",
   },

   -- Install parsers synchronously (only applied to `ensure_installed`)
   sync_install = false,

   -- Automatically install missing parsers when entering buffer
   auto_install = true,

   -- Ignore install for these parsers
   ignore_install = {},

   highlight = {
      enable = true,
      -- Disable for large files
      disable = function(lang, buf)
         local max_filesize = 100 * 1024 -- 100 KB
         local ok, stats = pcall(vim.loop.fs_stat, vim.api.nvim_buf_get_name(buf))
         if ok and stats and stats.size > max_filesize then
            return true
         end
      end,
      -- Setting this to true will run `:h syntax` and tree-sitter at the same time.
      additional_vim_regex_highlighting = false,
   },

   incremental_selection = {
      enable = true,
      keymaps = {
         init_selection = "<C-space>",
         node_incremental = "<C-space>",
         scope_incremental = false,
         node_decremental = "<bs>",
      },
   },

   indent = {
      enable = true,
      disable = { "yaml" }, -- YAML indentation can be problematic
   },

   -- Use treesitter for folding
   fold = {
      enable = true,
   },
})

-- Configure folding to use treesitter
vim.opt.foldmethod = "expr"
vim.opt.foldexpr = "nvim_treesitter#foldexpr()"
vim.opt.foldenable = false -- Don't fold by default