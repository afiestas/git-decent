package cmd

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/afiestas/git-decent/internal"
	"github.com/afiestas/git-decent/ui"
	u "github.com/afiestas/git-decent/utils"
	"github.com/spf13/cobra"
)

//go:embed pre-push-template.sh
var preCommitTpl []byte

// Check uinpushed commits for commits in the future
var prePushCmd = &cobra.Command{
	Use:    "pre-push",
	Short:  "Prevents pushign at undecent hours",
	Hidden: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		decentContext, ok := cmd.Context().Value(decentContextKey).(*DecentContext)
		if !ok {
			return fmt.Errorf("could not get context")
		}

		r := decentContext.gitRepo
		upstream := r.BranchUpstream(r.CurrentBranch())
		ui.Info("Upstream branch", upstream)

		aLog := fmt.Sprintf("%s...", upstream)
		log, err := r.LogWithRevision(aLog)
		if err != nil {
			return u.WrapE("Unable to get the log", err)
		}

		ui.Info("Unpushed commits:", fmt.Sprintf("%d", len(log)))
		if len(log) > 0 {
			commits := containsCommitInFuture(log)
			for _, commit := range commits {
				fmt.Println("Commit is in the future", commit.Message, commit.Date)
			}
		}

		s := decentContext.schedule
		now := time.Now()
		_, dMin := s.ClosestDecentMinute(now)
		if dMin == 0 {
			ui.Success("Allowed to push, decent time")
			return nil
		}

		current := now.Format("Mon 15:04")
		ui.PrintTemplate(fmt.Sprintf(`{{Bold (W "%s")}} {{W "is not a decent time."}}`, current))
		ui.PrintTemplate((`Use {{S "git push --no-verify"}} to skip the hook`))

		return errors.New("it is not a decent time")
	},
}

func containsCommitInFuture(log internal.GitLog) []internal.Commit {
	now := time.Now()
	commits := []internal.Commit{}
	for _, c := range log {
		if c.Date.After(now) {
			commits = append(commits, *c)
		}
	}

	return commits
}

var installPrePush = &cobra.Command{
	Use:   "install",
	Short: "Installs the pre-push hook",
	Long: `This hook prevents to push on undecent times. It is interactive
so yuo can always override and push`,

	RunE: func(cmd *cobra.Command, args []string) error {
		decentContext, ok := cmd.Context().Value(decentContextKey).(*DecentContext)
		if !ok {
			return fmt.Errorf("could not get the context")
		}
		repo := decentContext.gitRepo

		ui.Title("Install pre-push")
		hookPath := filepath.Join(repo.Dir, ".git/hooks/pre-push")
		if _, err := os.Stat(hookPath); err == nil {
			err := askIfInstall("pre-push", hookPath, repo)
			if err != nil {
				return u.WrapE("Could not edit the hook", err)
			}
			return nil
		}

		err := installHook(hookPath, preCommitTpl)
		if err != nil {
			return u.WrapE("could not install the hook", err)
		}

		ui.Success("Hook installed")
		return nil
	},
}
