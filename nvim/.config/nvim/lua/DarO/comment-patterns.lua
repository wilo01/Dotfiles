local M = {}

M.commentstrings = {
   bicep = "// %s",
   sql = "-- %s",
   lua = "-- %s",
   python = "# %s",
   bash = "# %s",
   sh = "# %s",
   zsh = "# %s",
   fish = "# %s",
   vim = '" %s',
   javascript = "// %s",
   typescript = "// %s",
   javascriptreact = "// %s",
   typescriptreact = "// %s",
   c = "// %s",
   cpp = "// %s",
   objc = "// %s",
   objcpp = "// %s",
   go = "// %s",
   rust = "// %s",
   zig = "// %s",
   java = "// %s",
   kotlin = "// %s",
   swift = "// %s",
   csharp = "// %s",
   scala = "// %s",
   yaml = "# %s",
   toml = "# %s",
   json = "// %s",
   jsonc = "// %s",
   json5 = "// %s",
   css = "/* %s */",
   scss = "// %s",
   sass = "// %s",
   less = "// %s",
   html = "<!-- %s -->",
   xml = "<!-- %s -->",
   vue = "<!-- %s -->",
   svelte = "<!-- %s -->",
   astro = "<!-- %s -->",
   markdown = "<!-- %s -->",
   dockerfile = "# %s",
   make = "# %s",
   cmake = "# %s",
   gitcommit = "# %s",
   gitconfig = "# %s",
   gitignore = "# %s",
   conf = "# %s",
   dotenv = "# %s",
   config = "# %s",
   ini = "; %s",
   r = "# %s",
   julia = "# %s",
   perl = "# %s",
   ruby = "# %s",
   php = "// %s",
   elixir = "# %s",
   erlang = "% %s",
   haskell = "-- %s",
   ocaml = "(* %s *)",
   fsharp = "// %s",
   clojure = "; %s",
   lisp = "; %s",
   scheme = "; %s",
   racket = "; %s",
   nix = "# %s",
   terraform = "# %s",
   hcl = "# %s",
   graphql = "# %s",
   proto = "// %s",
   dart = "// %s",
   groovy = "// %s",
   apex = "// %s",
   solidity = "// %s",
   vhdl = "-- %s",
   verilog = "// %s",
   systemverilog = "// %s",
   matlab = "% %s",
   octave = "% %s",
   tex = "% %s",
   latex = "% %s",
   bibtex = "% %s",
   plaintex = "% %s",
   asm = "; %s",
   nasm = "; %s",
   masm = "; %s",
   arduino = "// %s",
   glsl = "// %s",
   hlsl = "// %s",
   wgsl = "// %s",
}

function M.get_commentstring(ft)
   return M.commentstrings[ft] or vim.bo.commentstring or "// %s"
end

function M.get_comment_prefix(ft)
   local cs = M.get_commentstring(ft)
   local prefix = cs:match("^(.-)%s*%%s") or "//"
   return prefix:gsub("%s+$", "")
end

function M.get_comment_suffix(ft)
   local cs = M.get_commentstring(ft)
   local suffix = cs:match("%%s%s*(.-)$") or ""
   return suffix:gsub("^%s+", "")
end

function M.get_comment_prefix_spaced(ft)
   local prefix = M.get_comment_prefix(ft)
   if not prefix:match("%s$") and not prefix:match("%*$") then
      return prefix .. " "
   end
   return prefix
end

function M.get_comment_pattern(ft)
   local prefix = M.get_comment_prefix(ft)
   local escaped = prefix:gsub("([%^%$%(%)%%%.%[%]%*%+%-%?])", "%%%1")
   return escaped .. "%s*"
end

function M.get_comment_pattern_raw(ft)
   local prefix = M.get_comment_prefix(ft)
   if not prefix:match("[%*%)]$") then
      return prefix .. " "
   end
   return prefix
end

function M.has_block_comments(ft)
   return M.get_comment_suffix(ft) ~= ""
end

function M.get_block_comment(ft, content)
   local cs = M.get_commentstring(ft)
   return cs:format(content)
end

function M.create_todo_comment(ft, message)
   local prefix = M.get_comment_prefix_spaced(ft)
   local suffix = M.get_comment_suffix(ft)

   local todo = prefix .. "[ ] TODO: " .. message
   if suffix ~= "" then
      todo = todo .. " " .. suffix
   end

   return todo
end

function M.get_supported_filetypes()
   local filetypes = {}
   for ft, _ in pairs(M.commentstrings) do
      table.insert(filetypes, ft)
   end
   table.sort(filetypes)
   return filetypes
end

function M.debug_info(ft)
   ft = ft or vim.bo.filetype
   print("Filetype: " .. ft)
   print("Commentstring: " .. M.get_commentstring(ft))
   print("Prefix: '" .. M.get_comment_prefix(ft) .. "'")
   print("Prefix (spaced): '" .. M.get_comment_prefix_spaced(ft) .. "'")
   print("Suffix: '" .. M.get_comment_suffix(ft) .. "'")
   print("Pattern (escaped): '" .. M.get_comment_pattern(ft) .. "'")
   print("Pattern (raw): '" .. M.get_comment_pattern_raw(ft) .. "'")
   print("Has block comments: " .. tostring(M.has_block_comments(ft)))
   print("TODO example: " .. M.create_todo_comment(ft, "Fix this"))
end

return M
