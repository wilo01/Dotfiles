return {
   "mg979/vim-visual-multi",

   config = function()
      vim.g.VM_leader = '<Space>' -- Set the leader key to Space
      vim.g.VM_maps = {
         ["I BS"] = '',           -- disable backspace mapping
      }
   end
}
