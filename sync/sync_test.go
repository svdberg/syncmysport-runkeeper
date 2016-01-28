package sync

import (
	"fmt"
	"testing"
	"time"
)

func TestTsAtStartOfDay(t *testing.T) {
	//Mon Jan 2 15:04:05 -0700 MST 2006
	ti, _ := time.Parse("15:04:05 02-01-2006 MST", "03:00:00 11-01-2014 UTC")
	timestampAtStartOfDay := CalculateTsAtStartOfDay(int(ti.Unix()))
	timeStringAtStartOfDay := time.Unix(int64(timestampAtStartOfDay), 0).UTC().Format("15:04:05 02-01-2006 MST")

	expectedTime := "00:01:00 11-01-2014 UTC"

	if timeStringAtStartOfDay != expectedTime {
		t.Error(fmt.Sprintf("%s is not %s", timeStringAtStartOfDay, expectedTime))
	}
}
