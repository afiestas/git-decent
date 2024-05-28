package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var installCdm = &cobra.Command{
	Use:   "install",
	Short: "Installs git-hooks to make things automagic",
	Long: `This command will optionally install two hooks,
	one post-commit to amend the date and one pre-push to prevent pushes on
	undecent times`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Installing shit!")
	},
}
