-- ~/.config/nvim/lua/plugins/avante.lua
return {
   "yetone/avante.nvim",
   event = "VeryLazy",
   enabled = false,
   version = false,
   build = "make",
   opts = {
      provider = "claude",
      claude = {
         endpoint = "https://api.anthropic.com",
         model = "claude-3-haiku-20240307", -- Cheapest Claude 3 model
         -- model = "claude-3-5-sonnet-20241022", -- Best Claude model
         timeout = 30000,
         temperature = 0,
         max_tokens = 4096,
         disable_tools = true, -- optional: disables built-in tools Claude sometimes overuses
      },
      behaviour = {
         enable_cursor_planning_mode = true,
      },
      file_selector = {
         provider = "telescope", -- or 'fzf' or 'mini.pick'
         telescope = {
            show_preview = true,
         },
      },

   },
   dependencies = {
      "nvim-treesitter/nvim-treesitter",
      "stevearc/dressing.nvim",
      "nvim-lua/plenary.nvim",
      "MunifTanjim/nui.nvim",
      "echasnovski/mini.pick",         -- for file_selector provider mini.pick
      "nvim-telescope/telescope.nvim", -- for file_selector provider telescope
      "hrsh7th/nvim-cmp",              -- autocompletion for avante commands and mentions
      "ibhagwan/fzf-lua",              -- for file_selector provider fzf
      "nvim-tree/nvim-web-devicons",   -- or echasnovski/mini.icons
      "zbirenbaum/copilot.lua",        -- for providers='copilot'
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
               use_absolute_path = true,
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
