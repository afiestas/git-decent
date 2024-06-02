package cmd

import (
	"fmt"
	"os"

	"github.com/afiestas/git-decent/internal"
	"github.com/afiestas/git-decent/ui"
	u "github.com/afiestas/git-decent/utils"
	"github.com/spf13/cobra"
)

var installCdm = &cobra.Command{
	Use:   "install",
	Short: "Installs git-hooks to make things automagic",
	Long: `This command will optionally install two hooks,
	one post-commit to amend the date and one pre-push to prevent pushes on
	undecent times`,

	RunE: func(cmd *cobra.Command, args []string) error {
		err := installPostCommit.RunE(cmd, args)
		if err != nil {
			return err
		}

		fmt.Println()
		err = installPrePush.RunE(cmd, args)
		return err
	},
}

func askIfInstall(hook string, hookPath string, repo *internal.GitRepo) error {
	command := fmt.Sprintf("git decent %s", hook)
	tpl := fmt.Sprintf(`{{P "\nA %s hook already exists, we"}} {{Bold "won't"}} {{P "edit it"}}`, hook)
	ui.PrintTemplate(tpl)
	ui.Print("instead you can add this line manually\n")
	ui.PrintTemplate(fmt.Sprintf(`> %s {{S "(Copied 📋)"}}`, command))
	ui.Copy(command)

	a, err := ui.YesNoQuestion("\nDo you want to manually edit the hook?")
	if err != nil {
		return err
	}

	if !a {
		return nil
	}

	err = openEditor(hookPath, repo)
	if err != nil {
		return err
	}

	return nil
}

func installHook(hookPath string, content []byte) error {
	f, err := os.Create(hookPath)
	defer func() {
		err := f.Close()
		if err != nil {
			ui.PrintError(err)
		}
	}()

	if err != nil {
		return u.WrapE("Couldn't create post-commit file", err)
	}
	_, err = f.Write(content)
	if err != nil {
		return u.WrapE("Couldn't create post-commit file", err)
	}

	fStat, err := f.Stat()
	if err != nil {
		return u.WrapE("Couldn't create post-commit file", err)
	}
	newMode := fStat.Mode() | 0100
	err = f.Chmod(newMode)
	if err != nil {
		return u.WrapE("Couldn't create post-commit file", err)
	}
	return nil
}
