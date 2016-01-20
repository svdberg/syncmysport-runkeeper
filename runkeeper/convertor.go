package runkeeper

import (
	runkeeper "github.com/c9s/go-runkeeper"
	dm "github.com/svdberg/syncmysport-runkeeper/datamodel"
	"time"
)

const API = "API"

func ConvertToActivity(rkActivity *runkeeper.FitnessActivity) *dm.Activity {
	returnActivity := dm.CreateActivity()
	returnActivity.Type = rkActivity.Type
	returnActivity.StartTime = int(time.Time(rkActivity.StartTime).Unix())
	returnActivity.Duration = int(rkActivity.Duration)
	returnActivity.Name = rkActivity.Comment
	returnActivity.Notes = rkActivity.Comment //hmm dunno
	returnActivity.Private = false
	returnActivity.Stationary = rkActivity.HasMap
	returnActivity.AverageHeartRate = 0 //rkActivity.AverageHeartRate
	returnActivity.Calories = rkActivity.TotalCalories
	returnActivity.Distance = rkActivity.TotalDistance

	return returnActivity
}

func ConvertToRkActivity(activity *dm.Activity) *runkeeper.FitnessActivityNew {
	rkActivity := runkeeper.CreateNewFitnessActivity(activity.Name, float64(activity.Duration))

	rkActivity.Type = activity.Type
	rkActivity.StartTime = runkeeper.Time(time.Unix(int64(activity.StartTime), 0))
	rkActivity.Notes = activity.Name
	rkActivity.TotalDistance = activity.Distance
	rkActivity.AverageHeartRate = activity.AverageHeartRate
	rkActivity.TotalCalories = activity.Calories
	rkActivity.Source = activity.Source
	rkActivity.EntryMode = API

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
	rkHr := make([]runkeeper.HeartRate, len(hr))
	for i, h := range hr {
		rkHr[i] = runkeeper.HeartRate{h.Timestamp, h.Heartrate}
	}
	return rkHr
}
