return {
   "folke/snacks.nvim",
   priority = 1000,
   lazy = false,
   ---@type snacks.Config
   opts = {
      bigfile = { enabled = true },
      dashboard = {
         sections = {
            { section = "header" },
            { section = "startup", padding = 1 },
            { icon = " ", title = "Keymaps", section = "keys", indent = 3 },
            {
               icon = " ",
               desc = "Browse Repo",
               key = "b",
               action = function()
                  Snacks.gitbrowse()
               end,
               indent = 3,
            },
            { icon = " ", title = "Recent Files", section = "recent_files", indent = 3, padding = 1, pane = 2 },
            {
               icon = " ",
               title = "Git Status",
               section = "terminal",
               enabled = function()
                  return Snacks.git.get_root() ~= nil
               end,
               cmd = "hub --no-pager diff --stat -B -M -C",
               ttl = 5 * 60,
               indent = 3,
               padding = 1,
               pane = 2
            },
         },
      },
      notifier = {
         enabled = true,
         timeout = 3000,
      },
      quickfile = { enabled = true },
      statuscolumn = { enabled = true },
      words = { enabled = true },
      scope = { enabled = true },
      scroll = { enabled = false },
      input = { enabled = false },
      styles = {
         notification = {
            wo = { wrap = true }
         }
      }
   },
   keys = function()
      local Snacks = require("snacks")
      return {
         { "<leader>un", function() Snacks.notifier.hide() end,           desc = "Snacks Dismiss All Notifications" },
         { "<leader>bd", function() Snacks.bufdelete() end,               desc = "Snacks Delete Buffer" },
         { "<leader>gg", function() Snacks.lazygit() end,                 desc = "Snacks Lazygit" },
         { "<leader>gb", function() Snacks.git.blame_line() end,          desc = "Snacks Git Blame Line" },
         { "<leader>gr", function() Snacks.gitbrowse() end,               desc = "Snacks Git Browse" },
         { "<leader>gf", function() Snacks.lazygit.log_file() end,        desc = "Snacks Lazygit Current File History" },
         { "<leader>gl", function() Snacks.lazygit.log() end,             desc = "Snacks Lazygit Log (cwd)" },
         { "<leader>cR", function() Snacks.rename.rename_file() end,      desc = "Snacks Rename File" },
         { "<leader>t",  function() Snacks.terminal() end,                desc = "Snacks Toggle Terminal" },
         { "<leader>T",  function() Snacks.terminal() end,                desc = "Snacks which_key_ignore" },
         { "<leader>]",  function() Snacks.words.jump(vim.v.count1) end,  desc = "Snacks Next Reference",              mode = { "n", "t" } },
         { "<leader>[",  function() Snacks.words.jump(-vim.v.count1) end, desc = "Snacks Prev Reference",              mode = { "n", "t" } },
         {
            "<leader>N",
            desc = "Neovim News",
            function()
               Snacks.win({
                  file = vim.api.nvim_get_runtime_file("doc/news.txt", false)[1],
                  width = 0.6,
                  height = 0.6,
                  wo = {
                     spell = false,
                     wrap = false,
                     signcolumn = "yes",
                     statuscolumn = " ",
                     conceallevel = 3,
                  },
               })
            end,
         },
      }
   end,
   init = function()
      local Snacks = require("snacks") -- Require Snacks locally
      vim.api.nvim_create_autocmd("User", {
         pattern = "VeryLazy",
         callback = function()
            -- Setup some globals for debugging (lazy-loaded)
            _G.dd = function(...)
               Snacks.debug.inspect(...)
            end
            _G.bt = function()
               Snacks.debug.backtrace()
            end
            vim.print = _G.dd -- Override print to use snacks for `:=` command

            -- Create some toggle mappings using key binds
            Snacks.toggle.option("spell", { name = "Spelling" }):map("<leader>us")
            Snacks.toggle.option("wrap", { name = "Wrap" }):map("<leader>uw")
            Snacks.toggle.option("relativenumber", { name = "Relative Number" }):map("<leader>uL")
            Snacks.toggle.diagnostics():map("<leader>ud")
            Snacks.toggle.line_number():map("<leader>ul")
            Snacks.toggle.option("conceallevel", { off = 0, on = vim.o.conceallevel > 0 and vim.o.conceallevel or 2 })
                :map("<leader>uc")
            Snacks.toggle.treesitter():map("<leader>uT")
            Snacks.toggle.option("background", { off = "light", on = "dark", name = "Dark Background" }):map(
               "<leader>ub")
            Snacks.toggle.inlay_hints():map("<leader>uh")
         end,
      })
   end,
}
