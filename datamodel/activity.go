package datamodel

import (
	"fmt"
	"math"
	"time"
)

//time delta in seconds for Activities to be considered the same
const delta = float64(100)

type Activity struct {
	StartTime int    `json:"start_time"`
	Duration  int    `json:"duration"`
	Type      string `json:"type"`
	//Laps = lapList if lapList is not None else []
	//Stats = ActivityStatistics(distance=distance)
	//TZ = tz
	//FallbackTZ = fallbackTz
	Name       string
	Notes      string
	Private    bool
	Stationary bool
	GPS        []GPS
	HeartRate  []HeartRate
	//PrerenderedFormats = {}
	//Device = device
}

type HeartRate struct {
	Timestamp float64
	Heartrate int
}

/*
 			"timestamp": 0,
      "altitude": 0,
      "longitude": -70.95182336425782,
      "latitude": 42.312620297384676,
      "type": "start"
*/
type GPS struct {
	Timestamp float64
	Altitude  float64
	Longitude float64
	Latitude  float64
}

func (a Activity) String() string {
	return fmt.Sprintf("%s %s", time.Unix(int64(a.StartTime), 0).Format("02-01-2006 15:04 MST"), a.Name)
}

func (a Activity) ConsideredEqual(otherActivity *Activity) bool {
	startDelta := math.Abs(float64(a.StartTime - otherActivity.StartTime))
	endDelta := math.Abs(float64((a.StartTime + otherActivity.Duration) - (otherActivity.StartTime + otherActivity.Duration)))
	return startDelta < delta && endDelta < delta
}
