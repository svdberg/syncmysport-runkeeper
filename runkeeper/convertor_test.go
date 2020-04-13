package runkeeper

import (
	"fmt"
	"testing"
	"time"

	runkeeper "github.com/c9s/go-runkeeper"
	dm "github.com/svdberg/syncmysport-runkeeper/datamodel"
)

func TestFromRKConvertor(t *testing.T) {
	rkActivity := runkeeper.FitnessActivity{}
	rkActivity.Type = "Running"
	tim, _ := time.Parse("Mon, 2 Jan 2006 15:04:05", "Tue, 1 Mar 2011 07:00:00")
	rkActivity.StartTime = runkeeper.Time(tim)
	rkActivity.Duration = 10
	rkActivity.Source = "RunKeeper"
	rkActivity.EntryMode = "API"
	rkActivity.HasMap = true
	rkActivity.Uri = "/activities/40"
	rkActivity.Path = make([]runkeeper.Path, 2)
	rkActivity.Path[0] = runkeeper.Path{1.0, 0.0, "gps", 10.0, 11.0}
	rkActivity.Path[1] = runkeeper.Path{2.0, 0.0, "gps", 15.0, 13.0}
	rkActivity.HeartRate = make([]runkeeper.HeartRate, 1)
	rkActivity.HeartRate[0] = runkeeper.HeartRate{1.0, 2}

	expectedActivity := dm.CreateActivity()
	expectedActivity.StartTime = int(tim.Unix())
	expectedActivity.Duration = 10
	expectedActivity.Type = "Running"
	activity := ConvertToActivity(&rkActivity)

	if !activity.ConsideredEqual(expectedActivity) ||
		len(activity.GPS) != 2 ||
		len(activity.HeartRate) != 1 {

		t.Error(fmt.Sprintf("activity %s should match result activity", activity))
	}
}

func TestOtherActivityType(t *testing.T) {
	rkActivity := runkeeper.FitnessActivity{}
	rkActivity.Type = "Other"

	expectedActivity := dm.CreateActivity()
	expectedActivity.Type = "Activity"

	activity := ConvertToActivity(&rkActivity)
	if activity.Type != expectedActivity.Type {
		t.Error(fmt.Sprintf("Type %s of activity should match %s", activity.Type, expectedActivity.Type))
	}
}

func TestToRKConvertor(t *testing.T) {
	activity := dm.CreateActivity()
	activity.Type = "Activity"

	rkActivity := ConvertToRkActivity(activity)

	if rkActivity.Type != "Other" {
		t.Error(fmt.Sprintf("Type %s of activity should match \"Other\"", rkActivity.Type))
	}
}
