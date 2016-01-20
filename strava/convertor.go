package strava

import (
	stravalib "github.com/strava/go.strava"
	dm "github.com/svdberg/syncmysport-runkeeper/datamodel"
)

func ConvertToActivity(stravaActivity *stravalib.ActivityDetailed, timeStream *stravalib.StreamSet, gpsTrack *stravalib.StreamSet, hrTrack *stravalib.StreamSet) *dm.Activity {
	stvActivity := dm.CreateActivity()
	stvActivity.StartTime = int(stravaActivity.StartDate.Unix())
	stvActivity.Duration = stravaActivity.ElapsedTime
	stvActivity.Name = stravaActivity.Name
	stvActivity.Calories = stravaActivity.Calories
	stvActivity.Distance = stravaActivity.Distance
	stvActivity.AverageHeartRate = int(stravaActivity.AverageHeartrate)

	if stravaActivity.Type.String() == "Run" {
		stvActivity.Type = "Running"
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
