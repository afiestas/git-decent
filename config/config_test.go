/* SPDX-License-Identifier: MIT */
package config

import (
	"testing"
	"time"

	_ "github.com/afiestas/git-decent/testutils"
	"github.com/stretchr/testify/assert"
)

func TestGetGitRawConfig(t *testing.T) {

	options := map[string]string{
		"monday":  "09:00/17:00, 18:00/19:00",
		"tuesday": "10:00/11:00",
	}
	rawC, err := GetGitRawConfig(&options)
	assert.Nil(t, err, "No error is expected")

	expectedRawC := RawScheduleConfig{
		Days: map[time.Weekday]string{
			time.Monday:  "09:00/17:00, 18:00/19:00",
			time.Tuesday: "10:00/11:00",
		},
	}

	assert.Equal(t, rawC, expectedRawC)
}
