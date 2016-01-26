package main

import (
	cron "github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/robfig/cron"
	api "github.com/svdberg/syncmysport-runkeeper/api"
	sync "github.com/svdberg/syncmysport-runkeeper/sync"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const tsDelta = -45 //minutes
const secret = ".rk_app_secret"

//CONFIG
var (
	DbConnectionString string
	RkRedirectUri      string
	RkSecret           string //needed for oauth
	StvSecret          string //needed for oauth only
	StvRedirectUri     string
	Environment        string
)

func main() {
	//Load config vars
	var err error
	portString := os.Getenv("PORT")

	if portString == "" {
		log.Print("$PORT must be set, falling back to 8100")
	}
	port, nerr := strconv.Atoi(portString)
	if nerr != nil {
		log.Print("Error converting $PORT to an int: %q - Using default", err)
		port = 8100
	}
	DbConnectionString = os.Getenv("CLEARDB_DATABASE_URL")

	RkSecret = os.Getenv("RUNKEEPER_SECRET")
	if RkSecret == "" {
		//fallback to load from file
	}
	RkRedirectUri = os.Getenv("RUNKEEPER_REDIRECT")
	if RkRedirectUri == "" {
		//fallback to load from file
		RkRedirectUri = "http://localhost:4444/code"
	}

	StvSecret = os.Getenv("STRAVA_SECRET")
	if StvSecret == "" {
		//fallback to load from file
	}
	StvRedirectUri = os.Getenv("STRAVA_REDIRECT")
	if StvRedirectUri == "" {
		//fallback to load from file
		StvRedirectUri = "http://localhost:4444/code"
	}

	Environment = os.Getenv("ENVIRONMENT")

	//Start Scheduler
	log.Printf("Starting SyncMySport with config: Port: %d, DBString: %s, RKSecret: %s, RKRedirect: %s, StvSecret: %s, StvRedirect: %s",
		port, DbConnectionString, RkSecret, RkRedirectUri, StvSecret, StvRedirectUri)
	log.Print("Starting SyncTask Scheduler")
	c := cron.New()
	err = c.AddFunc("0 5/15 * * *", startSync) //every 15 minutes, starting 5 in
	if err != nil {
		log.Fatal("Error adding the job to the scheduler", err)
	}
	c.Start()

	//Start api
	log.Print("Launching REST API")
	api.Start(DbConnectionString, port, RkSecret, RkRedirectUri, StvSecret, StvRedirectUri)
}

func loadSecret() string {
	stat, _ := os.Stat(secret)
	var secret string
	if stat != nil {
		file, _ := os.Open(secret)
		fileContents, _ := ioutil.ReadAll(file)
		file.Close()
		secret = strings.TrimSpace(string(fileContents))
	}
	return secret
}

/*
 * The Core functionality of this application.
 * - Get RK Activities since timestamp
 * - Get STV Activities since timestamp
 * - compare and calc difference
 * - store in STV
 * - update timestamp for record
 */
func startSync() {
	repo := sync.CreateSyncDbRepo(DbConnectionString)
	allSyncs, err := repo.RetrieveAllSyncTasks()
	log.Printf("Retrieved %d sync tasks", len(allSyncs))
	if err != nil {
		//retrival failed, we log and return
		log.Print("ERROR: error retrieving Sync Tasks, db down?")
		return
	}
	for _, syncer := range allSyncs {
		log.Printf("Now syncing for task: %s, %s, %s", syncer.StravaToken, syncer.RunkeeperToken, time.Unix(int64(syncer.LastSeenTimestamp), 0))
		difference, nrItemsCreated, err := syncer.Sync(Environment)
		if err != nil {
			log.Print("ERROR: errors during sync, aborting this task: %s", syncer)
			return
		}
		log.Printf("Nr of Activities missing in RunKeeper: %d, Actvities created: %d", difference, nrItemsCreated)
		if difference == nrItemsCreated {
			log.Print("Updating last seen timestamp")
			//subtract 45 minutes to prevent activites being missed
			syncer.LastSeenTimestamp = int(time.Now().Add(time.Duration(tsDelta) * time.Minute).Unix())
			rowsUpdated, err := repo.UpdateSyncTask(syncer)
			if err != nil || rowsUpdated != 1 {
				log.Fatal("Error updating the SyncTask record with a new timestamp")
			}
		} else {
			log.Print("Something went wrong storing Activities, not updating timestamp so we will retry")
		}
	}
}
