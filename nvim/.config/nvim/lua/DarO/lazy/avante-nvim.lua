-- ~/.config/nvim/lua/plugins/avante.lua
return {
   "yetone/avante.nvim",
   event = "VeryLazy",
   version = false,
   build = "make",
   opts = {
      provider = "claude",
      claude = {
         endpoint = "https://api.anthropic.com",
         model = "claude-3-5-sonnet-20241022",
         timeout = 30000,
         temperature = 0,
         max_tokens = 4096,
         -- disable_tools = true, -- optional: disables built-in tools Claude sometimes overuses
      },
   },
   dependencies = {
      "nvim-treesitter/nvim-treesitter",
      "stevearc/dressing.nvim",
      "nvim-lua/plenary.nvim",
      "MunifTanjim/nui.nvim",
      "echasnovski/mini.pick",       -- or "nvim-telescope/telescope.nvim" for file selection
      "nvim-tree/nvim-web-devicons", -- or "echasnovski/mini.icons"
      "hrsh7th/nvim-cmp",
      -- "zbirenbaum/copilot.lua",      -- optional: Copilot integration
      {
         "HakonHarnes/img-clip.nvim",
         event = "VeryLazy",
         opts = {
            default = {
               embed_image_as_base64 = false,
               prompt_for_file_name = false,
               drag_and_drop = {
                  insert_mode = true,
               },
            },
         },
      },
      {
         "MeanderingProgrammer/render-markdown.nvim",
         opts = {
            file_types = { "markdown", "Avante" },
         },
         ft = { "markdown", "Avante" },
      },
   },
}
