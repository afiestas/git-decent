package internal

import (
	"fmt"
	"time"

	"github.com/afiestas/git-decent/config"
)

func Amend(commit Commit, log GitLog, schedule config.Schedule) Commit {
	date := commit.Date
	dm := config.DayMinute(date)
	dDay, nDays := schedule.ClosestDecentDay(date.Weekday())
	sDay := schedule.Days[dDay]
	if dDay == date.Weekday() && sDay.Minutes[dm] != nil {
		return commit
	}

	//If it has a before commit, take its date as the base
	// Then see if compression or recolocation is needed, if so apply
	// Check if date is within a good DEcentFrame, if nto look for the next and take that as base

	//If not, take the commit date.
	fmt.Printf("Decent Frame: %s\n", sDay.DecentFrames[0])
	fmt.Printf("Initial: %s(%d), Current: %s(%d), Elapsed: %d, Original date: %s\n", date.Weekday(), date.Weekday(), dDay, dDay, nDays, commit.Date)
	decentFrame := sDay.DecentFrames[0]
	hour := decentFrame.StartMinute / 60
	minute := decentFrame.StartMinute % 60
	fmt.Println("Hour", hour, "min", minute)
	proposedDate := time.Date(
		date.Year(),
		date.Month(),
		date.Day()+nDays,
		hour,
		minute,
		date.Second(),
		date.Nanosecond(),
		date.Location(),
	)

	if commit.Prev == nil {
		commit.Date = proposedDate
		return commit
	}

	dateDiff := int(commit.Date.Sub(commit.Prev.Date).Minutes())

	fmt.Println("DATE DIFF", dateDiff, commit.Prev.Date)
	if dateDiff > 30 {
		minute += dateDiff % 10
	} else {
		minute += dateDiff
	}

	proposedDate = time.Date(
		date.Year(),
		date.Month(),
		date.Day()+nDays,
		hour,
		minute,
		date.Second(),
		date.Nanosecond(),
		date.Location(),
	)

	commit.Date = proposedDate
	return commit
}
