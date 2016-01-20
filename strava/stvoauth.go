package strava

import (
	"encoding/json"
	"fmt"
	"github.com/strava/go.strava"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const port = 8080 // port of local demo server
const tokenfile = ".stv_bearer_token"
const secret = ".stv_app_secret"

var authenticator *strava.OAuthAuthenticator

func CheckForStvBearerToken() string {
	stat, _ := os.Stat(tokenfile)
	var bearerToken string
	if stat != nil {
		file, _ := os.Open(tokenfile)
		fileContents, _ := ioutil.ReadAll(file)
		file.Close()
		bearerToken = strings.TrimSpace(string(fileContents))
	}
	return bearerToken
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

func StartStvOAuth() {
	strava.ClientId = 9667
	strava.ClientSecret = loadSecret()

	// define a strava.OAuthAuthenticator to hold state.
	// The callback url is used to generate an AuthorizationURL.
	// The RequestClientGenerator can be used to generate an http.RequestClient.
	// This is usually when running on the Google App Engine platform.
	authenticator = &strava.OAuthAuthenticator{
		CallbackURL:            fmt.Sprintf("http://localhost:%d/exchange_token", port),
		RequestClientGenerator: nil,
	}

	http.HandleFunc("/", indexHandler)

	path, err := authenticator.CallbackPath()
	if err != nil {
		// possibly that the callback url set above is invalid
		fmt.Println(err)
		os.Exit(1)
	}
	http.HandleFunc(path, authenticator.HandlerFunc(oAuthSuccess, oAuthFailure))

	// start the server
	fmt.Printf("Visit http://localhost:%d/ to view the demo\n", port)
	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	OpenBrowserForStrava()
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// you should make this a template in your real application
	fmt.Fprintf(w, `<a href="%s">`, authenticator.AuthorizationURL("state1", strava.Permissions.Public, true))
	fmt.Fprint(w, `<img src="http://strava.github.io/api/images/ConnectWithStrava.png" />`)
	fmt.Fprint(w, `</a>`)
}

func OpenBrowserForStrava() {
	url := fmt.Sprintf(authenticator.AuthorizationURL("state1", strava.Permissions.Public, true))
	command := exec.Command("open", url)
	command.Run()
}

func oAuthSuccess(auth *strava.AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "SUCCESS:\nAt this point you can use this information to create a new user or link the account to one of your existing users\n")
	fmt.Fprintf(w, "State: %s\n\n", auth.State)
	fmt.Fprintf(w, "Access Token: %s\n\n", auth.AccessToken)

	fmt.Fprintf(w, "The Authenticated Athlete (you):\n")
	content, _ := json.MarshalIndent(auth.Athlete, "", " ")
	fmt.Fprint(w, string(content))
	file, _ := os.Create(tokenfile)
	file.WriteString(auth.AccessToken)
	file.Close()
}

func oAuthFailure(err error, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Authorization Failure:\n")

	// some standard error checking
	if err == strava.OAuthAuthorizationDeniedErr {
		fmt.Fprint(w, "The user clicked the 'Do not Authorize' button on the previous page.\n")
		fmt.Fprint(w, "This is the main error your application should handle.")
	} else if err == strava.OAuthInvalidCredentialsErr {
		fmt.Fprint(w, "You provided an incorrect client_id or client_secret.\nDid you remember to set them at the begininng of this file?")
	} else if err == strava.OAuthInvalidCodeErr {
		fmt.Fprint(w, "The temporary token was not recognized, this shouldn't happen normally")
	} else if err == strava.OAuthServerErr {
		fmt.Fprint(w, "There was some sort of server error, try again to see if the problem continues")
	} else {
		fmt.Fprint(w, err)
	}
}
