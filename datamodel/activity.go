package datamodel

import (
	"fmt"
	"math"
	"time"
)

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
	//GPS = gps
	//PrerenderedFormats = {}
	//Device = device
}

func (a Activity) String() string {
	return fmt.Sprintf("%s %s", time.Unix(int64(a.StartTime), 0).Format("02-01-2006 15:04 MST"), a.Name)
}

func (a Activity) ConsideredEqual(otherActivity *Activity) bool {
	startDelta := math.Abs(float64(a.StartTime - otherActivity.StartTime))
	endDelta := math.Abs(float64((a.StartTime + otherActivity.Duration) - (otherActivity.StartTime + otherActivity.Duration)))
	return startDelta < delta && endDelta < delta
}
