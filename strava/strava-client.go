package strava

import (
	"fmt"
	stravalib "github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/strava/go.strava"
	"strings"
)

type StravaClientInt interface {
	GetSTVActivitiesSince(timestamp int) ([]*stravalib.ActivitySummary, error)
	GetSTVDetailedActivity(activityId int64) (*stravalib.ActivityDetailed, error)
	GetSTVActivityStream(activityId int64, streamType string) (*stravalib.StreamSet, error)
}

type StravaDetailed stravalib.ActivityDetailed

type StravaClient struct {
	BearerToken string
	Client      *stravalib.Client
}

func CreateStravaClient(token string) StravaClientInt {
	client := stravalib.NewClient(token)
	return &StravaClient{token, client}
}

func (c StravaClient) GetSTVActivitiesSince(timestamp int) ([]*stravalib.ActivitySummary, error) {
	service := stravalib.NewCurrentAthleteService(c.Client)
	call := service.ListActivities()
	call.After(timestamp)
	activities, err := call.Do()
	return activities, err
}

func (c StravaClient) GetSTVDetailedActivity(activityId int64) (*stravalib.ActivityDetailed, error) {
	service := stravalib.NewActivitiesService(c.Client)
	call := service.Get(activityId)
	return call.Do()
}

func (c StravaClient) GetSTVActivityStream(activityId int64, streamType string) (*stravalib.StreamSet, error) {
	service := stravalib.NewActivityStreamsService(c.Client)
	var types = make([]stravalib.StreamType, 1)
	if streamType == "GPS" {
		types = append(types, stravalib.StreamTypes.Location)
	} else if streamType == "Heartrate" {
		types = append(types, stravalib.StreamTypes.HeartRate)
	} else if streamType == "Time" {
		types = append(types, stravalib.StreamTypes.Time)
	}
	stream, err := service.Get(activityId, types).Resolution("high").SeriesType("distance").Do()
	if err != nil && strings.Contains(err.Error(), "Record Not Found") {
		return nil, nil
	} else if err == nil {
		return stream, nil
	} else {
		return nil, err
	}
}

func (da StravaDetailed) String() string {
	return fmt.Sprintf("{Id: %d, Name: '%s'}", da.Id, da.Name)
}
