package strava

import (
	stravalib "github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/strava/go.strava"
	dm "github.com/svdberg/syncmysport-runkeeper/datamodel"
	"log"
	"time"
)

func ConvertToActivity(stravaActivity *stravalib.ActivityDetailed, timeStream *stravalib.StreamSet, gpsTrack *stravalib.StreamSet, hrTrack *stravalib.StreamSet) *dm.Activity {
	stvActivity := dm.CreateActivity()
	stvActivity.StartTime = int(stravaActivity.StartDate.Unix()) //UTC date
	//stvActivity.StartTime = int(stravaActivity.StartDateLocal.Unix()) //UTC +1 Date (local)
	log.Printf("STV Local date: %s, start date: %s, unix: %d", stravaActivity.StartDateLocal, stravaActivity.StartDate, stravaActivity.StartDate.Unix())
	stvActivity.Duration = stravaActivity.ElapsedTime
	stvActivity.Name = stravaActivity.Name
	stvActivity.Calories = stravaActivity.Calories
	stvActivity.Distance = stravaActivity.Distance
	stvActivity.AverageHeartRate = int(stravaActivity.AverageHeartrate)
	loc, err := time.LoadLocation(stravaActivity.TimeZone)
	if err != nil {
		timeInTZ := time.Time(stravaActivity.StartDate).In(loc)
		_, offsetInSeconds := timeInTZ.Zone()
		stvActivity.UtcOffSet = offsetInSeconds / 60 / 60
	} else {
		log.Print("Error reading location from strava Activity: %s", err)
	}

	if stravaActivity.Type.String() == "Run" {
		stvActivity.Type = "Running"
	} else if stravaActivity.Type.String() == "Ride" || stravaActivity.Type.String() == "EBikeRide" {
		stvActivity.Type = "Cycling"
	} else if stravaActivity.Type.String() == "Swim" {
		stvActivity.Type = "Swimming"
	} else {
		//I don't know, call it Activity
		stvActivity.Type = "Activity"
	}

	if gpsTrack != nil && gpsTrack.Location != nil && timeStream != nil {
		stvActivity.GPS = convertGPSTrack(gpsTrack, timeStream)
	}
	if hrTrack != nil && hrTrack.HeartRate != nil && timeStream != nil {
		stvActivity.HeartRate = convertHeartRateTrack(hrTrack, timeStream)
	}
	return stvActivity
}

func convertGPSTrack(sourceStream *stravalib.StreamSet, timeStream *stravalib.StreamSet) []dm.GPS {
	//merge the time stream + the location stream
	merged := mergeTimeAndLocation(timeStream.Time, sourceStream.Location)

	result := make([]dm.GPS, len(sourceStream.Location.Data))
	for index, gpsTime := range merged {
		alt := 0.0
		result[index] = dm.GPS{float64(gpsTime.Time), alt, gpsTime.Long, gpsTime.Lat}
	}
	return result
}

type GPSTime struct {
	Time int
	Lat  float64
	Long float64
}

func mergeTimeAndLocation(timeStream *stravalib.IntegerStream, locStream *stravalib.LocationStream) []GPSTime {
	merged := make([]GPSTime, len(timeStream.Data))
	for i, t := range timeStream.Data {
		latLong := locStream.Data[i]
		merged[i] = GPSTime{t, latLong[0], latLong[1]}
	}
	return merged
}

func convertHeartRateTrack(sourceStream *stravalib.StreamSet, timeStream *stravalib.StreamSet) []dm.HeartRate {
	result := make([]dm.HeartRate, len(sourceStream.HeartRate.Data))
	for index, hr := range sourceStream.HeartRate.Data {
		time := timeStream.Time.Data[index]
		result[index] = dm.HeartRate{float64(time), hr}
	}
	return result
}
