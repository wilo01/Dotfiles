-- lazydocker.nvim
return {
   "mgierada/lazydocker.nvim",
   dependencies = { "akinsho/toggleterm.nvim" },
   cmd = { "Lazydocker", "LazyDocker" },
   keys = {
      {
         "<leader>ld",
         function()
            require("lazydocker").open()
         end,
         desc = "Open Lazydocker floating window",
      },
   },
   config = function()
      require("lazydocker").setup({
         border = "curved", -- valid options are "single" | "double" | "shadow" | "curved"
      })
   end,
}
