package runkeeper

import (
	"testing"
	"time"

	timez "4d63.com/tz"
)

func TestTzOffsetCalculator(t *testing.T) {
	activity := &FitnessActivity{}
	activity.Path = make([]Path, 2)
	activity.Path[0] = Path{1.0, -73.935242, "gps", 40.730610, 11.0} //location in New York
	activity.Path[1] = Path{2.0, 0.0, "gps", 15.0, 13.0}

	location := calculateLocationOfActivity(activity)
	want, _ := timez.LoadLocation("America/New_York")

	if location.String() != want.String() {
		t.Fatalf("Got location %s, expected %s", location, want)
	}
}

func TestTzOffsetUpdate(t *testing.T) {
	activity := &FitnessActivity{}
	activity.UtcOffset = 0
	ti, _ := time.Parse(time.RFC822, "02 Jan 19 15:04 CET")
	activity.StartTime = Time(ti)

	location, _ := timez.LoadLocation("America/New_York")

	writeTimeOffsetFromUTC(activity, location)

	if activity.UtcOffset != -5 {
		t.Fatalf("Got UTC offset %d, expected -5", activity.UtcOffset)
	}
}
