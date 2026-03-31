local gh = function(x) return 'https://github.com/' .. x end

-----------------------------------------------------
-- 1. Build hooks (MUST come before vim.pack.add)
-----------------------------------------------------
vim.api.nvim_create_autocmd('PackChanged', { callback = function(ev)
   local name, kind = ev.data.spec.name, ev.data.kind
   if kind ~= 'install' and kind ~= 'update' then return end

   local builds = {
      ['telescope-fzf-native.nvim'] = function()
         local info = vim.pack.get({ name })[1]
         if info then vim.system({ 'make' }, { cwd = info.path }):wait() end
      end,
      ['CopilotChat.nvim'] = function()
         local info = vim.pack.get({ name })[1]
         if info then vim.system({ 'make', 'tiktoken' }, { cwd = info.path }):wait() end
      end,
      ['nvim-treesitter'] = function()
         if not ev.data.active then vim.cmd.packadd('nvim-treesitter') end
         vim.cmd('TSUpdate')
      end,
      ['markdown-preview.nvim'] = function()
         if not ev.data.active then vim.cmd.packadd('markdown-preview.nvim') end
         vim.fn['mkdp#util#install']()
      end,
   }

   if builds[name] then builds[name]() end
end })

-----------------------------------------------------
-- 2. Eager plugin declarations
-----------------------------------------------------
vim.pack.add({
   -- Foundation (dependencies first)
   gh('nvim-lua/plenary.nvim'),
   gh('echasnovski/mini.nvim'),
   gh('nvim-tree/nvim-web-devicons'),

   -- Colorscheme
   gh('folke/tokyonight.nvim'),

   -- Core UI
   gh('folke/snacks.nvim'),
   gh('stevearc/dressing.nvim'),

   -- Git
   gh('lewis6991/gitsigns.nvim'),
   gh('tpope/vim-fugitive'),
   gh('isakbm/gitgraph.nvim'),

   -- Treesitter
   gh('nvim-treesitter/nvim-treesitter'),

   -- LSP chain
   gh('williamboman/mason.nvim'),
   gh('williamboman/mason-lspconfig.nvim'),
   gh('WhoIsSethDaniel/mason-tool-installer.nvim'),
   gh('b0o/schemastore.nvim'),
   gh('joechrisellis/lsp-format-modifications.nvim'),
   gh('neovim/nvim-lspconfig'),
   gh('mfussenegger/nvim-lint'),

   -- Completion
   gh('hrsh7th/cmp-buffer'),
   gh('hrsh7th/cmp-path'),
   gh('hrsh7th/cmp-nvim-lsp'),
   gh('hrsh7th/nvim-cmp'),

   -- Telescope
   gh('nvim-telescope/telescope-file-browser.nvim'),
   gh('nvim-telescope/telescope-fzf-native.nvim'),
   { src = gh('nvim-telescope/telescope.nvim'), version = '0.1.8' },

   -- Navigation
   { src = gh('ThePrimeagen/harpoon'), version = 'harpoon2' },

   -- Editor enhancements
   gh('windwp/nvim-autopairs'),
   gh('laytan/cloak.nvim'),
   gh('stevearc/conform.nvim'),
   gh('brenoprata10/nvim-highlight-colors'),
   gh('mg979/vim-visual-multi'),
   gh('folke/trouble.nvim'),
   gh('folke/todo-comments.nvim'),
   gh('github/copilot.vim'),
   gh('akinsho/toggleterm.nvim'),

   -- Fun
   gh('eandrju/cellular-automaton.nvim'),
})

-----------------------------------------------------
-- 3. Plugin configs (order matters)
-----------------------------------------------------

-- Eager configs
require("DarO.plugins.colors")
require("DarO.plugins.snacks")
require("DarO.plugins.mini-nvim")
require("DarO.plugins.dressing")
require("DarO.plugins.gitsigns")
require("DarO.plugins.fugitive")
require("DarO.plugins.gitgraph")
require("DarO.plugins.treesitter")
require("DarO.plugins.lsp")
require("DarO.plugins.snippets")
require("DarO.plugins.telescope")
require("DarO.plugins.harpoon")
require("DarO.plugins.autopairs")
require("DarO.plugins.cloak")
require("DarO.plugins.conform")
require("DarO.plugins.highlight-color")
require("DarO.plugins.vim-visual-multi")
require("DarO.plugins.trouble")
require("DarO.plugins.todo-comments")

-- Deferred configs (set up autocommand/keymap triggers)
require("DarO.plugins.copilot-chat")
require("DarO.plugins.diffview")
require("DarO.plugins.lazydocker")
require("DarO.plugins.markdown-preview")
require("DarO.plugins.dotenv")
require("DarO.plugins.undotree")
require("DarO.plugins.vimbegood")
require("DarO.plugins.snipe")
