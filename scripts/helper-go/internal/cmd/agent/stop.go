package agent

import (
	"fmt"
	"strings"

	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

var stopKill bool

var stopCmd = &cobra.Command{
	Use:   "stop <TICKET-KEY>",
	Short: "Stop a ticket agent (graceful interrupt; --kill ends the session)",
	Long: `Sends Escape + Ctrl-C to the agent's tmux pane. With --kill the whole
tmux session is terminated. Worktrees and branches are never removed.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.ToUpper(args[0])
		if !sessionExists(key) {
			return fmt.Errorf("no tmux session named %s", key)
		}
		if stopKill {
			if err := killSession(key); err != nil {
				return err
			}
			fmt.Println(ui.SuccessMsg("session " + key + " killed"))
			return nil
		}
		if err := interruptSession(key); err != nil {
			return err
		}
		fmt.Println(ui.SuccessMsg("interrupt sent to " + key + " (session kept alive)"))
		return nil
	},
}

func init() {
	stopCmd.Flags().BoolVar(&stopKill, "kill", false, "kill the tmux session instead of interrupting")
}
