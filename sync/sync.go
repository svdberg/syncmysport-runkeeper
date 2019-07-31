package sync

import (
	"fmt"
	"log"
	"time"

	"github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/newrelic/go-agent"
	dm "github.com/svdberg/syncmysport-runkeeper/datamodel"
	rk "github.com/svdberg/syncmysport-runkeeper/runkeeper"
	stv "github.com/svdberg/syncmysport-runkeeper/strava"
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
func (st SyncTask) Sync(stvClient stv.StravaClientInt, rkClient rk.RunkeeperCientInt, txn newrelic.Transaction) (int, int, error) {
	//get activities from strava
	//normalize time to the start of the day, because Runkeeper only supports days as offset, not timestamps
	tsOfStartOfDay := calculateTsAtStartOfDay(st.LastSeenTimestamp)
	segment := newrelic.Segment{}
	segment.Name = "strava-retrieve-activities"
	segment.StartTime = newrelic.StartSegmentNow(txn)
	activities, err := stvClient.GetSTVActivitiesSince(tsOfStartOfDay)
	segment.End()
	if err != nil {
		log.Printf("Error: %e while retrieving Strava activitites since %s, aborting this run", err, time.Unix(int64(st.LastSeenTimestamp), 0))
		return 0, 0, err
	}
	stvDetailedActivities := dm.NewActivitySet()
	for _, actSummary := range activities {
		//get Detailed Actv
		detailedAct, _ := stvClient.GetSTVDetailedActivity(actSummary.Id)

		//get associated Streams
		timeStream, err := stvClient.GetSTVActivityStream(actSummary.Id, "Time")
		if err != nil {
			log.Printf("Error while retrieving time series from Strava: %s", err)
			return 0, 0, err
		}
		log.Printf("timeStream: %s", timeStream)

		//Optional streams
		// var locStream, hrStream, altStream *stv.StreamSet
		locStream, err := stvClient.GetSTVActivityStream(actSummary.Id, "GPS")
		if err != nil {
			log.Printf("Error while reading lat/long for activity id %d: %e", actSummary.Id, err)
		}
		log.Printf("locStream: %s", locStream)
		hrStream, err := stvClient.GetSTVActivityStream(actSummary.Id, "Heartrate")
		if err != nil {
			log.Printf("Error while reading Heartrate for activity id %d: %e", actSummary.Id, err)
		}

		altStream, err := stvClient.GetSTVActivityStream(actSummary.Id, "Altitude")
		if err != nil {
			log.Printf("Error while reading Altitude for activity id %d: %e", actSummary.Id, err)
		}

		stvDetailedActivities.Add(*stv.ConvertToActivity(detailedAct, timeStream, locStream, hrStream, altStream))
	}
	log.Printf("Got %d items from Strava", stvDetailedActivities.NumElements())
	for i := 0; i < stvDetailedActivities.NumElements(); i++ {
		log.Printf("Strava Activity: %s", stvDetailedActivities.Get(i))
	}
	//stvDetailedActivities should be in UTC

	//get activities from runkeeper
	segment = newrelic.Segment{}
	segment.Name = "runkeeper-retrieve-activities"
	segment.StartTime = newrelic.StartSegmentNow(txn)
	rkActivitiesOverview, err := rkClient.GetRKActivitiesSince(st.LastSeenTimestamp)
	rkDetailActivities := rkClient.EnrichRKActivities(rkActivitiesOverview)
	segment.End()
	//log.Printf("rk detail activities: %s", rkDetailActivities)

	rkActivities := dm.NewActivitySet()
	for _, item := range rkDetailActivities {
		rkActivities.Add(*rk.ConvertToActivity(&item))
	}
	if err != nil {
		log.Printf("%s", err)
	}
	log.Printf("Got %d items from RunKeeper", rkActivities.NumElements())
	// rkActivities are in UTC (or should be)

	for i := 0; i < rkActivities.NumElements(); i++ {
		log.Printf("Runkeeper Activity: %s", rkActivities.Get(i))
	}

	//caclulate difference
	itemsToSyncToRk := stvDetailedActivities.ApproxSubtract(rkActivities)
	log.Printf("Difference between Runkeeper and Strava is %d items", itemsToSyncToRk.NumElements())

	//write to runkeeper
	segment = newrelic.Segment{}
	segment.Name = "runkeeper-write-activities"
	segment.StartTime = newrelic.StartSegmentNow(txn)
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
			log.Printf("Something failed during the write to Runkeeper: %s", err)
			return itemsToSyncToRk.NumElements(), totalItemsCreated, err
		}
		if uri != "" {
			log.Printf("URI of activity: %s", uri)
			totalItemsCreated++
		}
	}
	segment.End()
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
