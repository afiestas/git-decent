package config

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduleFromRawError(t *testing.T) {
	raw := RawScheduleConfig{
		Days: map[time.Weekday]string{
			time.Monday:    "09:00",
			time.Tuesday:   "28:00/29:00",
			time.Wednesday: "10:00/00:00",
		},
	}

	_, err := NewScheduleFromRaw(&raw)
	assert.Error(t, err, "Parsing should fail")
	assert.ErrorContains(t, err, "time range format should be")
	assert.ErrorContains(t, err, "hour out of range")
	assert.ErrorContains(t, err, "time range is inverse")
}

func TestScheduleFromRaw(t *testing.T) {
	raw := RawScheduleConfig{
		Days: map[time.Weekday]string{
			//Space after , is part of the test to check that we are trimming
			time.Monday:  "09:00/17:00, 18:00/19:00",
			time.Tuesday: "10:00/11:00",
		},
	}
	s, err := NewScheduleFromRaw(&raw)
	assert.Nil(t, err)
	sminute := 9 * 60
	eminute := 17 * 60
	assert.NotNil(t, s.Days[time.Monday].Minutes[sminute])
	assert.Nil(t, s.Days[time.Monday].Minutes[sminute-1])
	assert.NotNil(t, s.Days[time.Monday].Minutes[eminute])
	assert.Nil(t, s.Days[time.Monday].Minutes[eminute+1])
	assert.Nil(t, s.Days[time.Monday].Minutes[18*60-1])
	assert.Nil(t, s.Days[time.Monday].Minutes[18*60-1])
	assert.Nil(t, s.Days[time.Monday].Minutes[20*60])
	assert.NotNil(t, s.Days[time.Tuesday].Minutes[10*60])
	assert.Nil(t, s.Days[time.Tuesday].Minutes[10*60-1])
	assert.NotNil(t, s.Days[time.Tuesday].Minutes[10*60+1])
	assert.NotNil(t, s.Days[time.Tuesday].Minutes[11*60])

	t.Run("Test Frames", func(t *testing.T) {
		assert.False(t, s.HasDecentTimeframe(time.Wednesday))
		assert.Len(t, s.DecentTimeFrames(time.Monday), 2)

		frames := s.DecentTimeFrames(time.Monday)

		assert.Equal(t, frames[0], *s.Days[time.Monday].Minutes[sminute])
		assert.Equal(t, frames[1], *s.Days[time.Monday].Minutes[18*60])
	})
}

func TestScheduleFromPlainText(t *testing.T) {
	plainText := `[decent]
	     Monday = 09:00/13:00, 14:00/17:00
	     Tuesday=09:00/13:00, 14:00/17:00
	     Wednesday= 09:00/13:00, 14:00/17:00
	     Thursday =09:00/13:00, 14:00/17:00
	     Friday = 09:00/13:00, 14:00/17:00
	`

	s, err := NewScheduleFromPlainText(strings.NewReader(plainText))
	assert.NoError(t, err)

	assert.Equal(t, time.Monday, s.Days[time.Saturday].ClosestDecentDay)
	assert.Equal(t, time.Monday, s.Days[time.Monday].ClosestDecentDay)
	assert.Equal(t, time.Tuesday, s.Days[time.Tuesday].ClosestDecentDay)
	assert.Equal(t, time.Wednesday, s.Days[time.Wednesday].ClosestDecentDay)
	assert.Equal(t, time.Thursday, s.Days[time.Thursday].ClosestDecentDay)
	assert.Equal(t, time.Friday, s.Days[time.Friday].ClosestDecentDay)
	assert.NotNil(t, s.Days[time.Monday].Minutes[60*9])
	assert.NotNil(t, s.Days[time.Tuesday].Minutes[60*10])
	assert.NotNil(t, s.Days[time.Wednesday].Minutes[60*13])
	assert.NotNil(t, s.Days[time.Thursday].Minutes[60*14])
	assert.NotNil(t, s.Days[time.Friday].Minutes[60*17])
	assert.Nil(t, s.Days[time.Saturday].Minutes[60*9])
}
func TestClosestDecentDay(t *testing.T) {
	raw := RawScheduleConfig{
		Days: map[time.Weekday]string{
			//Space after , is part of the test to check that we are trimming
			time.Monday:    "10:00/11:00",
			time.Wednesday: "10:00/11:00",
			time.Friday:    "10:00/11:00",
		},
	}

	schedule, err := NewScheduleFromRaw(&raw)
	require.NoError(t, err)

	d, nDays := schedule.ClosestDecentDay(time.Monday)
	assert.Equal(t, d, time.Monday)
	assert.Equal(t, nDays, 0)

	d, nDays = schedule.ClosestDecentDay(time.Tuesday)
	assert.Equal(t, d, time.Wednesday)
	assert.Equal(t, nDays, 1)

	d, nDays = schedule.ClosestDecentDay(time.Wednesday)
	assert.Equal(t, d, time.Wednesday)
	assert.Equal(t, nDays, 0)

	d, nDays = schedule.ClosestDecentDay(time.Thursday)
	assert.Equal(t, d, time.Friday)
	assert.Equal(t, nDays, 1)

	d, nDays = schedule.ClosestDecentDay(time.Saturday)
	assert.Equal(t, d, time.Monday)
	assert.Equal(t, nDays, 2)
}

func TestClosestDecentMinute(t *testing.T) {
	raw := RawScheduleConfig{
		Days: map[time.Weekday]string{
			time.Monday:    "10:00/11:00, 13:00/14:00",
			time.Wednesday: "10:00/11:00",
			time.Friday:    "10:00/11:00",
		},
	}

	schedule, err := NewScheduleFromRaw(&raw)
	require.NoError(t, err)

	layout := "2006-01-02 15:04:05"
	pTime, err := time.Parse(layout, "2024-01-28 18:30:00")
	assert.NoError(t, err)

	minute, nMins := schedule.ClosestDecentMinute(pTime)
	assert.Equal(t, 930, nMins)
	assert.Equal(t, 10*60, minute)

	pTime, err = time.Parse(layout, "2024-01-29 09:59:00")
	assert.NoError(t, err)

	minute, nMins = schedule.ClosestDecentMinute(pTime)
	assert.Equal(t, 1, nMins)
	assert.Equal(t, 10*60, minute)

	pTime, err = time.Parse(layout, "2024-01-29 11:01:00")
	assert.NoError(t, err)

	minute, nMins = schedule.ClosestDecentMinute(pTime)
	assert.Equal(t, 2*60-1, nMins)
	assert.Equal(t, 13*60, minute)

	pTime, err = time.Parse(layout, "2024-01-29 14:01:00")
	assert.NoError(t, err)

	minute, nMins = schedule.ClosestDecentMinute(pTime)
	assert.Equal(t, 2639, nMins)
	assert.Equal(t, 10*60, minute)
}

func BenchmarkNewScheduleFromRaw(b *testing.B) {
	raw := RawScheduleConfig{
		Days: map[time.Weekday]string{
			time.Monday:    "10:00/11:00, 13:00/14:00",
			time.Wednesday: "10:00/11:00",
			time.Friday:    "10:00/11:00",
		},
	}

	for i := 0; i < b.N; i++ {
		NewScheduleFromRaw(&raw)
	}
}
