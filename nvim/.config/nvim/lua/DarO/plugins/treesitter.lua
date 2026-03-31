-- nvim-treesitter now only handles parser installation.
-- Highlight and indent are native Neovim features (enabled by default).
local parsers = {
   "vimdoc", "javascript", "typescript", "c", "lua", "rust",
   "jsdoc", "bash", "markdown", "markdown_inline", "query",
   "vim", "html", "css", "json", "yaml", "python", "go",
   "templ", "toml", "tsx", "dockerfile", "vue",
   "svelte", "php", "java", "regex", "elixir", "heex", "eex",
   "latex", "scss", "typst"
}

-- Install missing parsers
local installed = require("nvim-treesitter").get_installed()
local installed_set = {}
for _, p in ipairs(installed) do
   installed_set[p] = true
end

local to_install = {}
for _, p in ipairs(parsers) do
   if not installed_set[p] then
      table.insert(to_install, p)
   end
end

if #to_install > 0 then
   require("nvim-treesitter").install(to_install)
end

vim.treesitter.language.register("bash", "dotenv")
