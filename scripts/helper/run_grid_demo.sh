#!/bin/bash
# Run the Google Sheets Grid TUI in demo mode

echo "🚀 Launching Google Sheets Grid TUI in demo mode..."
echo "📝 This is a preview of the improved design"
echo ""
echo "Keyboard shortcuts:"
echo "  ← → : Navigate months"
echo "  ↑ ↓ : Move between cells"  
echo "  Enter : Edit cell"
echo "  S : Sync (demo)"
echo "  H : Help"
echo "  Q : Quit"
echo ""
echo "Press any key to start..."
read -n 1

# Run the TUI
cd /home/dariuszw/.Dotfiles/scripts/helper
python -c "
from src.helper_cli.tui.sheets_grid_app import GoogleSheetsGridApp
app = GoogleSheetsGridApp()
app.run()
"