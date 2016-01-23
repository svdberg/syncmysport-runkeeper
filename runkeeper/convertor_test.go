package runkeeper

import (
	"fmt"
	runkeeper "github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/c9s/go-runkeeper"
	dm "github.com/svdberg/syncmysport-runkeeper/datamodel"
	"testing"
	"time"
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

	resultActivity := dm.Activity{int(tim.Unix()), 10, "Running", "", "", false, false}
	activity := ConvertToActivity(&rkActivity)

	if !activity.ConsideredEqual(&resultActivity) {
		t.Error(fmt.Sprintf("activity %s should match result activity", activity))
	}
}

func TestToRKConvertor(t *testing.T) {
}
