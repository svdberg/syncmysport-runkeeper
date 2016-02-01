package runkeeper

import (
	runkeeper "github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/c9s/go-runkeeper"
	"time"
)

type RunkeeperCientInt interface {
	PostActivity(activity *runkeeper.FitnessActivityNew) (string, error)
	EnrichRKActivity(activitySummary *runkeeper.FitnessActivity) (*runkeeper.FitnessActivity, error)
	EnrichRKActivities(activities *runkeeper.FitnessActivityFeed) []runkeeper.FitnessActivity
	GetRKActivitiesSince(timestamp int) (*runkeeper.FitnessActivityFeed, error)
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

func (c RkClient) EnrichRKActivities(activities *runkeeper.FitnessActivityFeed) []runkeeper.FitnessActivity {
	result := make([]runkeeper.FitnessActivity, activities.Size)
	for i, act := range activities.Items {
		a, _ := c.EnrichRKActivity(&act)
		result[i] = *a
	}
	return result
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
