return {
   "mbbill/undotree",
   keys = {
      { "<leader>u", vim.cmd.UndotreeToggle, desc = "Toggle UndoTree" }
   },
   cmd = { "UndotreeToggle", "UndotreeShow", "UndotreeHide", "UndotreeFocus" },
}
