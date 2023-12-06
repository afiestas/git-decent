/* SPDX-License-Identifier: MIT */
package config

import (
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
