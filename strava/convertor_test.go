package strava

import (
	"fmt"
	stravalib "github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/strava/go.strava"
	dm "github.com/svdberg/syncmysport-runkeeper/datamodel"
	"testing"
	"time"
)

func TestConvertor(t *testing.T) {
	//expected
	theTime, _ := time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Jan 2 15:04:05 -0700 MST 2006")
	duration := 10000
	theType := "Running"
	name := "Test Activity"
	notes := "Some Random\n Bullshit"

	stvActivity := stravalib.ActivityDetailed{}
	stvActivity.Name = name
	stvActivity.ElapsedTime = duration
	stvActivity.StartDate = theTime
	stvActivity.Description = notes
	stvActivity.Type = stravalib.ActivityTypes.Run

	activity := ConvertToActivity(&stvActivity, nil, nil, nil)
	resultActivity := dm.CreateActivity()
	resultActivity.StartTime = int(theTime.Unix())
	resultActivity.Duration = duration
	resultActivity.Type = theType
	resultActivity.Name = name

	fmt.Printf("%s\n", activity)
	fmt.Printf("%s\n", resultActivity)

	//this compare probably doesnt work as expected
	if !activity.ConsideredEqual(resultActivity) {
		t.Error("activity should match resultActivity")
	}
}

func TestGPSStreamConvertor(t *testing.T) {
}
