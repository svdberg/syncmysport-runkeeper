package datamodel

import (
	"fmt"
	"math"
	"time"
)

//time delta in seconds for Activities to be considered the same
const delta = float64(100)
const Source = "SyncMySport"

type ActivitySet struct {
	slice []Activity
}

//functions for set
func (set *ActivitySet) Add(p Activity) {
	if !set.Contains(p) {
		set.slice = append(set.slice, p)
	}
}

func (s *ActivitySet) Get(index int) *Activity {
	return &s.slice[index]
}

func (p1 Activity) Equals(p2 Activity) bool {
	return (p1.Name == p2.Name) && (p1.StartTime == p2.StartTime)
}

func (set ActivitySet) Contains(p Activity) bool {
	for _, v := range set.slice {
		if v.Equals(p) {
			return true
		}
	}
	return false
}

func (set ActivitySet) ApproxContains(p Activity) bool {
	for _, v := range set.slice {
		if v.ConsideredEqual(&p) {
			return true
		}
	}
	return false
}

func (set ActivitySet) Subtract(other ActivitySet) *ActivitySet {
	difference := NewActivitySet()
	for _, elem := range set.slice {
		if !other.Contains(elem) {
			difference.Add(elem)
		}
	}
	return &difference
}

func (set ActivitySet) ApproxSubtract(other ActivitySet) *ActivitySet {
	difference := NewActivitySet()
	for _, elem := range set.slice {
		if !other.ApproxContains(elem) {
			difference.Add(elem)
		}
	}
	return &difference
}

func (set ActivitySet) NumElements() int {
	return len(set.slice)
}

func NewActivitySet() ActivitySet {
	return ActivitySet{(make([]Activity, 0, 10))}
}

type Activity struct {
	StartTime        int     `json:"start_time"`
	Duration         int     `json:"duration"`
	Distance         float64 `json:"distance"`
	Type             string  `json:"type"`
	Calories         float64 `json:"calories"`
	AverageHeartRate int     `json:"average_heartrate"`
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
	Source     string `json:"source"`
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

func CreateActivity() *Activity {
	activity := Activity{}
	activity.GPS = make([]GPS, 0)
	activity.HeartRate = make([]HeartRate, 0)
	activity.Source = Source
	return &activity
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
