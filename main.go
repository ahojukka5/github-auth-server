package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/golang/gddo/httputil/header"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

const ghClientIDEnv = "GITHUB_CLIENT_ID"
const ghClientSecretEnv = "GITHUB_CLIENT_SECRET"
const ghAuthUrl = "https://github.com/login/oauth/access_token"

// AuthInfo contains client_id, client_secret and code.
type AuthInfo struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
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
		println(msg)
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

	fmt.Fprintf(w, "authInfo: %+v", authInfo)

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
