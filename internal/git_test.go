/* SPDX-License-Identifier: MIT */
package internal

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var dFlag = flag.Bool("debug", false, "enable debug mode")

func debug(args ...string) {
	if !*dFlag {
		return
	}

	fmt.Println("[DEBUG]", args)
}
func createTempDir(t *testing.T, pattern string) string {
	dir, err := os.MkdirTemp("", pattern)
	if err != nil {
		require.NoError(t, err, "failed creatign temp directory")
	}

	debug("New temp dir", dir)
	t.Cleanup(func() {
		if !*dFlag {
			os.RemoveAll(dir)
		} else {
			debug("did not clena repo for debugging", dir)
		}
	})
	return dir
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
		dir := createTempDir(t, "git-decent-test-with-dir")
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

	t.Run("Log with random commit", func(t *testing.T) {
		repo, err := NewRepositoryBuilder(t).As(Working).WithRandomCommits(20).Build()
		require.NoError(t, err)
		require.NotNil(t, repo)

		commits, err := repo.Log()
		assert.NoError(t, err, "git log should not fail")
		assert.Len(t, commits, 20, "all 20 commits should be returned")
		assert.Len(t, commits[0].Files, 2)
	})
}

func TestPushToOrigin(t *testing.T) {
	const amountCommits = 10

	bare, err := NewRepositoryBuilder(t).As(Bare).Build()
	require.NoError(t, err)
	commits, err := bare.Log()
	require.Empty(t, commits)

	repo, err := NewRepositoryBuilder(t).As(Working).WithRandomCommits(amountCommits).WithOrigin(bare.Dir).Build()

	err = repo.Push()
	require.NoError(t, err)
	commits, err = bare.Log()
	require.NoError(t, err)
	assert.Len(t, commits, amountCommits)
}
