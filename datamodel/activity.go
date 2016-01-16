package datamodel

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
