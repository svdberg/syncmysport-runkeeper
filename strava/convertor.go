package strava

import (
	"log"
	"strings"
	"time"

	stravalib "github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/strava/go.strava"
	dm "github.com/svdberg/syncmysport-runkeeper/datamodel"
)

func ConvertToActivity(stravaActivity *stravalib.ActivityDetailed, timeStream *stravalib.StreamSet, gpsTrack *stravalib.StreamSet, hrTrack *stravalib.StreamSet, altTrack *stravalib.StreamSet) *dm.Activity {
	stvActivity := dm.CreateActivity()
	stvActivity.StartTime = int(stravaActivity.StartDate.Unix()) //UTC date
	log.Printf("STV Local date: %s, start date: %s, unix: %d", stravaActivity.StartDateLocal, stravaActivity.StartDate, stravaActivity.StartDate.Unix())
	stvActivity.Duration = stravaActivity.ElapsedTime
	stvActivity.Name = stravaActivity.Name
	stvActivity.Calories = stravaActivity.Calories
	stvActivity.Distance = stravaActivity.Distance
	stvActivity.AverageHeartRate = int(stravaActivity.AverageHeartrate)

	//properly format TZ
	tz := stravaActivity.TimeZone
	log.Printf("TZ: %s", tz)
	startOfTz := strings.Index(stravaActivity.TimeZone, ")")
	if startOfTz != -1 {
		tz = string(stravaActivity.TimeZone[startOfTz:])
	}

	loc, err := time.LoadLocation(tz)
	if err == nil {
		timeInTZ := time.Time(stravaActivity.StartDate).In(loc)
		_, offsetInSeconds := timeInTZ.Zone()
		stvActivity.UtcOffSet = offsetInSeconds / 60 / 60
		log.Printf("calculated offset: %d", offsetInSeconds)
	} else {
		log.Printf("Warning: reading location from strava Activity failed with: %e", err)
	}

	if stravaActivity.Type.String() == "Run" {
		stvActivity.Type = "Running"
	} else if stravaActivity.Type.String() == "Ride" || stravaActivity.Type.String() == "EBikeRide" || stravaActivity.Type.String() == "VirtualRide" {
		stvActivity.Type = "Cycling"
	} else if stravaActivity.Type.String() == "Swim" {
		stvActivity.Type = "Swimming"
	} else {
		//I don't know, call it Activity
		stvActivity.Type = "Activity"
	}

	if gpsTrack != nil && gpsTrack.Location != nil && timeStream != nil {
		stvActivity.GPS = convertGPSTrack(gpsTrack, timeStream, altTrack)
	}
	if hrTrack != nil && hrTrack.HeartRate != nil && timeStream != nil {
		stvActivity.HeartRate = convertHeartRateTrack(hrTrack, timeStream)
	}
	return stvActivity
}

func convertGPSTrack(sourceStream *stravalib.StreamSet, timeStream *stravalib.StreamSet, elevationStream *stravalib.StreamSet) []dm.GPS {
	//merge the time stream + the location stream
	merged := mergeTimeAndLocation(timeStream.Time, sourceStream.Location, elevationStream.Elevation)

	result := make([]dm.GPS, len(sourceStream.Location.Data))
	for index, gpsTime := range merged {
		result[index] = dm.GPS{float64(gpsTime.Time), gpsTime.Alt, gpsTime.Long, gpsTime.Lat}
	}
	return result
}

type GPSTime struct {
	Time int
	Lat  float64
	Long float64
	Alt  float64
}

func mergeTimeAndLocation(timeStream *stravalib.IntegerStream, locStream *stravalib.LocationStream, altStream *stravalib.DecimalStream) []GPSTime {
	merged := make([]GPSTime, len(timeStream.Data))
	for i, t := range timeStream.Data {
		latLong := locStream.Data[i]
		alt := altStream.Data[i]
		merged[i] = GPSTime{t, latLong[0], latLong[1], alt}
	}
	return merged
}

func convertHeartRateTrack(sourceStream *stravalib.StreamSet, timeStream *stravalib.StreamSet) []dm.HeartRate {
	result := make([]dm.HeartRate, len(sourceStream.HeartRate.Data))
	for index, hr := range sourceStream.HeartRate.Data {
		hrTime := timeStream.Time.Data[index]
		result[index] = dm.HeartRate{float64(hrTime), hr}
	}
	return result
}
