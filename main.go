package main

import (
	"fmt"
	dm "github.com/svdberg/syncmysport-runkeeper/datamodel"
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
	detailedActivities := make([]*dm.Activity, len(activities))
	for i, actSummary := range activities {
		//get Detailed Actv
		detailedAct, _ := stv.GetSTVDetailedActivity(token, actSummary.Id)
		//get associated Streams
		timeStream, err := stv.GetSTVActivityStream(token, actSummary.Id, "Time")
		if err != nil {
			log.Fatal("Error while retrieving time series from Strava: %s", err)
		}
		locStream, _ := stv.GetSTVActivityStream(token, actSummary.Id, "GPS")
		hrStream, _ := stv.GetSTVActivityStream(token, actSummary.Id, "Heartrate")

		detailedActivities[i] = stv.ConvertToActivity(detailedAct, timeStream, locStream, hrStream)
	}
	log.Println(detailedActivities)

	//try to write the last activity to RK
	log.Println("Writing last strava activity to RK")
	lastActivity := detailedActivities[len(detailedActivities)-1]
	lastActivity.Name = fmt.Sprintf("%s-%s", lastActivity.Name, "-SANDERTEST")
	log.Printf("Writing '%s' to RK", lastActivity)
	rkBearerToken := rk.CheckForBearerToken()
	rkActivity := rk.ConvertToRkActivity(lastActivity)
	uri, err := rk.PostActivity(rkActivity, rkBearerToken)
	if err != nil {
		log.Fatal("Error while creating RK Activity: %s", err)
	}
	log.Printf("URI of new activity: %s", uri)
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
