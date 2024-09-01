-- 'nvim-telescope/telescope.nvim', tag = '0.1.0',

local builtin = require('telescope.builtin')
vim.keymap.set('n', '<leader>pf', builtin.find_files, {})
vim.keymap.set('n', '<C-p>', builtin.git_files, {})
vim.keymap.set('n', '<leader>pc', function()
   builtin.grep_string({ search = vim.fn.input("Grep > ", vim.fn.getreg('+')) });
end)
vim.keymap.set('n', '<leader>ps', function()
   builtin.grep_string({ search = vim.fn.input("Grep > ") });
end)

require("telescope").setup {
   defaults = {
      file_ignore_patterns = {
         "node_modules"
      },
      preview = {
         mime_hook = function(filepath, bufnr, opts)
            local is_image = function(filepath)
               local image_extensions = { 'png', 'jpg' } -- Supported image formats
               local split_path = vim.split(filepath:lower(), '.', { plain = true })
               local extension = split_path[#split_path]
               return vim.tbl_contains(image_extensions, extension)
            end
            if is_image(filepath) then
               local term = vim.api.nvim_open_term(bufnr, {})
               local function send_output(_, data, _)
                  for _, d in ipairs(data) do
                     vim.api.nvim_chan_send(term, d .. '\r\n')
                  end
               end
               vim.fn.jobstart(
                  {
                     'catimg', filepath -- Terminal image viewer command
                  },
                  { on_stdout = send_output, stdout_buffered = true, pty = true })
            else
               require("telescope.previewers.utils").set_preview_message(bufnr, opts.winid, "Binary cannot be previewed")
            end
         end
      },
   }
}
