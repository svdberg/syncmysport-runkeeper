package runkeeper

import (
	runkeeper "github.com/c9s/go-runkeeper"
	dm "github.com/svdberg/syncmysport-runkeeper/datamodel"
	"time"
)

func ConvertToActivity(rkActivity *runkeeper.FitnessActivity) *dm.Activity {
	returnActivity := dm.Activity{}
	returnActivity.Type = rkActivity.Type
	returnActivity.StartTime = int(time.Time(rkActivity.StartTime).Unix())
	returnActivity.Duration = int(rkActivity.Duration)
	returnActivity.Name = rkActivity.Comment
	returnActivity.Notes = rkActivity.Comment //hmm dunno
	returnActivity.Private = false
	returnActivity.Stationary = rkActivity.HasMap

	return &returnActivity
}

func ConvertToRkActivity(activity *dm.Activity) *runkeeper.FitnessActivity {
	return nil
}
