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
         statuscolumn = { enabled = false },
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

   -- Treesitter (lazy-loaded on file read)
   {
      "nvim-treesitter/nvim-treesitter",
      build = ":TSUpdate",
      event = { "BufReadPost", "BufNewFile" },
      cmd = { "TSUpdate", "TSInstall", "TSBufEnable", "TSBufDisable", "TSModuleInfo" },
      config = function()
         require("ai.config.treesitter")
      end,
      dependencies = {
         -- Treesitter text objects (lazy-loaded with treesitter)
         {
            "nvim-treesitter/nvim-treesitter-textobjects",
            config = function()
               require("ai.config.treesitter-textobjects")
            end,
         },
      },
   },

   -- LSP Configuration (lazy-loaded on file type)
   {
      "neovim/nvim-lspconfig",
      event = { "BufReadPost", "BufNewFile" },
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

   -- Git integration
   {
      "lewis6991/gitsigns.nvim",
      event = { "BufReadPost", "BufNewFile" },
      opts = {
         signs = {
            add = { text = "+" },
            change = { text = "~" },
            delete = { text = "_" },
            topdelete = { text = "‾" },
            changedelete = { text = "~" },
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
      },
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

   -- Harpoon (lazy-loaded on key)
   {
      "ThePrimeagen/harpoon",
      branch = "harpoon2",
      keys = {
         { "<leader>a", function() require("harpoon"):list():add() end, desc = "Add to Harpoon" },
         { "<C-e>", function() require("harpoon").ui:toggle_quick_menu(require("harpoon"):list()) end, desc = "Harpoon menu" },
         { "<C-h>", function() require("harpoon"):list():select(1) end, desc = "Harpoon 1" },
         { "<C-t>", function() require("harpoon"):list():select(2) end, desc = "Harpoon 2" },
         { "<C-n>", function() require("harpoon"):list():select(3) end, desc = "Harpoon 3" },
         { "<C-s>", function() require("harpoon"):list():select(4) end, desc = "Harpoon 4" },
      },
      dependencies = { "nvim-lua/plenary.nvim" },
      config = function()
         require("harpoon"):setup()
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
               enabled = false,
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
      "norcalli/nvim-colorizer.lua",
      event = { "BufReadPost", "BufNewFile" },
      config = function()
         require("colorizer").setup()
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
      enabled = false,
      notify = false,
   },
   change_detection = {
      enabled = false,
      notify = false,
   },
})