/* SPDX-License-Identifier: MIT */
package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const g string = "git"

type GitRepo struct {
	Dir string
}

func NewGitRepo(dir string) (*GitRepo, error) {
	if i, err := os.Stat(dir); err != nil {
		return nil, fmt.Errorf("couldn't stat directory %w", err)
	} else if !i.IsDir() {
		return nil, fmt.Errorf("repository is not a dir %s", dir)
	}

	return &GitRepo{
		Dir: dir,
	}, nil
}

func (r *GitRepo) Command(arg ...string) *exec.Cmd {
	cmd := exec.Command(g, arg...)
	cmd.Dir = r.Dir

	return cmd
}

func (r *GitRepo) Init(bare bool) error {
	checkFile := ".git"

	args := [2]string{"init"}
	if bare {
		args[1] = "--bare"
		checkFile = "config"
	}

	gitDir := filepath.Join(r.Dir, checkFile)
	_, err := os.Stat(gitDir)
	if err == nil {
		return fmt.Errorf("repository already initialized in %s", r.Dir)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("couldn't stat the directory %w", err)
	}

	cmd := r.Command(args[:]...)
	output, err := cmd.Output()
	fmt.Println("OUTPUT", string(output))
	return err
}

func (r *GitRepo) IsBare() (bool, error) {
	cmd := r.Command("rev-parse", "--is-bare-repository")

	//TODO Caputre stderr&stdout for error reporting
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("couldn't execute git rev parse, output: %s, error: %w", output, err)
	}

	cleaned := strings.TrimSpace(string(output))
	return cleaned == "true", nil
}
