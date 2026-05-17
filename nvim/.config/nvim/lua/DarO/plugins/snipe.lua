local defer = require("DarO.defer")

defer.on_cmd({
   repo = "leath-dub/snipe.nvim",
   setup = function()
      require("snipe").setup({
         hints = {
            dictionary = "asfghl;wertyuiop",
         },
         navigate = {
            cancel_snipe = "<esc>",
            close_buffer = "d",
         },
         sort = "default",
      })
   end,
   keymaps = function(ensure)
      vim.keymap.set("n", "<leader>sn", function()
         if not ensure() then return end
         require("snipe").open_buffer_menu()
      end, { desc = "Open Snipe buffer menu" })
   end,
})
