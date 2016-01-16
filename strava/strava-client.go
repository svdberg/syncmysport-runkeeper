package strava

import (
	"fmt"
	stravalib "github.com/strava/go.strava"
)

type StravaDetailed stravalib.ActivityDetailed

func GetSTVActivitiesSince(bearerToken string, timestamp int) ([]*stravalib.ActivitySummary, error) {
	client := stravalib.NewClient(bearerToken)
	service := stravalib.NewCurrentAthleteService(client)
	call := service.ListActivities()
	call.After(timestamp)
	activities, err := call.Do()
	return activities, err
}

func GetSTVDetailedActivity(bearerToken string, activityId int64) (*stravalib.ActivityDetailed, error) {
	client := stravalib.NewClient(bearerToken)
	service := stravalib.NewActivitiesService(client)
	call := service.Get(activityId)
	return call.Do()
}

func (da StravaDetailed) String() string {
	return fmt.Sprintf("{Id: %d, Name: '%s'}", da.Id, da.Name)
}
