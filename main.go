package main

import (
	cron "github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/robfig/cron"
	api "github.com/svdberg/syncmysport-runkeeper/api"
	sync "github.com/svdberg/syncmysport-runkeeper/sync"
	"log"
	"os"
	"strconv"
	"time"
)

const tsDelta = -5 //minutes

//CONFIG
var (
	DbConnectionString string
	//RkSecret           string //needed for oauth
	//StvSecret          string //needed for oauth only
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
	dbUrl := os.Getenv("CLEARDB_DATABASE_URL")
	DbConnectionString = dbUrl

	//Start Scheduler
	log.Printf("Starting SyncMySport with config: Port: %d, DBString: %s", port, DbConnectionString)
	log.Print("Starting SyncTask Scheduler")
	c := cron.New()
	err = c.AddFunc("0 5/15 * * *", startSync) //every 15 minutes, starting 5 in
	if err != nil {
		log.Fatal("Error adding the job to the scheduler", err)
	}
	c.Start()

	//Start api
	log.Print("Launching REST API")
	api.Start(DbConnectionString, port)
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
		difference, nrItemsCreated, err := syncer.Sync()
		if err != nil {
			log.Print("ERROR: errors during sync, aborting this task: %s", syncer)
			return
		}
		log.Printf("Nr of Activities missing in RunKeeper: %d, Actvities created: %d", difference, nrItemsCreated)
		if difference == nrItemsCreated {
			log.Print("Updating last seen timestamp")
			//subtract 5 minutes to prevent activites being missed
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
