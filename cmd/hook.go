package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/afiestas/git-decent/internal"
	"github.com/afiestas/git-decent/ui"
	"github.com/spf13/cobra"
)

type lockFile string

const lockFileKey lockFile = "lockFile"

var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "To be used by the hook",
	Long: `The post commit hook was not designed to amend the last commit
but instead to notify third party systems that a commit has been made.
One of the nasty side effects of having a post-commit edit the commit is
that the system might get into an infinity loop state since amending a commit will
also issue a post-commit.
This hook command adds a file semaphore to prevent the infinite loop from happening.`,
	Hidden: true,

	PostRun: func(cmd *cobra.Command, args []string) {
		if fileLock, ok := cmd.Context().Value(lockFileKey).(string); ok {
			os.Remove(fileLock)
		}
	},

	Run: func(cmd *cobra.Command, args []string) {
		decentContext, ok := cmd.Context().Value(decentContextKey).(*DecentContext)
		if !ok {
			return
		}
		r := decentContext.gitRepo
		s := *decentContext.schedule

		log, err := r.LogWithRevision("-2")
		if err != nil {
			fmt.Println(ui.ErrorStyle.Styled("❌ couldn't get log from repo"))
			ui.PrintError(err)
			return
		}

		if len(log) == 0 {
			fmt.Println(ui.ErrorStyle.Styled("❌ git log seems to be empty"))
			return
		}

		fmt.Println(ui.InfoStyle.Styled("Schedule:"))
		ui.PrintSchedule(s)
		fmt.Println()

		var lastDate *time.Time = nil
		if len(log) > 1 {
			lastDate = &log[0].Date
		}
		commit := log[1]
		amended := internal.Amend(commit.Date, lastDate, s)
		ui.PrintAmend(commit.Date, amended, commit.Message)

		if commit.Date == amended {
			return
		}
		commit.Date = amended

		err = r.AmendDate(commit)
		if err != nil {
			fmt.Println(ui.ErrorStyle.Styled("❌ error while amending the date"))
			ui.PrintError(err)
			return
		}
	},
}
