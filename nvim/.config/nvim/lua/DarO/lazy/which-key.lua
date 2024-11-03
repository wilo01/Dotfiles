return {
   "folke/which-key.nvim",
   event = "VeryLazy",
   opts = {
      defaults = {
         ["z="] = { "Spell suggestions" },
         ["zg"] = { "Add word to spellfile as good word" },
         ["zG"] = { "Add word to internal word list as good word" },
         ["zw"] = { "Mark word as wrong (bad) in spellfile" },
         ["zW"] = { "Mark word as wrong (bad) in internal word list" },
         ["zug"] = { "Remove word from spellfile as good/bad" },
         ["zuG"] = { "Remove word from internal word list as good/bad" },
         ["]s"] = { "Next misspelled word" },
         ["[s"] = { "Previous misspelled word" },
         ["]S"] = { "Next bad word only" },
         ["[S"] = { "Previous bad word only" },
         ["]r"] = { "Next rare word" },
         ["[r"] = { "Previous rare word" },
      },
   },
   keys = {
      {
         "<leader>?",
         function()
            require("which-key").show({ global = false })
         end,
         desc = "Buffer Local Keymaps (which-key)",
      },
   },
}
