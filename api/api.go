package api

import (
	"encoding/json"
	"fmt"
	sync "github.com/svdberg/syncmysport-runkeeper/sync"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const startTime = -1 * time.Duration(1) * time.Hour * 24 * 365 //1 year ago
const ClientId = "73664cff18ed4800aab6cffc7ef8f4e1"

var (
	DbConnectionString string
	RkSecret           string
	RedirectUri        string
)

func Start(connString string, port int, secret string, redirect string) {
	DbConnectionString = connString
	RkSecret = secret
	RedirectUri = redirect
	portString := fmt.Sprintf(":%d", port)
	router := NewRouter()
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./api/static/")))
	log.Fatal(http.ListenAndServe(portString, router))
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
	formData["client_id"] = []string{ClientId}
	formData["client_secret"] = []string{RkSecret}
	formData["redirect_uri"] = []string{RedirectUri}
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

func SyncTaskIndex(response http.ResponseWriter, request *http.Request) {
}

func SyncTaskShow(response http.ResponseWriter, request *http.Request) {
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

	//Creation set the last know timestamp to 1 year ago
	syncTask.LastSeenTimestamp = int(time.Now().Add(startTime).Unix())
	db := sync.CreateSyncDbRepo(DbConnectionString)
	_, _, st, _ := db.StoreSyncTask(*syncTask)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(st); err != nil {
		panic(err)
	}
}
