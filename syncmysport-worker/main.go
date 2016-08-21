package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

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
)

// syncTaskJob would do whatever syncing is necessary in the background
func syncTaskJob(j *que.Job) error {
	var synctask sync.SyncTask
	err := json.Unmarshal(j.Args, &synctask)
	if err != nil {
		log.WithField("args", string(j.Args)).Error("Unable to unmarshal job arguments into SyncTask")
		return err
	}

	log.WithField("SyncTask", synctask).Info("Processing Synctask!")

	stvClientImpl := stv.CreateStravaClient(synctask.StravaToken)
	rkClientImpl := rk.CreateRKClient(synctask.RunkeeperToken)
	_, _, err = synctask.Sync(stvClientImpl, rkClientImpl)
	if err != nil {
		log.WithField("args", string(j.Args)).WithField("QueId", j.ID).Error("Error while syncing synctask.")
		return err
	}

	j.Delete()
	j.Done()

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
