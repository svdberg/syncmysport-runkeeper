package runkeeper

import (
	"fmt"
	"log"
	"time"

	runkeeper "github.com/svdberg/go-runkeeper"
)

type RunkeeperCientInt interface {
	PostActivity(activity *runkeeper.FitnessActivityNew) (string, error)
	EnrichRKActivity(activitySummary *runkeeper.FitnessActivity) (*runkeeper.FitnessActivity, error)
	EnrichRKActivities(activities *runkeeper.FitnessActivityFeed) ([]runkeeper.FitnessActivity, error)
	GetRKActivitiesSince(timestamp int) (*runkeeper.FitnessActivityFeed, error)
	ValidateToken(token string) bool
	DeAuthorize(token string) error
}

type RkClient struct {
	BearerToken string
	Client      *runkeeper.Client
}

func CreateRKClient(token string) RunkeeperCientInt {
	client := runkeeper.NewClient(token)
	return &RkClient{token, client}
}

func (c RkClient) PostActivity(activity *runkeeper.FitnessActivityNew) (string, error) {
	return c.Client.PostNewFitnessActivity(activity)
}

func (c RkClient) EnrichRKActivity(activitySummary *runkeeper.FitnessActivity) (*runkeeper.FitnessActivity, error) {
	var params runkeeper.Params
	params = make(map[string]interface{})
	return c.Client.GetFitnessActivity(activitySummary.Uri, &params)
}

//EnrichRKActivities Set the lat long for any activities we downloaded
func (c RkClient) EnrichRKActivities(activities *runkeeper.FitnessActivityFeed) ([]runkeeper.FitnessActivity, error) {
	if activities == nil {
		return nil, fmt.Errorf("passed in activities struct is nil")
	}
	result := make([]runkeeper.FitnessActivity, activities.Size)
	for i, act := range activities.Items {
		a, err := c.EnrichRKActivity(&act)
		if err != nil {
			return nil, fmt.Errorf("nil activity requested for enrichemnt")
		}
		result[i] = *a
	}
	return result, nil
}

func (c RkClient) GetRKActivitiesSince(timestamp int) (*runkeeper.FitnessActivityFeed, error) {
	//int to timestamp
	var ts int64
	ts = int64(timestamp)
	tm := time.Unix(ts, 0)
	var params runkeeper.Params
	params = make(map[string]interface{})
	params["noEarlierThan"] = tm.Format("2006-01-02")
	return c.Client.GetFitnessActivityFeed(&params)
}

func (c RkClient) ValidateToken(token string) bool {
	c.BearerToken = token
	_, err := c.Client.GetUser()
	if err != nil {
		log.Printf("(Expected) error while validating token %s : %s", token, err)
		return false
	}
	return true
}

func (c RkClient) DeAuthorize(token string) error {
	return c.Client.Deauthorize()
}
