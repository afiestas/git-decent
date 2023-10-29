/* SPDX-License-Identifier: MIT */
package internal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGitRepo(t *testing.T) {
	t.Run("When directory does not exists", func(t *testing.T) {
		_, err := NewGitRepo("some/random/dir")
		assert.ErrorContains(t, err, "couldn't stat directory")
	})
	t.Run("When directory is actually a file", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "git-decent-test-")
		require.NoError(t, err, "couldn't create temp file for test")
		defer os.Remove(tmpFile.Name())

		_, err = NewGitRepo(tmpFile.Name())
		assert.ErrorContains(t, err, "repository is not a dir")
	})

	dir, err := os.MkdirTemp("", "git-decent-")
	require.NoError(t, err, "couldn't get temp dir for fixtures")
	defer os.RemoveAll(dir)

	repo, err := NewGitRepo(dir)
	assert.NoError(t, err, "repo should be returned without errors")
	assert.NotEmpty(t, repo.Dir, "the returned repo should have the Dir initialized")
}

func TestCommandWrapper(t *testing.T) {
	repo := GitRepo{Dir: "some-dir"}
	cmd := repo.Command("some-cmd", "some-arg")
	assert.Equal(t, cmd.Dir, "some-dir", "cwd is to be set to the repo basedir")
}

func TestInit(t *testing.T) {
	t.Run("Initialize bare", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "git-decent-")
		require.NoError(t, err, "couldn't get temp dir for fixtures")
		// defer os.RemoveAll(dir)

		repo, err := NewGitRepo(dir)
		assert.NoError(t, err, "repo should be returned without errors")

		err = repo.Init(true)
		assert.NoError(t, err, "repo should initialized without errors")

		isBare, err := repo.IsBare()
		assert.NoError(t, err, "isBare should not return any errors")

		assert.True(t, isBare, "repo should be bare")

		err = repo.Init(true)
		assert.ErrorContains(t, err, "repository already initialized")
	})
}
