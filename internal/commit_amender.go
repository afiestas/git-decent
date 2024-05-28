package internal

import (
	"time"

	"github.com/afiestas/git-decent/config"
)

func Amend(date time.Time, lastDate *time.Time, schedule config.Schedule) time.Time {
	//When it is the first commit, just look for the closest decent frame
	if lastDate == nil {
		_, dMin := schedule.ClosestDecentMinute(date)
		return date.Add(time.Duration(dMin) * time.Minute)
	}

	//Instead of generating random numbers we use the already humandly random generated
	//minute when the commit was created
	min := date.Minute() % 10
	if min == 0 {
		min = randomMinute()
	}

	//If the commit being amended is before the previous commit, move it just after it
	if lastDate.After(date) {
		date = *lastDate
		date = date.Add(time.Duration(min) * time.Minute)
	}

	_, dMin := schedule.ClosestDecentMinute(date)
	if dMin > 0 {
		//Ass the noise to avoid many commits with time 0
		date = date.Add(time.Duration(dMin+min) * time.Minute)
	}

	return date
}
