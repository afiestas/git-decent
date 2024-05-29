package cmd

import (
	"context"
	"fmt"

	"github.com/afiestas/git-decent/cmd/repo"
	"github.com/afiestas/git-decent/cmd/security"
	"github.com/afiestas/git-decent/config"
	"github.com/afiestas/git-decent/internal"
	"github.com/afiestas/git-decent/ui"
	"github.com/spf13/cobra"
)

type contextKey string

const decentContextKey contextKey = "decentContext"

type DecentContext struct {
	gitRepo  *internal.GitRepo
	schedule *config.Schedule
}

func commandPreRun(cmd *cobra.Command, args []string) error {
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("error getting the verbose flag %w", err)
	}
	Ui.SetVerbose(verbose)

	err = security.Setup(Ui)

	if err != nil {
		//TODO: Handle somehow not to show the recurssion prevention
		return err
	}

	err = ui.Setup()
	if err != nil {
		return fmt.Errorf("couldn't setup the ui %w", err)
	}

	repo, schedule, err := repo.Setup(Ui)
	if err != nil {
		Ui.PrintError(err)
		return err
	}

	decentContext := &DecentContext{
		gitRepo:  repo,
		schedule: schedule,
	}

	ctx := context.WithValue(cmd.Context(), decentContextKey, decentContext)
	cmd.SetContext(ctx)

	return nil
}

func commandPostRun(cmd *cobra.Command, args []string) {
	ui.TearDown()
	repo.TearDown()
	security.TearDown()
}
