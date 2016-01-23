package main

import (
	"fmt"
	cron "github.com/robfig/cron"
	api "github.com/svdberg/syncmysport-runkeeper/api"
	dm "github.com/svdberg/syncmysport-runkeeper/datamodel"
	rk "github.com/svdberg/syncmysport-runkeeper/runkeeper"
	stv "github.com/svdberg/syncmysport-runkeeper/strava"
	sync "github.com/svdberg/syncmysport-runkeeper/sync"
	"log"
	"time"
)

const timestamp = 1452384000
const tsDelta = -5 //minutes

func main() {
	//Start Scheduler
	c := cron.New()
	err := c.AddFunc("0 5/15 * * *", startSync) //every 15 minutes, starting 5 in
	if err != nil {
		log.Fatal("Error adding the job to the scheduler", err)
	}
	c.Start()

	//Start api
	api.Start()
}

func startSync() {
	repo := sync.CreateSyncDbRepo()
	allSyncs, err := repo.RetrieveAllSyncTasks()
	log.Printf("Retrieved %d sync tasks", len(allSyncs))
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	for _, syncer := range allSyncs {
		log.Printf("Now syncing for task: %s, %s, %s", syncer.StravaToken, syncer.RunkeeperToken, time.Unix(int64(syncer.LastSeenTimestamp), 0))
		difference, nrItemsCreated := syncer.Sync()
		log.Printf("Nr of Activities missing in RunKeeper: %d, Actvities created: %d", difference, nrItemsCreated)
		if difference == nrItemsCreated {
			log.Print("Updating last seen timestamp")
			//subtract 5 minutes to prevent activites being missed
			syncer.LastSeenTimestamp = int(time.Now().Add(time.Duration(tsDelta) * time.Minute).Unix())
			rowsUpdated, err := repo.UpdateSyncTask(syncer)
			if err != nil || rowsUpdated != 1 {
				log.Fatal("Error updating the SyncTask record with a new timestamp")
			}
		} else {
			log.Print("Something went wrong storing Activities, not updating timestamp so we will retry")
		}
	}
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
