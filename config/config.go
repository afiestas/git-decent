/* SPDX-License-Identifier: MIT */
package config

import (
	"fmt"
	"time"

	"github.com/go-git/go-git/v5/config"
)

const section = "decent"

type RawScheduleConfig struct {
	Days map[time.Weekday]string
}

func (config *RawScheduleConfig) SetValue(day string, value string) error {
	switch day {
	case "Monday":
		config.Days[time.Monday] = value
	case "Tuesday":
		config.Days[time.Tuesday] = value
	case "Wednesday":
		config.Days[time.Wednesday] = value
	case "Thursday":
		config.Days[time.Thursday] = value
	case "Friday":
		config.Days[time.Friday] = value
	case "Saturday":
		config.Days[time.Saturday] = value
	case "Sunday":
		config.Days[time.Sunday] = value
	default:
		return fmt.Errorf("invalid day configured, got %s with value %s", day, value)
	}
	return nil
}

func GetGitRawConfig(c *config.Config) (RawScheduleConfig, error) {
	rawC := RawScheduleConfig{
		Days: make(map[time.Weekday]string),
	}

	if !c.Raw.HasSection(section) {
		return rawC, fmt.Errorf("can't find %s section in git config", section)
	}

	if len(c.Raw.Section("decent").Options) == 0 {
		return rawC, fmt.Errorf("section %s is empty, no schedule found", section)
	}

	o := c.Raw.Section("decent").Options
	for _, day := range o {
		rawC.SetValue(day.Key, day.Value)
	}

	return rawC, nil
}

func ParseRawConfig(rawC *RawScheduleConfig) error {

	return nil
}
