package main

import (
	strava "github.com/strava/go.strava"
)

func GetSTVActivitiesSince(bearerToken string, timestamp int) ([]*strava.ActivitySummary, error) {
	client := strava.NewClient(bearerToken)
	service := strava.NewCurrentAthleteService(client)
	call := service.ListActivities()
	call.After(timestamp)
	activities, err := call.Do()
	return activities, err
}
