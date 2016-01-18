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

func ConvertToRkActivity(activity *dm.Activity) *runkeeper.FitnessActivityNew {
	rkActivity := runkeeper.CreateNewFitnessActivity(activity.Name, float64(activity.Duration))

	rkActivity.Type = activity.Type
	rkActivity.StartTime = runkeeper.Time(time.Unix(int64(activity.StartTime), 0))
	rkActivity.Notes = activity.Name

	rkActivity.Path = convertToPath(activity.GPS)
	rkActivity.HeartRate = convertToHR(activity.HeartRate)
	return rkActivity
}

func convertToPath(gps []dm.GPS) []runkeeper.Path {
	rkPath := make([]runkeeper.Path, len(gps))
	for i, gp := range gps {
		rkPath[i] = runkeeper.Path{gp.Altitude, gp.Longitude, "gps", gp.Latitude, gp.Timestamp}
	}
	return rkPath
}

func convertToHR(hr []dm.HeartRate) []runkeeper.HeartRate {
	return make([]runkeeper.HeartRate, 0)
}
