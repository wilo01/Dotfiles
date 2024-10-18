return {
   {
      'hrsh7th/nvim-cmp',
      event = "InsertEnter",
      dependencies = {
         { 'hrsh7th/cmp-buffer' },
         { 'hrsh7th/cmp-path' },
         { 'saadparwaiz1/cmp_luasnip' },
         -- lsp support
         { 'hrsh7th/cmp-nvim-lsp' },
      },
      config = function()
         vim.opt.completeopt = { 'menu', 'menuone', 'noselect' }

         local cmp = require('cmp')
         local luasnip = require('luasnip')
         local select_opts = { behavior = cmp.SelectBehavior.Insert }

         cmp.setup({
            snippet = {
               expand = function(args)
                  luasnip.lsp_expand(args.body)
               end,
            },
            sources = {
               { name = 'path' },
               { name = 'nvim_lsp', keyword_length = 3 },
               -- { name = 'nvim_lsp_signature_help' },
               { name = 'buffer', keyword_length = 3 },
               { name = 'luasnip', keyword_length = 2 },
            },
            window = {
               documentation = cmp.config.window.bordered()
            },
            formatting = {
               fields = { 'menu', 'abbr', 'kind' },
               format = function(entry, item)
                  local menu_icon = {
                     nvim_lsp = '',
                     luasnip = '',
                     buffer = '',
                     path = '',
                  }

                  item.menu = menu_icon[entry.source.name]
                  return item
               end,
            },
            -- See :help cmp-mapping
            mapping = {
               ['<Up>'] = cmp.mapping.select_prev_item(select_opts, { desc = "Select previous item" }),
               ['<Down>'] = cmp.mapping.select_next_item(select_opts, { desc = "Select next item" }),

               ['<C-k>'] = cmp.mapping.select_prev_item(select_opts, { desc = "Select previous item" }),
               ['<C-j>'] = cmp.mapping.select_next_item(select_opts, { desc = "Select next item" }),

               ['<C-u>'] = cmp.mapping.scroll_docs(-4, { desc = "Scroll documentation up" }),
               ['<C-f>'] = cmp.mapping.scroll_docs(4, { desc = "Scroll documentation down" }),

               ['<C-e>'] = cmp.mapping.abort({ desc = "Abort completion" }),
               ['<CR>'] = cmp.mapping.confirm({ select = false, desc = "Confirm selection" }),

               ['<C-d>'] = cmp.mapping(function(fallback)
                  if luasnip.jumpable(1) then
                     luasnip.jump(1)
                  else
                     fallback()
                  end
               end, { 'i', 's', desc = "Jump forward in snippet" }),

               ['<C-b>'] = cmp.mapping(function(fallback)
                  if luasnip.jumpable(-1) then
                     luasnip.jump(-1)
                  else
                     fallback()
                  end
               end, { 'i', 's', desc = "Jump backward in snippet" }),

               ['<Tab>'] = cmp.mapping(function(fallback)
                  local col = vim.fn.col('.') - 1

                  if cmp.visible() then
                     cmp.select_next_item(select_opts)
                  elseif col == 0 or vim.fn.getline('.'):sub(col, col):match('%s') then
                     fallback()
                  else
                     cmp.complete()
                  end
               end, { 'i', 's', desc = "Select next item or trigger completion" }),

               ['<S-Tab>'] = cmp.mapping(function(fallback)
                  if cmp.visible() then
                     cmp.select_prev_item(select_opts)
                  else
                     fallback()
                  end
               end, { 'i', 's', desc = "Select previous item" }),
            },
         })
      end,
   },
}
