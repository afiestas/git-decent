package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type RepositoryBuilder struct {
	repoType      RepoType
	dir           string
	origin        string
	initialize    bool
	commits       []Commit
	randomCommits []Commit
}

func NewRepositoryBuilder(t *testing.T) *RepositoryBuilder {
	rb := &RepositoryBuilder{initialize: true}
	if t != nil {
		t.Cleanup(func() {
			if rb.dir != "" {
				os.RemoveAll(rb.dir)
			}
		})
	}
	return rb
}

func (rb *RepositoryBuilder) At(path string) *RepositoryBuilder {
	rb.dir = path
	return rb
}

func (rb *RepositoryBuilder) As(rt RepoType) *RepositoryBuilder {
	rb.initialize = true
	rb.repoType = rt
	return rb
}

func (rb *RepositoryBuilder) WithOrigin(remote string) *RepositoryBuilder {
	rb.origin = remote
	return rb
}

func (rb *RepositoryBuilder) AddCommit(commit *Commit) *RepositoryBuilder {
	rb.commits = append(rb.commits, *commit)
	return rb
}

func (rb *RepositoryBuilder) AddRandomCommit() *RepositoryBuilder {
	rb.randomCommits = append(rb.randomCommits, Commit{})
	return rb
}

func (rb *RepositoryBuilder) Build() (*GitRepo, error) {
	if rb.dir == "" {
		dir, err := os.MkdirTemp("", "git-decent-")
		if err != nil {
			return nil, fmt.Errorf("couldn't get temp dir for fixtures %w", err)
		}
		rb.dir = dir
	}

	repo, err := NewGitRepo(rb.dir)
	if err != nil {
		return nil, fmt.Errorf("couldn't get the GitRpeo %w", err)
	}

	if !rb.initialize {
		return repo, nil
	}

	err = repo.Init(rb.repoType)
	if err != nil {
		return nil, fmt.Errorf("couldn't init the GitRpeo %w", err)
	}

	if rb.origin != "" {
		err = repo.SetOrigin(rb.origin)
		if err != nil {
			return nil, fmt.Errorf("couldn't set origin in repo %w", err)
		}
	}

	for _, commit := range rb.commits {
		for _, file := range commit.Files {
			_, err := os.Stat(filepath.Join(rb.dir, file))
			if err != nil {
				return nil, fmt.Errorf("couldn't add file %w", err)
			}
		}
		err := repo.Commit(&commit)
		if err != nil {
			return nil, fmt.Errorf("couldn't create commit %w", err)
		}
	}

	return repo, nil
}

func createFileInRepo(baseDir string) (string, error) {
	i := 0
	prefix := "fixture_"
	fname := ""
	createFile := ""
	for {
		fname = fmt.Sprintf("%s_%d", prefix, i)
		createFile = filepath.Join(baseDir, fname)
		_, err := os.Stat(createFile)
		if err == nil {
			continue //File already exists
		} else if err != os.ErrNotExist {
			return "", fmt.Errorf("couldn't stat fixture file %w", err)
		}

		err = os.WriteFile(createFile, []byte(fname), 0666)

		if err != nil {
			return "", fmt.Errorf("couldn't write fixture file %w", err)
		}

		return fname, nil
	}
}

func newFixtureCommit(repo *GitRepo) (*Commit, error) {

	fname, err := createFileInRepo(repo.Dir)
	if err != nil {
		return nil, fmt.Errorf("couldn't create fixture file %w", err)
	}
	return &Commit{
		Message: fmt.Sprintf("Some commit message for %s", fname),
		Date:    time.Date(2000, 12, 20, 1, 2, 3, 4, time.UTC),
		Author:  "Git test test@git-decent.git",
		Files:   []string{"fname"},
	}, nil
}
