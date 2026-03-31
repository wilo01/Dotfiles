require("tokyonight").setup({
   style = "night",
   transparent = true,
   terminal_colors = true,
   styles = {
      comments = { italic = false },
      keywords = { italic = false },
      sidebars = "dark",
      floats = "dark",
   },
   on_colors = function(colors)
   end,
   on_highlights = function(hl, colors)
      hl.GitSignsCurrentLineBlame = { fg = "#7dcfff", italic = true }
      hl.GitSignsAddInline = { bg = "#35603a" }
      hl.GitSignsDeleteInline = { bg = "#603a35" }
      hl.GitSignsChangeInline = { bg = "#35603a" }
   end,
})
vim.cmd.colorscheme('tokyonight')
