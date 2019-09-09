package main

import (
	"encoding/json"
	"os"
	"time"

	log "github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	que "github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/bgentry/que-go"
	"github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/jackc/pgx"
	sync "github.com/svdberg/syncmysport-runkeeper/sync"
	shared "github.com/svdberg/syncmysport-runkeeper/syncmysport-shared"
)

//CONFIG
var (
	DbConnectionString string
)

//for queuing
var (
	qc      *que.Client
	pgxpool *pgx.ConnPool
)

func main() {
	//Load config vars
	var err error
	DbConnectionString = os.Getenv("CLEARDB_DATABASE_URL")
	dbURL := os.Getenv("DATABASE_URL")

	log.Printf("Starting SyncMySport-trigger with config: MysqlDBString: %s, PostgresDBString: %s",
		DbConnectionString, dbURL)

	success := false
	i := 0
	for i < 10 && !success {
		pgxpool, qc, err = shared.Setup(dbURL)
		if err != nil {
			if i == 4 {
				log.WithField("DATABASE_URL", dbURL).Fatal("Unable to setup queue / database")
			} else {
				log.Printf("Waiting 1 second for retry. I = %d", i)
				time.Sleep(1000 * time.Millisecond)
				i++
			}
		} else {
			success = true
		}
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
	err := repo.CreateTableIfNotExist()
	if err != nil {
		log.Fatalf("Error checking or creating the Sync database table: %s", err)
	}

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
			log.Fatalf("Error enqueuing job for sync: %s", err)
		}
	}
}
