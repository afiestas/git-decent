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
	cmd := repo.command("some-cmd", "some-arg")
	assert.Equal(t, cmd.Dir, "some-dir", "cwd is to be set to the repo basedir")
}

func TestInit(t *testing.T) {
	t.Run("Initialize bare", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "git-decent-")
		require.NoError(t, err, "couldn't get temp dir for fixtures")
		defer os.RemoveAll(dir)

		repo, err := NewGitRepo(dir)
		assert.NoError(t, err, "repo should be returned without errors")

		err = repo.Init(Bare)
		assert.NoError(t, err, "repo should initialized without errors")

		rT, err := repo.Type()
		assert.NoError(t, err, "Type should not return any errors")

		assert.Equal(t, rT, Bare, "repo should be bare")

		err = repo.Init(Bare)
		assert.ErrorContains(t, err, "repository already initialized")
	})
	t.Run("Initialize working", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "git-decent-")
		require.NoError(t, err, "couldn't get temp dir for fixtures")
		defer os.RemoveAll(dir)

		repo, err := NewGitRepo(dir)
		assert.NoError(t, err, "repo should be returned without errors")

		err = repo.Init(Working)
		assert.NoError(t, err, "repo should initialized without errors")

		rT, err := repo.Type()
		assert.NoError(t, err, "Type should not return any errors")

		assert.Equal(t, rT, Working, "repo should be working")

		err = repo.Init(Working)
		assert.ErrorContains(t, err, "repository already initialized")
	})
}

func TestFixtures(t *testing.T) {
	t.Run("Without dir", func(t *testing.T) {
		repo, err := NewRepositoryBuilder().Build()
		defer os.RemoveAll(repo.Dir)

		assert.NoError(t, err, "builder without step should always work")
		assert.NotEmpty(t, repo.Dir, "a dir should be created since none is passed")
		assert.DirExists(t, repo.Dir, "returned directory should exist")

		repo2, err := NewRepositoryBuilder().Build()
		assert.NoError(t, err, "builder without step should always work")
		assert.NotEqual(t, repo.Dir, repo2.Dir, "Repos withotu dir should use a different tempt dir")
	})

	t.Run("With dir", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "git-decent-test-with-dir")
		assert.NoError(t, err, "mkdirtemp should give us a directory without error")

		repo, err := NewRepositoryBuilder().At(dir).Build()
		assert.NoError(t, err, "builder without step should always work")
		assert.NotEmpty(t, repo.Dir, "a dir should be created since none is passed")
		assert.DirExists(t, repo.Dir, "returned directory should exist")
		assert.Equal(t, dir, repo.Dir)
	})
	t.Run("With Initialize", func(t *testing.T) {
		repo, err := NewRepositoryBuilder().As(Bare).Build()
		assert.NoError(t, err, "builder without step should always work")
		rt, err := repo.Type()
		assert.NoError(t, err, "getting the type shouldn't fail")
		assert.Equal(t, rt, Bare)

		repo, err = NewRepositoryBuilder().As(Working).Build()
		assert.NoError(t, err, "builder without step should always work")
		rt, err = repo.Type()
		assert.NoError(t, err, "getting the type shouldn't fail")
		assert.Equal(t, rt, Working)
	})

	t.Run("With origin", func(t *testing.T) {
		repo, err := NewRepositoryBuilder().WithOrigin("/foo/bar").Build()
		assert.NoError(t, err, "no error is expected")

		o, err := repo.Origin()
		assert.NoError(t, err, "no error is expected")

		assert.Equal(t, "/foo/bar", o)
	})
}
