#!/bin/bash

swayidle -w \
   timeout 3 "swaylock -i /usr/share/backgrounds/default-dark.png -f" \
   timeout 6 'swaymsg "output * power off"' \
   resume 'swaymsg "output * power on"; swaybg -i /usr/share/backgrounds/default-dark.png -m fill' \
   before-sleep "swaylock -i /usr/share/backgrounds/default-dark.png -f"
