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

func takeFirstNOrMaxHr(sl []HeartRate, n int) []HeartRate {
	if len(sl) <= n {
		return sl
	} else {
		if n >= 0 {
			return sl[:n]
		} else {
			empty := make([]HeartRate, 0)
			return empty
		}
	}
}

func takeFirstNOrMax(sl []GPS, n int) []GPS {
	if len(sl) <= n {
		return sl
	} else {
		if n >= 0 {
			return sl[:n]
		} else {
			empty := make([]GPS, 0)
			return empty
		}
	}
}

func (a Activity) String() string {
	gpsSubSel := fmt.Sprintf("%v...", takeFirstNOrMax(a.GPS, 5))
	hrSubSel := fmt.Sprintf("%v...", takeFirstNOrMaxHr(a.HeartRate, 5))
	return fmt.Sprintf("Activity: %s, time: %s,  GPS: %s, HR: %s", a.Name, time.Unix(int64(a.StartTime), 0).Format("02-01-2006 15:04 MST"), gpsSubSel, hrSubSel)
}

func (a Activity) ConsideredEqual(otherActivity *Activity) bool {
	startDelta := math.Abs(float64(a.StartTime - otherActivity.StartTime))
	endDelta := math.Abs(float64((a.StartTime + otherActivity.Duration) - (otherActivity.StartTime + otherActivity.Duration)))
	return startDelta < delta && endDelta < delta
}
