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

	t.Run("Clone", func(t *testing.T) {
		bare, err := NewRepositoryBuilder(t).As(Bare).Build()
		require.NoError(t, err)
		_, err = NewRepositoryBuilder(t).Clone(bare.Dir).Build()
		require.NoError(t, err)
	})

	t.Run("Clone + Push + Check upstream", func(t *testing.T) {
		bare, err := NewRepositoryBuilder(t).As(Bare).Build()
		require.NoError(t, err)
		repo, err := NewRepositoryBuilder(t).Clone(bare.Dir).WithRandomCommits(10).Build()
		require.NoError(t, err)

		assert.Equal(t, "", repo.BranchUpstream("main"))

		err = repo.Push()
		require.NoError(t, err)

		require.Equal(t, "main", repo.CurrentBranch())
		assert.Equal(t, "origin/main", repo.BranchUpstream("main"))
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

	t.Run("Log with commits with date", func(t *testing.T) {
		historyDates := []time.Time{
			time.Date(2022, 1, 1, 1, 0, 0, 0, time.Now().UTC().Location()),
			time.Date(2022, 1, 1, 2, 0, 0, 0, time.Now().UTC().Location()),
		}
		repo, err := NewRepositoryBuilder(t).As(Working).WithCommitsWithDates(historyDates).Build()
		require.NoError(t, err)
		log, err := repo.Log()
		assert.NoError(t, err)
		assert.Len(t, log, 2)

		assert.Equal(t, log[0].Date, historyDates[0])
		assert.Equal(t, log[1].Date, historyDates[1])
	})
}

func TestIsGitRepo(t *testing.T) {
	dir := createTempDir(t, "not-a-repo")
	repo, err := NewGitRepo(dir, *dFlag)
	assert.NoError(t, err)

	assert.False(t, repo.IsGitRepo())

	NewRepositoryBuilder(t).At(dir).MustBuild()
	assert.True(t, repo.IsGitRepo())
}

func TestConfig(t *testing.T) {
	r := NewRepositoryBuilder(t).As(Working).MustBuild()

	input := []byte(`[decent]
		Monday = 09:00/17:00, 18:00/19:00
		Tuesday = 10:00/11:00
`)

	cFile := filepath.Join(r.Dir, ".git/config")
	f, err := os.OpenFile(cFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	require.NoError(t, err)

	f.Write(input)
	f.Close()

	options, err := r.GetSectionOptions("not-existing")
	assert.Error(t, err)
	assert.Empty(t, options)
	options, err = r.GetSectionOptions("decent")
	assert.NoError(t, err)
	assert.Len(t, options, 2)

	assert.Equal(t, "09:00/17:00, 18:00/19:00", options["monday"])
	assert.Equal(t, "10:00/11:00", options["tuesday"])

	o, err := r.GetConfig("decent.Wednesday")
	assert.Empty(t, o)
	assert.Error(t, err)

	o, err = r.GetConfig("decent.Tuesday")
	assert.NoError(t, err)
	assert.Equal(t, "10:00/11:00", o)
}

func TestGetHook(t *testing.T) {
	r := NewRepositoryBuilder(t).As(Working).MustBuild()
	hook := filepath.Join(r.Dir, ".git/hooks/pre-commit")
	os.WriteFile(hook, []byte(`some hook`), 0644)

	content, err := r.GetHook("pre-commit")
	assert.NoError(t, err)
	assert.Equal(t, "some hook", content)

	content, err = r.GetHook("pre-commit-non-existent")
	assert.Error(t, err)
	assert.Empty(t, content)
}

func TestLogWithRevisionFromUpstream(t *testing.T) {
	bare, err := NewRepositoryBuilder(t).As(Bare).Build()
	require.NoError(t, err)
	repo, err := NewRepositoryBuilder(t).Clone(bare.Dir).WithRandomCommits(10).Build()
	require.NoError(t, err)

	err = repo.Push()
	require.NoError(t, err)

	c, err := NewFixtureCommit(repo)
	require.NoError(t, err)
	err = repo.Commit(c)
	require.NoError(t, err)

	aLog := fmt.Sprintf("%s...", repo.BranchUpstream("main"))
	cs, err := repo.LogWithRevision(aLog)
	assert.NoError(t, err)
	assert.Len(t, cs, 1)
	assert.Equal(t, cs[0].Message, c.Message)
}

func TestPushToOrigin(t *testing.T) {
	const amountCommits = 10

	bare, err := NewRepositoryBuilder(t).As(Bare).Build()
	require.NoError(t, err)
	commits, err := bare.Log()
	assert.Error(t, err)
	require.Empty(t, commits)

	repo, err := NewRepositoryBuilder(t).As(Working).WithRandomCommits(amountCommits).WithOrigin(bare.Dir).Build()
	assert.NoError(t, err)
	err = repo.Push()
	require.NoError(t, err)
	commits, err = bare.Log()
	require.NoError(t, err)
	assert.Len(t, commits, amountCommits)
}

func TestLogBeforeAndAfter(t *testing.T) {
	dir := createTempDir(t, "git-decent-test-log")
	os.WriteFile(filepath.Join(dir, "fixture"), []byte("test fixture"), 0666)
	os.WriteFile(filepath.Join(dir, "fixture2"), []byte("test fixture2"), 0666)
	c := Commit{
		Message: "Some commit message",
		Date:    time.Date(2000, 12, 20, 1, 2, 3, 4, time.UTC),
		Author:  "Git test <withcommits@git-decent.git>",
		Files:   []string{"fixture"},
	}
	c2 := c
	c2.Date = time.Date(2000, 12, 21, 1, 2, 3, 4, time.UTC)
	c2.Files = []string{"fixture2"}
	repo, err := NewRepositoryBuilder(t).At(dir).AddCommit(&c).AddCommit(&c2).Build()
	assert.NoError(t, err)
	assert.NotNil(t, repo)

	cs, err := repo.Log()
	assert.NoError(t, err)
	assert.NotEmpty(t, cs)
	assert.Nil(t, cs[0].Prev)
	assert.NotNil(t, cs[0].Next, "First commit should be linked to the second")
}

func TestRootCommitHash(t *testing.T) {
	repo := NewRepositoryBuilder(t).WithRandomCommits(2).MustBuild()
	log, err := repo.Log()
	assert.NoError(t, err)

	rootHash, err := repo.RootCommitHash()
	assert.NoError(t, err)
	assert.Equal(t, log[0].Hash, rootHash)
}

func TestAmendMultipleDatesWithRoot(t *testing.T) {
	repo := NewRepositoryBuilder(t).WithRandomCommits(5).MustBuild()
	log, err := repo.Log()
	assert.NoError(t, err)
	assert.NotEmpty(t, log)

	for key := range log {
		log[key].Date = time.Date(2022, 02, key+1, 0, 0, 0, 0, log[key].Date.Location())
	}

	err = repo.AmendDates(log)
	assert.NoError(t, err, "Amend dates should not return an error")

	amendedLog, _ := repo.Log()
	assert.NotEmpty(t, log)

	for key, commit := range amendedLog {
		assert.Equal(t, time.Date(2022, 02, key+1, 0, 0, 0, 0, commit.Date.Location()), commit.Date)
	}
}

func TestAmendMultipleDatesWithoutRoot(t *testing.T) {
	repo := NewRepositoryBuilder(t).WithRandomCommits(5).MustBuild()
	log, err := repo.LogWithRevision("-3")
	assert.NoError(t, err)
	assert.NotEmpty(t, log)

	for key := range log {
		log[key].Date = time.Date(2022, 02, key+1, 0, 0, 0, 0, log[key].Date.Location())
	}

	err = repo.AmendDates(log)
	assert.NoError(t, err, "Amend dates should not return an error")

	amendedLog, _ := repo.LogWithRevision("-3")
	assert.NotEmpty(t, log)

	for key, commit := range amendedLog {
		assert.Equal(t, time.Date(2022, 02, key+1, 0, 0, 0, 0, commit.Date.Location()), commit.Date)
	}
}

func TestAmendSingleCommitNotHead(t *testing.T) {
	repo := NewRepositoryBuilder(t).WithRandomCommits(3).MustBuild()
	log, err := repo.LogWithRevision("-2")
	assert.NoError(t, err)
	assert.Len(t, log, 2)

	log[1].Date = time.Date(2022, 02, 1, 0, 0, 0, 0, log[0].Date.Location())

	err = repo.AmendDate(log[0])
	assert.Error(t, err, "is not HEAD")
}

func TestAmendSingleCommit(t *testing.T) {
	repo := NewRepositoryBuilder(t).WithRandomCommits(2).MustBuild()
	log, err := repo.LogWithRevision("-1")
	assert.NoError(t, err)
	assert.Len(t, log, 1)

	origDate := log[0].Date
	changedDate := time.Date(2022, 02, 1, 0, 0, 0, 0, log[0].Date.Location())
	log[0].Date = changedDate

	err = repo.AmendDate(log[0])
	assert.NoError(t, err)

	changedLog, err := repo.LogWithRevision("-1")
	assert.NoError(t, err)
	assert.Len(t, log, 1)

	fmt.Println(changedLog[0].Date, origDate)
	assert.Equal(t, changedLog[0].Date, changedDate)
}
