package strava

import (
	stravalib "github.com/strava/go.strava"
	dm "github.com/svdberg/syncmysport-runkeeper/datamodel"
)

func ConvertToActivity(stravaActivity *stravalib.ActivityDetailed) *dm.Activity {
	stvActivity := dm.Activity{}
	stvActivity.StartTime = int(stravaActivity.StartDate.Unix())
	stvActivity.Duration = stravaActivity.ElapsedTime
	stvActivity.Name = stravaActivity.Name
	return &stvActivity
}
