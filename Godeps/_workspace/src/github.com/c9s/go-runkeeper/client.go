package runkeeper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	timez "4d63.com/tz"
	latlong "github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/bradfitz/latlong"
)

const (
	ContentTypeBackgroundActivity       = "application/vnd.com.runkeeper.BackgroundActivity+json"
	ContentTypeBackgroundActivitySet    = "application/vnd.com.runkeeper.BackgroundActivitySet+json"
	ContentTypeComment                  = "application/vnd.com.runkeeper.Comment+json"
	ContentTypeDiabetesMeasurementSet   = "application/vnd.com.runkeeper.DiabetesMeasurementSet+json"
	ContentTypeFitnessActivity          = "application/vnd.com.runkeeper.FitnessActivity+json"
	ContentTypeFitnessActivityFeed      = "application/vnd.com.runkeeper.FitnessActivityFeed+json"
	ContentTypeGeneralMeasurementSet    = "application/vnd.com.runkeeper.GeneralMeasurementSet+json"
	ContentTypeNutritionSet             = "application/vnd.com.runkeeper.NutritionSet+json"
	ContentTypeProfile                  = "application/vnd.com.runkeeper.Profile+json"
	ContentTypeSleepSet                 = "application/vnd.com.runkeeper.SleepSet+json"
	ContentTypeStrengthTrainingActivity = "application/vnd.com.runkeeper.StrengthTrainingActivity+json"
	ContentTypeUser                     = "application/vnd.com.runkeeper.User+json"
	ContentTypeWeightSet                = "application/vnd.com.runkeeper.WeightSet+json"
)
const (
	ContentTypeNewSleep    = "application/vnd.com.runkeeper.NewSleep+json"
	ContentTypeNewActivity = "application/vnd.com.runkeeper.NewFitnessActivity+json"
)

const baseUrl = "https://api.runkeeper.com"
const deAuthUrl = "https://runkeeper.com/apps/de-authorize"

type Client struct {
	AccessToken string
	CookieJar   *cookiejar.Jar
	*http.Client
}

func NewClient(accessToken string) *Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	return &Client{accessToken, jar, &http.Client{Jar: jar}}
}

func (self *Client) Deauthorize() error {
	//https://runkeeper.com/apps/de-authorize
	tokenString := fmt.Sprintf("{\"access_token\":\"%s\"}", self.AccessToken)
	body := strings.NewReader(tokenString)
	req, err := http.NewRequest("POST", deAuthUrl, body)
	if err != nil {
		return err
	}
	_, err = self.Do(req)
	if err != nil {
		return err
	}
	return nil
}

/**
result should be a struct pointer
*/
func parseJsonResponse(resp *http.Response, result interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, result)
}

func (self *Client) createBaseRequest(method string, url string, contentType string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, baseUrl+url, body)
	if method != "POST" {
		req.Header.Add("Accept", contentType)
	} else {
		req.Header.Add("Content-Type", contentType)
	}
	req.Header.Add("Authorization", "Bearer "+self.AccessToken)
	if err != nil {
		return nil, err
	}
	return req, nil
}

type Params map[string]interface{}

func (self *Client) GetRequestParams(userParams *Params) url.Values {
	params := url.Values{}
	if userParams != nil {
		for key, val := range *userParams {
			switch t := val.(type) {
			case int, int8, int16, int32, int64:
				params.Set(key, strconv.Itoa(t.(int)))
			case string:
				params.Set(key, t)
			case []byte:
				params.Set(key, string(t))
			default:
				params.Set(key, t.(string))
			}
		}
	}
	return params
}

func (self *Client) GetUser() (*User, error) {
	req, err := self.createBaseRequest("GET", "/user", ContentTypeUser, nil)
	if err != nil {
		return nil, err
	}
	resp, err := self.Do(req)
	if err != nil {
		return nil, err
	}
	var user = User{}
	defer resp.Body.Close()
	if err := parseJsonResponse(resp, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

/**
Valid params:
	page: (int)
	pageSize: (int)
*/
func (self *Client) GetFitnessActivityFeed(userParams *Params) (*FitnessActivityFeed, error) {
	params := self.GetRequestParams(userParams)
	req, err := self.createBaseRequest("GET", "/fitnessActivities?"+params.Encode(), ContentTypeFitnessActivityFeed, nil)
	if err != nil {
		return nil, err
	}

	resp, err := self.Do(req)
	if err != nil {
		return nil, err
	}
	var activities = FitnessActivityFeed{}
	defer resp.Body.Close()
	if err := parseJsonResponse(resp, &activities); err != nil {
		return nil, err
	}
	return &activities, nil
}

func (self *Client) GetFitnessActivity(activityUri string, userParams *Params) (*FitnessActivity, error) {
	params := self.GetRequestParams(userParams)
	req, err := self.createBaseRequest("GET", activityUri+"?"+params.Encode(), ContentTypeFitnessActivity, nil)
	if err != nil {
		return nil, err
	}

	resp, err := self.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var activity = &FitnessActivity{}

	// re-fill the details
	if err := parseJsonResponse(resp, activity); err != nil {
		return nil, err
	}

	//Now that we know the activity, Calculate the TZ the activity was in, and change the StartTime to UTC,
	//and store the UTC offset.

	if activity.UtcOffset == 0 {
		activitiesLocation := calculateLocationOfActivity(activity)
		writeTimeOffsetFromUTC(activity, activitiesLocation)
	}
	return activity, nil
}

func correctTimeForOffsetFromUTC(activity *FitnessActivity) {
	correctedTime := Time(time.Time(activity.StartTime).Add(time.Duration(-1*activity.UtcOffset) * time.Hour))
	activity.StartTime = correctedTime
}

func writeTimeOffsetFromUTC(activity *FitnessActivity, location *time.Location) {
	//in case of UTC we dont do anything, we assume the time is already in UTC
	if location != time.UTC {
		timeInTZ := time.Time(activity.StartTime).In(location)
		_, offsetInSeconds := timeInTZ.Zone()
		activity.UtcOffset = offsetInSeconds / 60 / 60
	} else {
		activity.UtcOffset = 0
	}
}

func calculateLocationOfActivity(activity *FitnessActivity) *time.Location {
	//get the first lat/long from the GPS track
	if len(activity.Path) > 0 {
		latLong := activity.Path[0]
		timeZone := latlong.LookupZoneName(latLong.Latitude, latLong.Longitude)
		location, err := timez.LoadLocation(timeZone)
		if err != nil {
			fmt.Printf("Error while getting location from activity for TZ: %e", err)
			return time.UTC
		}
		return location
	} else {
		//Assume UTC??
		return time.UTC
	}
	//cant happen
	return nil
}

func (self *Client) PostNewFitnessActivity(activity *FitnessActivityNew) (string, error) {
	payload, err := json.Marshal(activity)
	req, err := self.createBaseRequest("POST", "/fitnessActivities", ContentTypeNewActivity, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}

	resp, err := self.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.Status == "201 Created" {
		return resp.Header.Get("Location"), nil
	}

	return "", fmt.Errorf("Activity not created, no 201 returned, got %s", resp.Status)

}
