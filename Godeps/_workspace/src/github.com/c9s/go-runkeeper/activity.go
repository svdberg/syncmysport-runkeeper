package runkeeper

import (
	"fmt"
	"time"

	timez "4d63.com/tz"
)

/*
{
	"size": 40,
	"items": [
		{
			"type": "Running",
			"start_time": "Tue, 1 Mar 2011 07:00:00",
			"total_distance": 70,
			"duration": 10,
			"source": "RunKeeper",
			"entry_mode": "API",
			"has_map": "true",
			"uri": "/activities/40"
		},
		{
		"type": "Running",
		"start_time": "Thu, 3 Mar 2011 07:00:00",
		"total_distance": 70,
		"duration": 10,
		"source": "RunKeeper",
		"entry_mode": "Web",
		"has_map": "true",
		"uri": "/activities/39"
		},
		{
		"type": "Running",
		"startTime": "Sat, 5 Mar 2011 11:00:00",
		"total_distance": 70,
		"duration": 10,
		"source": "RunKeeper",
		"entry_mode": "API",
		"has_map": "true",
		"uri": "/activities/38"
		},
		{
		"type": "Running",
		"startTime": "Mon, 7 Mar 2011 07:00:00",
		"total_distance": 70,
		"duration": 10,
		"source": "RunKeeper",
		"entry_mode": "API",
		"has_map": "false",
		"uri": "/activities/37"
		},
		⋮
	],
	"previous": "https://api.runkeeper.com/user/1234567890/activities?page=2"
}
*/

type FitnessActivityFeed struct {
	Size     int64             `json:"size"`
	Items    []FitnessActivity `json:"items"`
	Previous string            `json:"previous"`
}

/*
	{
		"type": "Running",
		"start_time": "Tue, 1 Mar 2011 07:00:00",
		"total_distance": 70,
		"duration": 10,
		"source": "RunKeeper",
		"entry_mode": "API",
		"has_map": "true",
		"uri": "/activities/40"
	}
*/
type FitnessActivity struct {
	Type          string  `json:"type"`
	StartTime     Time    `json:"start_time"`
	UtcOffset     int     `json:"utc_offset"`
	TotalDistance float64 `json:"total_distance"`
	Duration      float64 `json:"duration"`
	Source        string  `json:"source"`
	HasMap        bool    `json:"has_map"`
	HasPath       bool    `json:"has_path"`
	EntryMode     string  `json:"entry_mode"`
	Uri           string  `json:"uri"`
	Notes         string  `json:"notes"`

	// Details
	Climb            float64 `json:"climb"`
	Comment          string  `json:"comments"`           // "comments" : "/fitnessActivities/318671963/comments",
	UserID           int64   `json:"userID"`             // "userID" : 24207205,
	IsLive           bool    `json:"is_live"`            // "is_live" : false,
	Equipment        string  `json:"equipment"`          // "equipment" : "None",
	TotalCalories    float64 `json:"total_calories"`     // "total_calories" : 22,
	AverageHeartRate int     `json:"average_heart_rate"` // "average_heart_rate" : 122,

	Share    string `json:"share"`     // "share" : "Everyone",
	ShareMap string `json:"share_map"` // "share_map" : "Friends",

	Distance  []Distance  `json:"distance"` // "distance" : [ { "distance" : 0, "timestamp" : 0 }, ... ]
	Path      []Path      `json:"path"`
	HeartRate []HeartRate `json:"heart_rate"`
}

