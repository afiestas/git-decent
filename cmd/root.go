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

	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println("Couldn't get cwd", err)
		}
		r, err := internal.NewGitRepo(cwd)
		if err != nil {
			fmt.Println("Couldn't get cwd", err)
		}

		fmt.Println("Gettign diff")
		upstream := r.BranchUpstream(r.CurrentBranch())
		fmt.Println("Upstream ranch", upstream)
		aLog := fmt.Sprintf("%s...", upstream)
		log, err := r.LogWithRevision(aLog)
		if err != nil {
			fmt.Println("AAAA", err)
		}

		fmt.Println("Unpushed commits:", len(log))

		ops, _ := r.GetSectionOptions("decent")
		rc, _ := config.GetGitRawConfig(&ops)
		s, _ := config.NewScheduleFromRaw(&rc)
		fmt.Println(s)

		var lastDate *time.Time = nil
		for k, commit := range log {
			if commit.Prev != nil {
				commit.Prev = log[k-1]
				lastDate = &commit.Prev.Date
			}

			amended := internal.Amend(commit.Date, lastDate, s)
			fmt.Println("Commit", commit.Message)
			fmt.Println("\tOriginal", commit.Date.Format(time.DateTime))
			fmt.Println("\tAmended", amended.Format(time.DateTime))
			commit.Date = amended
			log[k] = commit
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
