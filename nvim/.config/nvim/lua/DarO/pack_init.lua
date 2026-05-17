local gh = require("DarO.utils").gh

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

   if builds[name] then
      local ok, err = pcall(builds[name])
      if not ok then
         vim.notify(name .. " build failed: " .. tostring(err), vim.log.levels.ERROR)
      end
   end
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
   { src = gh('nvim-telescope/telescope.nvim'), version = 'v0.2.2' },

   -- Navigation
   { src = gh('ThePrimeagen/harpoon'), version = 'harpoon2' },

   -- Editor enhancements
   gh('windwp/nvim-autopairs'),
   gh('laytan/cloak.nvim'),
   gh('stevearc/conform.nvim'),
   gh('brenoprata10/nvim-highlight-colors'),
   gh('mg979/vim-visual-multi'),
   gh('folke/trouble.nvim'),
   gh('github/copilot.vim'),
   gh('akinsho/toggleterm.nvim'),
})

-----------------------------------------------------
-- 2b. Scheduled (non-critical, off startup path)
-----------------------------------------------------
vim.schedule(function()
   local ok, err = pcall(function()
      vim.pack.add({
         gh('eandrju/cellular-automaton.nvim'),
         gh('folke/todo-comments.nvim'),
      })
      require("DarO.plugins.todo-comments")
   end)
   if not ok then
      vim.notify("Deferred plugin load failed (todo-comments/cellular-automaton): " .. tostring(err), vim.log.levels.WARN)
   end
end)

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
require("DarO.plugins.copilot")
require("DarO.plugins.snippets")
require("DarO.plugins.telescope")
require("DarO.plugins.harpoon")
require("DarO.plugins.autopairs")
require("DarO.plugins.cloak")
require("DarO.plugins.conform")
require("DarO.plugins.highlight-color")
require("DarO.plugins.vim-visual-multi")
require("DarO.plugins.trouble")

-- Deferred configs (set up autocommand/keymap triggers)
require("DarO.plugins.copilot-chat")
require("DarO.plugins.diffview")
require("DarO.plugins.lazydocker")
require("DarO.plugins.markdown-preview")
require("DarO.plugins.dotenv")
require("DarO.plugins.undotree")
require("DarO.plugins.vimbegood")
require("DarO.plugins.snipe")

-----------------------------------------------------
-- 5. Pack management commands
-----------------------------------------------------
local pack_dir = vim.fn.stdpath('data') .. '/site/pack/core/opt'

vim.api.nvim_create_user_command('PackClean', function()
   local input = vim.fn.input('Delete all plugins and re-download? (y/N): ')
   if input:lower() ~= 'y' then
      vim.notify('Cancelled', vim.log.levels.INFO)
      return
   end
   vim.fn.delete(pack_dir, 'rf')
   vim.notify('Plugins removed. Restart nvim to re-download.', vim.log.levels.WARN)
end, { desc = 'Remove all plugins for a clean re-download on next startup' })

vim.api.nvim_create_user_command('PackUpdate', function()
   local ok, err = pcall(vim.pack.update)
   if not ok then
      vim.notify("PackUpdate error: " .. tostring(err), vim.log.levels.ERROR)
   end
end, { desc = 'Update all plugins' })

vim.api.nvim_create_user_command('PackStatus', function()
   local ok, err = pcall(function()
      local plugins = vim.pack.get(nil, { info = true })
      local lines = {}
      for _, p in ipairs(plugins) do
         local spec = p.spec or {}
         local status = p.active and '✓' or '✗'
         local pinned = spec.version or ''
         local tags = p.tags or {}
         local latest_tag = tags[#tags] or ''
         local version = pinned ~= '' and pinned or latest_tag
         table.insert(lines, string.format(
            '%s %-30s %s  %s', status, spec.name or "?", (p.rev or "unknown"):sub(1, 8), version
         ))
      end
      table.sort(lines)
      vim.notify(string.format('%d plugins\n%s', #plugins, table.concat(lines, '\n')), vim.log.levels.INFO)
   end)
   if not ok then
      vim.notify("PackStatus error: " .. tostring(err), vim.log.levels.ERROR)
   end
end, { desc = 'Show installed plugins with status' })

vim.keymap.set('n', '<leader>pc', '<cmd>PackClean<CR>', { desc = 'Pack: clean reinstall' })
vim.keymap.set('n', '<leader>pr', '<cmd>restart<CR>', { desc = 'Pack: restart nvim' })
vim.keymap.set('n', '<leader>pu', '<cmd>PackUpdate<CR>', { desc = 'Pack: update all' })
vim.keymap.set('n', '<leader>ps', '<cmd>PackStatus<CR>', { desc = 'Pack: status' })
