package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/afiestas/git-decent/config"
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

		if state := r.State(); state != internal.Clean {
			fmt.Println(ui.ErrorStyle.Styled(fmt.Sprintf("❌ can't operate while %s is in progress", state)))
			return
		}

		branch := r.CurrentBranch()
		if branch == "HEAD" {
			fmt.Println(ui.ErrorStyle.Styled("❌ can't operate in detached head"))
			return
		}

		log, err := r.LogWithRevision("-2")
		if err != nil {
			fmt.Println(ui.ErrorStyle.Styled("❌ couldn't get log from repo"))
			Ui.PrintError(err)
			return
		}

		if len(log) == 0 {
			fmt.Println(ui.ErrorStyle.Styled("❌ git log seems to be empty"))
			return
		}

		ops, _ := r.GetSectionOptions("decent")
		s, err := config.NewScheduleFromMap(ops)
		if err != nil {
			Ui.PrintError(err)
			return
		}

		fmt.Println(ui.InfoStyle.Styled("Schedule:"))
		Ui.PrintSchedule(s)
		fmt.Println()

		var lastDate *time.Time = nil
		if len(log) > 1 {
			lastDate = &log[0].Date
		}
		commit := log[1]
		commitDate := commit.Date
		amended := internal.Amend(commit.Date, lastDate, s)
		sameDay := amended.Day() == commit.Date.Day()
		sameTime := amended.Minute() == commitDate.Minute() && amended.Hour() == commitDate.Hour()

		fmt.Println("✨", commit.Message)
		day := commitDate.Format("Mon")
		if !sameDay {
			day = ui.AccentStyle.Styled(day)
		}
		timeStr := commitDate.Format("15:04")
		if !sameTime {
			timeStr = ui.SecondaryStyle.Styled(timeStr)
		}

		fmt.Printf(
			"    %s %s %s ",
			commitDate.Format(time.DateOnly),
			day,
			timeStr,
		)
		if amended == commit.Date {
			fmt.Printf("✅")
		} else {
			day := amended.Format("Mon")
			if !sameDay {
				day = ui.AccentStyle.Styled(day)
			}
			time := amended.Format("15:04")
			if !sameTime {
				time = ui.SecondaryStyle.Styled(time)
			}
			fmt.Printf("➡️ %s %s",
				day,
				time,
			)
		}
		fmt.Println()

		commit.Date = amended

		err = r.AmendDate(commit)
		if err != nil {
			fmt.Println(ui.ErrorStyle.Styled("❌ error while amending the date"))
			Ui.PrintError(err)
			return
		}
	},
}
