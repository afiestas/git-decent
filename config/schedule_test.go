package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
}
