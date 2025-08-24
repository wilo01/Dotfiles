-- Treesitter text objects configuration
-- Provides smart text objects based on syntax tree

require("nvim-treesitter.configs").setup({
   textobjects = {
      select = {
         enable = true,
         lookahead = true, -- Automatically jump forward to textobj
         keymaps = {
            -- You can use the capture groups defined in textobjects.scm
            ["af"] = "@function.outer",
            ["if"] = "@function.inner",
            ["ac"] = "@class.outer",
            ["ic"] = "@class.inner",
            ["al"] = "@loop.outer",
            ["il"] = "@loop.inner",
            ["ab"] = "@block.outer",
            ["ib"] = "@block.inner",
            ["ap"] = "@parameter.outer",
            ["ip"] = "@parameter.inner",
            ["as"] = "@statement.outer",
            ["is"] = "@statement.inner",
            ["am"] = "@comment.outer",
            ["im"] = "@comment.inner",
         },
         selection_modes = {
            ["@parameter.outer"] = "v", -- charwise
            ["@function.outer"] = "V", -- linewise
            ["@class.outer"] = "<c-v>", -- blockwise
         },
      },
      move = {
         enable = true,
         set_jumps = true, -- whether to set jumps in the jumplist
         goto_next_start = {
            ["]m"] = "@function.outer",
            ["]]"] = "@class.outer",
            ["]l"] = "@loop.outer",
            ["]s"] = "@statement.outer",
            ["]p"] = "@parameter.inner",
            ["]b"] = "@block.outer",
         },
         goto_next_end = {
            ["]M"] = "@function.outer",
            ["]["] = "@class.outer",
            ["]L"] = "@loop.outer",
            ["]S"] = "@statement.outer",
            ["]P"] = "@parameter.outer",
            ["]B"] = "@block.outer",
         },
         goto_previous_start = {
            ["[m"] = "@function.outer",
            ["[["] = "@class.outer",
            ["[l"] = "@loop.outer",
            ["[s"] = "@statement.outer",
            ["[p"] = "@parameter.inner",
            ["[b"] = "@block.outer",
         },
         goto_previous_end = {
            ["[M"] = "@function.outer",
            ["[]"] = "@class.outer",
            ["[L"] = "@loop.outer",
            ["[S"] = "@statement.outer",
            ["[P"] = "@parameter.outer",
            ["[B"] = "@block.outer",
         },
      },
      swap = {
         enable = true,
         swap_next = {
            ["<leader>sp"] = "@parameter.inner",
            ["<leader>sf"] = "@function.outer",
         },
         swap_previous = {
            ["<leader>sP"] = "@parameter.inner",
            ["<leader>sF"] = "@function.outer",
         },
      },
      lsp_interop = {
         enable = true,
         border = "rounded",
         peek_definition_code = {
            ["<leader>df"] = "@function.outer",
            ["<leader>dF"] = "@class.outer",
         },
      },
   },
})