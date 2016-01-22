package sync

import (
	dm "github.com/svdberg/syncmysport-runkeeper/datamodel"
	rk "github.com/svdberg/syncmysport-runkeeper/runkeeper"
	stv "github.com/svdberg/syncmysport-runkeeper/strava"
	"log"
)

/*
 * Returns the set of activities that are in RunKeeper, but not in Strava.
 * So if the set of Runkeeper activites is A, and the set of Strava activities is B,
 * this function calculates B\A.
 */
func CalculateRKDifference(rkActivities dm.ActivitySet, stvActivities dm.ActivitySet) *dm.ActivitySet {
	return rkActivities.ApproxSubtract(stvActivities)
}

type SyncTask struct {
	StravaToken       string
	RunkeeperToken    string
	LastSeenTimestamp int
	uid               int64
}

func CreateSyncTask(rkToken string, stvToken string, lastSeenTS int) *SyncTask {
	return &SyncTask{stvToken, rkToken, lastSeenTS, -1}
}

func (st SyncTask) Sync() {
	//get activities from strava
	stvClient := stv.CreateStravaClient(st.StravaToken)
	activities, _ := stvClient.GetSTVActivitiesSince(st.LastSeenTimestamp)
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
	rkClient := rk.CreateRKClient(st.RunkeeperToken)
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
	for i := 0; i < itemsToSyncToRk.NumElements(); i++ {
		log.Printf("Now storing item %s to RunKeeper", itemsToSyncToRk.Get(i))
		//rkClient.PostActivity(rk.ConvertToRkActivity(itemsToSyncToRk.Get(i)))
	}
}
