package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/afiestas/git-decent/config"
	"github.com/afiestas/git-decent/internal"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

type contextKey string

const decentContextKey contextKey = "decentContext"

type DecentContext struct {
	restoreConsole func() error
	gitRepo        *internal.GitRepo
}

func commandPreRun(cmd *cobra.Command, args []string) error {
	//windows compatibility
	restoreConsole, err := termenv.EnableVirtualTerminalProcessing(termenv.DefaultOutput())
	if err != nil {
		panic(err)
	}

	repo, err := getRepoReady()
	if err != nil {
		printError(err)
		return err
	}
	if repo == nil {
		return fmt.Errorf("unable to configure the repository")
	}

	_, err = repo.LogWithRevision("-1")
	if err != nil {
		return fmt.Errorf(errorStyle.Styled("❌ No commits found"))
	}

	decentContext := &DecentContext{
		restoreConsole: restoreConsole,
		gitRepo:        repo,
	}

	ctx := context.WithValue(cmd.Context(), decentContextKey, decentContext)
	cmd.SetContext(ctx)

	return nil
}

func commandPostRun(cmd *cobra.Command, args []string) {
	cmd.Context()
	if decentContext, ok := cmd.Context().Value(decentContextKey).(*DecentContext); ok {
		defer decentContext.restoreConsole()
	}
}

func getRepoReady() (*internal.GitRepo, error) {
	r, err := getRepo()
	if err != nil {
		return nil, err
	}

	ops, _ := r.GetSectionOptions("decent")

	if len(ops) == 0 {
		asnwer, err := yesNoQuestion("Git decent is not configured, do you want to do it now?")
		if err != nil {
			printError(err)
			return nil, err
		}
		if !asnwer {
			return nil, nil
		}

		initConfiguration(r)
	}

	return r, nil
}

func initConfiguration(repo *internal.GitRepo) error {
	rawC, err := openGitEditor()
	if err != nil {
		return err
	}

	for x := time.Monday; x < time.Saturday; x++ {
		if len(rawC.Days[x]) == 0 {
			continue
		}
		err = repo.SetConfig("decent."+strings.Title(x.String()), rawC.Days[x])
		if err != nil {
			return err
		}
	}
	if len(rawC.Days[time.Sunday]) > 0 {
		err = repo.SetConfig("decent."+strings.Title(time.Sunday.String()), rawC.Days[time.Sunday])
		if err != nil {
			return err
		}
	}
	return nil
}

func openGitEditor() (*config.RawScheduleConfig, error) {
	gitcfg := exec.Command("git", "var", "GIT_EDITOR")
	editorName, err := gitcfg.Output()
	if err != nil {
		return nil, fmt.Errorf("openGitEditor coudn't fetch the GIT_EDITOR var")
	}

	if len(editorName) == 0 {
		return nil, fmt.Errorf("openGitEditor empty editor configured")
	}

	var args []string

	editorCmd := strings.Fields(string(editorName))
	cmd := editorCmd[0]
	if len(editorCmd) > 0 {
		args = editorCmd[1:]
	}

	f, err := os.CreateTemp(os.TempDir(), "schedule-tempalte")
	if err != nil {
		return nil, fmt.Errorf("openGitEditor can't create tmp file %w", err)
	}

	defer func() {
		f.Close()
		os.RemoveAll(f.Name())
	}()

	f.WriteString(configTemplate)
	f.Seek(0, 0)

	args = append(args, f.Name())
	for {
		editor := exec.Command(cmd, args...)
		editor.Stdin = os.Stdin
		editor.Stdout = os.Stdout
		editor.Stderr = os.Stderr

		err = editor.Run()
		if err != nil {
			return nil, err
		}

		rawC, err := config.NewScheduleFromPlainText(f)
		if err == nil {
			_, err := config.NewScheduleFromRaw(rawC)
			if err == nil {
				return rawC, nil
			}
		}

		fmt.Println("the configuration coudln't be parsed", err)
		answer, err := yesNoQuestion("Do you want to edit it again?")
		if err != nil {
			return nil, err
		}
		if !answer {
			return nil, nil
		}
	}
}

func getRepo() (*internal.GitRepo, error) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("❌", errorStyle.Styled("Couldn't get cwd"), err)
		return nil, err
	}
	r, err := internal.NewGitRepo(cwd)
	if err != nil {
		fmt.Println("❌", errorStyle.Styled("Couldn't open the repository"), err)
		return nil, err
	}
	if !r.IsGitRepo() {
		fmt.Println(errorStyle.Styled("❌ Not a git repository"))
		fmt.Println("The directory", secondaryStyle.Styled(cwd), "does not appear to be a git repo")
		return nil, err
	}

	return r, nil
}

func yesNoQuestion(question string) (bool, error) {
	fmt.Println(primaryStyle.Styled(question), primaryStyle.Bold().Styled("(Y/n)"))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	input = strings.TrimSpace(input)

	switch strings.ToLower(input) {
	case "n":
		return false, nil
	default:
		return true, err
	}
}

func printSchedule(schedule config.Schedule) {
	for x := time.Monday; x <= time.Saturday; x++ {
		s := schedule.Days[x].DecentFrames.String()
		if len(s) == 0 {
			s = "↪️ " + schedule.Days[x].ClosestDecentDay.String()
		}
		fmt.Printf("📅 %-10s %s\n", x.String()+":", s)
	}
	s := schedule.Days[0].DecentFrames.String()
	if len(s) == 0 {
		s = "↪️ " + schedule.Days[0].ClosestDecentDay.String()
	}
	fmt.Printf("📅 %-10s %s\n", time.Sunday.String()+":", s)

}

func printError(err error) {
	var commandError *internal.CommandError
	switch {
	case errors.As(err, &commandError):
		fmt.Println("   ", secondaryStyle.Bold().Styled("Command:"), primaryStyle.Styled(commandError.Command))
		if len(commandError.Stdout) > 0 {
			fmt.Println("   ", secondaryStyle.Bold().Styled("Stdout:"), primaryStyle.Styled(commandError.Stdout))
		}
		if len(commandError.Stderr) > 0 {
			fmt.Println("   ", secondaryStyle.Bold().Styled("Stderr:"), primaryStyle.Styled(commandError.Stderr))
		}
	default:
		fmt.Println("   ", primaryStyle.Styled(err.Error()))
	}

}
