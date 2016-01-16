package datamodel

import (
	"fmt"
	"testing"
	"time"
)

func TestConsideredEqual(t *testing.T) {
	startTime := time.Now()
	duration := 1000000

	activity1 := Activity{}
	activity1.StartTime = int(startTime.Unix())
	activity1.Duration = duration

	activity2 := Activity{}
	activity2.StartTime = int(startTime.Unix())
	activity2.Duration = duration

	if !activity1.ConsideredEqual(&activity2) {
		t.Error("Activities should be considered equal")
	}
}

func TestConsideredEqualOutsideOfDelta(t *testing.T) {
	startTimeOne := time.Now()
	startTimeTwo := time.Now().Add(time.Duration(delta+1) * time.Second)
	duration := 1000000

	activity1 := Activity{}
	activity1.StartTime = int(startTimeOne.Unix())
	activity1.Duration = duration

	activity2 := Activity{}
	activity2.StartTime = int(startTimeTwo.Unix())
	activity2.Duration = duration

	if activity1.ConsideredEqual(&activity2) {
		t.Error("Activities should NOT be considered equal when outside of delta range")
	}
}

func TestString(t *testing.T) {
	now := time.Now()
	nowString := now.Format("02-01-2006 15:04 MST")
	activity := Activity{}
	activity.Name = "test-act"
	activity.StartTime = int(now.Unix())

	res := fmt.Sprintf("%s", activity)
	if res != fmt.Sprintf("%s test-act", nowString) {
		t.Error(fmt.Sprintf("%s is not equal to \"%s test-act\"", res, nowString))
	}
}
