package strava

import (
	"fmt"
	stravalib "github.com/strava/go.strava"
	dm "github.com/svdberg/syncmysport-runkeeper/datamodel"
	"testing"
	"time"
)

func TestConvertor(t *testing.T) {
	theTime, _ := time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Jan 2 15:04:05 -0700 MST 2006")
	duration := 10000
	theType := "Run"
	name := "Test Activity"
	notes := "Some Random\n Bullshit"

	stvActivity := stravalib.ActivityDetailed{}
	stvActivity.Name = name
	stvActivity.ElapsedTime = duration
	stvActivity.StartDate = theTime
	stvActivity.Description = notes
	stvActivity.Type = stravalib.ActivityTypes.Run

	activity := ConvertToActivity(&stvActivity)
	resultActivity := dm.Activity{int(theTime.Unix()), duration, theType, name, notes, false, false}

	fmt.Printf("%s\n", activity)
	fmt.Printf("%s\n", resultActivity)

	//this compare probably doesnt work as expected
	if *activity != resultActivity {
		t.Error("activity should match resultActivity")
	}
}
