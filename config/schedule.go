/* SPDX-License-Identifier: MIT */
package config

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type TimeFrame struct {
	StartMinute int
	EndMinute   int
	nextFrame   *TimeFrame
}

// A list of day minutes that are work times
type DayMinutes [1440]*TimeFrame
type Day struct {
	Minutes      DayMinutes
	DecentFrames []*TimeFrame
}
type Schedule struct {
	Days [7]Day
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

			timeFrame := TimeFrame{StartMinute: sMinute, EndMinute: eMinute}
			if l := len(s.Days[d].DecentFrames); l > 0 {
				s.Days[d].DecentFrames[l-1].nextFrame = &timeFrame
			}
			s.Days[d].DecentFrames = append(s.Days[d].DecentFrames, &timeFrame)

			for m := sMinute; m <= eMinute; m++ {
				s.Days[d].Minutes[m] = &timeFrame
			}

		}
	}

	if len(errs) > 0 {
		return s, errors.Join(errs...)
	}
	return s, nil
}

func (s *Schedule) HasDecentTimeframe(day time.Weekday) bool {
	return false
}

func (s *Schedule) DecentTimeFrames(day time.Weekday) []*TimeFrame {
	return s.Days[day].DecentFrames
}
