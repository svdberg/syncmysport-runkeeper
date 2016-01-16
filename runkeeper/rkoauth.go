package runkeeper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const (
	RedirectUri  = "http://localhost:4444/code"
	ClientId     = "73664cff18ed4800aab6cffc7ef8f4e1"
	ClientSecret = "76f5b6465f3b4c5f8aec9a29574d787d"
	tokenfile    = ".rk_bearer_token"
)

func OpenBrowser() {
	url := fmt.Sprintf("https://runkeeper.com/apps/authorize?client_id=%s&response_type=code&redirect_uri=%s", ClientId, RedirectUri)
	command := exec.Command("open", url)
	command.Run()
}

func OAuthCallbackServerHelloServer(w http.ResponseWriter, req *http.Request) {
	code := req.URL.Query().Get("code")
	go ObtainBearerToken(code)
	w.Write([]uint8("Called Back!\n"))
}

func ObtainBearerToken(code string) {
	tokenUrl := "https://runkeeper.com/apps/token"
	formData := make(map[string][]string)
	formData["grant_type"] = []string{"authorization_code"}
	formData["code"] = []string{code}
	formData["client_id"] = []string{ClientId}
	formData["client_secret"] = []string{ClientSecret}
	formData["redirect_uri"] = []string{RedirectUri}
	client := new(http.Client)
	response, err := client.PostForm(tokenUrl, formData)
	responseJson := make(map[string]string)
	if err == nil {
		responseBody, _ := ioutil.ReadAll(response.Body)
		json.Unmarshal(responseBody, &responseJson)
		file, _ := os.Create(tokenfile)
		file.WriteString(responseJson["access_token"])
		file.Close()
	} else {
		fmt.Print(err)
	}
}

func CheckForBearerToken() string {
	stat, _ := os.Stat(tokenfile)
	var bearerToken string
	if stat != nil {
		log.Println("found RK token")
		file, _ := os.Open(tokenfile)
		fileContents, _ := ioutil.ReadAll(file)
		file.Close()
		bearerToken = strings.TrimSpace(string(fileContents))
	}
	return bearerToken
}

func LaunchOAuth() {
	fmt.Print("No bearer token found, going through the OAuth process.\n")
	http.HandleFunc("/code", OAuthCallbackServerHelloServer)
	go http.ListenAndServe(":4444", nil)
	OpenBrowser()
}
