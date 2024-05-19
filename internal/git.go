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
type GitLog []*Commit

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
	Prev    *Commit
	Next    *Commit
}

type CommandError struct {
	error
	Command string
	Stdout  string
	Stderr  string
}

func NewGitRepoWithoutGlobalConfig(dir string) (*GitRepo, error) {
	repo, err := newGitRepo(dir)
	if err != nil {
		return nil, err
	}
	repo.configDir = filepath.Join(repo.Dir, ".globalconfig")
	return repo, err
}

func NewGitRepo(dir string) (*GitRepo, error) {
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

func (r *GitRepo) commandWithEnv(env []string, arg ...string) (string, error) {

	cmd := exec.Command(g, arg...)
	cmd.Dir = r.Dir

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, env...)

	if r.configDir != "" {
		cmd.Env = append(cmd.Env, "GIT_CONFIG_GLOBAL="+r.configDir)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	if err != nil {
		command := cmd.String()
		return "", &CommandError{
			error:   fmt.Errorf("%s error %s %s %w", command, cmd.Stdout, cmd.Stderr, err),
			Stdout:  stdoutBuf.String(),
			Stderr:  stderrBuf.String(),
			Command: command,
		}

	}
	return stdoutBuf.String(), nil
}

func (r *GitRepo) command(arg ...string) (string, error) {
	return r.commandWithEnv([]string{}, arg...)
}

func (r *GitRepo) IsGitRepo() bool {
	output, err := r.command("rev-parse", "--is-inside-work-tree")
	if err != nil {
		return false
	}

	cleaned := strings.TrimSpace(string(output))
	return cleaned == "true"
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

func (r *GitRepo) CurrentBranch() string {
	output, err := r.command("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func (r *GitRepo) BranchUpstream(branch string) string {
	out, err := r.command("rev-parse", "--abbrev-ref", fmt.Sprintf("%s@{u}", branch))
	if err != nil {
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
		fmt.Sprintf("--date=%s", commit.Date.Format("2006-01-02T15:04:05-0700")),
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

func (r *GitRepo) GetConfig(option string) (string, error) {
	out, err := r.command("config", "--get", option)
	if err != nil {
		return "", fmt.Errorf("git config failed, seciton does not exists? %w", err)
	}
	return strings.TrimSpace(out), nil
}

func (r *GitRepo) GetSectionOptions(name string) (map[string]string, error) {
	ops := map[string]string{}
	out, err := r.command("config", "--get-regexp", fmt.Sprintf("^%s.*", name))
	if err != nil {
		return ops, fmt.Errorf("git config failed, seciton does not exists? %w", err)
	}

	rOps := strings.Split(strings.TrimSpace(out), "\n")
	for _, option := range rOps {
		option = strings.TrimSpace(option)
		parts := strings.SplitN(option, " ", 2)
		if len(parts) != 2 {
			return ops, fmt.Errorf("git config option with invalid value (%s)", option)
		}

		key := strings.Replace(parts[0], fmt.Sprintf("%s.", name), "", 1)
		value := parts[1]
		ops[key] = value
	}

	return ops, nil
}

func (r *GitRepo) Push() error {
	_, err := r.command("push", "--all")
	return err
}

func (r *GitRepo) RootCommitHash() (string, error) {
	params := []string{"rev-list", "--max-parents=0", "HEAD"}
	output, err := r.command(params...)
	if err != nil {
		return "", err
	}
	output = strings.TrimSpace(output)

	return output, nil
}

func (r *GitRepo) AmendDates(log GitLog) error {
	hash, err := r.RootCommitHash()
	if err != nil {
		return fmt.Errorf("failed to obtain the root commit hash: %w", err)
	}

	cmd := []string{"rebase", "--interactive"}
	if hash == log[0].Hash {
		cmd = append(cmd, "--root")
	} else {
		cmd = append(cmd, fmt.Sprintf("HEAD~%d", len(log)))
	}

	var builder strings.Builder
	for _, commit := range log {
		newDate := commit.Date.Format("Mon, 02 Jan 2006 15:04:05 -0700")
		builder.WriteString(fmt.Sprintf(
			"pick %s\nexec GIT_COMMITTER_DATE=\"%s\" git commit --amend --no-edit --date=\"%s\"\n",
			commit.Hash[:7],
			newDate,
			newDate,
		))
	}

	file, err := os.CreateTemp(os.TempDir(), "git-decent-amend")
	if err != nil {
		return fmt.Errorf("failed to create tmp file for rebase todo: %w", err)
	}

	defer os.Remove(file.Name())

	_, err = file.WriteString(builder.String())

	if err != nil {
		file.Close()
		return fmt.Errorf("failed to write the todo file: %w", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("failed to close the root commit hash: %w", err)
	}

	editorCommand := fmt.Sprintf("GIT_SEQUENCE_EDITOR=cp %s", file.Name())
	_, err = r.commandWithEnv([]string{editorCommand}, cmd...)
	return err
}

func (r *GitRepo) LogWithRevision(revisionRange string) (GitLog, error) {
	return r.log(revisionRange)
}

func (r *GitRepo) Log() (GitLog, error) {
	return r.log()
}

func (r *GitRepo) log(args ...string) (GitLog, error) {
	params := []string{"log", "--pretty=format:%H%x1f%an%x1f%ai%x1f%s%x1f", "--name-only", "--reverse"}
	params = append(params, args...)
	output, err := r.command(params...)
	if err != nil {
		return make(GitLog, 0), fmt.Errorf("couldn't execute git log %w", err)
	}

	return parseLog(output)
}

func parseLog(output string) (GitLog, error) {
	commits := GitLog{}
	var lastCommit *Commit

	rawCommits := strings.Split(output, "\n\n")
	for _, rawCommit := range rawCommits {
		parts := strings.Split(rawCommit, "\x1f")
		if len(parts) < 4 {
			continue
		}
		commit := Commit{
			Hash:    parts[0],
			Author:  parts[1],
			Message: parts[3],
			Prev:    lastCommit,
		}

		dateStr := parts[2]
		files := parts[4]

		date, err := time.Parse("2006-01-02 15:04:05 -0700", dateStr)
		if err != nil {
			fmt.Println("[WARN]: couldn't parse date from commit log")
		}

		commit.Date = date.UTC()

		pFiles := strings.Split(files, "\n")
		for _, pFile := range pFiles {
			if pFile != "" {
				commit.Files = append(commit.Files, pFile)
			}
		}

		if lastCommit != nil {
			lastCommit.Next = &commit
			commit.Prev = lastCommit
		}

		lastCommit = &commit
		commits = append(commits, &commit)
	}

	return commits, nil
}
