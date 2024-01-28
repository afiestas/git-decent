/* SPDX-License-Identifier: MIT */
package config

import (
	"fmt"
	"time"
)

const section = "decent"

type RawScheduleConfig struct {
	Days map[time.Weekday]string
}

func (config *RawScheduleConfig) SetValue(day string, value string) error {
	switch day {
	case "monday":
		config.Days[time.Monday] = value
	case "tuesday":
		config.Days[time.Tuesday] = value
	case "wednesday":
		config.Days[time.Wednesday] = value
	case "thursday":
		config.Days[time.Thursday] = value
	case "friday":
		config.Days[time.Friday] = value
	case "saturday":
		config.Days[time.Saturday] = value
	case "sunday":
		config.Days[time.Sunday] = value
	default:
		return fmt.Errorf("invalid day configured, got %s with value %s", day, value)
	}
	return nil
}

func GetGitRawConfig(options *map[string]string) (RawScheduleConfig, error) {
	rawC := RawScheduleConfig{
		Days: make(map[time.Weekday]string),
	}

	for day, value := range *options {
		rawC.SetValue(day, value)
	}

	return rawC, nil
}
