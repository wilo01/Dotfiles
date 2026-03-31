local Snacks = require("snacks")

---@type snacks.Config
Snacks.setup({
   bigfile = { enabled = true },
   dashboard = {
      sections = {
         {
            section = "header",
            enabled = function()
               return not vim.g.hide_startup_info
            end,
         },
         { icon = " ", title = "Keymaps", section = "keys", indent = 3 },
         {
            icon = " ",
            desc = "Browse Repo",
            key = "b",
            action = function()
               Snacks.gitbrowse()
            end,
            indent = 3,
         },
         { icon = " ", title = "Recent Files", cwd = true, section = "recent_files", indent = 3, padding = 1, pane = 2 },
         {
            icon = " ",
            title = "Git Status",
            section = "terminal",
            enabled = function()
               return Snacks.git.get_root() ~= nil
            end,
            cmd = "git --no-pager diff --stat -B -M -C",
            ttl = 5 * 60,
            indent = 3,
            padding = 1,
            pane = 2
         },
      },
      styles = {
         snacks_image = {
            relative = "editor",
            col = -1,
         },
      },
      image = {
         enabled = true,
         force = true,
         doc = {
            inline = true,
            float = true,
            max_width = 60,
            max_height = 30,
         },
      },
   },
   notifier = {
      enabled = true,
      timeout = 3000,
   },
   quickfile = { enabled = true },
   words = { enabled = true },
   statuscolumn = { enabled = false },
   scope = { enabled = false },
   scroll = { enabled = false },
   input = { enabled = false },
   styles = {
      notification = {
         wo = { wrap = true }
      }
   },
   gh = {
      enabled = true,
      keys = {
         select  = { "<cr>", "gh_actions", desc = "Select Action" },
         edit    = { "i", "gh_edit", desc = "Edit" },
         comment = { "a", "gh_comment", desc = "Add Comment" },
         close   = { "c", "gh_close", desc = "Close" },
         reopen  = { "o", "gh_reopen", desc = "Reopen" },
      },
   },
   picker = {
      enabled = true,
      sources = {
         gh_issue = {},
         gh_pr = {},
      }
   }
})

-- Keymaps (was keys = function())
vim.keymap.set("n", "<leader>n",  function() Snacks.notifier.show_history() end,   { desc = "Notification History" })
vim.keymap.set("n", "<leader>un", function() Snacks.notifier.hide() end,           { desc = "Snacks Dismiss All Notifications" })
vim.keymap.set("n", "<leader>bd", function() Snacks.bufdelete() end,               { desc = "Snacks Delete Buffer" })
vim.keymap.set("n", "<leader>gg", function() Snacks.lazygit() end,                 { desc = "Snacks Lazygit" })
vim.keymap.set("n", "<leader>gb", function() Snacks.git.blame_line() end,          { desc = "Snacks Git Blame Line" })
vim.keymap.set("n", "<leader>ob", function() Snacks.gitbrowse() end,               { desc = "Snacks Git Open in Browser" })
vim.keymap.set("n", "<leader>gf", function() Snacks.lazygit.log_file() end,        { desc = "Snacks Lazygit Current File History" })
vim.keymap.set("n", "<leader>gl", function() Snacks.lazygit.log() end,             { desc = "Snacks Lazygit Log (cwd)" })
vim.keymap.set("n", "<leader>gi", function() Snacks.picker.gh_issue() end,                  { desc = "GitHub Issues (open)" })
vim.keymap.set("n", "<leader>gI", function() Snacks.picker.gh_issue({ state = "all" }) end, { desc = "GitHub Issues (all)" })
vim.keymap.set("n", "<leader>gp", function() Snacks.picker.gh_pr() end,                     { desc = "GitHub Pull Requests (open)" })
vim.keymap.set("n", "<leader>gP", function() Snacks.picker.gh_pr({ state = "all" }) end,    { desc = "GitHub Pull Requests (all)" })
vim.keymap.set("n", "<leader>cR", function() Snacks.rename.rename_file() end,      { desc = "Snacks Rename File" })
vim.keymap.set("n", "<leader>te", function() Snacks.terminal() end,                { desc = "Snacks Toggle Terminal" })
vim.keymap.set({ "n", "t" }, "<leader>]",  function() Snacks.words.jump(vim.v.count1) end,  { desc = "Snacks Next Reference" })
vim.keymap.set({ "n", "t" }, "<leader>[",  function() Snacks.words.jump(-vim.v.count1) end, { desc = "Snacks Prev Reference" })
vim.keymap.set("n", "<leader>N", function()
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
end, { desc = "Neovim News" })

-- Image hover for markdown
vim.api.nvim_create_autocmd("CursorHold", {
   pattern = "*.md",
   callback = function()
      Snacks.image.hover();
   end,
})

-- Toggles and debug helpers (was VeryLazy, now UIEnter)
vim.api.nvim_create_autocmd("UIEnter", {
   once = true,
   callback = function()
      _G.dd = function(...)
         Snacks.debug.inspect(...)
      end
      _G.bt = function()
         Snacks.debug.backtrace()
      end
      vim.print = _G.dd

      Snacks.toggle.option("spell", { name = "Spelling" }):map("<leader>us")
      Snacks.toggle.option("wrap", { name = "Wrap" }):map("<leader>uw")
      Snacks.toggle.option("relativenumber", { name = "Relative Number" }):map("<leader>uL")
      Snacks.toggle.diagnostics():map("<leader>ud")
      Snacks.toggle.line_number():map("<leader>ul")
      Snacks.toggle.option("conceallevel", { off = 0, on = vim.o.conceallevel > 0 and vim.o.conceallevel or 2 })
          :map("<leader>uc")
      Snacks.toggle.treesitter():map("<leader>uT")
      Snacks.toggle.option("background", { off = "light", on = "dark", name = "Dark Background" }):map("<leader>ub")
      Snacks.toggle.inlay_hints():map("<leader>uh")
   end,
})
