package strava

import "testing"

func TestStreamsCall(t *testing.T) {
	client := CreateStravaClient("4b112d2d0dea7d7193966485fb41ab56a8917af3")
	gpsStream, err := client.GetSTVActivityStream(2577480635, "GPS")
	if gpsStream == nil || len(gpsStream.Location.Data) == 0 || err != nil {
		t.Fatalf("GPS stream was empty")
	}
}