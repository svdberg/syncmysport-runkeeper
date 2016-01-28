package sync

import (
	"fmt"
	"testing"
	"time"
)

func testTsAtStartOfDay(t *testing.T) {
	//Mon Jan 2 15:04:05 -0700 MST 2006
	ti, _ := time.Parse("15:04:05 _2-_1-2006 MST", "03:00:00 11-01-2014 UTC")
	timestampAtStartOfDay := calculateTsAtStartOfDay(int(ti.Unix()))
	timeStringAtStartOfDay := time.Unix(int64(timestampAtStartOfDay), 0).Format("15:04:05 _2-_1-2006 MST")
	expectedTime := "00:01:00 11-01-2014 UTC"
	if timeStringAtStartOfDay != expectedTime {
		t.Error(fmt.Sprintf("%s is not %s", timeStringAtStartOfDay, expectedTime))
	}
}
