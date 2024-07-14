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

var rootCmd = &cobra.Command{
	Use:   "git-decent",
	Short: "Ammends your commits so it looks like you are behaving",
	Long: `Git-Decent is a small tool designed to help you, night owls,
maintain appearances while working during unconventional hours...`,

	PersistentPreRunE: commandPreRun,

	SilenceUsage:  true,
	SilenceErrors: true,

	Run: func(cmd *cobra.Command, args []string) {
		decentContext, ok := cmd.Context().Value(decentContextKey).(*DecentContext)
		if !ok {
			return
		}
		r := decentContext.gitRepo
		s := *decentContext.schedule

		ui.Title("Schedule:")
		ui.PrintSchedule(s)
		fmt.Println()

		ui.Title("Current status")
		upstream := r.BranchUpstream(r.CurrentBranch())
		ui.Info("Upstream branch", upstream)

		aLog := fmt.Sprintf("%s...", upstream)
		log, err := r.LogWithRevision(aLog)
		if err != nil {
			ui.PrintError(err)
			return
		}

		ui.Info("Unpushed commits:", fmt.Sprintf("%d", len(log)))
		if len(log) == 0 {
			return
		}

		amendedCount := 0
		var lastRealDate *time.Time = nil
		var lastDate *time.Time = nil
		for k, commit := range log {
			if commit.Prev != nil {
				commit.Prev = log[k-1]
				lastDate = &commit.Prev.Date
			}
			commitDate := commit.Date
			amended := internal.Amend(commitDate, lastDate, lastRealDate, 0, s)
			lastRealDate = &commit.Date

			ui.PrintAmend(commitDate, amended, commit.Message)
			if amended != commitDate {
				amendedCount += 1
			}

			commit.Date = amended
			log[k] = commit
		}

		ui.Info("Amended commits:", fmt.Sprintf("%d", amendedCount))
		if amendedCount == 0 {
			return
		}

		answer, err := ui.YesNoQuestion("Do you want to ament the dates?")
		if err != nil {
			ui.PrintError(err)
			return
		}

		if !answer {
			return
		}

		err = r.AmendDates(log)
		if err != nil {
			fmt.Println("❌", ui.ErrorStyle.Styled("Error amending the dates"))
			ui.PrintError(err)
		}
	},
}

func Execute() {
	postCommitCmd.AddCommand(installPostCommit)
	rootCmd.AddCommand(postCommitCmd)
	prePushCmd.AddCommand(installPrePush)
	rootCmd.AddCommand(prePushCmd)
	rootCmd.AddCommand(amendCmd)
	rootCmd.AddCommand(installCdm)
	rootCmd.AddCommand(configCmd)
	err := rootCmd.Execute()
	commandPostRun()

	if err != nil {
		ui.PrintError(err)
		os.Exit(1)
	}

}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
}
