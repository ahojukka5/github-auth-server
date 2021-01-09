package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/golang/gddo/httputil/header"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

const ghClientIDEnv = "GITHUB_CLIENT_ID"
const ghClientSecretEnv = "GITHUB_CLIENT_SECRET"
const ghAuthURL = "https://github.com/login/oauth/access_token"

// AuthInfo contains client_id, client_secret and code.
type AuthInfo struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
}

// AuthResponse is what Github sends as a response for a request
type AuthResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorURI         string `json:"error_uri"`
}

// UserInfo contains logged user info.
type UserInfo struct {
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
	Token     string `json:"token"`
}

// GithubAuth is route that takes `client_id` and `code` and returns `login`,
// `email` and `token`, if authentication is succesful.
func GithubAuth(w http.ResponseWriter, r *http.Request) {

	value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
	if value != "application/json" {
		msg := "Content-Type is not application/json"
		http.Error(w, msg, http.StatusUnsupportedMediaType)
		return
	}

	var authInfo AuthInfo
	err := json.NewDecoder(r.Body).Decode(&authInfo)
	switch {

	case err == io.EOF:
		msg := "Bad request: empty body"
		http.Error(w, msg, http.StatusBadRequest)
		return

	case err != nil:
		msg := "Bad request: " + err.Error()
		http.Error(w, msg, http.StatusBadRequest)
		return

	}

	authInfo.ClientSecret = os.Getenv(ghClientSecretEnv)

	if authInfo.ClientID == "" {
		clientID, exists := os.LookupEnv(ghClientIDEnv)
		if exists {
			authInfo.ClientID = clientID
		} else {
			msg := "Bad request: missing parameter `client_id`"
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

	}

	if authInfo.Code == "" {
		msg := "Bad request: missing parameter 'code'"
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	authJSON, err := json.Marshal(&authInfo)
	if err != nil {
		println(err)
		return
	}

	request, err := http.NewRequest("POST", ghAuthURL, bytes.NewBuffer(authJSON))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	client := &http.Client{}
	print("Sending login request to " + ghAuthURL + " ... ")
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	println(response.Status)

	if response.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			panic(err)
		}
		w.WriteHeader(response.StatusCode)
		w.Write(bodyBytes)
		return
	}

}

func main() {
	if _, exists := os.LookupEnv(ghClientSecretEnv); !exists {
		println("Environment variable " + ghClientSecretEnv + " must be set.")
		return
	}
	app := mux.NewRouter()
	app.HandleFunc("/", GithubAuth).Methods("POST")
	println("Server listening on port 8080 ...")
	err := http.ListenAndServe(":8080", cors.Default().Handler(app))
	if err != nil {
		panic(err)
	}
}
