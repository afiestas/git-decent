package cmd

import (
	"fmt"
	"time"

	"github.com/afiestas/git-decent/config"
	"github.com/afiestas/git-decent/internal"
	"github.com/afiestas/git-decent/ui"
	"github.com/afiestas/git-decent/utils"
	u "github.com/afiestas/git-decent/utils"
	"github.com/spf13/cobra"
)

var amendCmd = &cobra.Command{
	Use:   "amend",
	Short: "Amends the last commit to a decent date if needed",
	Long: `It will amend the last commit (no matter if it is pushed or not)
if the date is not decent.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		decentContext, ok := cmd.Context().Value(decentContextKey).(*DecentContext)
		if !ok {
			return fmt.Errorf("couldn't obtian repo from context")
		}

		r := decentContext.gitRepo
		s := *decentContext.schedule

		commit, err := amend(r, &s)
		if err != nil {
			return err
		}

		if commit == nil {
			return nil
		}

		asnwer, err := ui.YesNoQuestion("Are you sure you want to amend the commit?")
		if err != nil {
			return err
		}

		if !asnwer {
			return nil
		}

		err = r.AmendDate(commit)
		if err != nil {
			return utils.WrapE("error while amending the date", err)
		}

		return nil
	},
}

func amend(repo *internal.GitRepo, schedule *config.Schedule) (*internal.Commit, error) {
	log, err := repo.LogWithRevision("-2")
	if err != nil {
		return nil, u.WrapE("couldn't get log from repo", err)
	}

	if len(log) == 0 {
		return nil, fmt.Errorf("git log seems to be empty")
	}

	fmt.Println(ui.InfoStyle.Styled("Schedule:"))
	ui.PrintSchedule(*schedule)
	fmt.Println()

	var lastRealDate *time.Time = nil
	var lastDate *time.Time = nil
	if len(log) > 1 {
		lastDate = &log[0].Date
		//TODO: Change this to fetch the date from the saved file
		lastRealDate = lastDate
	}

	commit := log[1]
	amended := internal.Amend(commit.Date, lastDate, lastRealDate, 0, *schedule)
	ui.PrintAmend(commit.Date, amended, commit.Message)

	if commit.Date == amended {
		return nil, nil
	}
	commit.Date = amended

	return commit, nil
}
