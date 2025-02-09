return {
   "ThePrimeagen/harpoon",
   branch = "harpoon2",
   dependencies = { "nvim-lua/plenary.nvim" },
   config = function()
      local harpoon = require("harpoon")
      local harpoon_auto_add_enabled = false
      local list = harpoon:list()
      harpoon:setup()

      local conf = require("telescope.config").values
      local function toggle_telescope(harpoon_files)
         local file_paths = {}
         for _, item in ipairs(harpoon_files.items) do
            table.insert(file_paths, item.value)
         end

         require("telescope.pickers").new({}, {
            prompt_title = "Harpoon",
            finder = require("telescope.finders").new_table({
               results = file_paths,
            }),
            previewer = conf.file_previewer({}),
            sorter = conf.generic_sorter({}),
         }):find()
      end

      vim.keymap.set("n", "<C-t>", function()
         toggle_telescope(harpoon:list())
      end, { desc = "Open harpoon window with telescope" })

      vim.keymap.set("n", "<C-e>", function()
         harpoon.ui:toggle_quick_menu(list)
      end, { desc = "Toggle Harpoon quick menu" })

      vim.keymap.set("n", "<leader>a", function()
         list:add()
         add_file_and_notify()
      end, { desc = "Add file to Harpoon list and notify" })

      for i = 1, 10 do
         vim.keymap.set("n", "<leader>" .. (i % 10), function()
            list:select(i)
         end, { desc = "Select " .. i .. " item in Harpoon list" })
      end

      vim.keymap.set("n", "<C-h>", function()
         list:prev()
      end, { desc = "Go to previous item in Harpoon list" })

      vim.keymap.set("n", "<C-l>", function()
         list:next()
      end, { desc = "Go to next item in Harpoon list" })

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
