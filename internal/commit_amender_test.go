package internal

import (
	"fmt"
	"testing"
	"time"

	"github.com/afiestas/git-decent/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseDate(dateStr string, t *testing.T) time.Time {
	layout := "2006-01-02 15:04:05"
	pTime, err := time.Parse(layout, dateStr)
	assert.NoError(t, err)

	return pTime
}
func TestSingleUndecentCommit(t *testing.T) {
	historyDates := []time.Time{
		parseDate("2024-01-28 18:30:00", t),
	}

	repo, err := NewRepositoryBuilder(t).As(Working).WithCommitsWithDates(historyDates).Build()
	require.NoError(t, err)
	require.NotNil(t, repo)

	log, err := repo.Log()
	require.NoError(t, err)

	commit := log[0]

	raw := config.RawScheduleConfig{
		Days: map[time.Weekday]string{
			time.Monday: "09:00/17:00",
		},
	}
	schedule, err := config.NewScheduleFromRaw(&raw)
	require.NoError(t, err)

	amended := Amend(*commit, log, schedule)
	assert.NotEqual(t, amended.Date, commit.Date)

	assert.Equal(t, time.Monday, amended.Date.Weekday())
	assert.Equal(t, 29, amended.Date.Day())
	assert.Equal(t, 9, amended.Date.Hour())
	assert.Equal(t, 0, amended.Date.Minute())
}

func TestTwoCloseUndecentCommits(t *testing.T) {
	historyDates := []time.Time{
		parseDate("2024-01-28 18:30:00", t),
		parseDate("2024-01-28 18:35:00", t),
	}

	repo, err := NewRepositoryBuilder(t).As(Working).WithCommitsWithDates(historyDates).Build()
	require.NoError(t, err)
	require.NotNil(t, repo)

	log, err := repo.Log()
	require.NoError(t, err)

	fmt.Println(log)
	commit := log[0]

	raw := config.RawScheduleConfig{
		Days: map[time.Weekday]string{
			time.Monday: "09:00/17:00",
		},
	}
	schedule, err := config.NewScheduleFromRaw(&raw)
	require.NoError(t, err)

	amended := Amend(*log[0], log, schedule)
	amended2 := Amend(*log[1], log, schedule)

	assert.NotEqual(t, amended.Date, commit.Date)
	assert.Equal(t, time.Monday, amended.Date.Weekday())
	assert.Equal(t, 29, amended.Date.Day())
	assert.Equal(t, 9, amended.Date.Hour())
	assert.Equal(t, 0, amended.Date.Minute())

	assert.Equal(t, time.Monday, amended2.Date.Weekday())
	assert.Equal(t, 29, amended2.Date.Day())
	assert.Equal(t, 9, amended2.Date.Hour())
	assert.Equal(t, 5, amended2.Date.Minute())
}

func TestCompressionUndecentCommits(t *testing.T) {
	historyDates := []time.Time{
		parseDate("2024-01-28 18:30:00", t),
		parseDate("2024-01-28 23:59:00", t),
	}

	repo, err := NewRepositoryBuilder(t).As(Working).WithCommitsWithDates(historyDates).Build()
	require.NoError(t, err)
	require.NotNil(t, repo)

	log, err := repo.Log()
	require.NoError(t, err)

	fmt.Println(log)
	commit := log[0]

	raw := config.RawScheduleConfig{
		Days: map[time.Weekday]string{
			time.Monday: "09:00/17:00",
		},
	}
	schedule, err := config.NewScheduleFromRaw(&raw)
	require.NoError(t, err)

	amended := Amend(*log[0], log, schedule)
	amended2 := Amend(*log[1], log, schedule)

	assert.NotEqual(t, amended.Date, commit.Date)
	assert.Equal(t, time.Monday, amended.Date.Weekday())
	assert.Equal(t, 29, amended.Date.Day())
	assert.Equal(t, 9, amended.Date.Hour())
	assert.Equal(t, 0, amended.Date.Minute())

	assert.Equal(t, time.Monday, amended2.Date.Weekday())
	assert.Equal(t, 29, amended2.Date.Day())
	assert.Equal(t, 9, amended2.Date.Hour())
	assert.Equal(t, 9, amended2.Date.Minute())
}

func TestCommitInAmendedRange(t *testing.T) {
	historyDates := []time.Time{
		parseDate("2024-01-28 18:30:00", t),
		parseDate("2024-01-28 23:59:00", t),
		parseDate("2024-01-29 09:00:00", t),
	}

	repo, err := NewRepositoryBuilder(t).As(Working).WithCommitsWithDates(historyDates).Build()
	require.NoError(t, err)
	require.NotNil(t, repo)

	log, err := repo.Log()
	require.NoError(t, err)

	fmt.Println(log)
	commit := log[0]

	raw := config.RawScheduleConfig{
		Days: map[time.Weekday]string{
			time.Monday: "09:00/17:00",
		},
	}
	schedule, err := config.NewScheduleFromRaw(&raw)
	require.NoError(t, err)

	amended := Amend(*log[0], log, schedule)
	amended2 := Amend(*log[1], log, schedule)
	amended3 := Amend(*log[2], log, schedule)

	assert.NotEqual(t, amended.Date, commit.Date)
	assert.Equal(t, time.Monday, amended.Date.Weekday())
	assert.Equal(t, 29, amended.Date.Day())
	assert.Equal(t, 9, amended.Date.Hour())
	assert.Equal(t, 0, amended.Date.Minute())

	assert.Equal(t, time.Monday, amended2.Date.Weekday())
	assert.Equal(t, 29, amended2.Date.Day())
	assert.Equal(t, 9, amended2.Date.Hour())
	assert.Equal(t, 9, amended2.Date.Minute())

	assert.Equal(t, time.Monday, amended3.Date.Weekday())
	assert.Equal(t, 29, amended3.Date.Day())
	assert.Equal(t, 9, amended3.Date.Hour())
	assert.Equal(t, 15, amended3.Date.Minute())
}
