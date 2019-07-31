package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/newrelic/go-agent"
	rk "github.com/svdberg/syncmysport-runkeeper/runkeeper"
	stv "github.com/svdberg/syncmysport-runkeeper/strava"
	sync "github.com/svdberg/syncmysport-runkeeper/sync"
	shared "github.com/svdberg/syncmysport-runkeeper/syncmysport-shared"

	log "github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/bgentry/que-go"
	"github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/jackc/pgx"
)

var (
	qc      *que.Client
	pgxpool *pgx.ConnPool
	app     newrelic.Application
)

const tsDelta = -45 //minutes

func init() {
	key := os.Getenv("NEW_RELIC_KEY")
	if key != "" {
		config := newrelic.NewConfig("SyncMySport-worker", key)
		var err error
		app, err = newrelic.NewApplication(config)
		if err != nil {
			log.Info("Error creating New Relic app instance, continuing...")
		}
	}
}

// syncTaskJob would do whatever syncing is necessary in the background
func syncTaskJob(j *que.Job) error {
	defer j.Delete()
	defer j.Done()

	var txn newrelic.Transaction
	if app != nil {
		txn = app.StartTransaction("sync-task", nil, nil)
		txn.SetName("sync job")
		defer txn.End()
	}

	var synctask sync.SyncTask
	err := json.Unmarshal(j.Args, &synctask)
	if err != nil {
		log.WithField("args", string(j.Args)).Error("Unable to unmarshal job arguments into SyncTask")
		return err
	}

	log.WithField("SyncTask", synctask).Info("Processing Synctask!")

	stvClientImpl := stv.CreateStravaClient(synctask.StravaToken)
	rkClientImpl := rk.CreateRKClient(synctask.RunkeeperToken)
	itemsCreatedRk, totalItems, err := synctask.Sync(stvClientImpl, rkClientImpl, txn)
	if err != nil {
		if app != nil {
			app.RecordCustomEvent("sync_error_event", map[string]interface{}{"error": err})
		}
		if txn != nil {
			txn.NoticeError(err)
		}
		log.WithField("args", string(j.Args)).WithField("QueId", j.ID).Error("Error while syncing synctask.")
		return err //don't update Timestamp, so we will retry
	}
	//update last executed timestamp
	log.Print("Updating last seen timestamp")
	//subtract 45 minutes to prevent activites being missed
	synctask.LastSeenTimestamp = int(time.Now().Add(time.Duration(tsDelta) * time.Minute).Unix())
	rowsUpdated, err := repo.UpdateSyncTask(synctask)
	if err != nil || rowsUpdated != 1 {
		log.Fatal("Error updating the SyncTask record with a new timestamp")
	}

	if app != nil {
		app.RecordCustomEvent("sync_items_created", map[string]interface{}{
			"rk-items-created": itemsCreatedRk,
			"total-items":      totalItems})
	}

	return nil
}

func main() {
	log.Print("Starting worker")

	var err error
	dbURL := os.Getenv("DATABASE_URL")
	pgxpool, qc, err = shared.Setup(dbURL)
	if err != nil {
		log.WithField("DATABASE_URL", dbURL).Fatal("Errors setting up the queue / database: ", err)
	}
	defer pgxpool.Close()

	wm := que.WorkMap{
		shared.SyncTaskJob: syncTaskJob,
	}

	// 2 worker go routines
	// LIMITED to 1 per worker.. otherwise we overload Runkeeper
	// and I guess create activties twice.
	workers := que.NewWorkerPool(qc, wm, 1)

	// Catch signal so we can shutdown gracefully
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	go workers.Start()

	// Wait for a signal
	sig := <-sigCh
	log.WithField("signal", sig).Info("Signal received. Shutting down.")

	workers.Shutdown()
}
