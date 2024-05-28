/* SPDX-License-Identifier: MIT */
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/afiestas/git-decent/config"
	"github.com/afiestas/git-decent/internal"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "git-decent",
	Short: "Ammends your commits so it looks like you are behaving",
	Long: `Git-Decent is a small tool designed to help you, night owls,
maintain appearances while working during unconventional hours...`,

	PersistentPreRunE: commandPreRun,
	PersistentPostRun: commandPostRun,

	Run: func(cmd *cobra.Command, args []string) {
		decentContext, ok := cmd.Context().Value(decentContextKey).(*DecentContext)
		if !ok {
			return
		}
		r := decentContext.gitRepo

		ops, _ := r.GetSectionOptions("decent")
		s, err := config.NewScheduleFromMap(ops)
		if err != nil {
			printError(err)
			return
		}

		fmt.Println(infoStyle.Styled("Schedule:"))
		printSchedule(s)
		fmt.Println()

		fmt.Println(infoStyle.Styled("Current status"))
		upstream := r.BranchUpstream(r.CurrentBranch())
		fmt.Println("Upstream branch", secondaryStyle.Styled(upstream))

		aLog := fmt.Sprintf("%s...", upstream)
		log, err := r.LogWithRevision(aLog)
		if err != nil {
			printError(err)
			return
		}

		fmt.Println("Unpushed commits:", secondaryStyle.Styled(fmt.Sprintf("%d", len(log))))
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
				amendedCount += 1
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
			log[k] = commit
		}

		fmt.Println("Amended commits:", secondaryStyle.Styled(fmt.Sprintf("%d", amendedCount)))
		if amendedCount == 0 {
			return
		}

		answer, err := yesNoQuestion("Do you want to ament the dates?")
		if err != nil {
			printError(err)
			return
		}

		if !answer {
			return
		}

		err = r.AmendDates(log)
		if err != nil {
			fmt.Println("❌", errorStyle.Styled("Error amending the dates"))
			printError(err)
		}
	},
}

func Execute() {
	rootCmd.AddCommand(hookCmd)
	rootCmd.AddCommand(installCdm)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
