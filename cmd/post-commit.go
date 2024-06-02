package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/afiestas/git-decent/internal"
	"github.com/afiestas/git-decent/ui"
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
	Hidden: true,

	PostRun: func(cmd *cobra.Command, args []string) {
		if fileLock, ok := cmd.Context().Value(lockFileKey).(string); ok {
			os.Remove(fileLock)
		}
	},

	Run: func(cmd *cobra.Command, args []string) {
		decentContext, ok := cmd.Context().Value(decentContextKey).(*DecentContext)
		if !ok {
			return
		}
		r := decentContext.gitRepo
		s := *decentContext.schedule

		log, err := r.LogWithRevision("-2")
		if err != nil {
			fmt.Println(ui.ErrorStyle.Styled("❌ couldn't get log from repo"))
			ui.PrintError(err)
			return
		}

		if len(log) == 0 {
			fmt.Println(ui.ErrorStyle.Styled("❌ git log seems to be empty"))
			return
		}

		fmt.Println(ui.InfoStyle.Styled("Schedule:"))
		ui.PrintSchedule(s)
		fmt.Println()

		var lastDate *time.Time = nil
		if len(log) > 1 {
			lastDate = &log[0].Date
		}
		commit := log[1]
		amended := internal.Amend(commit.Date, lastDate, s)
		ui.PrintAmend(commit.Date, amended, commit.Message)

		if commit.Date == amended {
			return
		}
		commit.Date = amended

		err = r.AmendDate(commit)
		if err != nil {
			fmt.Println(ui.ErrorStyle.Styled("❌ error while amending the date"))
			ui.PrintError(err)
			return
		}
	},
}

var installPostCommit = &cobra.Command{
	Use:   "install",
	Short: "Installs the post-commit hook",
	Long: `This command will try to install the post-commit hook,
	If a hook already exists, manual edit must be done`,

	Run: func(cmd *cobra.Command, args []string) {
		decentContext, ok := cmd.Context().Value(decentContextKey).(*DecentContext)
		if !ok {
			return
		}
		repo := decentContext.gitRepo

		ui.BlinkingTitle("\n⚠️⚠️⚠️ BE AWARE ⚠️⚠️⚠️")
		ui.PrintTemplate(`{{W "You are about to install a"}} {{Bold ( W "post-commit hook")}} {{W "that will"}}`)
		ui.PrintTemplate(`{{W "automatically amend your commit to a decent time if needed.\n"}}`)
		ui.PrintTemplate(`{{W "Be aware that this hook is"}} {{Bold (W "NOT MEANT")}} {{W "to amend the commit"}}`)
		ui.Warning("so using this hook is out of spec, use it at your own risk.")

		hookPath := filepath.Join(repo.Dir, ".git/hooks/post-commit")
		if _, err := os.Stat(hookPath); err == nil {
			ui.PrintTemplate(`{{P "\nA post-commit hook already exists, we"}} {{Bold "won't"}} {{P "edit it"}}`)
			ui.Print("instead you can add this line manually\n")
			ui.PrintTemplate(`> git decent post-commit {{S "(Copied 📋)"}}`)
			ui.Copy("git decent post-commit")
			a, err := ui.YesNoQuestion("\nDo you want to manually edit the hook?")
			if err != nil {
				ui.PrintError(err)
				return
			}

			if !a {
				return
			}
			err = openEditor(hookPath, repo)
			if err != nil {
				ui.PrintError(err)
				return
			}
			return
		}
		ui.YesNoQuestion("\nDo you want to install the hook?")

		f, err := os.Create(hookPath)
		defer func() {
			err := f.Close()
			if err != nil {
				ui.PrintError(err)
			}
		}()

		if err != nil {
			u.WrapE("Couldn't create post-commit file", err)
			return
		}
		_, err = f.Write(postCommitTpl)
		if err != nil {
			u.WrapE("Couldn't create post-commit file", err)
			return
		}

		fStat, err := f.Stat()
		if err != nil {
			u.WrapE("Couldn't create post-commit file", err)
			return
		}
		newMode := fStat.Mode() | 0100
		err = f.Chmod(newMode)
		if err != nil {
			u.WrapE("Couldn't create post-commit file", err)
			return
		}

		ui.Success("Hook installed")
	},
}
