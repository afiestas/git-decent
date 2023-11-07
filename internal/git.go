/* SPDX-License-Identifier: MIT */
package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const g string = "git"

type RepoType int

const (
	Working RepoType = iota
	Bare
)

type GitRepo struct {
	Dir string
}

type Commit struct {
	Message string
	Date    time.Time
	Author  string
	Files   []string
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

func (r *GitRepo) command(arg ...string) *exec.Cmd {
	fmt.Println(arg)
	cmd := exec.Command(g, arg...)
	cmd.Dir = r.Dir
	return cmd
}

func (r *GitRepo) Init(rT RepoType) error {

	checkFile := ".git/config"
	args := make([]string, 0, 2)
	args = append(args, "init")

	if rT == Bare {
		args = append(args, "--bare")
		checkFile = "config"
	}

	gitDir := filepath.Join(r.Dir, checkFile)
	_, err := os.Stat(gitDir)
	if err == nil {
		return fmt.Errorf("repository already initialized in %s", r.Dir)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("couldn't stat the directory %w", err)
	}

	cmd := r.command(args[:]...)
	output, err := cmd.CombinedOutput()
	fmt.Println("OUTPUT", string(output), cmd, err)
	return err
}

func (r *GitRepo) Type() (RepoType, error) {
	cmd := r.command("rev-parse", "--is-bare-repository")

	//TODO Caputre stderr&stdout for error reporting
	output, err := cmd.Output()
	if err != nil {
		return Working, fmt.Errorf("couldn't execute git rev parse, output: %s, error: %w", output, err)
	}

	cleaned := strings.TrimSpace(string(output))
	if cleaned == "true" {
		return Bare, nil
	}
	return Working, nil
}

func (r *GitRepo) SetOrigin(url string) error {
	cmd := r.command("remote", "add", "origin", url)
	o, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("couldn't get the remote for origin =%s= %w", o, err)
	}

	return nil
}

func (r *GitRepo) Origin() (string, error) {
	cmd := r.command("remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("couldn't get the remote for origin %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func (r *GitRepo) Commit(commit *Commit) error {
	cmd := r.command(append([]string{"add"}, commit.Files...)...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("couldn't add files %s, %w", output, err)
	}

	cmd = r.command(
		"commit",
		"-m", commit.Message,
		fmt.Sprintf("--author=%s", commit.Author),
		fmt.Sprintf("--date=%s", commit.Date.Format("2006-01-02T15:04:05-07:00")),
	)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("couldn't create a commit %s, %w", output, err)
	}

	return nil
}
