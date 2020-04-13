package strava

import (
	"fmt"
	"testing"
	"time"

	timez "4d63.com/tz"
	stravalib "github.com/strava/go.strava"
	dm "github.com/svdberg/syncmysport-runkeeper/datamodel"
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

	activity := ConvertToActivity(&stvActivity, nil, nil, nil, nil)
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

func assertEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if a == b {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Fatal(message)
}

func TestTZOffsetCalculation(t *testing.T) {
	inputTZ := "(GMT-05:00) America/New_York"
	sampleActivity := stravalib.ActivityDetailed{}
	sampleActivity.StartDate, _ = time.Parse(time.RFC822, "02 Jan 19 15:04 CET")
	offset := getTZOffsetForLocation(inputTZ, sampleActivity.StartDate)
	fmt.Printf("calculated TZ: %d", offset)
	assertEqual(t, offset, -5, "Offset for New York should be -5 hours")
}

func TestTZLib(t *testing.T) {
	tz := "America/New_York"
	timezone, e := timez.LoadLocation(tz)
	if e == nil {
		fmt.Printf("%s", timezone)
	} else {
		t.Fatalf("Error while loading location: %e", e)
	}

}

func TestGPSStreamConvertor(t *testing.T) {
}
