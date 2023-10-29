/* SPDX-License-Identifier: MIT */
package config

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// A list of day minutes that are work times
type Schedule struct {
	Days [7][1440]bool
}

type ParseDayError struct {
	error
	Day time.Weekday
}

func NewScheduleFromRaw(config *RawScheduleConfig) (Schedule, error) {
	errs := []error{}

	s := Schedule{}
	for d := time.Sunday; d <= time.Saturday; d++ {
		v, exists := config.Days[d]
		if !exists {
			continue
		}

		tRange := strings.Split(v, ",")
		for _, r := range tRange {
			hMinutes := strings.Split(r, "/")
			if len(hMinutes) < 2 {
				errs = append(
					errs,
					ParseDayError{
						Day:   d,
						error: fmt.Errorf("time range format should be 04:00/18:00 but instead %s given", r),
					},
				)
				continue
			}

			sMin := strings.TrimSpace(hMinutes[0])
			sTime, err := time.Parse("15:04", sMin)
			if err != nil {
				errs = append(
					errs,
					err,
				)
				continue
			}

			eMin := strings.TrimSpace(hMinutes[1])
			eTime, err := time.Parse("15:04", eMin)
			if err != nil {
				errs = append(
					errs,
					err,
				)
				continue
			}

			if sTime.After(eTime) {
				errs = append(errs, fmt.Errorf("time range is inverse: %s is after %s", sMin, eMin))
			}

			sMinute := sTime.Hour()*60 + sTime.Minute()
			eMinute := eTime.Hour()*60 + eTime.Minute()

			for m := sMinute; m <= eMinute; m++ {
				s.Days[d][m] = true
			}
		}
	}

	if len(errs) > 0 {
		return s, errors.Join(errs...)
	}
	return s, nil
}
