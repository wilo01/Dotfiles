return {
   "folke/tokyonight.nvim",
   config = function()
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
         end,
      })
      vim.cmd.colorscheme('tokyonight')
   end
}
