package main

import (
	"fmt"
	dm "github.com/svdberg/syncmysport-runkeeper/datamodel"
	rk "github.com/svdberg/syncmysport-runkeeper/runkeeper"
	stv "github.com/svdberg/syncmysport-runkeeper/strava"
	sync "github.com/svdberg/syncmysport-runkeeper/sync"
	"log"
	"time"
)

const timestamp = 1452384000

func main() {
	stvToken := GetStravaToken()
	//rkToken := GetRkToken()
	//getRkActivities()
	//getSTVActivities()
	//syncer := sync.CreateSyncTask(rkToken, stvToken, timestamp)
	repo := sync.CreateSyncDbRepo()
	syncer, err := repo.RetrieveSyncTaskByToken(stvToken)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	syncer.Sync()
}

func GetStravaToken() string {
	token := stv.CheckForStvBearerToken()
	if token == "" {
		stv.StartStvOAuth()

		for token == "" {
			time.Sleep(500000000)
			token = stv.CheckForStvBearerToken()
		}
	}
	return token
}

func GetRkToken() string {
	bearerToken := rk.CheckForBearerToken()
	if bearerToken == "" {
		rk.LaunchOAuth()
		for bearerToken == "" {
			time.Sleep(500000000)
			bearerToken = rk.CheckForBearerToken()
		}
	}
	return bearerToken
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
	stvClient := stv.CreateStravaClient(token)
	activities, _ := stvClient.GetSTVActivitiesSince(timestamp)
	detailedActivities := make([]*dm.Activity, len(activities))
	for i, actSummary := range activities {
		//get Detailed Actv
		detailedAct, _ := stvClient.GetSTVDetailedActivity(actSummary.Id)
		log.Printf("Detailed STV Act: %s in TZ: %s", detailedAct, detailedAct.TimeZone)
		//get associated Streams
		timeStream, err := stvClient.GetSTVActivityStream(actSummary.Id, "Time")
		if err != nil {
			log.Fatal("Error while retrieving time series from Strava: %s", err)
		}
		locStream, _ := stvClient.GetSTVActivityStream(actSummary.Id, "GPS")
		hrStream, _ := stvClient.GetSTVActivityStream(actSummary.Id, "Heartrate")

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
	rkClient := rk.CreateRKClient(rkBearerToken)
	uri, err := rkClient.PostActivity(rkActivity)
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

	rkClient := rk.CreateRKClient(bearerToken)
	activities, err := rkClient.GetRKActivitiesSince(timestamp)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(activities)
}
