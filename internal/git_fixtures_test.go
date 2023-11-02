package internal

import (
	"fmt"
	"os"
)

type RepositoryBuilder struct {
	repoType   RepoType
	dir        string
	origin     string
	initialize bool
}

func NewRepositoryBuilder() *RepositoryBuilder {
	return &RepositoryBuilder{initialize: true}
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

	return repo, nil
}
