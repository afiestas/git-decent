/* SPDX-License-Identifier: MIT */
package internal

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTempDir(t *testing.T, pattern string) string {
	dir, err := os.MkdirTemp("", pattern)
	if err != nil {
		assert.NoError(t, err, "failed creatign temp directory")
	}

	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

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

	repo, err := NewGitRepo(createTempDir(t, "git-decent-"))
	assert.NoError(t, err, "repo should be returned without errors")
	assert.NotEmpty(t, repo.Dir, "the returned repo should have the Dir initialized")
}

func TestInit(t *testing.T) {
	t.Run("Initialize bare", func(t *testing.T) {
		repo, err := NewGitRepo(createTempDir(t, "git-decent-"))
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
		repo, err := NewGitRepo(createTempDir(t, "git-decent-"))
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
		repo, err := NewRepositoryBuilder(t).Build()
		defer os.RemoveAll(repo.Dir)

		assert.NoError(t, err, "builder without step should always work")
		assert.NotEmpty(t, repo.Dir, "a dir should be created since none is passed")
		assert.DirExists(t, repo.Dir, "returned directory should exist")

		repo2, err := NewRepositoryBuilder(t).Build()
		assert.NoError(t, err, "builder without step should always work")
		assert.NotEqual(t, repo.Dir, repo2.Dir, "Repos withotu dir should use a different tempt dir")
	})

	t.Run("With dir", func(t *testing.T) {
		dir := createTempDir(t, "git-decent-test-with-dir")
		repo, err := NewRepositoryBuilder(t).At(dir).Build()
		assert.NoError(t, err, "builder without step should always work")
		assert.NotEmpty(t, repo.Dir, "a dir should be created since none is passed")
		assert.DirExists(t, repo.Dir, "returned directory should exist")
		assert.Equal(t, dir, repo.Dir)
	})
	t.Run("With Initialize", func(t *testing.T) {
		repo, err := NewRepositoryBuilder(t).As(Bare).Build()
		assert.NoError(t, err, "builder without step should always work")
		rt, err := repo.Type()
		assert.NoError(t, err, "getting the type shouldn't fail")
		assert.Equal(t, rt, Bare)

		repo, err = NewRepositoryBuilder(t).As(Working).Build()
		assert.NoError(t, err, "builder without step should always work")
		rt, err = repo.Type()
		assert.NoError(t, err, "getting the type shouldn't fail")
		assert.Equal(t, rt, Working)
	})

	t.Run("With origin", func(t *testing.T) {
		repo, err := NewRepositoryBuilder(t).WithOrigin("/foo/bar").Build()
		assert.NoError(t, err, "no error is expected")

		o, err := repo.Origin()
		assert.NoError(t, err, "no error is expected")

		assert.Equal(t, "/foo/bar", o)
	})

	t.Run("With commit file not exists", func(t *testing.T) {
		c := Commit{
			Message: "Some commit message",
			Date:    time.Date(2000, 12, 20, 1, 2, 3, 4, time.UTC),
			Author:  "Git test",
			Files:   []string{"/some/file/that/does/not/exists"},
		}
		_, err := NewRepositoryBuilder(t).AddCommit(&c).Build()
		assert.ErrorIs(t, err, os.ErrNotExist)
	})
	t.Run("With commit", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "git-decent-test-with-dir")
		// os.RemoveAll(dir)
		assert.NoError(t, err, "mkdirtemp should give us a directory without error")

		os.WriteFile(filepath.Join(dir, "fixture"), []byte("test fixture"), 0666)
		c := Commit{
			Message: "Some commit message",
			Date:    time.Date(2000, 12, 20, 1, 2, 3, 4, time.UTC),
			Author:  "Git test <withcommits@git-decent.git>",
			Files:   []string{"fixture"},
		}
		repo, err := NewRepositoryBuilder(t).At(dir).AddCommit(&c).Build()
		assert.NoError(t, err)
		assert.NotNil(t, repo)
	})
}
