return {
   "CopilotC-Nvim/CopilotChat.nvim",
   dependencies = {
      { "github/copilot.vim" }, -- Or use 'zbirenbaum/copilot.lua' if preferred
      { "nvim-lua/plenary.nvim", branch = "master" },
   },
   enabled = true,
   build = "make tiktoken", -- Only on MacOS or Linux
   cmd = { "CopilotChat", "CopilotChatOpen", "CopilotChatToggle", "CopilotChatModels" },
   keys = {
      { "<c-s>",     "<CR>", ft = "copilot-chat", desc = "Submit Prompt", remap = true },
      { "<leader>A", "",     desc = "+ai",        mode = { "n", "v" } },
      {
         "<leader>gc",
         function()
            return require("CopilotChat").toggle()
         end,
         desc = "Toggle (CopilotChat)",
         mode = { "n", "v" },
      },
      {
         "<leader>gx",
         function()
            return require("CopilotChat").reset()
         end,
         desc = "Clear (CopilotChat)",
         mode = { "n", "v" },
      },
      {
         "<leader>GP",
         function()
            require("CopilotChat").select_prompt()
         end,
         desc = "Prompt Actions (CopilotChat)",
         mode = { "n", "v" },
      },
      {
         "<leader>gm",
         function()
            require("CopilotChat").select_model()
         end,
         desc = "Select model (CopilotChat)",
         mode = { "n", "v" },
      },
   },
   opts = {
      model = "claude-sonnet-4", -- List model names with `:CopilotChatModels` command
      agent = "copilot",
      remember_as_sticky = true,
      auto_insert_mode = true,
      selection = function(source)
         local select = require("CopilotChat.select")
         return select.visual(source) or select.buffer(source)
      end,
      mappings = {
         complete = {
            insert = '<CR>',
         },
         close = {
            normal = 'q',
            insert = '<C-c>',
         },
         reset = {
            normal = '<C-l>',
            insert = '<C-l>',
         },
         submit_prompt = {
            normal = '<CR>',
            insert = '<C-s>',
         },
         toggle_sticky = {
            normal = 'grr',
         },
         clear_stickies = {
            normal = 'grx',
         },
         accept_diff = {
            normal = '<C-y>',
            insert = '<C-y>',
         },
         jump_to_diff = {
            normal = 'gj',
         },
         quickfix_answers = {
            normal = 'gqa',
         },
         quickfix_diffs = {
            normal = 'gqd',
         },
         yank_diff = {
            normal = 'gy',
            register = '"',
         },
         show_diff = {
            normal = 'gd',
            full_diff = false,
         },
         show_info = {
            normal = 'gi',
         },
         show_context = {
            normal = 'gc',
         },
         show_help = {
            normal = 'gh',
         },
      },

      system_prompt = 'COPILOT_INSTRUCTIONS',
      context = nil,
      sticky = nil,

      temperature = 0.1,
      headless = false,
      stream = nil,
      callback = nil,
      window = {
         layout = 'float',
         width = 0.8,
         height = 0.8,

         relative = 'editor',
         border = 'single',
         row = nil,
         col = nil,
         title = 'Copilot Chat',
         footer = nil,
         zindex = 1,
      },

      show_help = true,
      highlight_selection = true,
      highlight_headers = true,
      references_display = 'virtual',
      auto_follow_cursor = true,
      clear_chat_on_new_prompt = false,

      debug = false,
      log_level = 'info', -- Log level to use, 'trace', 'debug', 'info', 'warn', 'error', 'fatal'
      proxy = nil,        -- [protocol://]host[:port] Use this proxy
      allow_insecure = false,

      chat_autocomplete = false,

      log_path = vim.fn.stdpath('state') .. '/CopilotChat.log',
      history_path = vim.fn.stdpath('data') .. '/copilotchat_history',

      question_header = '# User ',
      answer_header = '# Copilot ',
      error_header = '# Error ',
      separator = '───',

      -- -- default providers
      -- -- see config/providers.lua for implementation
      -- providers = {
      --    copilot = {
      --    },
      --    github_models = {
      --    },
      --    copilot_embeddings = {
      --    },
      -- },

      -- -- default contexts
      -- -- see config/contexts.lua for implementation
      -- contexts = {
      --    buffer = {
      --    },
      --    buffers = {
      --    },
      --    file = {
      --    },
      --    files = {
      --    },
      --    git = {
      --    },
      --    url = {
      --    },
      --    register = {
      --    },
      --    quickfix = {
      --    },
      --    system = {
      --    }
      -- },

      -- default prompts
      -- see config/prompts.lua for implementation
      prompts = {
         Explain = {
            prompt = 'Write an explanation for the selected code as paragraphs of text.',
            system_prompt = 'COPILOT_EXPLAIN',
         },
         Review = {
            prompt = 'Review the selected code.',
            system_prompt = 'COPILOT_REVIEW',
         },
         Fix = {
            prompt =
            'There is a problem in this code. Identify the issues and rewrite the code with fixes. Explain what was wrong and how your changes address the problems.',
         },
         Optimize = {
            prompt =
            'Optimize the selected code to improve performance and readability. Explain your optimization strategy and the benefits of your changes.',
         },
         Docs = {
            prompt = 'Please add documentation comments to the selected code.',
         },
         Tests = {
            prompt = 'Please generate tests for my code.',
         },
         Commit = {
            prompt =
            'Write commit message for the change with commitizen convention. Keep the title under 50 characters and wrap message at 72 characters. Format as a gitcommit code block. If there are other bullet points do not change them but add your points bellow to the list. Please do not duplicate the previous logs',
            context = 'git:staged',
         },
      },
   },
}
