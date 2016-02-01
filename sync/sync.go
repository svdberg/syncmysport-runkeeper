package sync

import (
	"fmt"
	dm "github.com/svdberg/syncmysport-runkeeper/datamodel"
	rk "github.com/svdberg/syncmysport-runkeeper/runkeeper"
	stv "github.com/svdberg/syncmysport-runkeeper/strava"
	"log"
	"time"
)

type Syncer interface {
	Sync() (int, int, error)
}

/*
 * Returns the set of activities that are in RunKeeper, but not in Strava.
 * So if the set of Runkeeper activites is A, and the set of Strava activities is B,
 * this function calculates B\A.
 */
func CalculateRKDifference(rkActivities dm.ActivitySet, stvActivities dm.ActivitySet) *dm.ActivitySet {
	return rkActivities.ApproxSubtract(stvActivities)
}

type SyncTask struct {
	StravaToken       string `json:"stv_token"`
	RunkeeperToken    string `json:"rk_token"`
	LastSeenTimestamp int    `json:"last_seen_ts"`
	Uid               int64  `json:"id"`
	Environment       string `json:"environment"`
}

func CreateSyncTask(rkToken string, stvToken string, lastSeenTS int, environment string) *SyncTask {
	return &SyncTask{stvToken, rkToken, lastSeenTS, -1, environment}
}

/*
 * return the Total difference and the number of Activites created
 */
func (st SyncTask) Sync(stvClient stv.StravaClientInt, rkClient rk.RunkeeperCientInt) (int, int, error) {
	//get activities from strava
	//normalize time to the start of the day, because Runkeeper only supports days as offset, not timestamps
	tsOfStartOfDay := calculateTsAtStartOfDay(st.LastSeenTimestamp)
	activities, err := stvClient.GetSTVActivitiesSince(tsOfStartOfDay)
	if err != nil {
		log.Print("Error retrieving Strava activitites since %s, aborting this run", st.LastSeenTimestamp)
		return 0, 0, err
	}
	stvDetailedActivities := dm.NewActivitySet()
	for _, actSummary := range activities {
		//get Detailed Actv
		detailedAct, _ := stvClient.GetSTVDetailedActivity(actSummary.Id)
		//get associated Streams
		timeStream, err := stvClient.GetSTVActivityStream(actSummary.Id, "Time")
		if err != nil {
			log.Fatal("Error while retrieving time series from Strava: %s", err)
		}
		locStream, _ := stvClient.GetSTVActivityStream(actSummary.Id, "GPS")
		hrStream, _ := stvClient.GetSTVActivityStream(actSummary.Id, "Heartrate")

		stvDetailedActivities.Add(*stv.ConvertToActivity(detailedAct, timeStream, locStream, hrStream))
	}
	log.Printf("Got %d items from Strava", stvDetailedActivities.NumElements())
	for i := 0; i < stvDetailedActivities.NumElements(); i++ {
		log.Printf("Strava Activity: %s", stvDetailedActivities.Get(i))
	}

	//get activities from runkeeper
	rkActivitiesOverview, err := rkClient.GetRKActivitiesSince(st.LastSeenTimestamp)
	rkDetailActivities := rkClient.EnrichRKActivities(rkActivitiesOverview)
	//log.Printf("rk detail activities: %s", rkDetailActivities)

	rkActivities := dm.NewActivitySet()
	for _, item := range rkDetailActivities {
		rkActivities.Add(*rk.ConvertToActivity(&item))
	}
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Got %d items from RunKeeper", rkActivities.NumElements())
	for i := 0; i < rkActivities.NumElements(); i++ {
		log.Printf("Runkeeper Activity: %s", rkActivities.Get(i))
	}

	//caclulate difference
	itemsToSyncToRk := stvDetailedActivities.ApproxSubtract(rkActivities)
	log.Printf("Difference between Runkeeper and Strava is %d items", itemsToSyncToRk.NumElements())

	//write to runkeeper
	totalItemsCreated := 0
	for i := 0; i < itemsToSyncToRk.NumElements(); i++ {
		log.Printf("Now storing item %s to RunKeeper", itemsToSyncToRk.Get(i))
		var (
			uri string
			err error
		)
		if st.Environment == "Prod" {
			uri, err = rkClient.PostActivity(rk.ConvertToRkActivity(itemsToSyncToRk.Get(i)))
		} else {
			log.Print("Assuming DEBUG/TEST mode, not actually writing to Runkeeper")
			err = nil
			uri = "fake_uri"
		}

		if err != nil {
			log.Fatal("Something failed during the write to Runkeeper: %s", err)
		}
		if uri != "" {
			log.Printf("URI of activity: %s", uri)
			totalItemsCreated++
		}
	}
	return itemsToSyncToRk.NumElements(), totalItemsCreated, nil
}

func calculateTsAtStartOfDay(timestamp int) int {
	timeAtTimestamp := time.Unix(int64(timestamp), 0).UTC()
	year, month, day := timeAtTimestamp.Date()
	//	Mon Jan 2 15:04:05 -0700 MST 2006
	ts, err := time.Parse("2006-1-2 15:04:05 MST", fmt.Sprintf("%d-%d-%d 00:00:00 UTC", year, month, day))
	if err != nil {
		log.Printf("Error parsing time from y/m/d to ts of current day: %s", err)
	}
	ts = ts.Add(time.Duration(1) * time.Minute)
	return int(ts.Unix())
}
