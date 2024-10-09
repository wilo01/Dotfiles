return {
   "ThePrimeagen/harpoon",
   branch = "harpoon2",
   dependencies = { "nvim-lua/plenary.nvim" },
   config = function()
      local harpoon = require("harpoon")
      local harpoon_auto_add_enabled = false
      local list = harpoon:list()
      harpoon:setup()

      vim.keymap.set("n", "<C-e>", function() harpoon.ui:toggle_quick_menu(list) end)
      vim.keymap.set("n", "<leader>a", function()
         list:add()
         add_file_and_notify()
      end)

      vim.keymap.set("n", "<C-1>", function() list:select(1) end)
      vim.keymap.set("n", "<C-2>", function() list:select(2) end)
      vim.keymap.set("n", "<C-3>", function() list:select(3) end)
      vim.keymap.set("n", "<C-4>", function() list:select(4) end)

      vim.keymap.set("n", "<C-h>", function() list:prev() end)
      vim.keymap.set("n", "<C-l>", function() list:next() end)

      function _G.add_file_and_notify()
         local file_name = vim.fn.expand('%:t')
         list:add()
         vim.notify("File: " .. file_name .. " added to Harpoon", vim.log.levels.INFO)
      end

      -- Function to toggle the autocmd
      function _G.toggle_harpoon_auto_add()
         if harpoon_auto_add_enabled then
            vim.cmd [[
              augroup HarpoonAutoAdd
                autocmd!
              augroup END
            ]]
            vim.notify("Harpoon auto-add disabled", vim.log.levels.INFO)
         else
            vim.cmd [[
              augroup HarpoonAutoAdd
                autocmd!
                autocmd BufWritePost * lua add_file_and_notify()
              augroup END
            ]]
            vim.notify("Harpoon auto-add enabled", vim.log.levels.INFO)
         end
         harpoon_auto_add_enabled = not harpoon_auto_add_enabled
      end

      -- Command to toggle the autocmd
      vim.api.nvim_create_user_command("HarpoonAutoAdd", toggle_harpoon_auto_add, {})
   end
}
