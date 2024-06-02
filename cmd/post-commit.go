package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/afiestas/git-decent/ui"
	"github.com/afiestas/git-decent/utils"
	u "github.com/afiestas/git-decent/utils"
	"github.com/spf13/cobra"
)

//go:embed post-commit-template.sh
var postCommitTpl []byte

type lockFile string

const lockFileKey lockFile = "lockFile"

var postCommitCmd = &cobra.Command{
	Use:   "post-commit",
	Short: "To be used by the hook",
	Long: `The post commit hook was not designed to amend the last commit
but instead to notify third party systems that a commit has been made.
One of the nasty side effects of having a post-commit edit the commit is
that the system might get into an infinity loop state since amending a commit will
also issue a post-commit.
This hook command adds a file semaphore to prevent the infinite loop from happening.`,
	Hidden:        true,
	SilenceErrors: true,
	SilenceUsage:  true,

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

		err = r.AmendDate(commit)
		if err != nil {
			return utils.WrapE("error while amending the date", err)
		}

		return nil
	},
}

var installPostCommit = &cobra.Command{
	Use:   "install",
	Short: "Installs the post-commit hook",
	Long: `This command will try to install the post-commit hook,
	If a hook already exists, manual edit must be done`,

	RunE: func(cmd *cobra.Command, args []string) error {
		decentContext, ok := cmd.Context().Value(decentContextKey).(*DecentContext)
		if !ok {
			return fmt.Errorf("could not get the context")
		}
		repo := decentContext.gitRepo

		ui.BlinkingTitle("\n⚠️⚠️⚠️ BE AWARE ⚠️⚠️⚠️")
		ui.PrintTemplate(`{{W "You are about to install a"}} {{Bold ( W "post-commit hook")}} {{W "that will"}}`)
		ui.PrintTemplate(`{{W "automatically amend your commit to a decent time if needed.\n"}}`)
		ui.PrintTemplate(`{{W "Be aware that this hook is"}} {{Bold (W "NOT MEANT")}} {{W "to amend the commit"}}`)
		ui.Warning("so using this hook is out of spec, use it at your own risk.")

		hookPath := filepath.Join(repo.Dir, ".git/hooks/post-commit")
		if _, err := os.Stat(hookPath); err == nil {
			err := askIfInstall("post-commit", hookPath, repo)
			if err != nil {
				return u.WrapE("Could not edit the hook", err)
			}
			return nil
		}
		ui.YesNoQuestion("\nDo you want to install the hook?")

		err := installHook(hookPath, postCommitTpl)
		if err != nil {
			return u.WrapE("could not install the hook", err)
		}

		ui.Success("Hook installed")
		return nil
	},
}
