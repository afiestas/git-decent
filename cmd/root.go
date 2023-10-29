/* SPDX-License-Identifier: MIT */
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "git-decent",
	Short: "Ammends your commits so it looks like you are behaving",
	Long: `Git-Decent is a small tool designed to help you, night owls,
maintain appearances while working during unconventional hours...`,

	Run: func(cmd *cobra.Command, args []string) {

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
