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

func TestApproxSetDifference(t *testing.T) {
	set1 := NewActivitySet()
	activity1 := CreateActivity()
	activity1.Name = "Activity1"
	duration := 10000000
	activity1.StartTime = int(time.Now().Unix() - int64(duration))
	activity1.Duration = duration
	set1.Add(*activity1)

	activity2 := CreateActivity()
	activity2.Name = "Activity2"
	activity2.StartTime = int(time.Now().Unix())
	activity2.Duration = duration
	set1.Add(*activity2)

	set2 := NewActivitySet()
	set2.Add(*activity2)

	activity3 := CreateActivity()
	activity3.Name = "Activity3"
	activity3.StartTime = int(time.Now().Unix())
	activity3.Duration = duration
	set2.Add(*activity3)

	//{1,2}\{2,3} => {1}
	resultSet := set1.ApproxSubtract(set2)

	if resultSet.NumElements() != 1 || !resultSet.ApproxContains(*activity1) {
		t.Error(fmt.Sprintf("%d is not equal to 1", len(resultSet.slice)))
	}
}

func TestEqualsWithRKTimezones(t *testing.T) {
	activityWithTZ := CreateActivity()
	timeWithTZ, _ := time.Parse(time.RFC822, "02 Jan 19 15:04 CET")
	activityWithTZ.StartTime = int(timeWithTZ.Unix())
	activityWithTZ.UtcOffSet = -5

	activityWithOutTZ := CreateActivity()
	timeWithoutTZ, _ := time.Parse(time.RFC822, "02 Jan 19 10:04 CET")
	activityWithOutTZ.StartTime = int(timeWithoutTZ.Unix()) //this is "UTC"
	activityWithOutTZ.UtcOffSet = 0

	if !activityWithTZ.ConsideredEqual(activityWithOutTZ) {
		t.Errorf("WithTZ -> WithOutTZ: %s was not equal to %s", activityWithTZ, activityWithOutTZ)
	}

	if !activityWithOutTZ.ConsideredEqual(activityWithTZ) {
		t.Errorf("WithOutTZ -> WithTZ -> %s was not equal to %s", activityWithOutTZ, activityWithTZ)
	}
}

func TestSetDifference(t *testing.T) {
	set1 := NewActivitySet()
	activity1 := CreateActivity()
	activity1.Name = "Activity1"
	activity1.StartTime = int(time.Now().Unix())
	set1.Add(*activity1)

	activity2 := CreateActivity()
	activity2.Name = "Activity2"
	activity2.StartTime = int(time.Now().Unix())
	set1.Add(*activity2)

	set2 := NewActivitySet()
	set2.Add(*activity2)

	activity3 := CreateActivity()
	activity3.Name = "Activity3"
	activity3.StartTime = int(time.Now().Unix())
	set2.Add(*activity3)

	//{1,2}\{2,3} => {1}
	resultSet := set1.Subtract(set2)

	if resultSet.NumElements() != 1 || !resultSet.Contains(*activity1) {
		t.Error(fmt.Sprintf("%d is not equal to 1", len(resultSet.slice)))
	}
}

func TestString(t *testing.T) {
	now := time.Now()
	nowString := now.Format("02-01-2006 15:04 MST")
	activity := CreateActivity()
	activity.Name = "test-act"
	activity.StartTime = int(now.Unix())
	activity.Type = "Running"
	activity.Duration = 10
	endString := now.Add(time.Duration(10) * time.Second).Format("02-01-2006 15:04 MST")

	expectedFormatString := "Activity: test-act, Type: Running, start-time: %s,  end-time: %s, utc_offset: %d, duration: %ds, GPS: []..., HR: []..."
	expectedString := fmt.Sprintf(expectedFormatString, nowString, endString, 0, 10)

	res := fmt.Sprintf("%s", activity)
	if res != expectedString {
		t.Error(fmt.Sprintf("%s is not equal to \"%s\"", res, expectedString))
	}
}
