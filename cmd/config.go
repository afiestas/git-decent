package cmd

import (
	"fmt"

	"github.com/afiestas/git-decent/ui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "It prints the current git decent config",
	RunE: func(cmd *cobra.Command, args []string) error {
		decentContext, ok := cmd.Context().Value(decentContextKey).(*DecentContext)
		if !ok {
			return fmt.Errorf("could not get context")
		}

		schedule := decentContext.schedule
		ui.Title("\nSchedule")
		ui.PrintSchedule(*schedule)

		return nil
	},
}
