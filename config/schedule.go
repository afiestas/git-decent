/* SPDX-License-Identifier: MIT */
package config

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

type TimeFrame struct {
	StartMinute int
	EndMinute   int
}

type HourMin struct {
	hour   int
	minute int
}

func (t TimeFrame) String() string {
	return fmt.Sprintf("%02d:%02d - %02d:%02d", t.StartMinute/60, t.StartMinute%60, t.EndMinute/60, t.EndMinute%60)
}

// A list of day minutes that are work times
type DayMinutes [1440]*TimeFrame
type Day struct {
	Minutes          DayMinutes
	DecentFrames     []TimeFrame
	ClosestDecentDay time.Weekday
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
	first := -1
	daysWithout := 0
	for d := time.Sunday; d <= time.Saturday; d++ {
		v, exists := config.Days[d]
		if !exists {
			daysWithout++
			continue
		}

		if first == -1 {
			first = int(d)
		}
		s.Days[d].ClosestDecentDay = d
		tRange := parseFrames(v)
		s.Days[d].DecentFrames = make([]TimeFrame, len(tRange))
		for k, r := range tRange {
			if len(r) != 11 {
				err := fmt.Errorf("time range format should be 04:00/18:00 but instead %s given", r)
				errs = append(errs, ParseDayError{Day: d, error: err})
				continue
			}

			sMin := r[:5]
			sTime, err := parseTime(sMin)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			eMin := r[6:]
			eTime, err := parseTime(eMin)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			sMinute := sTime.hour*60 + sTime.minute
			eMinute := eTime.hour*60 + eTime.minute

			if sMinute > eMinute {
				errs = append(errs, fmt.Errorf("time range is inverse: %s is after %s", sMin, eMin))
			}

			timeFrame := TimeFrame{StartMinute: sMinute, EndMinute: eMinute}
			s.Days[d].DecentFrames[k] = timeFrame

			for m := sMinute; m <= eMinute; m++ {
				s.Days[d].Minutes[m] = &timeFrame
			}

		}
		for x := daysWithout; x > 0; x-- {
			s.Days[int(d)-x].ClosestDecentDay = d
		}
		daysWithout = 0
	}
	if len(s.Days[time.Saturday].DecentFrames) == 0 {
		s.Days[time.Saturday].ClosestDecentDay = time.Weekday(first)
	}

	if len(errs) > 0 {
		return s, errors.Join(errs...)
	}
	return s, nil
}

// "10:00/11:00, 13:00/14:00"
func parseFrames(rawFrames string) []string {
	var frames []string
	s := 0
	for c := 0; c < len(rawFrames); c++ {
		if rawFrames[c] == ' ' {
			s = c + 1
			continue
		}
		if rawFrames[c] == ',' {
			frames = append(frames, rawFrames[s:c])
			s = c + 1
		}
	}

	frames = append(frames, rawFrames[s:])
	return frames
}

// 15:04
func parseTime(time string) (HourMin, error) {
	hourMin := HourMin{}

	if len(time) > 5 || time[2] != ':' {
		return hourMin, fmt.Errorf("incorrect time format, expected 15:04 but given %s", time)
	}

	hour, err := strconv.Atoi(time[:2])
	if err != nil {
		return hourMin, fmt.Errorf("couldn't parse hour: %s error: %w", time[:2], err)
	}

	min, err := strconv.Atoi(time[3:])
	if err != nil {
		return hourMin, fmt.Errorf("couldn't parse minutes: %s error: %w", time[:3], err)
	}

	if hour < 0 || hour > 23 {
		return hourMin, fmt.Errorf("hour out of range, hour: %d", hour)
	}

	if min < 0 || min > 59 {
		return hourMin, fmt.Errorf("time out of range, hour: minute: %d", min)
	}

	hourMin.hour = hour
	hourMin.minute = min

	return hourMin, nil
}

func (s *Schedule) HasDecentTimeframe(day time.Weekday) bool {
	return false
}

func (s *Schedule) DecentTimeFrames(day time.Weekday) []TimeFrame {
	return s.Days[day].DecentFrames
}

func (s *Schedule) ClosestDecentDay(day time.Weekday) (time.Weekday, int) {
	next := s.Days[day].ClosestDecentDay
	return next, (int(next-day) + 7) % 7
}

func (s *Schedule) ClosestDecentMinute(date time.Time) (int, int) {
	wDay := date.Weekday()
	dMin := DayMinute(date)
	frameMin := s.Days[wDay].Minutes[dMin]
	if frameMin != nil {
		return dMin, 0
	}

	for _, frame := range s.Days[wDay].DecentFrames {
		if dMin <= frame.EndMinute {
			return frame.StartMinute, frame.StartMinute - dMin
		}
	}

	day, nDay := s.ClosestDecentDay(date.Weekday() + 1)
	nDay++
	frame := &s.Days[day].DecentFrames[0]
	hoursToaDd := 24 - date.Hour()

	hoursToaDd += (nDay - 1) * 24
	return frame.StartMinute, hoursToaDd*60 + frame.StartMinute - date.Minute()

}

func (s Schedule) String() string {
	ss := ""
	for day, sch := range s.Days {
		ss = fmt.Sprintf("%s%s has %d decent frames next %s\n\t", ss, time.Weekday(day), len(sch.DecentFrames), sch.ClosestDecentDay)
		for _, timeFrame := range sch.DecentFrames {
			ss = fmt.Sprintf("%s%s, ", ss, timeFrame)
		}
		ss = fmt.Sprintf("%s\n", ss)
	}

	return ss
}
