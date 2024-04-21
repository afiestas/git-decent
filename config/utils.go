package config

import "time"

func DayMinute(t time.Time) int {
	return t.Hour()*60 + t.Minute()
}
