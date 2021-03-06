package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	mux "github.com/gorilla/mux"
	strava "github.com/svdberg/go.strava"
	rk "github.com/svdberg/syncmysport-runkeeper/runkeeper"
	stv "github.com/svdberg/syncmysport-runkeeper/strava"
	sync "github.com/svdberg/syncmysport-runkeeper/sync"
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
	StaticPath         string
	authenticator      *strava.OAuthAuthenticator
)

func Start(connString string, port int, secretRk string, redirectRk string, secretStv string, redirectStv string, env string, staticPath string) {
	DbConnectionString = connString
	RkSecret = secretRk
	RedirectUriRk = redirectRk
	RedirectUriStv = redirectStv
	StvSecret = secretStv
	portString := fmt.Sprintf(":%d", port)
	Environment = env
	StaticPath = staticPath

	strava.ClientId = 9667
	strava.ClientSecret = StvSecret

	//for strava
	authenticator = &strava.OAuthAuthenticator{
		CallbackURL:            RedirectUriStv,
		RequestClientGenerator: nil,
	}

	log.Printf("callback url: %s", authenticator.AuthorizationURL("state1", strava.Permissions.Public, true))

	db := sync.CreateSyncDbRepo(DbConnectionString)
	err := db.CreateTableIfNotExist()
	if err != nil {
		log.Fatalf("Error checking or creating the Sync database table: %s", err.Error())
	}

	router := NewRouter()
	router.Methods("GET").Path("/exchange_token").Name("STVOAuthCallback").Handler(authenticator.HandlerFunc(oAuthSuccess, oAuthFailure))
	router.PathPrefix("/").Handler(NewHttpsRedirectFileHandler(http.Dir(StaticPath)))

	log.Fatal(http.ListenAndServe(portString, router))
}

func ActiveUsersShow(response http.ResponseWriter, request *http.Request) {
	db := sync.CreateSyncDbRepo(DbConnectionString)
	userCount, err := db.CountActiveUsers()
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
	}

	response.Header().Set("Content-Type", "application/json; charset=UTF-8")
	//return as json
	response.Write([]byte(fmt.Sprintf("{\"active-users\" : %d }", userCount)))
}

