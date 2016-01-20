package runkeeper

import (
	runkeeper "github.com/c9s/go-runkeeper"
	"time"
)

type RkClient struct {
	BearerToken string
	Client      *runkeeper.Client
}

func CreateRKClient(token string) *RkClient {
	client := runkeeper.NewClient(token)
	return &RkClient{token, client}
}

func (c RkClient) PostActivity(activity *runkeeper.FitnessActivityNew) (string, error) {
	return c.Client.PostNewFitnessActivity(activity)
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
