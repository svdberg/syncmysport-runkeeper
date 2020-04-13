package strava

import (
	"fmt"
	"strings"

	stravalib "github.com/strava/go.strava"
)

type StravaClientInt interface {
	GetSTVActivitiesSince(timestamp int) ([]*stravalib.ActivitySummary, error)
	GetSTVDetailedActivity(activityId int64) (*stravalib.ActivityDetailed, error)
	GetSTVActivityStream(activityId int64, streamType string) (*stravalib.StreamSet, error)
	ValidateToken(token string) bool
	RefreshToken(refresh_token string) (string, string, error)
	DeAuthorize(token string) error
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

func CreateStravaClientWithSecretAndId(token string, clientId int, secret string) StravaClientInt {
	stravalib.ClientSecret = secret
	stravalib.ClientId = clientId
	client := stravalib.NewClient(token)
	return &StravaClient{token, client}
}

func (c StravaClient) RefreshToken(refresh_token string) (string, string, error) {
	access_token, refresh_token, err := stravalib.NewOAuthService(c.Client).RefreshToken(refresh_token).Do()
	if err != nil {
		return "", "", err
	}
	return access_token, refresh_token, nil
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
	//improve this so we collect all streams in one call
	service := stravalib.NewActivityStreamsService(c.Client)
	var types = make([]stravalib.StreamType, 1)
	if streamType == "GPS" {
		types = append(types, stravalib.StreamTypes.Location)
	} else if streamType == "Heartrate" {
		types = append(types, stravalib.StreamTypes.HeartRate)
	} else if streamType == "Time" {
		types = append(types, stravalib.StreamTypes.Time)
	} else if streamType == "Altitude" {
		types = append(types, stravalib.StreamTypes.Elevation)
	}
	stream, err := service.Get(activityId, types).Do()
	if err != nil && strings.Contains(err.Error(), "Record Not Found") {
		return nil, nil
	} else if err == nil {
		return stream, nil
	} else {
		return nil, err
	}
}

func (c StravaClient) ValidateToken(token string) bool {
	//create a new client to validate this token
	verifyClient := stravalib.NewClient(token)
	service := stravalib.NewCurrentAthleteService(verifyClient)
	_, err := service.Get().Do()
	if err != nil {
		return false
	}
	return true
}

func (c StravaClient) DeAuthorize(token string) error {
	deAuthClient := stravalib.NewClient(token)
	return stravalib.NewOAuthService(deAuthClient).Deauthorize().Do()
}

func (da StravaDetailed) String() string {
	return fmt.Sprintf("{Id: %d, Name: '%s'}", da.Id, da.Name)
}
