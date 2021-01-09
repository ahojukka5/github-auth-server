package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/golang/gddo/httputil/header"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

const ghClientIDEnv = "GITHUB_CLIENT_ID"
const ghClientSecretEnv = "GITHUB_CLIENT_SECRET"
const ghAuthURL = "https://github.com/login/oauth/access_token"
const ghUserInfoURL = "https://api.github.com/user"

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
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
}

// UserInfo contains logged user info.
type UserInfo struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Token string `json:"token"`
}

// GithubAuth is route that takes `client_id` and `code` and returns `login`,
// `email` and `token`, if authentication is succesful.
func GithubAuth(w http.ResponseWriter, r *http.Request) {

	debug := os.Getenv("GITHUB_AUTH_SERVER_DEBUG") != ""

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
		panic(err)
	}

	request, err := http.NewRequest("POST", ghAuthURL, bytes.NewBuffer(authJSON))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	if debug {
		requestDump, err := httputil.DumpRequest(request, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(requestDump))
	}

	client := &http.Client{}
	print("Sending login request to " + ghAuthURL + " ... ")
	response, err := client.Do(request)

	if debug {
		responseDump, err := httputil.DumpResponse(response, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(responseDump))
	}

	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	println(response.Status)

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	if debug {
		println(string(bodyBytes))
	}

	var authResponse AuthResponse
	err = json.Unmarshal(bodyBytes, &authResponse)
	if err != nil {
		msg := "Bad response from authentication: " + err.Error()
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	if response.StatusCode != http.StatusOK || authResponse.Error != "" {
		w.WriteHeader(response.StatusCode)
		w.Write(bodyBytes)
		return
	}

	// Send request to fetch user name and email address

	request, err = http.NewRequest("GET", ghUserInfoURL, &bytes.Buffer{})
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "token "+authResponse.AccessToken)

	print("Sending login request to " + ghUserInfoURL + " ... ")
	response, err = client.Do(request)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	println(response.Status)
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
