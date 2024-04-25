package internal

import (
	"fmt"
	"math"
	"time"

	"github.com/afiestas/git-decent/config"
)

func Amend(commit Commit, log GitLog, schedule config.Schedule) Commit {
	date := commit.Date
	fmt.Println("Starting amend:", date)
	dm := config.DayMinute(date)
	dDay, frame, nDays := schedule.ClosestDecentFrame(date)
	sDay := schedule.Days[dDay]
	if commit.Prev == nil && dDay == date.Weekday() && sDay.Minutes[dm] != nil {
		fmt.Println("It-s all good, returning")
		return commit
	}

	//If it has a before commit, take its date as the base
	// Then see if compression or recolocation is needed, if so apply
	// Check if date is within a good DEcentFrame, if nto look for the next and take that as base

	//If not, take the commit date.
	fmt.Printf("Decent Frame: %s\n", sDay.DecentFrames[frame])
	fmt.Printf("Initial: %s(%d), Current: %s(%d), Elapsed: %d, Original date: %s\n", date.Weekday(), date.Weekday(), dDay, dDay, nDays, commit.Date)

	decentFrame := sDay.DecentFrames[frame]
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
	} else {
		fmt.Println("Got previous commit, checking it")
		pDay := commit.Prev.Date.Weekday()
		if dDay != pDay {
			fmt.Println("Day is different, so jsut returning", dDay, pDay)
			commit.Date = proposedDate
			return commit
		}

		pMinute := config.DayMinute(commit.Prev.Date)
		fmt.Println("AAAA", commit.Prev.Date, pDay, pMinute)
		prevFrame := schedule.Days[pDay].Minutes[pMinute]
		if prevFrame == nil || decentFrame.StartMinute != prevFrame.StartMinute {
			fmt.Println("Frame is different, so jsut returning", decentFrame, prevFrame)
			commit.Date = proposedDate
			return commit
		}

		fmt.Println("Same day samne frame", dDay, decentFrame)
	}

	minutesToAdd := int(commit.Date.Sub(commit.Prev.Date).Minutes())

	fmt.Println("DATE DIFF", minutesToAdd, commit.Date, commit.Prev.Date)
	if minutesToAdd > 30 {
		minutesToAdd = minutesToAdd % 10
	} else if minutesToAdd < 0 {
		minutesToAdd = int(math.Abs(float64(minutesToAdd)))
		cMin := commit.Date.Minute() % 10
		if cMin == 0 {
			cMin = 1
		}
		minutesToAdd += cMin
		fmt.Println("IN THE PAST", hour, minute, minutesToAdd)
	} else {
		return commit
	}

	proposedDate = commit.Date.Add(time.Duration(minutesToAdd) * time.Minute)

	fmt.Println("Proposed date", proposedDate)
	commit.Date = proposedDate
	return commit
}
