package main

import (
	"log"
	"os"

	// strava "github.com/strava/go.strava"
	"github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/strava/go.strava"

	stv "github.com/svdberg/syncmysport-runkeeper/strava"
	"github.com/svdberg/syncmysport-runkeeper/sync"
)

/*
Migrate existing users to short-lived tokens and refresh tokens.
To do so, make a request to the
	POST https://www.strava.com/oauth/token
endpoint described in refreshing expired tokens, specifying
	grant_type = refresh_token
and
	refresh_token = {Forever_Access_Token_For_User}
Note: This endpoint will allow a forever token to be used as a refresh token,
during the migration period only.
*/
func main() {
	log.Print("Starting token migration")

	//get all the tokens
	dbConnectionString := os.Getenv("CLEARDB_DATABASE_URL")
	log.Printf("syntask connection string: %s", dbConnectionString)
	repo := sync.CreateSyncDbRepo(dbConnectionString)

	allSyncs, err := repo.RetrieveAllSyncTasks()
	log.Printf("Retrieved %d sync tasks", len(allSyncs))
	if err != nil {
		log.Fatalf("Error retrieving the sync tasks from db, aborting. %v", err)
	}

	strava.ClientId = 9667
	strava.ClientSecret = os.Getenv("STRAVA_SECRET")

	for _, sync := range allSyncs {
		//exchange the token for a short lived one and update the sync task in the repo
		client := stv.CreateStravaClient(sync.StravaToken)

		accessToken, refreshToken, err := client.RefreshToken(sync.StravaToken) //one time exchange, only works during migration period
		if err != nil {
			log.Printf("Error exchanging ever lasting token for refresh token in sync task %s: %s", sync, err)
			continue
		}

		sync.StravaToken = accessToken
		sync.StravaRefreshToken = refreshToken

		c, err := repo.UpdateSyncTask(sync)
		if err != nil || c < 1 {
			log.Printf("Error updating sync task: %s, skipping this one.. %s", sync, err)
		}
		log.Printf("Moved sync task %s to refresh tokens", sync)
	}

}
