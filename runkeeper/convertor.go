package runkeeper

import (
	"log"
	"time"

	runkeeper "github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/c9s/go-runkeeper"
	dm "github.com/svdberg/syncmysport-runkeeper/datamodel"
)

const API = "API"

func ConvertToActivity(rkActivity *runkeeper.FitnessActivity) *dm.Activity {
	returnActivity := dm.CreateActivity()
	if rkActivity.Type == "Other" {
		returnActivity.Type = "Activity"
	} else {
		returnActivity.Type = rkActivity.Type
	}

	//RK time is 'Local'
	correctedTime := time.Time(rkActivity.StartTime).Add(time.Duration(rkActivity.UtcOffset) * time.Hour)
	log.Printf("RK Local date: %s, start date: %s, unix: %d, offset: %d", time.Time(rkActivity.StartTime), correctedTime, time.Time(rkActivity.StartTime).Unix(), rkActivity.UtcOffset)
	returnActivity.StartTime = int(time.Time(correctedTime).Unix())
	returnActivity.UtcOffSet = rkActivity.UtcOffset
	returnActivity.Duration = int(rkActivity.Duration)
	returnActivity.Name = rkActivity.Notes
	returnActivity.Notes = rkActivity.Notes
	returnActivity.Private = false
	returnActivity.Stationary = rkActivity.HasMap
	returnActivity.AverageHeartRate = rkActivity.AverageHeartRate
	returnActivity.Calories = rkActivity.TotalCalories
	returnActivity.Distance = rkActivity.TotalDistance
	returnActivity.GPS = convertFromPath(rkActivity.Path)
	returnActivity.HeartRate = convertFromHR(rkActivity.HeartRate)

	//log.Printf("INPUT: %s, OUTPUT: %s", rkActivity, returnActivity)
	return returnActivity
}

func ConvertToRkActivity(activity *dm.Activity) *runkeeper.FitnessActivityNew {
	rkActivity := runkeeper.CreateNewFitnessActivity(activity.Name, float64(activity.Duration))

	rkActivity.Type = activity.Type
	//runkeeper only nows the following types:
	//Running, Cycling, Mountain Biking, Walking,
	//Hiking, Downhill Skiing, Cross-Country Skiing,
	//Snowboarding, Skating, Swimming, Wheelchair, Rowing, Elliptical, Other
	//
	//check if Type is one of these, otherwise Other.
	rkKnownTypes := map[string]string{
		"Running":              "Running",
		"Cycling":              "Cycling",
		"Mountain Biking":      "Mountain Biking",
		"Walking":              "Walking",
		"Hiking":               "Hiking",
		"Downhill Skiing":      "Downhill Skiing",
		"Cross-Country Skiing": "Cross-Country Skiing",
		"Snowboarding":         "Snowboarding",
		"Skating":              "Skating",
		"Swimming":             "Swimming",
		"Wheelchair":           "Wheelchair",
		"Rowing":               "Rowing",
		"Elliptical":           "Elliptical",
		"Other":                "Other"}

	_, ok := rkKnownTypes[activity.Type]
	if !ok {
		rkActivity.Type = "Other"
	}

	//runkeeper times are in local timezones, so covert back to the local time
	rkLocalLocation := time.FixedZone("rkZone", activity.UtcOffSet*60*60)
	rkActivity.StartTime = runkeeper.Time(time.Unix(int64(activity.StartTime), 0).In(rkLocalLocation))
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

func convertFromPath(path []runkeeper.Path) []dm.GPS {
	dmPath := make([]dm.GPS, len(path))
	for i, rp := range path {
		dmPath[i] = dm.GPS{rp.Timestamp, rp.Altitude, rp.Longitude, rp.Latitude}
	}
	return dmPath
}

func convertToHR(hr []dm.HeartRate) []runkeeper.HeartRate {
	rkHr := make([]runkeeper.HeartRate, len(hr))
	for i, h := range hr {
		rkHr[i] = runkeeper.HeartRate{h.Timestamp, h.Heartrate}
	}
	return rkHr
}

func convertFromHR(rkHr []runkeeper.HeartRate) []dm.HeartRate {
	dmHr := make([]dm.HeartRate, len(rkHr))
	for i, h := range rkHr {
		dmHr[i] = dm.HeartRate{h.TimeStamp, h.HearRateNr}
	}
	return dmHr
}
