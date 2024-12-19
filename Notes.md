# NOTES: TODOs
- [ ] Add autosave on neovim inactivity
- [ ] Fix annoying issue with LSP formatting -> Does not work in large files
- [ ] Add ignore witch key or add longer timeout
- [ ] Improve -> Disable plugins via snacks or command -> use "comp = false"
- [ ] Fix issue with hardtime enabled -> Remap Ctrl+N in visual mode
- [X] Add "multiple grep", "multiple search parameters"
- [X] Add something to exclude search
- [x] Add git diff when jumping via gitsigns hunks -> Works with J / K and <leader>va
- [X] Search breaks and does not search across the project directory
- [X] Fix issue with telescope
- [X] Show current branch in tmux / neovim for example maintenance/12.1AV -> Used tmux instead
- [X] Fix issue with case sensitive search
- [X] Implement easy rename of the files
- [X] Use current rename ? Drop it ? -> Use <leader>pb -> R - for rename
- [X] Implement easy path copy -> Use <leader>fp
- [X] Implement white spaces formatting
# NOTES: Check if remap is already been used
- Inside neovim type this command
```bash
    :map <keybind>
```
# NOTES: FIX for pavucontrol
pavucontrol a gtk application? Try these two commands:
```bash
    gsettings set org.gnome.desktop.interface color-scheme prefer-dark
    gsettings set org.gnome.desktop.interface gtk-theme Adwaita-dark
```
# NOTES: TODO case insensitive
- [X] Fix issue with case sensitive search
# NOTES: TODO Easy rename file
- [X] Implement easy rename of the files
- [X] Use current rename ? Drop it ? -> Use <leader>pb -> R - for rename
# NOTES: TODO add file name path for easy copy
- [How can I see the full path of the current file?](https://vi.stackexchange.com/questions/104/how-can-i-see-the-full-path-of-the-current-file)
- [X] Implement easy path copy -> Use <leader>fp
```lua
    :!echo %
    :!echo %:t
    :!echo %:p
    :!echo %:p:h
```
# NOTES: TODO formatting white spaces
- [X] Implement white spaces formatting
```lua
    local augroup = vim.api.nvim_create_augroup
    local TheDaroGroup = augroup('TheDarO', {})
    local autocmd = vim.api.nvim_create_autocmd
    -- Remove white spaces at the end
    autocmd({"BufWritePre"}, {
        group = TheDaroGroup,
        pattern = "*",
        command = [[%s/\s\+$//e]],
    })
```
# NOTES: Git hooks example
- Command to edit
```sh
    git config --local -e
``
- Example
```js
    [hooks]
       enableLocalHooks = true
       enableGlobalHooks = true
       hooksLocalPath = ~/Dev/Private
       hooksLocalFilename = Notes
```
# 2024-06-14 13:07
## TMP
JIRA: #Linux

Logs:
- Updated kitty config from Linux Machine

.config/kitty/kitty.conf
- Commit branch HASH [f70e7de5ed86caee530035300ce8c523da4098c5](Linux)

# 2024-06-14 13:08
## TMP
JIRA: #Linux

Logs:
- Updated to new Changes
- Updated sed to work on Linux Machine

.git-template/hooks/prepare-commit-msg
- Commit branch HASH [fa122a3e16fc3e24d99aca5633f716552c23548b](Linux)

# 2024-06-14 13:12
## master
JIRA: #master

Logs:
- Notes update

Notes.md
Pictures/search_cluster-safe_user_portlet_options.png
- Commit branch HASH [c80972035e3b219584bed5f20bdfa7d56bfcf472](master)

# 2024-06-17 10:36
## master
JIRA: #master

Logs:
- Added new file

Stash_Notes.md
- Commit branch HASH [7afdf78be27883c30261464bd6751c69a65a470f](master)

# 2024-06-17 14:19
## master
JIRA: #master

Logs:
- Updated notes about SSO and kiosk_chrome_app access list

Notes.md
- Commit branch HASH [f1bf6d433ead56df3b630fe8e1c21cfc7dcde2b7](master)

# 2024-06-17 17:06
## master
JIRA: #master

Logs:
- SSO instructions

Notes.md
- Commit branch HASH [be5fb8de717e68b6a85a99bfda702da5fb8c5ead](master)

# 2024-06-18 14:57
## master
JIRA: #master

Logs:
- Updated notes with new MR

Notes.md
Stash_Notes.md
- Commit branch HASH [bc9590b311cb880a68745c8de801e1769a783064](master)

# 2024-06-19 11:47
## master
JIRA: #master

Logs:
- Fixing typo in db table name
- Specify what db to be used to get into Clients env

Notes.md
- Commit branch HASH [bfc7166e640efdddb7f8e717fa19f522295c4888](master)

# 2024-06-19 14:48
## master
JIRA: #master

Logs:
- Updated notes with kiosk chromebox_id
- How to backup oracle table

Notes.md
- Commit branch HASH [791f893db5cfc1fe8284217d7ced4ec5db6272ef](master)

# 2024-06-20 23:57
## TMP
JIRA: #Linux

Logs:
- Initial config for snippets

.config/nvim/after/plugin/luasnip.lua
.config/nvim/nvim_notes.md
- Commit branch HASH [2434f7258384d557da5bda04e1c08a28670ea283](Linux)

# 2024-06-21 00:12
## TMP
JIRA: #Linux

Logs:
- Updated snippets for easier reading

.config/nvim/after/plugin/luasnip.lua
- Commit branch HASH [e8fcf48a6129e280c5ad190a06cf3fa65c75aa47](Linux)

# 2024-06-21 10:52
## master
JIRA: #master

Logs:
- Commit

Notes.md
Stash_Notes.md
- Commit branch HASH [b2fc98c69d578e89de7141a066418b33b00a42d7](master)

# 2024-06-21 13:43
## master
JIRA: #master

Logs:
- Removed image and background Item type from Add Item button
- Removed edit for image and background
- Use add Image instead to upload and handle image and background
- Added new style .blank-space-full

Notes.md
- Commit branch HASH [ffdc502fdd373ea198c8dc857c49297e98b1293d](master)

# 2024-06-21 13:44
## master
JIRA: #master

Logs:
- JIRA: #vis-236-label-layout-issues-with-image-type-item-on-the-label-layout-portlet

Notes.md
- Commit branch HASH [af5dc97d5897858ce7e3f207b77edb45a23d54f0](master)

# 2024-06-21 14:01
## master
JIRA: #master

Logs:
- Cherry pick to 12.1AV VIS-233

Notes.md
- Commit branch HASH [de1dc42b92fb8b496f461b62abb4dde24b77ee9d](master)

# 2024-06-21 22:28
## TMP
JIRA: #Linux

Logs:
- Updated luasnip with new snippets
- Fixed issue on visual mode delayed cursor going down
- The issue was with remap for visual mode using jk

.config/nvim/after/plugin/luasnip.lua
.config/nvim/lua/DarO/remap.lua
- Commit branch HASH [b9c405a526c8d843b9732a401929c0919d105150](Linux)

# 2024-06-21 23:40
## TMP
JIRA: #Linux

Logs:
- Removed yadm that was breaking changes

.config/nvim/after/plugin/gitsigns.lua
.config/nvim/plugin/packer_compiled.lua
- Commit branch HASH [53ee9b98432801aff112f7c484c8268768999312](Linux)

# 2024-06-22 00:09
## TMP
JIRA: #Linux

Logs:
- Formatting

.config/nvim/after/plugin/autopairs.lua
.config/nvim/after/plugin/comment.lua
.config/nvim/after/plugin/gitsigns.lua
.config/nvim/after/plugin/lsp.lua
.config/nvim/after/plugin/rainbow-delimiters.lua
.config/nvim/after/plugin/telescope.lua
.config/nvim/after/plugin/treesitter.lua
.config/nvim/after/plugin/vim-visual-multi.lua
.config/nvim/lua/DarO/packer.lua
.config/nvim/lua/DarO/remap.lua
.tmux.conf
- Commit branch HASH [93e0c4536235d41890189dc30dd1eea33e77726a](Linux)

# 2024-06-24 09:19
## master
JIRA: #master

Logs:
- Updated Notes

Notes.md
- Commit branch HASH [1b5579819d41c38ad7e930566c41ec8c1a8123a0](master)

# 2024-06-24 09:21
## master
JIRA: #master

Logs:
- Added Jira-Snippet plugin for personal needs

Jira-Snippet
- Commit branch HASH [5258337e8d6664cb17a95e8641f4e337ce96b362](master)

# 2024-06-24 12:29
## master
JIRA: #master

Logs:
- Updates notes for MIPS

Notes.md
- Commit branch HASH [75c8b68dc416c88f7629c628555b329cee927ef1](master)

# 2024-06-26 11:59
## master
JIRA: #master

Logs:
- Updating notes

Notes.md
- Commit branch HASH [b1d9e298bcb81ab01999916f5be2595582a61751](master)

# 2024-06-27 09:22
## master
JIRA: #master

Logs:
- Kiosk json log

Notes.md
- Commit branch HASH [1e597e577ce4e93db2abe74cba60dcf4baff9fb7](master)

# 2024-06-28 12:32
## master
JIRA: #master

Logs:
- MR for VIS-392
- Some notes

Notes.md
Stash_Notes.md
- Commit branch HASH [5cbc098f5ec6351faf774c85adbcf908d15b492a](master)

# 2024-06-28 14:53
## master
JIRA: #master

Logs:
- Updated the notes about CSP
- Added picture

Notes.md
Pictures/VIS-321.png
- Commit branch HASH [fc740408fdc5b3809ee2d26118e4c5d1b279d557](master)

# 2024-07-01 11:47
## master
JIRA: #master

Logs:
- NOTES: Liquibase remove only once example based on ca_badge_layout_items
- Removing items does not require "--precondition-sql-check expectedResult"
- You just need to be aware that this item will be removed when image is null
- This needs validation on frontend just to work correctly
- so there is no point to do the check when you still going to delete it anyway
- also you can't predict how many records there will be in the database

Notes.md
- Commit branch HASH [78ec64ac3e90f80ae1d8c9e2ceb589686ff08e2d](master)

# 2024-07-02 09:43
## master
JIRA: #master

Logs:
- Notes for TDS More Options button search
- Notes for ID Ware workflows

Notes.md
Stash_Notes.md
- Commit branch HASH [92b91bd86c7db1f5f99f38d7b9c1fd1df2dedbaf](master)

# 2024-07-03 12:28
## master
JIRA: #master

Logs:
- Temporary object
- Notes about ID ware

Notes.md
- Commit branch HASH [d900946d2b87f5924880de3530a78cbfc1220c08](master)

# 2024-07-03 15:56
## master
JIRA: #master

Logs:
- Updated in notes

Notes.md
- Commit branch HASH [97eb896c8c877777c8b9106a1cd4bbdd291c03cb](master)

# 2024-07-04 21:59
## master
JIRA: #master

Logs:
- Update notes about the liquibase

Notes.md
Stash_Notes.md
- Commit branch HASH [327cc9ff5af8408e64cdef9e0b2b1ee8a2aafa19](master)

# 2024-07-07 15:16
## master
JIRA: #master

Logs:
- Updated notes

Notes.md
- Commit branch HASH [1811c8df2da9277392289448951b4765dcdfe824](master)

# 2024-07-07 19:54
## TMP
JIRA: #Linux

Logs:
- Added start-up time on boot

.config/nvim/init.lua
- Commit branch HASH [55eb241bd83d59a3bf2a231e8266eff0bd16135b](Linux)

# 2024-07-07 20:54
## TMP
JIRA: #Linux

Logs:
- Fixed one remap

.config/nvim/after/plugin/lsp.lua
.config/nvim/plugin/packer_compiled.lua
- Commit branch HASH [9bf2474538b141ffebc5075a35b77a59aa0711eb](Linux)

# 2024-07-08 09:38
## master
JIRA: #master

Logs:
- Updated PL/SQL script for ca_badge_layout_items

Notes.md
- Commit branch HASH [9a0f5010758b6413423200ece43fb1c3d78316b1](master)

# 2024-07-08 09:59
## master
JIRA: #master

Logs:
- Updated PL/SQL code for ca_badge_layout_items.sql

Notes.md
- Commit branch HASH [3786245a4f1d6e3e23217665554a027c7ff6a0ec](master)

# 2024-07-08 13:14
## TMP
JIRA: #Linux

Logs:
- Added new feature on save
- Adding edited files to harpoon on save

.config/nvim/after/plugin/harpoon.lua
- Commit branch HASH [e4ade83876b1e966dfbfe939121e33ba896ec92f](Linux)

# 2024-07-08 13:30
## master
JIRA: #master

Logs:
- Updated precognition for liquibase scrip
- Commit for VIS-236

Notes.md
Stash_Notes.md
- Commit branch HASH [94ea59fa8de3098f80e97e412368390f12f615a0](master)

# 2024-07-08 14:18
## TMP
JIRA: #Linux

Logs:
- Updated code for start-up printing

.config/nvim/init.lua
- Commit branch HASH [d9e0a8d11bfd8bb3143174e8d062f3ee132cd07c](Linux)

# 2024-07-09 08:49
## TMP
JIRA: #Linux

Logs:
- Added formatting on selection

.config/nvim/lua/DarO/remap.lua
- Commit branch HASH [96da358eff04bcd3a94be173e0e28265cf82a190](Linux)

# 2024-07-10 07:28
## TMP
JIRA: #Linux

Logs:
- Updated snippets for markdown

.config/nvim/after/plugin/luasnip.lua
- Commit branch HASH [778ae741b25c0e329f5fdd542eba618352adaff9](Linux)

# 2024-07-10 08:04
## TMP
JIRA: #Linux

Logs:
- Adding nvchad init config

.config/nvchad/.stylua.toml
.config/nvchad/LICENSE
.config/nvchad/init.lua
.config/nvchad/lazy-lock.json
.config/nvchad/lua/chadrc.lua
.config/nvchad/lua/configs/conform.lua
.config/nvchad/lua/configs/lazy.lua
.config/nvchad/lua/configs/lspconfig.lua
.config/nvchad/lua/mappings.lua
.config/nvchad/lua/options.lua
.config/nvchad/lua/plugins/init.lua
- Commit branch HASH [919b55f828f1472cb7275c2a23c38201c236e757](Linux)

# 2024-07-10 08:05
## TMP
JIRA: #Linux

Logs:
- Updated nvims switcher script

.zshrc
- Commit branch HASH [07bd8389861fda5ab71672d01f1fc4131f89cce7](Linux)

# 2024-07-10 11:56
## master
JIRA: #master

Logs:
- Commits and updating notes

Notes.md
Stash_Notes.md
- Commit branch HASH [b27afcf544a947103e7a67bbe338abdcdb6f0112](master)

# 2024-07-10 15:53
## master
JIRA: #master

Logs:
- Cherry-pick for VIS-429

Notes.md
- Commit branch HASH [6bf6640bf1052daf0dd5d005c9890ff7180402d0](master)

# 2024-07-12 08:38
## master
JIRA: #master

Logs:
- Saved some sql scripts

Notes.md
Stash_Notes.md
- Commit branch HASH [6adaa72980d504ed62d85e465d8e4a8086ee99e0](master)

# 2024-07-12 22:46
## TMP
JIRA: #Linux

Logs:
- Restore write on delay for RFID testing

helper.py
- Commit branch HASH [6bd22b593428c482d481f3d92b7a0c393f0d3ebb](Linux)

# 2024-07-13 14:47
## TMP
JIRA: #Linux

Logs:
- Removing nvchad config
- Little changes in config

.config/nvchad/.stylua.toml
.config/nvchad/LICENSE
.config/nvchad/init.lua
.config/nvchad/lazy-lock.json
.config/nvchad/lua/chadrc.lua
.config/nvchad/lua/configs/conform.lua
.config/nvchad/lua/configs/lazy.lua
.config/nvchad/lua/configs/lspconfig.lua
.config/nvchad/lua/mappings.lua
.config/nvchad/lua/options.lua
.config/nvchad/lua/plugins/init.lua
.config/nvim/lua/DarO/init.lua
.config/nvim/lua/DarO/remap.lua
- Commit branch HASH [a8752202a8fd89c1bfc7126928ae358df5236f47](Linux)

# 2024-07-13 14:54
## TMP
JIRA: #Linux

Logs:
- Added new config (Still in development)
- Added GO to exports for .zshrc

.config/nvim-prime/.gitignore
.config/nvim-prime/README.md
.config/nvim-prime/init.lua
.config/nvim-prime/lazy-lock.json
.config/nvim-prime/lua/DarO/init.lua
.config/nvim-prime/lua/DarO/lazy/cloak.lua
.config/nvim-prime/lua/DarO/lazy/colors.lua
.config/nvim-prime/lua/DarO/lazy/comment.lua
.config/nvim-prime/lua/DarO/lazy/fugitive.lua
.config/nvim-prime/lua/DarO/lazy/gitsigns.lua
.config/nvim-prime/lua/DarO/lazy/harpoon.lua
.config/nvim-prime/lua/DarO/lazy/init.lua
.config/nvim-prime/lua/DarO/lazy/local.lua.bak
.config/nvim-prime/lua/DarO/lazy/lsp.lua
.config/nvim-prime/lua/DarO/lazy/neotest.lua.bak
.config/nvim-prime/lua/DarO/lazy/snippets.lua
.config/nvim-prime/lua/DarO/lazy/telescope.lua
.config/nvim-prime/lua/DarO/lazy/treesitter.lua
.config/nvim-prime/lua/DarO/lazy/trouble.lua
.config/nvim-prime/lua/DarO/lazy/undotree.lua
.config/nvim-prime/lua/DarO/lazy/vim-visual-multi.lua
.config/nvim-prime/lua/DarO/lazy/vimbegood.lua
.config/nvim-prime/lua/DarO/lazy/zenmode.lua
.config/nvim-prime/lua/DarO/lazy_init.lua
.config/nvim-prime/lua/DarO/remap.lua
.config/nvim-prime/lua/DarO/set.lua
.config/nvim-prime/scripts/tests/minimal.vim
.zshrc
- Commit branch HASH [10d67e97b5a42d6d3f8b0a48ff9e2a4057c1a8b2](Linux)

# 2024-07-13 15:00
## TMP
JIRA: #Linux

Logs:
- Removed useless files
- Formatting

.config/nvim-prime/README.md
.config/nvim-prime/init.lua
.config/nvim-prime/lua/DarO/init.lua
.config/nvim-prime/lua/DarO/lazy/cloak.lua
.config/nvim-prime/lua/DarO/lazy/colors.lua
.config/nvim-prime/lua/DarO/lazy/comment.lua
.config/nvim-prime/lua/DarO/lazy/fugitive.lua
.config/nvim-prime/lua/DarO/lazy/init.lua
.config/nvim-prime/lua/DarO/lazy/lsp.lua
.config/nvim-prime/lua/DarO/lazy/snippets.lua
.config/nvim-prime/lua/DarO/lazy/telescope.lua
.config/nvim-prime/lua/DarO/lazy/treesitter.lua
.config/nvim-prime/lua/DarO/lazy/trouble.lua
.config/nvim-prime/lua/DarO/lazy/undotree.lua
.config/nvim-prime/lua/DarO/lazy/vim-visual-multi.lua
.config/nvim-prime/lua/DarO/lazy/vimbegood.lua
.config/nvim-prime/lua/DarO/lazy/zenmode.lua
.config/nvim-prime/lua/DarO/lazy_init.lua
.config/nvim-prime/scripts/tests/minimal.vim
- Commit branch HASH [544495a5e6e3d44cb9d2d853f6cc7975c7723eb1](Linux)

# 2024-07-14 18:43
## TMP
JIRA: #Linux

Logs:
- Updated snippets
- Added file types to include for white space formatting

.config/nvim-prime/lua/DarO/init.lua
.config/nvim-prime/lua/DarO/lazy/snippets.lua
- Commit branch HASH [2dbd45359751b502deed82e0b71492794f73ad4d](Linux)

# 2024-07-14 21:48
## TMP
JIRA: #Linux

Logs:
- Updated harpoon scripts to match with harpoon2

.config/nvim-prime/lua/DarO/lazy/harpoon.lua
- Commit branch HASH [7c98245a14c749341981b76553fe93cd4adb4b0f](Linux)

# 2024-07-14 21:49
## TMP
JIRA: #Linux

Logs:
- Updated LSP config
- Auto install eslint and tsserver

.config/nvim-prime/lua/DarO/lazy/lsp.lua
- Commit branch HASH [40049e5ab5ae9b785baa1639c9265ea51d6aa33a](Linux)

# 2024-07-14 21:51
## TMP
JIRA: #Linux

Logs:
- Updated snippets
- Still doesn't fully work

.config/nvim-prime/lua/DarO/lazy/snippets.lua
- Commit branch HASH [9de70bd8bca773d5c22d9533569b44482b84893f](Linux)

# 2024-07-14 21:52
## TMP
JIRA: #Linux

Logs:
- lazy-lock update

.config/nvim-prime/lazy-lock.json
- Commit branch HASH [89c6e57231ac8dd7b870ab994fbe0174972ff8e6](Linux)

# 2024-07-14 22:38
## TMP
JIRA: #Linux

Logs:
- Fixed issues with completion
- Fixed issues with snippets

.config/nvim-prime/lua/DarO/lazy/snip-nvim-cmp.lua
.config/nvim-prime/lua/DarO/lazy/snippets.lua
- Commit branch HASH [62a63e77ada012dee2d374dea275f0b629ebc9e0](Linux)

# 2024-07-14 23:12
## TMP
JIRA: #Linux

Logs:
- Added .md file to white space formatting
- Fixed message printing on boot

.config/nvim-prime/init.lua
.config/nvim-prime/lua/DarO/init.lua
- Commit branch HASH [ae645bdad66c014909b3e36c45b00ba9c9e0a27a](Linux)

# 2024-07-15 08:02
## TMP
JIRA: #Linux

Logs:
- Little changes

.config/nvim-prime/lazy-lock.json
.config/nvim-prime/lua/DarO/init.lua
.config/nvim-prime/lua/DarO/lazy/init.lua
.config/nvim-prime/lua/DarO/lazy/snip-nvim-cmp.lua
.config/nvim-prime/lua/DarO/lazy/telescope.lua
.config/nvim-prime/lua/DarO/lazy/undotree.lua
.config/nvim-prime/lua/DarO/lazy/vim-visual-multi.lua
.config/nvim-prime/lua/DarO/lazy/vimbegood.lua
.config/nvim-prime/lua/DarO/lazy/zenmode.lua
- Commit branch HASH [79a1c4c71e5d1e0422469f61c9057d58a86a1b05](Linux)

# 2024-07-15 08:56
## master
JIRA: #master

Logs:
- Update notes with commit message

Notes.md
Stash_Notes.md
- Commit branch HASH [ac4d31a28d0da03930c2df377eedf8870694d6ee](master)

# 2024-07-15 10:41
## TMP
JIRA: #Linux

Logs:
- Added markdown-preview into the lazy config

.config/nvim-prime/lazy-lock.json
.config/nvim-prime/lua/DarO/lazy/markdown-preview.lua
- Commit branch HASH [9ee0696ca6164acd208196de9837a98b744998b0](Linux)

# 2024-07-16 21:05
## TMP
JIRA: #Linux

Logs:
- Updated snippets, missing semicolon
- Updated telescope search

.config/nvim-prime/lua/DarO/lazy/snippets.lua
.config/nvim-prime/lua/DarO/lazy/telescope.lua
- Commit branch HASH [882f5785d7edf5d4bbd8f85f200359940d1e828b](Linux)

# 2024-07-16 21:24
## master
JIRA: #master

Logs:
- Update notes about the Apache CSP location

Notes.md
- Commit branch HASH [976bcd03cd4720a9145d4e71ea6590bab809b702](master)

# 2024-07-16 22:49
## TMP
JIRA: #Linux

Logs:
- Multiple features update
- Added autopairs
- Added highlight-color
- Updated remaps
- Added utils

.config/nvim-prime/lazy-lock.json
.config/nvim-prime/lua/DarO/autocmds.lua
.config/nvim-prime/lua/DarO/init.lua
.config/nvim-prime/lua/DarO/lazy/autopairs.lua
.config/nvim-prime/lua/DarO/lazy/comment.lua
.config/nvim-prime/lua/DarO/lazy/dressing.lua
.config/nvim-prime/lua/DarO/lazy/highlight-color.lua
.config/nvim-prime/lua/DarO/lazy/snip-nvim-cmp.lua
.config/nvim-prime/lua/DarO/lazy/telescope.lua
.config/nvim-prime/lua/DarO/lazy/telescope.lua.bak
.config/nvim-prime/lua/DarO/remap.lua
.config/nvim-prime/lua/DarO/set.lua
.config/nvim-prime/lua/DarO/utils.lua
- Commit branch HASH [c992e5a5b984ecca7bb50d0ba21c8286e7d89eb7](Linux)

# 2024-07-16 23:22
## TMP
JIRA: #Linux

Logs:
- Little changes in telescope and remap

.config/nvim-prime/lazy-lock.json
.config/nvim-prime/lua/DarO/lazy/telescope.lua
.config/nvim-prime/lua/DarO/remap.lua
- Commit branch HASH [34a6ded9d551a81769a4a217fd41f11a7a224609](Linux)

# 2024-07-24 15:55
## master
JIRA: #master

Logs:
- Notes update and commits

Notes.md
Stash_Notes.md
- Commit branch HASH [471ec3c831080423c5906593b8fdfa390669d06c](master)

# 2024-07-28 10:20
## TMP
JIRA: #Linux

Logs:
- Changes in telescope

.config/nvim-prime/lazy-lock.json
.config/nvim-prime/lua/DarO/lazy/telescope.lua
- Commit branch HASH [de87a90899696b6131b4c63c275ae29757392048](Linux)

# 2024-07-28 14:29
## master
JIRA: #master

Logs:
-

Notes.md
Stash_Notes.md
- Commit branch HASH [24f78d1023910cf8f0e8c9387b9a771d334dfbc0](master)

# 2024-07-28 14:31
## master
JIRA: #master

Logs:
- Added file for one issue SUITE-9274

SUITE-9274.txt
- Commit branch HASH [540da32c984e3d8085934831808488a807f2c4b7](master)

# 2024-07-28 14:32
## TMP
JIRA: #Linux

Logs:
- Added new Notes and stashed list

Notes.md
Stash_list.md
- Commit branch HASH [48be239b3881150bbb0fd29f37d9f463f7454b70](Linux)

# 2024-07-28 14:33
## TMP
JIRA: #Linux

Logs:
- Note update

Notes.md
Stash_list.md
- Commit branch HASH [8c77ff916953103a9049f573daf02912cc4197f0](Linux)

# 2024-08-01 16:00
## master
JIRA: #master

Logs:
- Updates in notes

Notes.md
Stash_Notes.md
- Commit branch HASH [52b12e35c809ca94483243847a19915c6cb0d653](master)

# 2024-08-01 21:13
## master
JIRA: #master

Logs:
- added custom_default_data.sql

custom_default_data.sql
- Commit branch HASH [b2cb10bbdadbf7a6d740023e4830123c83a4549b](master)

# 2024-08-02 11:05
## master
JIRA: #master

Logs:
- Added Docker image

Docker/coreaccessbase.tar
Docker/custom_default_data.sql
Docker/oracle-instantclient-basic-23.5.0.24.07-1.el9.x86_64.rpm
Notes.md
Stash_Notes.md
- Commit branch HASH [9c366e12189a47c85458bff3e73189b0b21dd994](master)
