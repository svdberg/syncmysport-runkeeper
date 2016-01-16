package main

import (
	rk "github.com/svdberg/syncmysport-runkeeper/runkeeper"
	stv "github.com/svdberg/syncmysport-runkeeper/strava"
	"log"
	"time"
)

const timestamp = 1452384000

func main() {
	getRkActivities()
	getSTVActivities()
}

func getSTVActivities() {
	token := stv.CheckForStvBearerToken()
	if token == "" {
		stv.StartStvOAuth()

		for token == "" {
			time.Sleep(500000000)
			token = stv.CheckForStvBearerToken()
		}
	}
	log.Println("Getting activities")
	activities, _ := stv.GetSTVActivitiesSince(token, timestamp)
	log.Println(activities)
}

func getRkActivities() {
	bearerToken := rk.CheckForBearerToken()
	if bearerToken == "" {
		rk.LaunchOAuth()
		for bearerToken == "" {
			time.Sleep(500000000)
			bearerToken = rk.CheckForBearerToken()
		}
	}

	activities, err := rk.GetRKActivitiesSince(bearerToken, timestamp)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(activities)
}
