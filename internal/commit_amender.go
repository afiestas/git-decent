package internal

import (
	"fmt"
	"time"

	"github.com/afiestas/git-decent/config"
	"github.com/afiestas/git-decent/utils"
)

func Amend(date time.Time, lastDate *time.Time, schedule config.Schedule) time.Time {
	db := utils.DebugBlock{Title: "⏰ Amending: " + date.String()}
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
	min := date.Minute() % 10
	if min == 0 {
		min = randomMinute()
	}

	//If the commit being amended is before the previous commit, move it just after it
	if lastDate.After(date) {
		date = *lastDate
		db.AddLine("LastDate is after:", "true")
		db.AddLine("AddedNoise:", fmt.Sprint(min))
		date = date.Add(time.Duration(min) * time.Minute)
	}

	_, dMin := schedule.ClosestDecentMinute(date)
	if dMin > 0 {
		//Add the noise to avoid many commits with time 0
		db.AddLine("AddedNoise:", fmt.Sprint(min))
		date = date.Add(time.Duration(dMin+min) * time.Minute)
	}

	db.Print()
	return date
}
