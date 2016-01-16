package main

import (
	runkeeper "github.com/c9s/go-runkeeper"
	"time"
)

//func PostActivity(activity *Activity, bearerToken string) {
//	client := runkeeper.NewClient(bearerToken)
//	client.
//}

func GetRKActivitiesSince(bearerToken string, timestamp int) (*runkeeper.FitnessActivityFeed, error) {
	//int to timestamp
	var ts int64
	ts = int64(timestamp)
	tm := time.Unix(ts, 0)
	var params runkeeper.Params
	params = make(map[string]interface{})
	params["noEarlierThan"] = tm.Format("2006-01-02")
	client := runkeeper.NewClient(bearerToken)
	return client.GetFitnessActivityFeed(&params)
}
