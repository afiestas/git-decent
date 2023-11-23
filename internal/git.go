/* SPDX-License-Identifier: MIT */
package internal

import (
	"bytes"
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
	Dir       string
	configDir string
}

type Commit struct {
	Hash    string
	Message string
	Date    time.Time
	Author  string
	Files   []string
}

func NewGitRepoWithoutGlobalConfig(dir string) (*GitRepo, error) {
	repo, err := newGitRepo(dir)
	if err != nil {
		return nil, err
	}
	repo.configDir = filepath.Join(repo.Dir, ".globalconfig")
	return repo, err
}

func NewGitRepo(dir string, noGlobalConfig bool) (*GitRepo, error) {
	return newGitRepo(dir)
}

func newGitRepo(dir string) (*GitRepo, error) {
	if i, err := os.Stat(dir); err != nil {
		return nil, fmt.Errorf("couldn't stat directory %w", err)
	} else if !i.IsDir() {
		return nil, fmt.Errorf("repository is not a dir %s", dir)
	}

	return &GitRepo{
		Dir: dir,
	}, nil
}

func (r *GitRepo) command(arg ...string) (string, error) {
	cmd := exec.Command(g, arg...)
	cmd.Dir = r.Dir
	if r.configDir != "" {
		cmd.Env = append(cmd.Env, "GIT_CONFIG_GLOBAL="+r.configDir)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git command error %s %s %w", cmd.Stdout, cmd.Stderr, err)
	}
	return stdoutBuf.String(), nil
}

func (r *GitRepo) Init(rT RepoType) error {

	checkFile := ".git/config"
	args := make([]string, 0, 3)
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

	args = append(args, "--initial-branch=main")
	_, err = r.command(args[:]...)
	return err
}

func (r *GitRepo) Type() (RepoType, error) {
	output, err := r.command("rev-parse", "--is-bare-repository")
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
	_, err := r.command("remote", "add", "origin", url)
	if err != nil {
		return fmt.Errorf("couldn't get the remote for origin %w", err)
	}

	return nil
}

func (r *GitRepo) Origin() (string, error) {
	output, err := r.command("remote", "get-url", "origin")
	if err != nil {
		return "", fmt.Errorf("couldn't get the remote for origin %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func (r *GitRepo) BranchUpstream(branch string) string {
	out, err := r.command("rev-parse", "--abbrev-ref", fmt.Sprintf("%s@{u}", branch))
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return strings.TrimSpace(out)
}

func (r *GitRepo) Clone(remote string) error {
	_, err := r.command("clone", remote, ".")
	return err
}

func (r *GitRepo) Commit(commit *Commit) error {
	output, err := r.command(append([]string{"add"}, commit.Files...)...)
	if err != nil {
		return fmt.Errorf("couldn't add files %s, %w", output, err)
	}

	_, err = r.command(
		"commit",
		"-m", commit.Message,
		fmt.Sprintf("--author=%s", commit.Author),
		fmt.Sprintf("--date=%s", commit.Date.Format("2006-01-02T15:04:05-07:00")),
	)
	if err != nil {
		return fmt.Errorf("couldn't create a commit %w", err)
	}

	return nil
}

func (r *GitRepo) SetConfig(key string, value string) error {
	_, err := r.command("config", "--local", key, value)
	return err
}

func (r *GitRepo) Push() error {
	_, err := r.command("push", "--all")
	return err
}

func (r *GitRepo) LogWithRevision(revisionRange string) ([]Commit, error) {
	return r.log(revisionRange)
}

func (r *GitRepo) Log() ([]Commit, error) {
	return r.log()
}

func (r *GitRepo) log(args ...string) ([]Commit, error) {
	params := []string{"log", "--pretty=format:%H%x1f%an%x1f%ai%x1f%s%x1f", "--name-only"}
	params = append(params, args...)
	output, err := r.command(params...)
	if err != nil {
		return make([]Commit, 0), fmt.Errorf("couldn't execute git log %w", err)
	}

	return parseLog(output)
}

func parseLog(output string) ([]Commit, error) {
	commits := []Commit{}
	rawCommits := strings.Split(output, "\n\n")
	for _, rawCommit := range rawCommits {
		parts := strings.Split(rawCommit, "\x1f")
		if len(parts) < 4 {
			fmt.Println("[WARN]: Couldn't parse commit from log")
			continue
		}
		commit := Commit{
			Hash:    parts[0],
			Author:  parts[1],
			Message: parts[3],
		}

		dateStr := parts[2]
		files := parts[4]

		date, err := time.Parse("2006-01-02 15:04:05 -0700", dateStr)
		if err != nil {
			fmt.Println("[WARN]: couldn't parse date from commit log")
		}

		commit.Date = date

		pFiles := strings.Split(files, "\n")
		for _, pFile := range pFiles {
			if pFile != "" {
				commit.Files = append(commit.Files, pFile)
			}
		}
		// fmt.Printf("%#v \n", commit)

		commits = append(commits, commit)
	}

	return commits, nil
}

//Check if tracking remote, if so compare to it
//Otherwise, assume it is a local branch and log from it
//the configure base
