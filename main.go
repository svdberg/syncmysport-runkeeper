package main

import (
	"encoding/json"
	log "github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	que "github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/bgentry/que-go"
	"github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/jackc/pgx"
	sync "github.com/svdberg/syncmysport-runkeeper/sync"
	shared "github.com/svdberg/syncmysport-runkeeper/syncmysport-shared"
	"os"
	"strconv"
	"time"
)

const tsDelta = -45 //minutes

//CONFIG
var (
	DbConnectionString string
	RkRedirectUri      string
	RkSecret           string //needed for oauth
	StvSecret          string //needed for oauth only
	StvRedirectUri     string
	Environment        string
)

//for queuing
var (
	qc      *que.Client
	pgxpool *pgx.ConnPool
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

	Environment = os.Getenv("ENVIRONMENT")

	log.Printf("Starting SyncMySport with config: Port: %d, DBString: %s, RKSecret: %s, RKRedirect: %s, StvSecret: %s, StvRedirect: %s",
		port, DbConnectionString, RkSecret, RkRedirectUri, StvSecret, StvRedirectUri)

	dbURL := os.Getenv("DATABASE_URL")
	pgxpool, qc, err = shared.Setup(dbURL)
	if err != nil {
		log.WithField("DATABASE_URL", dbURL).Fatal("Unable to setup queue / database")
	}

	defer pgxpool.Close()
	startSync()
}

// queueIndexRequest into the que as an encoded JSON object
func queueSyncTask(job sync.SyncTask) error {
	enc, err := json.Marshal(job)
	if err != nil {
		return err
	}

	j := que.Job{
		Type: shared.SyncTaskJob,
		Args: enc,
	}

	return qc.Enqueue(&j)
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
		//retrieval failed, we log and return
		log.Print("ERROR: error retrieving Sync Tasks, db down?")
		return
	}
	for _, syncer := range allSyncs {
		//que SyncTask for worker
		log.Printf("Now syncing for task: %s, %s, %s", syncer.StravaToken, syncer.RunkeeperToken, time.Unix(int64(syncer.LastSeenTimestamp), 0))
		err := queueSyncTask(syncer)
		if err != nil {
			log.Fatal("Error enqueuing job for sync: %s", err)
		}

		log.Print("Updating last seen timestamp")
		//subtract 45 minutes to prevent activites being missed
		syncer.LastSeenTimestamp = int(time.Now().Add(time.Duration(tsDelta) * time.Minute).Unix())
		rowsUpdated, err := repo.UpdateSyncTask(syncer)
		if err != nil || rowsUpdated != 1 {
			log.Fatal("Error updating the SyncTask record with a new timestamp")
		}
	}
}
