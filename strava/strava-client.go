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

func GetSTVActivityStream(bearerToken string, activityId int64, streamType string) (*stravalib.StreamSet, error) {
	client := stravalib.NewClient(bearerToken)
	service := stravalib.NewActivityStreamsService(client)
	var types = make([]stravalib.StreamType, 1)
	if streamType == "GPS" {
		types = append(types, stravalib.StreamTypes.Location)
	} else if streamType == "Heartrate" {
		types = append(types, stravalib.StreamTypes.HeartRate)
	}
	return service.Get(activityId, types).Resolution("10000").SeriesType("distance").Do()
}

func (da StravaDetailed) String() string {
	return fmt.Sprintf("{Id: %d, Name: '%s'}", da.Id, da.Name)
}
