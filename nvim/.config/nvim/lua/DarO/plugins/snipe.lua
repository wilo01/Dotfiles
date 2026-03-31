local gh = function(x) return 'https://github.com/' .. x end
local loaded = false

vim.keymap.set("n", "<leader>sn", function()
   if not loaded then
      loaded = true
      vim.pack.add({ gh('leath-dub/snipe.nvim') })
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
   end
   require("snipe").open_buffer_menu()
end, { desc = "Open Snipe buffer menu" })
