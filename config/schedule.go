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

func (t TimeFrame) String() string {
	return fmt.Sprintf("%02d:%02d - %02d:%02d", t.StartMinute/60, t.StartMinute%60, t.EndMinute/60, t.EndMinute%60)
}

// A list of day minutes that are work times
type DayMinutes [1440]*TimeFrame
type Day struct {
	Minutes      DayMinutes
	DecentFrames []*TimeFrame
}

func (d Day) String() string {
	str := "Minutes With frame:"
	for k, v := range d.Minutes {
		if v != nil {
			str = fmt.Sprintf("%s %d", str, k)
		}
	}
	str = fmt.Sprintf("%s\nFrames", str)
	for _, v := range d.DecentFrames {
		str = fmt.Sprintf("%s\n\t%s", str, v)
	}

	return str
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

func (s *Schedule) ClosestDecentDay(day time.Weekday) (time.Weekday, int) {
	n := 0
	if len(s.Days[day].DecentFrames) > 0 {
		return day, n
	}

	for nextDay := (day + 1) % 7; nextDay != day; nextDay = (nextDay + 1) % 7 {
		n++
		if len(s.Days[nextDay].DecentFrames) > 0 {
			return nextDay, n
		}
	}

	return day, n
}

func (s *Schedule) ClosestDecentFrame(date time.Time) (time.Weekday, int, int) {
	day := date.Weekday()
	minute := DayMinute(date)
	closestDay, passedDays := s.ClosestDecentDay(day)
	fmt.Println("Closest day", day, closestDay, passedDays)
	if passedDays == 0 {
		fmt.Println("Passed day is 0, checking frames", date)
		for k, frame := range s.Days[day].DecentFrames {
			fmt.Println("\t", minute, frame.EndMinute)
			if minute <= frame.EndMinute {
				return day, k, 0
			}
		}

		closestDay, passedDays = s.ClosestDecentDay(day + 1)
		return closestDay, 0, passedDays + 1

	}

	return closestDay, 0, int(closestDay) - int(day)
}

func (s Schedule) String() string {
	ss := ""
	for day, sch := range s.Days {
		ss = fmt.Sprintf("%s%s has %d decent frames\n\t", ss, time.Weekday(day), len(sch.DecentFrames))
		for _, timeFrame := range sch.DecentFrames {
			ss = fmt.Sprintf("%s%s, ", ss, timeFrame)
		}
		ss = fmt.Sprintf("%s\n", ss)
	}

	return ss
}
