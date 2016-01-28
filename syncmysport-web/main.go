package main

import (
	api "github.com/svdberg/syncmysport-runkeeper/api"
	"log"
	"os"
	"strconv"
)

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

	//Start api
	log.Print("Launching REST API")
	api.Start(DbConnectionString, port, RkSecret, RkRedirectUri, StvSecret, StvRedirectUri, Environment)
}
