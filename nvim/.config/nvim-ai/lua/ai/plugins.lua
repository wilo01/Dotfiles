-- Optimized Plugin Management with Lazy.nvim
-- Aggressive lazy-loading for maximum performance

-- Bootstrap lazy.nvim
local lazypath = vim.fn.stdpath("data") .. "/lazy/lazy.nvim"
if not vim.loop.fs_stat(lazypath) then
   vim.fn.system({
      "git",
      "clone",
      "--filter=blob:none",
      "https://github.com/folke/lazy.nvim.git",
      "--branch=stable",
      lazypath,
   })
end
vim.opt.rtp:prepend(lazypath)

-- Plugin specifications
local plugins = {
   -- Color scheme (loaded immediately for visual consistency)
   {
      "folke/tokyonight.nvim",
      priority = 1000,
      lazy = false,
      config = function()
         require("tokyonight").setup({
            style = "night",
            transparent = true,
            terminal_colors = true,
            styles = {
               comments = { italic = false },
               keywords = { italic = false },
               sidebars = "dark",
               floats = "dark",
            },
         })
         vim.cmd.colorscheme("tokyonight")
      end,
   },

   -- Snacks.nvim (dashboard, bigfile, etc.)
   {
      "folke/snacks.nvim",
      priority = 1000,
      lazy = false,
      opts = {
         bigfile = { enabled = true },
         dashboard = {
            enabled = true,
            sections = {
               { section = "header" },
               { section = "keys", gap = 1, padding = 1 },
               { section = "startup" },
            },
            preset = {
               keys = {
                  { icon = " ", key = "f", desc = "Find File", action = ":lua Snacks.dashboard.pick('files')" },
                  { icon = " ", key = "n", desc = "New File", action = ":ene | startinsert" },
                  { icon = " ", key = "g", desc = "Find Text", action = ":lua Snacks.dashboard.pick('live_grep')" },
                  { icon = " ", key = "r", desc = "Recent Files", action = ":lua Snacks.dashboard.pick('oldfiles')" },
                  { icon = " ", key = "c", desc = "Config", action = ":lua Snacks.dashboard.pick('files', {cwd = vim.fn.stdpath('config')})" },
                  { icon = " ", key = "s", desc = "Restore Session", section = "session" },
                  { icon = "󰒲 ", key = "L", desc = "Lazy", action = ":Lazy", enabled = package.loaded.lazy ~= nil },
                  { icon = " ", key = "q", desc = "Quit", action = ":qa" },
               },
            },
         },
         notifier = { enabled = true },
         quickfile = { enabled = true },
         statuscolumn = { enabled = true },
         words = { enabled = true },
      },
      config = function(_, opts)
         local ok, Snacks = pcall(require, "snacks")
         if not ok then return end
         
         Snacks.setup(opts)
         
         -- Set up keymaps only if Snacks loaded
         local map = vim.keymap.set
         map("n", "<leader>z", function() Snacks.zen() end, { desc = "Toggle Zen Mode" })
         map("n", "<leader>Z", function() Snacks.zen.zoom() end, { desc = "Toggle Zoom" })
         map("n", "<leader>.", function() Snacks.scratch() end, { desc = "Toggle Scratch Buffer" })
         map("n", "<leader>S", function() Snacks.scratch.select() end, { desc = "Select Scratch Buffer" })
         map("n", "<leader>n", function() Snacks.notifier.show_history() end, { desc = "Notification History" })
         map("n", "<leader>bd", function() Snacks.bufdelete() end, { desc = "Delete Buffer" })
         map("n", "<leader>cR", function() Snacks.rename() end, { desc = "Rename File" })
         map("n", "<leader>gB", function() Snacks.gitbrowse() end, { desc = "Git Browse" })
         map("n", "<leader>gb", function() Snacks.git.blame_line() end, { desc = "Git Blame Line" })
         map("n", "<leader>gf", function() Snacks.lazygit.log_file() end, { desc = "Lazygit Current File History" })
         map("n", "<leader>gg", function() Snacks.lazygit() end, { desc = "Lazygit" })
         map("n", "<leader>gl", function() Snacks.lazygit.log() end, { desc = "Lazygit Log (cwd)" })
         map("n", "<leader>un", function() Snacks.notifier.hide() end, { desc = "Dismiss All Notifications" })
         map({ "n", "t" }, "]]", function() Snacks.words.jump(vim.v.count1) end, { desc = "Next Reference" })
         map({ "n", "t" }, "[[", function() Snacks.words.jump(-vim.v.count1) end, { desc = "Prev Reference" })
      end,
   },

   -- Treesitter (load immediately for proper syntax highlighting)
   {
      "nvim-treesitter/nvim-treesitter",
      build = ":TSUpdate",
      lazy = false,  -- Load immediately to ensure syntax highlighting works
      priority = 900,  -- High priority to load early
      config = function()
         require("ai.config.treesitter")
         -- Ensure highlighting is enabled globally
         vim.api.nvim_create_autocmd("FileType", {
            pattern = "*",
            callback = function()
               pcall(vim.cmd, "TSBufEnable highlight")
            end,
         })
      end,
      dependencies = {
         -- Treesitter text objects
         {
            "nvim-treesitter/nvim-treesitter-textobjects",
            config = function()
               require("ai.config.treesitter-textobjects")
            end,
         },
      },
   },

   -- LSP Configuration (load immediately for proper language server support)
   {
      "neovim/nvim-lspconfig",
      lazy = false,  -- Load immediately to ensure LSP is available
      priority = 850,  -- Load after treesitter
      dependencies = {
         -- Mason for LSP server management
         {
            "williamboman/mason.nvim",
            cmd = { "Mason", "MasonInstall", "MasonUninstall", "MasonUninstallAll", "MasonLog" },
            build = ":MasonUpdate",
            config = function()
               require("mason").setup({
                  ui = {
                     border = "rounded",
                     icons = {
                        package_installed = "✓",
                        package_pending = "➜",
                        package_uninstalled = "✗",
                     },
                  },
               })
            end,
         },
         {
            "williamboman/mason-lspconfig.nvim",
            config = function()
               require("mason-lspconfig").setup({
                  ensure_installed = { "lua_ls", "pyright", "ts_ls", "rust_analyzer" },
                  automatic_installation = true,
               })
            end,
         },
         -- Better Lua development
         {
            "folke/lazydev.nvim",
            ft = "lua",
            opts = {
               library = {
                  { path = "luvit-meta/library", words = { "vim%.uv" } },
               },
            },
         },
      },
      config = function()
         require("ai.config.lsp")
      end,
   },

   -- Completion (lazy-loaded on insert mode)
   {
      "hrsh7th/nvim-cmp",
      event = { "InsertEnter", "CmdlineEnter" },
      dependencies = {
         "hrsh7th/cmp-nvim-lsp",
         "hrsh7th/cmp-buffer",
         "hrsh7th/cmp-path",
         "hrsh7th/cmp-cmdline",
         {
            "L3MON4D3/LuaSnip",
            version = "v2.*",
            build = "make install_jsregexp",
            dependencies = {
               "saadparwaiz1/cmp_luasnip",
               "rafamadriz/friendly-snippets",
            },
            config = function()
               require("luasnip.loaders.from_vscode").lazy_load()
            end,
         },
      },
      config = function()
         require("ai.config.cmp")
      end,
   },

   -- Telescope (lazy-loaded on command)
   {
      "nvim-telescope/telescope.nvim",
      cmd = "Telescope",
      keys = {
         { "<leader>ff", "<cmd>Telescope find_files<CR>", desc = "Find files" },
         { "<leader>fg", "<cmd>Telescope live_grep<CR>", desc = "Live grep" },
         { "<leader>fb", "<cmd>Telescope buffers<CR>", desc = "Buffers" },
         { "<leader>fh", "<cmd>Telescope help_tags<CR>", desc = "Help tags" },
         { "<leader>fr", "<cmd>Telescope oldfiles<CR>", desc = "Recent files" },
         { "<leader>fc", "<cmd>Telescope commands<CR>", desc = "Commands" },
         { "<leader>fk", "<cmd>Telescope keymaps<CR>", desc = "Keymaps" },
         { "<leader>fs", "<cmd>Telescope git_status<CR>", desc = "Git status" },
         { "<leader>gc", "<cmd>Telescope git_commits<CR>", desc = "Git commits" },
         { "<leader>gb", "<cmd>Telescope git_branches<CR>", desc = "Git branches" },
         { "<C-p>", "<cmd>Telescope git_files<CR>", desc = "Git files" },
         { "<leader>sd", "<cmd>Telescope diagnostics<CR>", desc = "Diagnostics" },
      },
      dependencies = {
         "nvim-lua/plenary.nvim",
         {
            "nvim-telescope/telescope-fzf-native.nvim",
            build = "make",
            cond = function()
               return vim.fn.executable("make") == 1
            end,
         },
      },
      config = function()
         require("ai.config.telescope")
      end,
   },

   -- Git integration (enhanced configuration)
   {
      "lewis6991/gitsigns.nvim",
      event = { "BufReadPost", "BufNewFile" },
      config = function()
         local gitsigns = require("gitsigns")
         
         gitsigns.setup({
            signs = {
               add = { text = "│" },
               change = { text = "│" },
               delete = { text = "_" },
               topdelete = { text = "‾" },
               changedelete = { text = "~" },
               untracked = { text = "┆" },
            },
            signcolumn = true, -- Toggle with `:Gitsigns toggle_signs`
            numhl = false,     -- Toggle with `:Gitsigns toggle_numhl`
            linehl = false,    -- Toggle with `:Gitsigns toggle_linehl`
            word_diff = false, -- Toggle with `:Gitsigns toggle_word_diff`
            watch_gitdir = {
               interval = 1000,
               follow_files = true,
            },
            attach_to_untracked = true,
            current_line_blame = true, -- Toggle with `:Gitsigns toggle_current_line_blame`
            current_line_blame_opts = {
               virt_text = true,
               virt_text_pos = "eol", -- 'eol' | 'overlay' | 'right_align'
               delay = 100,
               ignore_whitespace = false,
            },
            current_line_blame_formatter = "<author>, <author_time:%Y-%m-%d> - <summary>",
            sign_priority = 6,
            update_debounce = 100,
            status_formatter = nil,  -- Use default
            max_file_length = 40000, -- Disable if file is longer than this (in lines)
            preview_config = {
               -- Options passed to nvim_open_win
               border = "single",
               style = "minimal",
               relative = "cursor",
               row = 0,
               col = 1,
            },
            on_attach = function(bufnr)
               local gs = package.loaded.gitsigns
               
               local function map(mode, l, r, opts)
                  opts = opts or {}
                  opts.buffer = bufnr
                  vim.keymap.set(mode, l, r, opts)
               end
               
               -- Navigation
               map("n", "]c", function()
                  if vim.wo.diff then return "]c" end
                  vim.schedule(function() gs.next_hunk() end)
                  return "<Ignore>"
               end, { expr = true, desc = "Next hunk" })
               
               map("n", "[c", function()
                  if vim.wo.diff then return "[c" end
                  vim.schedule(function() gs.prev_hunk() end)
                  return "<Ignore>"
               end, { expr = true, desc = "Previous hunk" })
               
               -- Actions
               map("n", "<leader>hs", gs.stage_hunk, { desc = "Stage hunk" })
               map("n", "<leader>hr", gs.reset_hunk, { desc = "Reset hunk" })
               map("v", "<leader>hs", function() gs.stage_hunk({ vim.fn.line("."), vim.fn.line("v") }) end, { desc = "Stage hunk" })
               map("v", "<leader>hr", function() gs.reset_hunk({ vim.fn.line("."), vim.fn.line("v") }) end, { desc = "Reset hunk" })
               map("n", "<leader>hS", gs.stage_buffer, { desc = "Stage buffer" })
               map("n", "<leader>hu", gs.undo_stage_hunk, { desc = "Undo stage hunk" })
               map("n", "<leader>hR", gs.reset_buffer, { desc = "Reset buffer" })
               map("n", "<leader>hp", gs.preview_hunk, { desc = "Preview hunk" })
               map("n", "<leader>hb", function() gs.blame_line({ full = true }) end, { desc = "Blame line" })
               map("n", "<leader>tb", gs.toggle_current_line_blame, { desc = "Toggle blame" })
               map("n", "<leader>hd", gs.diffthis, { desc = "Diff this" })
               map("n", "<leader>hD", function() gs.diffthis("~") end, { desc = "Diff this ~" })
               map("n", "<leader>td", gs.toggle_deleted, { desc = "Toggle deleted" })
            end,
         })
         
         -- Autocmd for removing trailing whitespaces on changed lines
         local augroup = vim.api.nvim_create_augroup
         local TheDaroGroup = augroup('TheDarO', {})
         local autocmd = vim.api.nvim_create_autocmd
         
         autocmd({ "BufWritePre" }, {
            group = TheDaroGroup,
            pattern = "*",
            callback = function()
               local bufnr = vim.api.nvim_get_current_buf()
               local hunk_lines = gitsigns.get_hunks(bufnr)
               
               if hunk_lines and #hunk_lines > 0 then
                  for _, hunk in ipairs(hunk_lines) do
                     local start_line = hunk.added and hunk.added.start or nil
                     local count = hunk.added and hunk.added.count or 0
                     if start_line and count > 0 then
                        local end_line = start_line + count - 1
                        if start_line <= end_line then
                           vim.api.nvim_buf_call(bufnr, function()
                              vim.cmd(string.format("%d,%ds/\\s\\+$//e", start_line, end_line))
                           end)
                        end
                     end
                  end
               end
            end,
         })
      end,
   },

   {
      "tpope/vim-fugitive",
      cmd = { "Git", "G", "Gstatus", "Gblame", "Gpush", "Gpull", "Gcommit", "Glog", "Gdiff", "Gwrite" },
      keys = {
         { "<leader>gs", "<cmd>Git<CR>", desc = "Git status" },
         { "<leader>gd", "<cmd>Gdiff<CR>", desc = "Git diff" },
         { "<leader>gl", "<cmd>Glog<CR>", desc = "Git log" },
         { "<leader>gp", "<cmd>Git push<CR>", desc = "Git push" },
      },
   },

   -- Mini.nvim modules (lightweight alternatives)
   {
      "echasnovski/mini.nvim",
      version = false,
      event = "VeryLazy",
      config = function()
         -- Comment (replaces kommentary)
         require("mini.comment").setup({
            options = {
               ignore_blank_line = true,
               start_of_line = false,
               pad_comment_parts = true,
            },
            mappings = {
               comment = "gc",
               comment_line = "gcc",
               comment_visual = "gc",
               textobject = "gc",
            },
         })

         -- Surround (lightweight alternative to vim-surround)
         require("mini.surround").setup({
            mappings = {
               add = "sa",
               delete = "sd",
               find = "sf",
               find_left = "sF",
               highlight = "sh",
               replace = "sr",
               update_n_lines = "sn",
            },
         })

         -- Pairs (auto-pairs)
         require("mini.pairs").setup({
            modes = { insert = true, command = true, terminal = false },
            skip_next = [=[[%w%%%'%[%"%.%`%$]]=],
            skip_ts = { "string" },
            skip_unbalanced = true,
            markdown = true,
         })

         -- AI (text objects for arguments/indents)
         require("mini.ai").setup({
            n_lines = 500,
            custom_textobjects = {
               o = require("mini.ai").gen_spec.treesitter({
                  a = { "@block.outer", "@conditional.outer", "@loop.outer" },
                  i = { "@block.inner", "@conditional.inner", "@loop.inner" },
               }, {}),
               f = require("mini.ai").gen_spec.treesitter({ a = "@function.outer", i = "@function.inner" }, {}),
               c = require("mini.ai").gen_spec.treesitter({ a = "@class.outer", i = "@class.inner" }, {}),
            },
         })
      end,
   },

   -- Harpoon (enhanced configuration)
   {
      "ThePrimeagen/harpoon",
      branch = "harpoon2",
      keys = {
         { "<leader>a", desc = "Add to Harpoon" },
         { "<C-e>", desc = "Harpoon menu" },
         { "<C-t>", desc = "Harpoon telescope" },
         { "<C-h>", desc = "Previous Harpoon" },
         { "<C-l>", desc = "Next Harpoon" },
         { "<leader>1", desc = "Harpoon 1" },
         { "<leader>2", desc = "Harpoon 2" },
         { "<leader>3", desc = "Harpoon 3" },
         { "<leader>4", desc = "Harpoon 4" },
         { "<leader>5", desc = "Harpoon 5" },
         { "<leader>6", desc = "Harpoon 6" },
         { "<leader>7", desc = "Harpoon 7" },
         { "<leader>8", desc = "Harpoon 8" },
         { "<leader>9", desc = "Harpoon 9" },
         { "<leader>0", desc = "Harpoon 10" },
      },
      dependencies = { "nvim-lua/plenary.nvim", "nvim-telescope/telescope.nvim" },
      config = function()
         local harpoon = require("harpoon")
         local harpoon_auto_add_enabled = false
         local list = harpoon:list()
         harpoon:setup()
         
         local conf = require("telescope.config").values
         local extensions = require("harpoon.extensions")
         
         harpoon:extend(extensions.builtins.highlight_current_file())
         harpoon:extend(extensions.builtins.navigate_with_number())
         
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
            local file_name = vim.fn.expand('%:t')
            vim.notify("File: " .. file_name .. " added to Harpoon", vim.log.levels.INFO)
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
                   autocmd BufWritePost * lua list:add(); vim.notify("File: " .. vim.fn.expand('%:t') .. " added to Harpoon", vim.log.levels.INFO)
                 augroup END
               ]]
               vim.notify("Harpoon auto-add enabled", vim.log.levels.INFO)
            end
            harpoon_auto_add_enabled = not harpoon_auto_add_enabled
         end
         
         -- Command to toggle the autocmd
         vim.api.nvim_create_user_command("HarpoonAutoAdd", toggle_harpoon_auto_add, {})
      end,
   },


   -- Trouble (diagnostics UI - lazy-loaded)
   {
      "folke/trouble.nvim",
      cmd = { "Trouble", "TroubleToggle" },
      keys = {
         { "<leader>xx", "<cmd>Trouble diagnostics toggle<CR>", desc = "Diagnostics (Trouble)" },
         { "<leader>xX", "<cmd>Trouble diagnostics toggle filter.buf=0<CR>", desc = "Buffer Diagnostics (Trouble)" },
         { "<leader>cs", "<cmd>Trouble symbols toggle focus=false<CR>", desc = "Symbols (Trouble)" },
         { "<leader>cl", "<cmd>Trouble lsp toggle focus=false win.position=right<CR>", desc = "LSP Definitions / references / ... (Trouble)" },
         { "<leader>xL", "<cmd>Trouble loclist toggle<CR>", desc = "Location List (Trouble)" },
         { "<leader>xQ", "<cmd>Trouble qflist toggle<CR>", desc = "Quickfix List (Trouble)" },
      },
      opts = {},
   },

   -- Which-key (lazy-loaded, helps with keybindings discovery)
   {
      "folke/which-key.nvim",
      event = "VeryLazy",
      init = function()
         vim.o.timeout = true
         vim.o.timeoutlen = 300
      end,
      opts = {
         plugins = {
            marks = false,
            registers = false,
            spelling = {
               enabled = true,
            },
            presets = {
               operators = false,
               motions = false,
               text_objects = false,
               windows = false,
               nav = false,
               z = false,
               g = false,
            },
         },
      },
   },

   -- Copilot (disabled to avoid errors)
   {
      "github/copilot.vim",
      event = "InsertEnter",
      enabled = false, -- Disabled to prevent autocmd errors
      cond = function()
         return vim.fn.executable("node") == 1
      end,
      config = function()
         vim.g.copilot_no_tab_map = true
         vim.g.copilot_assume_mapped = true
         vim.api.nvim_set_keymap("i", "<C-j>", 'copilot#Accept("<CR>")', { silent = true, expr = true, script = true })
         vim.api.nvim_set_keymap("i", "<C-]>", "<Plug>(copilot-next)", {})
         vim.api.nvim_set_keymap("i", "<C-[>", "<Plug>(copilot-previous)", {})
      end,
   },

   -- CopilotChat (disabled since copilot.vim is disabled)
   {
      "CopilotC-Nvim/CopilotChat.nvim",
      branch = "main",
      cmd = "CopilotChat",
      event = "VeryLazy",
      enabled = false, -- Disabled since copilot.vim is disabled
      dependencies = {
         { "github/copilot.vim" },
         { "nvim-lua/plenary.nvim" },
      },
      opts = {
         debug = false,
      },
   },

   -- Avante.nvim (AI assistance)
   {
      "yetone/avante.nvim",
      enabled = false,  -- Disabled like in main config to avoid SSH askpass issues
      event = "VeryLazy",
      build = "make",
      opts = {},
      dependencies = {
         "nvim-treesitter/nvim-treesitter",
         "stevearc/dressing.nvim",
         "nvim-lua/plenary.nvim",
         "MunifTanjim/nui.nvim",
         "nvim-tree/nvim-web-devicons",
      },
   },

   {
      "MunifTanjim/nui.nvim",
      lazy = true,
   },

   -- Markdown preview (lazy-loaded on filetype)
   {
      "iamcco/markdown-preview.nvim",
      cmd = { "MarkdownPreviewToggle", "MarkdownPreview", "MarkdownPreviewStop" },
      ft = { "markdown" },
      build = function()
         vim.fn["mkdp#util#install"]()
      end,
      keys = {
         { "<leader>mp", "<cmd>MarkdownPreviewToggle<CR>", desc = "Markdown Preview" },
      },
   },

   -- Todo comments (lazy-loaded)
   {
      "folke/todo-comments.nvim",
      event = { "BufReadPost", "BufNewFile" },
      dependencies = { "nvim-lua/plenary.nvim" },
      opts = {
         signs = false,
      },
   },

   -- Additional plugins from original config
   {
      "nvim-tree/nvim-web-devicons",
      lazy = true,
   },

   {
      "laytan/cloak.nvim",
      event = "VeryLazy",
      optional = true,
      config = function()
         local ok, cloak = pcall(require, "cloak")
         if ok then
            cloak.setup({
               enabled = true,
               cloak_character = "*",
               highlight_group = "Comment",
               patterns = {
                  {
                     file_pattern = {
                        ".env*",
                        "wrangler.toml",
                        ".dev.vars",
                     },
                     cloak_pattern = "=.+",
                  },
               },
            })
         end
      end,
   },

   {
      "isakbm/gitgraph.nvim",
      cmd = "GitGraph",
      dependencies = { "nvim-tree/nvim-web-devicons" },
      config = function()
         require("gitgraph").setup()
      end,
   },

   {
      "nvim-telescope/telescope-file-browser.nvim",
      dependencies = { "nvim-telescope/telescope.nvim", "nvim-lua/plenary.nvim" },
      keys = {
         { "<leader>fe", "<cmd>Telescope file_browser<CR>", desc = "File browser" },
      },
   },

   {
      "tpope/vim-dotenv",
      event = "VeryLazy",
      enabled = false, -- Disabled due to netrw event error
   },

   {
      "brenoprata10/nvim-highlight-colors",
      lazy = false,  -- Load immediately for color highlighting
      priority = 800,  -- Load after LSP
      opts = {
         render = "background", -- or "foreground" or "virtual"
         enable_named_colors = true,
         enable_tailwind = false,
         custom_colors = {},
      },
      config = function(_, opts)
         require("nvim-highlight-colors").setup(opts)
         -- Ensure color highlighting is active for all buffers
         vim.api.nvim_create_autocmd("BufEnter", {
            callback = function()
               require("nvim-highlight-colors").turnOn()
            end,
         })
      end,
   },

   {
      "leath-dub/snipe.nvim",
      keys = {
         { "gb", function() require("snipe").open_buffer_menu() end, desc = "Open Snipe buffer menu" },
      },
      opts = {},
   },

   {
      "akinsho/toggleterm.nvim",
      version = "*",
      cmd = { "ToggleTerm", "TermExec" },
      keys = {
         { "<C-\\>", "<cmd>ToggleTerm<CR>", desc = "Toggle terminal" },
         { "<leader>tt", "<cmd>ToggleTerm<CR>", desc = "Toggle terminal" },
      },
      config = function()
         require("toggleterm").setup({
            size = 20,
            open_mapping = [[<C-\>]],
            direction = "float",
            float_opts = {
               border = "curved",
            },
         })
      end,
   },

   {
      "mg979/vim-visual-multi",
      branch = "master",
      event = { "BufReadPost", "BufNewFile" },
   },

   {
      "eandrju/cellular-automaton.nvim",
      cmd = "CellularAutomaton",
      keys = {
         { "<leader>fml", "<cmd>CellularAutomaton make_it_rain<CR>", desc = "Make it rain" },
      },
   },

   {
      "ThePrimeagen/vim-be-good",
      cmd = "VimBeGood",
   },

   {
      "codethread/qmk.nvim",
      ft = { "c", "cpp" },
      config = function()
         local qmk = require("qmk")
         qmk.setup({
            name = "LAYOUT_split_3x6_3",
            comment_preview = {
               position = "top",
            },
            layout = {
               "x x x x x x _ _ _ x x x x x x",
               "x x x x x x _ _ _ x x x x x x",
               "x x x x x x _ _ _ x x x x x x",
               "_ _ _ x x x _ _ _ x x x _ _ _",
            },
         })
      end,
   },

   {
      "mbbill/undotree",
      cmd = "UndotreeToggle",
      keys = {
         { "<leader>u", "<cmd>UndotreeToggle<CR>", desc = "Toggle undotree" },
      },
   },
   
   {
      "sindrets/diffview.nvim",
      cmd = { "DiffviewOpen", "DiffviewClose", "DiffviewToggleFiles", "DiffviewFocusFiles" },
      keys = {
         { "<leader>dvo", "<cmd>DiffviewOpen<CR>", desc = "Open diff view" },
         { "<leader>dvc", "<cmd>DiffviewClose<CR>", desc = "Close diff view" },
      },
      config = function()
         require("diffview").setup()
      end,
   },
   
   {
      "folke/zen-mode.nvim",
      cmd = "ZenMode",
      keys = {
         { "<leader>zz", "<cmd>ZenMode<CR>", desc = "Toggle Zen Mode" },
      },
      opts = {
         window = {
            width = 120,
            options = {
               signcolumn = "no",
               number = false,
               relativenumber = false,
               cursorline = false,
               cursorcolumn = false,
               foldcolumn = "0",
               list = false,
            },
         },
      },
   },
   
   {
      "m4xshen/hardtime.nvim",
      event = "VeryLazy",
      dependencies = { "MunifTanjim/nui.nvim", "nvim-lua/plenary.nvim" },
      opts = {
         max_count = 4,
         disable_mouse = false,
         disabled_keys = {
            ["<Up>"] = {},
            ["<Down>"] = {},
            ["<Left>"] = {},
            ["<Right>"] = {},
         },
      },
   },

   {
      "nvim-lua/plenary.nvim",
      lazy = true,
   },
}

-- Initialize lazy.nvim with optimized settings
require("lazy").setup(plugins, {
   defaults = {
      lazy = true,
   },
   performance = {
      cache = {
         enabled = true,
      },
      reset_packpath = true,
      rtp = {
         reset = true,
         disabled_plugins = {
            "gzip",
            "matchit",
            "matchparen",
            "netrwPlugin",
            "tarPlugin",
            "tohtml",
            "tutor",
            "zipPlugin",
         },
      },
   },
   git = {
      url_format = "git@github.com:%s.git", -- Use SSH instead of HTTPS
   },
   install = {
      colorscheme = { "tokyonight" },
   },
   checker = {
      enabled = true,
      notify = true,
   },
   change_detection = {
      enabled = true,
      notify = true,
   },
})