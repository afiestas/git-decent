/* SPDX-License-Identifier: MIT */
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/afiestas/git-decent/internal"
	"github.com/afiestas/git-decent/ui"
	"github.com/spf13/cobra"
)

var Ui ui.UserInterface

var rootCmd = &cobra.Command{
	Use:   "git-decent",
	Short: "Ammends your commits so it looks like you are behaving",
	Long: `Git-Decent is a small tool designed to help you, night owls,
maintain appearances while working during unconventional hours...`,

	PersistentPreRunE: commandPreRun,
	PersistentPostRun: commandPostRun,

	SilenceUsage: true,

	Run: func(cmd *cobra.Command, args []string) {
		decentContext, ok := cmd.Context().Value(decentContextKey).(*DecentContext)
		if !ok {
			return
		}
		r := decentContext.gitRepo
		s := *decentContext.schedule

		Ui.Title("Schedule:")
		Ui.PrintSchedule(s)
		fmt.Println()

		Ui.Title("Current status")
		upstream := r.BranchUpstream(r.CurrentBranch())
		Ui.Info("Upstream branch", upstream)

		aLog := fmt.Sprintf("%s...", upstream)
		log, err := r.LogWithRevision(aLog)
		if err != nil {
			Ui.PrintError(err)
			return
		}

		Ui.Info("Unpushed commits:", fmt.Sprintf("%d", len(log)))
		if len(log) == 0 {
			return
		}

		amendedCount := 0
		var lastDate *time.Time = nil
		for k, commit := range log {
			if commit.Prev != nil {
				commit.Prev = log[k-1]
				lastDate = &commit.Prev.Date
			}
			commitDate := commit.Date
			amended := internal.Amend(commitDate, lastDate, s)
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
				amendedCount += 1
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
			log[k] = commit
		}

		Ui.Info("Amended commits:", fmt.Sprintf("%d", amendedCount))
		if amendedCount == 0 {
			return
		}

		answer, err := Ui.YesNoQuestion("Do you want to ament the dates?")
		if err != nil {
			Ui.PrintError(err)
			return
		}

		if !answer {
			return
		}

		err = r.AmendDates(log)
		if err != nil {
			fmt.Println("❌", ui.ErrorStyle.Styled("Error amending the dates"))
			Ui.PrintError(err)
		}
	},
}

func Execute() {
	rootCmd.AddCommand(hookCmd)
	rootCmd.AddCommand(installCdm)
	err := rootCmd.Execute()
	if err != nil {
		Ui.PrintError(err)
		os.Exit(1)
	}

}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
}
