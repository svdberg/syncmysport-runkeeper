package main

import (
	"log"
	"runkeeper/runkeeper"
	"strava/strava"
	"time"
)

const timestamp = 1452384000

func main() {
	//	getRkActivities()
	getSTVActivities()
}

func getSTVActivities() {
	token := CheckForStvBearerToken()
	if token == "" {
		StartStvOAuth()

		for token == "" {
			time.Sleep(500000000)
			log.Println("MEH")
			token = CheckForStvBearerToken()
		}
	}
	log.Println("Getting activities")
	activities, _ := GetSTVActivitiesSince(token, timestamp)
	log.Println(activities)
}

func getRkActivities() {
	bearerToken := CheckForBearerToken()
	if bearerToken == "" {
		LaunchOAuth()
		for bearerToken == "" {
			time.Sleep(500000000)
			bearerToken = CheckForBearerToken()
		}
	}

	activities, err := GetRKActivitiesSince(bearerToken, timestamp)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(activities)
}