/**
type	String	The type of activity, as one of the following values: Running, Cycling, Mountain Biking, Walking, Hiking, Downhill Skiing, Cross-Country Skiing, Snowboarding, Skating, Swimming, Wheelchair, Rowing, Elliptical, Other
secondary_type	String	The secondary type of the activity, as a free-form string (max. 64 characters). This field is used only if the type field is "Other."
equipment	String	The equipment used to complete this activity, as one of the following values: None, Treadmill, Stationary Bike, Elliptical, Row Machine. (Optional; if not specified, None is assumed.)
start_time	String	The starting time for the activity (e.g., Sat, 1 Jan 2011 00:00:00)
total_distance	Double	The total distance traveled, in meters (optional)
duration	Double	The duration of the activity, in seconds
average_heart_rate	Integer	The user’s average heart rate, in beats per minute (optional)
heart_rate	HeartRate[]	The sequence of time-stamped heart rate measurements (optional)
total_calories	Double	The total calories burned (optional)
notes	String	Any notes that the user has associated with the activity
path	WGS84[]	The sequence geographical points along the route (must have at least a start point and end point; omit this field to indicate that no map exists for the activity)
post_to_facebook	Boolean	True to post this activity to Facebook, false to prevent posting (optional; if not specified, the user’s default preference is used)
post_to_twitter	Boolean	True to post this activity to Twitter, false to prevent posting (optional; if not specified, the user’s default preference is used)
detect_pauses	Boolean	True to automatically detect and insert pause points into the supplied path, false otherwise (optional; if not specified, no pause detection is performed)
*/
//Rules
//If Path = [] it should not be in the JSON
type FitnessActivityNew struct {
	Type             string      `json:"type"`
	SecondaryType    string      `json:"secondary_type,omitempty"`
	Equipment        string      `json:"equipment"` // "equipment" : "None",
	StartTime        Time        `json:"start_time"`
	TotalDistance    float64     `json:"total_distance"`
	Duration         float64     `json:"duration"`
	AverageHeartRate int         `json:"average_heart_rate"`
	Source           string      `json:"source"`
	EntryMode        string      `json:"entry_mode"`
	HeartRate        []HeartRate `json:"heart_rate"`
	TotalCalories    float64     `json:"total_calories"`
	Notes            string      `json:"notes"`
	Path             []Path      `json:"path,omitempty"`
	PostToFacebook   bool        `json:"post_to_facebook"`
	PostToTwitter    bool        `json:"post_to_twitter"`
	DetectPauses     bool        `json:"detect_pauses"`
}

//title and duration are the minimum requirements
func CreateNewFitnessActivity(title string, duration float64) *FitnessActivityNew {
	activity := FitnessActivityNew{}
	activity.Notes = title
	activity.SecondaryType = ""
	activity.Duration = duration
	activity.Equipment = "None"
	activity.Path = make([]Path, 0)
	activity.HeartRate = make([]HeartRate, 0)
	activity.PostToFacebook = false
	activity.PostToTwitter = false
	return &activity
}

type Distance struct {
	Distance  float64 `json:"distance"`  // "distance" : 0,
	Timestamp float64 `json:"timestamp"` // : 0
}

/*
   {
      "altitude" : 37,
      "longitude" : 121.371254,
      "type" : "gps",
      "timestamp" : 3.629,
      "latitude" : 24.942796
   },
*/
type Path struct {
	Altitude  float64 `json:"altitude"`
	Longitude float64 `json:"longitude"` // 121.37
	Type      string  `json:"type"`      // gps
	Latitude  float64 `json:"latitude"`
	Timestamp float64 `json:"timestamp"`
}

type HeartRate struct {
	TimeStamp  float64 `json:"timestamp"`
	HearRateNr int     `json:"heart_rate"`
}

type Time time.Time

// Unmarshal "Tue, 1 Mar 2011 07:00:00"
func (self *Time) UnmarshalJSON(data []byte) (err error) {
	if len(data) > 1 && data[0] == '"' && data[len(data)-1] == '"' {
		loc, _ := timez.LoadLocation("UTC")
		t, err := time.ParseInLocation("Mon, _2 Jan 2006 15:04:05", string(data[1:len(data)-1]), loc)
		if err != nil {
			return err
		}
		*self = Time(t)
	}
	return nil
}

func (self *Time) MarshalJSON() ([]byte, error) {
	//Mon Jan 2 15:04:05 -0700 MST 2006
	str := fmt.Sprintf("\"%s\"", time.Time(*self).Format("Mon, _2 Jan 2006 15:04:05"))
	return []byte(str), nil
}

/*
 */

/*
func (self *MetricValue) UnmarshalJSON(data []byte) error {
	if data[0] == '"' {
		var val int64
		if _, err := fmt.Sscanf(string(data), "\"%d\"", &val); err != nil {
			return err
		}
		*self = MetricValue(val)
	} else {
		val, err := strconv.Atoi(string(data))
		if err != nil {
			return err
		}
		*self = MetricValue(val)
	}
	return nil
}
*/
/*
type Duration float64
func ParseDurationInSeconds(duration string) (Duration, error) {
	var hours, minutes int
	var seconds float64
	if _, err := fmt.Sscanf(duration, "\"%02d:%02d:%f\"", &hours, &minutes, &seconds); err != nil {
		return 0, err
	}
	return Duration(float64(hours)*60*60 + float64(minutes)*60 + seconds), nil
}
func (self *Duration) UnmarshalText(data []byte) (err error) {
	// Fractional seconds are handled implicitly by Parse.
	*self, err = ParseDurationInSeconds(string(data))
	return err
}
func (self *Duration) UnmarshalJSON(data []byte) (err error) {
	// Fractional seconds are handled implicitly by Parse.
	*self, err = ParseDurationInSeconds(string(data))
	return err
}
*/