// Strava oAuth handler. handles the response of a Success flow
func oAuthSuccess(auth *strava.AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
	db := sync.CreateSyncDbRepo(DbConnectionString)
	task, err := db.FindSyncTaskByToken(auth.AccessToken)
	if err != nil {
		log.Printf("Error loading token %s from database, aborting...", auth.AccessToken)
		w.WriteHeader(http.StatusInternalServerError)
	}

	//task can either already exist for Strava only, or for both Strava and Runkeeper

	runkeeperToken := auth.State //We pass the RK token in through the JS in the frontend
	if task == nil && (runkeeperToken == "" || runkeeperToken == "undefined") {
		syncTask := sync.CreateSyncTask("", "", "", "", -1, Environment)
		syncTask.StravaToken = auth.AccessToken
		syncTask.LastSeenTimestamp = nowMinusOneHourInUnix()
		_, _, _, err := db.StoreSyncTask(*syncTask)
		if err == nil {
			cookie := &http.Cookie{Name: "strava", Value: fmt.Sprintf("%s", syncTask.StravaToken), Expires: time.Now().Add(356 * 24 * time.Hour), HttpOnly: false}
			cookie.Domain = "www.syncmysport.com"
			http.SetCookie(w, cookie)
		} else {
			log.Printf("Error while creating a new SyncTask: %s, err: %s", syncTask, err)
		}

	} else {

		if task == nil {
			//find the task for runkeeper
			task, err = db.FindSyncTaskByToken(runkeeperToken)
			if err != nil {
				log.Printf("Error retrieving the RK based Task on Strava Auth")
			}
			log.Printf("Found: %s for Runkeeper SyncTask.", task)
			task.StravaToken = auth.AccessToken
		}

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

		task.StravaToken = auth.AccessToken
		var i int
		i, err = db.UpdateSyncTask(*task)
		if i != 1 || err != nil {
			log.Printf("Error while updating synctask %s with token %s", task, auth.AccessToken)
		}
	}
	//redirect back to connect
	http.Redirect(w, r, "https://www.syncmysport.com/connect.html", 303) //replace by env var
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
	stvToken := request.URL.Query().Get("state")
	syncTask, err := ObtainBearerToken(code, stvToken)
	if err != nil {
		//report 40x
		response.WriteHeader(http.StatusBadRequest)
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

	http.Redirect(response, request, "https://www.syncmysport.com/connect.html", 303)
}

func ObtainBearerToken(code string, stvToken string) (*sync.SyncTask, error) {
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

		var task *sync.SyncTask
		if stvToken != "" && stvToken != "undefined" {
			task, err = db.FindSyncTaskByToken(stvToken)
		} else {
			task, err = db.FindSyncTaskByToken(token)
		}
		if task == nil || err != nil {
			syncTask := sync.CreateSyncTask("", "", "", "", -1, Environment)
			syncTask.RunkeeperToken = token
			syncTask.LastSeenTimestamp = nowMinusOneHourInUnix()
			db.StoreSyncTask(*syncTask)
			return syncTask, nil
		} else { //existing task
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

func RunkeeperDeauthorize(response http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(request.Body, 1048576))
	if err != nil {
		//malformed request
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	responseJson := make(map[string]string)
	json.Unmarshal(body, &responseJson)

	token := responseJson["access_token"]

	db := sync.CreateSyncDbRepo(DbConnectionString)
	task, err := db.FindSyncTaskByToken(token)
	if err != nil {
		response.WriteHeader(http.StatusAccepted)
		return
	}
	task.RunkeeperToken = ""
	_, err = db.UpdateSyncTask(*task)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	response.WriteHeader(http.StatusAccepted)
}

func nowMinusOneHourInUnix() int {
	now := time.Now().UTC()
	nowMinusOneHour := now.Add(time.Duration(1) * time.Hour)
	return int(nowMinusOneHour.Unix())
}

func TokenDisassociate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]
	log.Printf("Disassociating token %s", token)

	if token != "" {
		//validate this token against Strava
		stvClientImpl := stv.CreateStravaClient(token)
		rkClientImpl := rk.CreateRKClient(token)
		authInStrava := stvClientImpl.ValidateToken(token)
		authInRunkeeper := rkClientImpl.ValidateToken(token)

		if authInStrava {
			log.Printf("Token %s is valid for Strava", token)

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
			log.Printf("Removed Strava token from task %d", task.Uid)

			//We should also revoke auth at Strava
			err = stvClientImpl.DeAuthorize(token)
			if err != nil {
				log.Printf("Error while deauthorizing at strava: %s", err)
			}

			//drop cookie
			log.Printf("Removing cookie..")
			cookie := &http.Cookie{Name: "strava", Value: "", MaxAge: -1} //MaxAge will remove the cookie
			cookie.Domain = "www.syncmysport.com"
			http.SetCookie(w, cookie)

			w.Write([]byte("OK")) //200 OK
			return                //hmm
		}
		if authInRunkeeper {
			log.Printf("Token %s is valid for Runkeeper", token)

			//remove from db
			db := sync.CreateSyncDbRepo(DbConnectionString)
			task, err := db.FindSyncTaskByToken(token)
			if err != nil {
				//return 5xx?
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			task.RunkeeperToken = ""
			db.UpdateSyncTask(*task)
			log.Printf("Removed Runkeeper token from task %d", task.Uid)

			//We should also revoke auth at Runkeeper
			err = rkClientImpl.DeAuthorize(token)
			if err != nil {
				log.Printf("Error while deauthorizing at runkeeper: %s", err)
			}

			w.Write([]byte("OK")) //200 OK
			return                //hmm
		} else {
			log.Printf("Token %s is already no longer valid for Strava or Runkeeper", token)
		}
	}
	w.Write([]byte("OK")) //200 OK
}

func SyncTaskCreate(w http.ResponseWriter, r *http.Request) {
	//TODO. Check if there already is a task with either of the keys
	syncTask := sync.CreateSyncTask("", "", "", "", -1, Environment)
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
