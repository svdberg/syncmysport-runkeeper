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
	activity := Activity{}
	activity.Name = "test-act"
	activity.StartTime = int(now.Unix())

	res := fmt.Sprintf("%s", activity)
	if res != fmt.Sprintf("Activity: test-act, time: %s,  GPS: []..., HR: []...", nowString) {
		t.Error(fmt.Sprintf("%s is not equal to \"Activity: test-act, time: %s,  GPS: []..., HR: []...\"", res, nowString))
	}
}
