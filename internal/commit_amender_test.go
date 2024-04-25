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
func TestAmendSingleUndecentCommit(t *testing.T) {
	historyDates := []time.Time{
		parseDate("2024-01-28 18:30:00", t),
	}

	repo := NewRepositoryBuilder(t).As(Working).WithCommitsWithDates(historyDates).MustBuild()

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

func TestAmendTwoCloseUndecentCommits(t *testing.T) {
	historyDates := []time.Time{
		parseDate("2024-01-28 18:30:00", t),
		parseDate("2024-01-28 18:35:00", t),
	}

	repo := NewRepositoryBuilder(t).As(Working).WithCommitsWithDates(historyDates).MustBuild()

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
	log[0] = &amended
	log[1].Prev = log[0]
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

func TestAmendCompressionUndecentCommits(t *testing.T) {
	historyDates := []time.Time{
		parseDate("2024-01-28 18:30:00", t),
		parseDate("2024-01-28 23:59:00", t),
	}

	repo := NewRepositoryBuilder(t).As(Working).WithCommitsWithDates(historyDates).MustBuild()

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
	log[0] = &amended
	log[1].Prev = log[0]
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

func TestAmendCommitInAmendedRange(t *testing.T) {
	historyDates := []time.Time{
		parseDate("2024-01-28 18:30:00", t),
		parseDate("2024-01-28 23:59:00", t),
		parseDate("2024-01-29 09:00:00", t),
	}

	repo := NewRepositoryBuilder(t).As(Working).WithCommitsWithDates(historyDates).MustBuild()

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
	log[0] = &amended
	log[1].Prev = log[0]
	amended2 := Amend(*log[1], log, schedule)
	log[1] = &amended2
	log[2].Prev = log[1]

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
	assert.Equal(t, 10, amended3.Date.Minute())
}

func TestAmendOverflow(t *testing.T) {
	historyDates := []time.Time{
		parseDate("2024-01-29 17:00:00", t),
		parseDate("2024-01-29 16:55:00", t),
		parseDate("2024-01-29 17:50:00", t),
		parseDate("2024-01-29 20:51:00", t),
	}

	repo := NewRepositoryBuilder(t).As(Working).WithCommitsWithDates(historyDates).MustBuild()

	log, err := repo.Log()
	require.NoError(t, err)

	fmt.Println(log)
	commit := log[0]

	raw := config.RawScheduleConfig{
		Days: map[time.Weekday]string{
			time.Monday:  "09:00/17:00, 18:00/19:00",
			time.Tuesday: "09:00/17:00",
		},
	}
	schedule, err := config.NewScheduleFromRaw(&raw)
	require.NoError(t, err)

	amended := Amend(*log[0], log, schedule)
	amended2 := Amend(*log[1], log, schedule)
	log[1] = &amended2
	log[2].Prev = log[1]
	amended3 := Amend(*log[2], log, schedule)

	assert.Equal(t, amended.Date, commit.Date)
	assert.Equal(t, time.Monday, amended.Date.Weekday())
	assert.Equal(t, 29, amended.Date.Day())
	assert.Equal(t, 17, amended.Date.Hour())
	assert.Equal(t, 0, amended.Date.Minute())

	assert.Equal(t, time.Monday, amended2.Date.Weekday())
	assert.Equal(t, 29, amended2.Date.Day())
	assert.Equal(t, 17, amended2.Date.Hour())
	assert.Equal(t, 5, amended2.Date.Minute())

	assert.Equal(t, time.Monday, amended3.Date.Weekday())
	assert.Equal(t, 29, amended3.Date.Day())
	assert.Equal(t, 18, amended3.Date.Hour())
	assert.Equal(t, 0, amended3.Date.Minute())
}
