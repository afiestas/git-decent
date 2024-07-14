package internal

import (
	"fmt"
	"math"
	"time"

	"github.com/afiestas/git-decent/config"
	"github.com/afiestas/git-decent/utils"
)

func Amend(date time.Time, lastDate *time.Time, lastRealDate *time.Time, threshold int, schedule config.Schedule) time.Time {
	db := utils.DebugBlock{Title: "⏰ Amending: " + date.String()}
	db.AddLine("Threshold", fmt.Sprint(threshold))

	//When it is the first commit, just look for the closest decent frame
	if lastDate == nil {
		db.AddLine("LastDate:", "None")
		_, dMin := schedule.ClosestDecentMinute(date)
		db.AddLine("ClosestDecentMinute:", fmt.Sprint(dMin))
		db.Print()
		return date.Add(time.Duration(dMin) * time.Minute)
	}

	db.AddLine("LastDate:", lastDate.String())

	//Instead of generating random numbers we use the already humandly random generated
	//minute when the commit was created
	noise := date.Minute() % 10
	if noise == 0 {
		noise = randomMinute()
	}

	interval := 0
	if lastRealDate != nil {
		db.AddLine("LastRealDate:", lastRealDate.String())
		interval = int(math.Floor(date.Sub(*lastRealDate).Minutes()))
	}

	db.AddLine(fmt.Sprintf("Interval: %d", interval))

	//If the interval exceeds the threshold or if it is negative or equal to 0
	if interval > threshold || interval <= 0 {
		db.AddLine("Bad internval", fmt.Sprint(interval))
		interval = noise
	}

	//If the commit being amended is before the previous commit, move it just after it
	if lastDate.After(date) {
		date = *lastDate
		db.AddLine("LastDate is after:", "true")
		db.AddLine("Interval added:", fmt.Sprint(interval))
		date = date.Add(time.Duration(interval) * time.Minute)
	}

	_, dMin := schedule.ClosestDecentMinute(date)
	if dMin > 0 {
		//Add the noise to avoid many commits with time 0
		db.AddLine("Moved to a different frame", fmt.Sprint(dMin))
		db.AddLine("AddedNoise:", fmt.Sprint(noise))
		date = date.Add(time.Duration(dMin+noise) * time.Minute)
	}

	db.AddLine("Final date", date.String())
	db.Print()
	return date
}
