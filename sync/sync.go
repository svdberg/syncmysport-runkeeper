package sync

import (
	dm "github.com/svdberg/syncmysport-runkeeper/datamodel"
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
	lastSeenTimestamp int
}

func (st SyncTask) Sync() {
	//get activities from strava

	//get activities from runkeeper

	//make two sets

	//caclulate difference

	//write to runkeeper
}
