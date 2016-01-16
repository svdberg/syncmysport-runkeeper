package strava

import (
	stravalib "github.com/strava/go.strava"
	dm "github.com/svdberg/syncmysport-runkeeper/datamodel"
)

func convertToActivity(stravaActivity *stravalib.ActivityDetailed) *dm.Activity {
	return nil
}
