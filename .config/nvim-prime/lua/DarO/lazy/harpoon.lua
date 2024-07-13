return {
   "ThePrimeagen/harpoon",
   branch = "harpoon2",
   config = function()
      local mark = require("harpoon.mark")
      local ui = require("harpoon.ui")

      vim.keymap.set("n", "<leader>a", mark.add_file)
      vim.keymap.set("n", "<C-e>", ui.toggle_quick_menu)

      vim.keymap.set("n", "<C-h>", function() ui.nav_prev() end)
      vim.keymap.set("n", "<C-l>", function() ui.nav_next() end)

      function _G.add_file_and_notify()
         local file_name = vim.fn.expand('%:t')
         mark.add_file()
         vim.notify("File: " .. file_name .. " added to Harpoon", vim.log.levels.INFO)
      end

      vim.cmd [[
        augroup HarpoonAutoAdd
          autocmd!
          autocmd BufWritePost * lua add_file_and_notify()
        augroup END
      ]]
   end
}
