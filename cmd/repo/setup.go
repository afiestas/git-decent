package repo

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/afiestas/git-decent/config"
	"github.com/afiestas/git-decent/internal"
	"github.com/afiestas/git-decent/ui"
)

//go:embed config-template.ini
var configTemplate string
var Ui ui.UserInterface

func Setup(ui ui.UserInterface) (*internal.GitRepo, *config.Schedule, error) {
	Ui = ui
	repo, err := getRepo()
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't setup the repository %w", err)
	}

	_, err = repo.LogWithRevision("-1")
	if err != nil {
		return nil, nil, fmt.Errorf("coudlnt' get the git log %w", err)
	}

	schedule, err := getSchedule(repo)
	if err != nil {
		return nil, nil, err
	}

	return repo, schedule, nil
}

func TearDown() {

}

func getSchedule(r *internal.GitRepo) (*config.Schedule, error) {
	ops, _ := r.GetSectionOptions("decent")

	if len(ops) == 0 {
		asnwer, err := Ui.YesNoQuestion("Git decent is not configured, do you want to do it now?")
		if err != nil {
			return nil, err
		}
		if !asnwer {
			return nil, nil
		}

		initConfiguration(r)
	}

	r.GetSectionOptions("decent")
	s, err := config.NewScheduleFromMap(ops)
	if err != nil {
		return nil, err
	}
	return &s, nil
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
		answer, err := Ui.YesNoQuestion("Do you want to edit it again?")
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
		return nil, fmt.Errorf("couldn't getRepo %w", err)
	}

	r, err := internal.NewGitRepo(cwd, Ui.IsVerbose())
	if err != nil {
		return nil, fmt.Errorf("couldn't open the repository: %w", err)
	}

	if !r.IsGitRepo() {
		err = fmt.Errorf(fmt.Sprintf("the directory %s is not a git repository", cwd))
		return nil, err
	}

	return r, nil
}
