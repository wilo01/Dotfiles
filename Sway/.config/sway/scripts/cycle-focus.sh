#!/bin/bash
# Cycle focus within workspace with wrap-around
# Usage: cycle-focus.sh [right|left]
DIR="${1:-right}"
swaymsg "focus_wrapping workspace; focus $DIR; focus_wrapping yes"
