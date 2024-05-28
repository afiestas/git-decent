package internal

import (
	"testing"
	"time"

	"github.com/afiestas/git-decent/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sut(t *testing.T, fixtures amendFixture, decentFrames map[time.Weekday]string) {
	repo := NewRepositoryBuilder(t).As(Working).WithCommitsWithDates(fixtures.initialDates).MustBuild()
	log, err := repo.Log()
	assert.NoError(t, err)

	schedule, err := config.NewScheduleFromRaw(&config.RawScheduleConfig{Days: decentFrames})
	require.NoError(t, err)

	var lastDate *time.Time = nil
	for k, commit := range log {
		if commit.Prev != nil {
			commit.Prev = log[k-1]
			lastDate = &commit.Prev.Date
		}

		amended := Amend(commit.Date, lastDate, schedule)
		commit.Date = amended
		log[k] = commit
	}

	assertAmendedLog(t, fixtures.amendedDates, log)
}

func assertAmendedLog(t *testing.T, amendedDates []time.Time, log GitLog) {
	assert.NotEmpty(t, log)

	for k, commit := range log {
		assert.Equalf(t,
			amendedDates[k],
			commit.Date,
			"Amdeded commit (%d) is not like expected",
			k,
		)
	}
}

type amendFixture struct {
	initialDates []time.Time
	amendedDates []time.Time
}

func makeFixtures(t *testing.T, initial []string, amended []string) amendFixture {
	if len(initial) != len(amended) {
		t.Errorf("Initial(%d) dates array must be the same sas as amended(%d)", len(initial), len(amended))
	}

	parseDate := func(dateStr string, t *testing.T) time.Time {
		layout := "2006-01-02 15:04:05"
		pTime, err := time.Parse(layout, dateStr)
		assert.NoError(t, err)

		return pTime
	}

	fixtures := amendFixture{}
	for k, date := range initial {
		fixtures.initialDates = append(fixtures.initialDates, parseDate(date, t))
		fixtures.amendedDates = append(fixtures.amendedDates, parseDate(amended[k], t))
	}

	return fixtures
}

var tests = []struct {
	name        string                  // name of the test case
	initial     []string                // initial commit times as strings
	amended     []string                // expected amended commit times as strings
	decentSlots map[time.Weekday]string // mapping of weekdays to time slots
}{
	{
		name:    "Amend Single Undecent Commit",
		initial: []string{"2024-01-28 18:30:00"},
		amended: []string{"2024-01-29 09:00:00"},
		decentSlots: map[time.Weekday]string{
			time.Monday: "09:00/17:00",
		},
	},
	{
		name:    "Amend Two Close Undecent Commits",
		initial: []string{"2024-01-28 18:30:00", "2024-01-28 18:35:00"},
		amended: []string{"2024-01-29 09:00:00", "2024-01-29 09:05:00"},
		decentSlots: map[time.Weekday]string{
			time.Monday: "09:00/17:00",
		},
	},
	{
		name:    "Amend Compression Undecent Commits",
		initial: []string{"2024-01-28 18:30:00", "2024-01-28 23:59:00"},
		amended: []string{"2024-01-29 09:00:00", "2024-01-29 09:09:00"},
		decentSlots: map[time.Weekday]string{
			time.Monday: "09:00/17:00",
		},
	},
	{
		name:    "Amend Commit In Amended Range",
		initial: []string{"2024-01-28 18:30:00", "2024-01-28 23:59:00", "2024-01-29 09:00:00"},
		amended: []string{"2024-01-29 09:00:00", "2024-01-29 09:09:00", "2024-01-29 09:14:00"},
		decentSlots: map[time.Weekday]string{
			time.Monday: "09:00/17:00",
		},
	},
	{
		name:    "Amend Overflow",
		initial: []string{"2024-01-29 17:00:00", "2024-01-29 16:55:00", "2024-01-29 17:50:00", "2024-01-29 20:51:00"},
		amended: []string{"2024-01-29 17:00:00", "2024-01-29 18:05:00", "2024-01-29 18:10:00", "2024-01-30 09:01:00"},
		decentSlots: map[time.Weekday]string{
			time.Monday:  "09:00/17:00, 18:00/19:00",
			time.Tuesday: "09:00/17:00",
		},
	},
}

func TestAmendCommits(t *testing.T) {
	testRandom = true
	defer func() {
		testRandom = false
	}()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixtures := makeFixtures(t, tc.initial, tc.amended)
			sut(t, fixtures, tc.decentSlots)
		})
	}
}
