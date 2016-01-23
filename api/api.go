package api

import (
	"encoding/json"
	sync "github.com/svdberg/syncmysport-runkeeper/sync"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const startTime = -1 * time.Duration(1) * time.Hour * 24 * 365 //1 year ago

func Start() {
	router := NewRouter()
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./api/static/")))
	log.Fatal(http.ListenAndServe(":8100", router))
}

func SyncTaskIndex(response http.ResponseWriter, request *http.Request) {
}

func SyncTaskShow(response http.ResponseWriter, request *http.Request) {
}

func SyncTaskCreate(w http.ResponseWriter, r *http.Request) {
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
	db := sync.CreateSyncDbRepo()
	_, _, st, _ := db.StoreSyncTask(*syncTask)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(st); err != nil {
		panic(err)
	}
}
