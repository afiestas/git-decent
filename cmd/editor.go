package cmd

import (
	"os"
	"os/exec"
	"strings"

	"github.com/afiestas/git-decent/internal"
)

func openEditor(path string, repo *internal.GitRepo) error {
	editorName, err := repo.GetVar("GIT_EDITOR")
	if err != nil {
		return err
	}

	var args []string

	editorCmd := strings.Fields(string(editorName))
	cmd := editorCmd[0]
	if len(editorCmd) > 0 {
		args = editorCmd[1:]
	}
	args = append(args, path)

	editor := exec.Command(cmd, args...)
	editor.Stdin = os.Stdin
	editor.Stdout = os.Stdout
	editor.Stderr = os.Stderr

	err = editor.Run()
	if err != nil {
		return err
	}

	return nil
}
