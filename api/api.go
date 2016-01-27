package api

import (
	"encoding/json"
	"fmt"
	"github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/strava/go.strava"
	sync "github.com/svdberg/syncmysport-runkeeper/sync"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const startTime = -1 * time.Duration(1) * time.Hour * 24 * 365 //1 year ago
const RkClientId = "73664cff18ed4800aab6cffc7ef8f4e1"

var (
	DbConnectionString string
	RkSecret           string
	StvSecret          string
	RedirectUriRk      string
	RedirectUriStv     string
	authenticator      *strava.OAuthAuthenticator
)

func Start(connString string, port int, secretRk string, redirectRk string, secretStv string, redirectStv string) {
	DbConnectionString = connString
	RkSecret = secretRk
	RedirectUriRk = redirectRk
	RedirectUriStv = redirectStv
	StvSecret = secretStv
	portString := fmt.Sprintf(":%d", port)

	strava.ClientId = 9667
	strava.ClientSecret = StvSecret

	//for strava
	authenticator = &strava.OAuthAuthenticator{
		CallbackURL:            RedirectUriStv,
		RequestClientGenerator: nil,
	}

	log.Printf("callback url: %s", authenticator.AuthorizationURL("state1", strava.Permissions.Public, true))
	router := NewRouter()
	router.Methods("GET").Path("/exchange_token").Name("STVOAuthCallback").Handler(authenticator.HandlerFunc(oAuthSuccess, oAuthFailure))
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./api/static/")))
	log.Fatal(http.ListenAndServe(portString, router))
}

func oAuthSuccess(auth *strava.AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "SUCCESS:\nAt this point you can use this information to create a new user or link the account to one of your existing users\n")
	//fmt.Fprintf(w, "State: %s\n\n", auth.State)
	//fmt.Fprintf(w, "Access Token: %s\n\n", auth.AccessToken)

	//fmt.Fprintf(w, "The Authenticated Athlete (you):\n")
	//content, _ := json.MarshalIndent(auth.Athlete, "", " ")
	//fmt.Fprint(w, string(content))

	db := sync.CreateSyncDbRepo(DbConnectionString)
	task, err := db.FindSyncTaskByToken(auth.AccessToken)
	if task == nil || err != nil {
		syncTask := sync.CreateSyncTask("", "", -1)
		syncTask.StravaToken = auth.AccessToken
		syncTask.LastSeenTimestamp = nowMinusOneHourInUnix()
		_, _, _, err := db.StoreSyncTask(*syncTask)
		if err != nil {
			expire := time.Now().AddDate(0, 0, 1)
			cookie := http.Cookie{"test", "smscookie", "/", "www.syncmysport.com", expire, expire.Format(time.UnixDate), 86400, true, true,
				fmt.Sprintf("strava=%s", syncTask.StravaToken),
				[]string{fmt.Sprintf("strava=%s", syncTask.StravaToken)}}

			http.SetCookie(w, &cookie)
		}

	} else {
		//update cookie
		expire := time.Now().AddDate(0, 1, 0) // one month
		cookie := http.Cookie{"strava", fmt.Sprintf("%s", task.StravaToken), "/", "www.syncmysport.com", expire, expire.Format(time.UnixDate), 86400, true, true,
			fmt.Sprintf("strava=%s", task.StravaToken),
			[]string{fmt.Sprintf("strava=%s", task.StravaToken)}}

		http.SetCookie(w, &cookie)

		if task.StravaToken != auth.AccessToken {
			task.StravaToken = auth.AccessToken
			db.UpdateSyncTask(*task)
		} else {
			log.Printf("Token %s is already stored for task id: %d", auth.AccessToken, task.Uid)
		}
	}
	//redirect back to connect
	http.Redirect(w, r, "/connect.html", 303)
}

func oAuthFailure(err error, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Authorization Failure:\n")

	// some standard error checking
	if err == strava.OAuthAuthorizationDeniedErr {
		fmt.Fprint(w, "The user clicked the 'Do not Authorize' button on the previous page.\n")
		fmt.Fprint(w, "This is the main error your application should handle.")
	} else if err == strava.OAuthInvalidCredentialsErr {
		fmt.Fprint(w, "You provided an incorrect client_id or client_secret.\nDid you remember to set them at the begininng of this file?")
	} else if err == strava.OAuthInvalidCodeErr {
		fmt.Fprint(w, "The temporary token was not recognized, this shouldn't happen normally")
	} else if err == strava.OAuthServerErr {
		fmt.Fprint(w, "There was some sort of server error, try again to see if the problem continues")
	} else {
		fmt.Fprint(w, err)
	}
}
func OAuthCallback(response http.ResponseWriter, request *http.Request) {
	code := request.URL.Query().Get("code")
	go ObtainBearerToken(code)
	//redirect to sign up page with Acknowledgement of success..
	response.Write([]uint8("Called Back!\n"))
}

func ObtainBearerToken(code string) {
	tokenUrl := "https://runkeeper.com/apps/token"
	formData := make(map[string][]string)
	formData["grant_type"] = []string{"authorization_code"}
	formData["code"] = []string{code}
	formData["client_id"] = []string{RkClientId}
	formData["client_secret"] = []string{RkSecret}
	formData["redirect_uri"] = []string{RedirectUriRk}
	client := new(http.Client)
	response, err := client.PostForm(tokenUrl, formData)
	responseJson := make(map[string]string)
	if err == nil {
		responseBody, _ := ioutil.ReadAll(response.Body)
		json.Unmarshal(responseBody, &responseJson)
		token := responseJson["access_token"]
		db := sync.CreateSyncDbRepo(DbConnectionString)
		task, err := db.FindSyncTaskByToken(token)
		if task == nil || err != nil {
			syncTask := sync.CreateSyncTask("", "", -1)
			syncTask.RunkeeperToken = token
			syncTask.LastSeenTimestamp = nowMinusOneHourInUnix()
			db.StoreSyncTask(*syncTask)
		} else {
			if task.RunkeeperToken != token {
				task.RunkeeperToken = token
				db.UpdateSyncTask(*task)
			} else {
				log.Printf("Token %s is already stored for task id: %d", token, task.Uid)
			}
		}
	} else {
		fmt.Print(err)
	}
}

func TestCookie(response http.ResponseWriter, request *http.Request) {
	cookie := &http.Cookie{Name: "test", Value: "tcookie", Expires: time.Now().Add(356 * 24 * time.Hour), HttpOnly: false}
	http.SetCookie(response, cookie)

	fmt.Fprintf(response, "State: %s\n\n", "Hello Cookie")

	http.Redirect(response, request, "/index.html", 303)
}

func SyncTaskIndex(response http.ResponseWriter, request *http.Request) {
}

func SyncTaskShow(response http.ResponseWriter, request *http.Request) {
}

func nowMinusOneHourInUnix() int {
	now := time.Now().UTC()
	nowMinusOneHour := now.Add(time.Duration(1) * time.Hour)
	return int(nowMinusOneHour.Unix())
}

func SyncTaskCreate(w http.ResponseWriter, r *http.Request) {
	//TODO. Check if there already is a task with either of the keys
	syncTask := sync.CreateSyncTask("", "", -1)
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, &syncTask); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}

	//Creation set the last know timestamp to 1 hour ago
	syncTask.LastSeenTimestamp = nowMinusOneHourInUnix()

	db := sync.CreateSyncDbRepo(DbConnectionString)
	_, _, st, _ := db.StoreSyncTask(*syncTask)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(st); err != nil {
		panic(err)
	}
}
