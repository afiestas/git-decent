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
		fName := filepath.Join(os.TempDir(), "git-decent-hook-lock-file")
		f, err := os.OpenFile(fName, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)

		if err != nil {
			fmt.Println(errorStyle.Styled("❌ couldn't create lock file"))
			printError(err)
			return
		}
		err = f.Close()
		if err != nil {
			fmt.Println(errorStyle.Styled("❌ couldn't close lock file"))
			printError(err)
			return
		}

		cleanup := func() {
			if f != nil {
				fmt.Println("Cleaning up")
				os.Remove(f.Name())
			}
		}
		fmt.Println("Lock file", f.Name())
		ctx := context.WithValue(cmd.Context(), lockFileKey, f.Name())
		cmd.SetContext(ctx)

		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovered from panic: %v\n", r)
				cleanup()
				os.Exit(1)
			}
		}()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGABRT)
		go func() {
			sig := <-sigCh
			fmt.Printf("Received signal: %s, cleaning up...\n", sig)
			cleanup()
			os.Exit(1)
		}()

		err = commandPreRun(cmd, args)
		if err != nil {
			printError(err)
			return
		}

		decentContext, ok := cmd.Context().Value(decentContextKey).(*DecentContext)
		if !ok {
			return
		}
		r := decentContext.gitRepo
		log, err := r.LogWithRevision("-1")
		if err != nil {
			fmt.Println(errorStyle.Styled("❌ couldn't get log from repo"))
			printError(err)
			return
		}

		if len(log) == 0 {
			fmt.Println(errorStyle.Styled("❌ git log seems to be empty"))
			return
		}

		ops, _ := r.GetSectionOptions("decent")
		s, err := config.NewScheduleFromMap(ops)
		if err != nil {
			printError(err)
			return
		}

		fmt.Println(infoStyle.Styled("Schedule:"))
		printSchedule(s)
		fmt.Println()

		commit := log[0]
		commitDate := commit.Date
		amended := internal.Amend(commit.Date, nil, s)
		sameDay := amended.Day() == commit.Date.Day()
		sameTime := amended.Minute() == commitDate.Minute() && amended.Hour() == commitDate.Hour()

		fmt.Println("✨", commit.Message)
		day := commitDate.Format("Mon")
		if !sameDay {
			day = accentStyle.Styled(day)
		}
		timeStr := commitDate.Format("15:04")
		if !sameTime {
			timeStr = secondaryStyle.Styled(timeStr)
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
				day = accentStyle.Styled(day)
			}
			time := amended.Format("15:04")
			if !sameTime {
				time = secondaryStyle.Styled(time)
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
			fmt.Println(errorStyle.Styled("❌ error while amending the date"))
			printError(err)
			return
		}
	},
}
