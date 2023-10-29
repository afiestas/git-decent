/* SPDX-License-Identifier: MIT */
package config

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/config"
	"github.com/stretchr/testify/assert"
)

func TestGetGitRawConfig(t *testing.T) {
	c := config.NewConfig()
	_, err := GetGitRawConfig(c)
	assert.ErrorContains(t, err, "section in git config")

	input := []byte(`[decent]`)
	err = c.Unmarshal(input)
	assert.Nil(t, err)

	_, err = GetGitRawConfig(c)
	assert.ErrorContains(t, err, "is empty, no schedule found")

	input = []byte(`[decent]
		Monday = 09:00/17:00, 18:00/19:00
		Tuesday = 10:00/11:00
	`)
	err = c.Unmarshal(input)
	assert.Nil(t, err)
	rawC, err := GetGitRawConfig(c)
	assert.Nil(t, err, "No error is expected")

	expectedRawC := RawScheduleConfig{
		Days: map[time.Weekday]string{
			time.Monday:  "09:00/17:00, 18:00/19:00",
			time.Tuesday: "10:00/11:00",
		},
	}

	assert.Equal(t, rawC, expectedRawC)
}

func TestScheduleFromRawError(t *testing.T) {
	raw := RawScheduleConfig{
		Days: map[time.Weekday]string{
			time.Monday:    "09:00",
			time.Tuesday:   "28:00/29:00",
			time.Wednesday: "10:00/00:00",
		},
	}

	_, err := NewScheduleFromRaw(&raw)
	fmt.Println(err)
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
	assert.Equal(t, s.Days[time.Monday][sminute], true)
	assert.Equal(t, s.Days[time.Monday][sminute-1], false)
	assert.Equal(t, s.Days[time.Monday][eminute], true)
	assert.Equal(t, s.Days[time.Monday][eminute+1], false)
	assert.Equal(t, s.Days[time.Monday][18*60-1], false)
	assert.Equal(t, s.Days[time.Monday][18*60-1], false)
	assert.Equal(t, s.Days[time.Monday][20*60], false)
	assert.Equal(t, s.Days[time.Tuesday][10*60], true)
	assert.Equal(t, s.Days[time.Tuesday][10*60-1], false)
	assert.Equal(t, s.Days[time.Tuesday][10*60+1], true)
	assert.Equal(t, s.Days[time.Tuesday][11*60], true)
}
