local wezterm = require 'wezterm'
local config = wezterm.config_builder()

config.color_scheme = 'Tokyo Night'

config.font = wezterm.font 'MesloLGS NF'
config.font_size = 12

config.window_padding = {
   left = 0,
   right = 0,
   top = 0,
   bottom = 0,
}

config.term = 'screen-256color'
config.default_prog = { "tmux" }
config.hide_tab_bar_if_only_one_tab = true
config.tab_bar_at_bottom = true
config.use_ime = false

return config
