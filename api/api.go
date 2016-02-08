package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/strava/go.strava"
	//rk "github.com/svdberg/syncmysport-runkeeper/runkeeper"
	stv "github.com/svdberg/syncmysport-runkeeper/strava"
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
	Environment        string
	authenticator      *strava.OAuthAuthenticator
)

func Start(connString string, port int, secretRk string, redirectRk string, secretStv string, redirectStv string, env string) {
	DbConnectionString = connString
	RkSecret = secretRk
	RedirectUriRk = redirectRk
	RedirectUriStv = redirectStv
	StvSecret = secretStv
	portString := fmt.Sprintf(":%d", port)
	Environment = env

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
	db := sync.CreateSyncDbRepo(DbConnectionString)
	task, err := db.FindSyncTaskByToken(auth.AccessToken)
	if task == nil || err != nil {
		syncTask := sync.CreateSyncTask("", "", -1, Environment)
		syncTask.StravaToken = auth.AccessToken
		syncTask.LastSeenTimestamp = nowMinusOneHourInUnix()
		_, _, _, err := db.StoreSyncTask(*syncTask)
		if err != nil {
			cookie := &http.Cookie{Name: "strava", Value: fmt.Sprintf("%s", syncTask.StravaToken), Expires: time.Now().Add(356 * 24 * time.Hour), HttpOnly: false}
			cookie.Domain = "www.syncmysport.com"
			http.SetCookie(w, cookie)
		}

	} else {
		//update cookie
		cookie := &http.Cookie{Name: "strava", Value: fmt.Sprintf("%s", task.StravaToken), Expires: time.Now().Add(356 * 24 * time.Hour), HttpOnly: false}
		cookie.Domain = "www.syncmysport.com"
		http.SetCookie(w, cookie)

		//if runkeeper is set, set that cookie to
		if task.RunkeeperToken != "" {
			cookie = &http.Cookie{Name: "runkeeper", Value: fmt.Sprintf("%s", task.RunkeeperToken), Expires: time.Now().Add(356 * 24 * time.Hour), HttpOnly: false}
			cookie.Domain = "www.syncmysport.com"
			http.SetCookie(w, cookie)
		}

		if task.StravaToken != auth.AccessToken {
			task.StravaToken = auth.AccessToken
			db.UpdateSyncTask(*task)
		} else {
			log.Printf("Token %s is already stored for task id: %d", auth.AccessToken, task.Uid)
		}
	}
	//redirect back to connect
	http.Redirect(w, r, "http://www.syncmysport.com/connect.html", 303) //replace by env var
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
	syncTask, err := ObtainBearerToken(code)
	if err != nil {
		//report 40x
	}
	//if strava is set, set that cookie to
	if syncTask.StravaToken != "" {
		cookie := &http.Cookie{Name: "strava", Value: fmt.Sprintf("%s", syncTask.StravaToken), Expires: time.Now().Add(356 * 24 * time.Hour), HttpOnly: false}
		cookie.Domain = "www.syncmysport.com"
		http.SetCookie(response, cookie)
	}

	//redirect to sign up page with Acknowledgement of success..
	cookie := &http.Cookie{Name: "runkeeper", Value: fmt.Sprintf("%s", syncTask.RunkeeperToken), Expires: time.Now().Add(356 * 24 * time.Hour), HttpOnly: false}
	cookie.Domain = "www.syncmysport.com"
	http.SetCookie(response, cookie)

	http.Redirect(response, request, "http://www.syncmysport.com/connect.html", 303)
}

func ObtainBearerToken(code string) (*sync.SyncTask, error) {
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
			syncTask := sync.CreateSyncTask("", "", -1, Environment)
			syncTask.RunkeeperToken = token
			syncTask.LastSeenTimestamp = nowMinusOneHourInUnix()
			db.StoreSyncTask(*syncTask)
			return syncTask, nil
		} else {
			if task.RunkeeperToken != token {
				task.RunkeeperToken = token
				db.UpdateSyncTask(*task)
			} else {
				log.Printf("Token %s is already stored for task id: %d", token, task.Uid)
			}
			return task, nil
		}
	} else {
		fmt.Print(err)
	}
	return nil, errors.New("Not happened")
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

func TokenDisassociate(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	if token != "" {
		//validate this token against Strava
		stvClientImpl := stv.CreateStravaClient(token)
		authInStrava := stvClientImpl.ValidateToken(token)

		if authInStrava {
			//remove from db
			db := sync.CreateSyncDbRepo(DbConnectionString)
			task, err := db.FindSyncTaskByToken(token)
			if err != nil {
				//return 5xx?
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			task.StravaToken = ""
			db.UpdateSyncTask(*task)

			//We should also revoke auth at Strava
			stvClientImpl.DeAuthorize(token)

			//drop cookie
			cookie := &http.Cookie{Name: "strava", Value: "", MaxAge: -1} //MaxAge will remove the cookie
			cookie.Domain = "www.syncmysport.com"
			http.SetCookie(w, cookie)

			w.Write([]byte("OK")) //200 OK
			return                //hmm
		}
	}
	w.Write([]byte("OK")) //200 OK
	//validate this token against Runkeeper
}

func SyncTaskCreate(w http.ResponseWriter, r *http.Request) {
	//TODO. Check if there already is a task with either of the keys
	syncTask := sync.CreateSyncTask("", "", -1, Environment)
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
