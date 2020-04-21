package sync

import (
	"fmt"
	"testing"
	"time"

	runkeeper "github.com/svdberg/go-runkeeper"
	stravalib "github.com/svdberg/go.strava"
	rk "github.com/svdberg/syncmysport-runkeeper/runkeeper"
)

func TestTsAtStartOfDay(t *testing.T) {
	//Mon Jan 2 15:04:05 -0700 MST 2006
	ti, _ := time.Parse("15:04:05 02-01-2006 MST", "03:00:00 11-01-2014 UTC")
	timestampAtStartOfDay := calculateTsAtStartOfDay(int(ti.Unix()))
	timeStringAtStartOfDay := time.Unix(int64(timestampAtStartOfDay), 0).UTC().Format("15:04:05 02-01-2006 MST")

	expectedTime := "00:01:00 11-01-2014 UTC"

	if timeStringAtStartOfDay != expectedTime {
		t.Error(fmt.Sprintf("%s is not %s", timeStringAtStartOfDay, expectedTime))
	}
}

//keeping track
var activtiesCreated = make([]*runkeeper.FitnessActivityNew, 1)
var now = time.Now()

//mock runkeeper
type stubRK struct{}

func (rk stubRK) PostActivity(activity *runkeeper.FitnessActivityNew) (string, error) {
	activtiesCreated[0] = activity
	return "fake_uri", nil
}
func (rk stubRK) EnrichRKActivity(activitySummary *runkeeper.FitnessActivity) (*runkeeper.FitnessActivity, error) {
	return nil, nil
}
func (rk stubRK) EnrichRKActivities(activities *runkeeper.FitnessActivityFeed) ([]runkeeper.FitnessActivity, error) {
	return make([]runkeeper.FitnessActivity, 0), nil
}
func (rk stubRK) GetRKActivitiesSince(timestamp int) (*runkeeper.FitnessActivityFeed, error) {
	emptyFeed := &runkeeper.FitnessActivityFeed{0, make([]runkeeper.FitnessActivity, 0), ""}
	return emptyFeed, nil
}

func (rk stubRK) DeAuthorize(s string) error {
	return nil
}

func (rk stubRK) ValidateToken(s string) bool {
	return true
}

var stubRKImpl rk.RunkeeperCientInt = &stubRK{}

//mock stv
type stubSTV struct{}

func (stv stubSTV) DeAuthorize(s string) error {
	return nil
}

func (stv stubSTV) ValidateToken(s string) bool {
	return true
}

func (stv stubSTV) GetSTVActivitiesSince(timestamp int) ([]*stravalib.ActivitySummary, error) {
	results := make([]*stravalib.ActivitySummary, 1)
	activity := &stravalib.ActivitySummary{}
	activity.Id = 666
	results[0] = activity
	return results, nil
}
func (stv stubSTV) GetSTVDetailedActivity(activityId int64) (*stravalib.ActivityDetailed, error) {
	detailedAct := &stravalib.ActivityDetailed{}
	detailedAct.Id = activityId
	detailedAct.StartDate = now
	detailedAct.Type = stravalib.ActivityTypes.Run
	detailedAct.MovingTime = 3600
	detailedAct.ElapsedTime = 3600
	return detailedAct, nil
}
func (stv stubSTV) GetSTVActivityStream(activityId int64, streamType string) (*stravalib.StreamSet, error) {
	return nil, nil
}

func (stv stubSTV) RefreshToken(input string) (string, string, error) {
	return "new_ac_blah_blah", "new_rf_sdfsdfsdfsdfsd", nil
}

var stubStvImpl stubSTV

/*
 * Test a basic scenario from STV -> Runkeeper without any GPS or HR data.
 * Assumes activity in Local TZ
 */
func TestBasicSync(t *testing.T) {
	rkToken := "abcdef"
	stToken := "ghijkz"
	stRefresh := "sdfsdfsd"
	lastSeen := int(time.Now().Unix())
	syncTask := CreateSyncTask(rkToken, stRefresh, stToken, "", lastSeen, "Prod")
	syncTask.Sync(stubStvImpl, stubRKImpl, nil)

	expectedActivity := runkeeper.FitnessActivityNew{}
	expectedActivity.Type = "Running"
	expectedActivity.StartTime = runkeeper.Time(now)

	//RK actvitites are created in local time
	createdTimeString := time.Time(activtiesCreated[0].StartTime).Local().Format("Mon Jan 2 15:04:05 -0700 MST 2006")
	expectedTimeString := time.Time(expectedActivity.StartTime).Format("Mon Jan 2 15:04:05 -0700 MST 2006")
	if len(activtiesCreated) != 1 || createdTimeString != expectedTimeString {
		t.Error(fmt.Sprintf("%s is not %s", time.Time(activtiesCreated[0].StartTime).Local(), time.Time(expectedActivity.StartTime)))
	}
}
