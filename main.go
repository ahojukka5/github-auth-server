package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"

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

	var authInfo AuthInfo
	authInfo.ClientSecret = os.Getenv(ghClientSecretEnv)

	code, exists := r.URL.Query()["code"]
	if !exists || len(code[0]) < 1 {
		msg := "Missing query parameter 'code'"
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	authInfo.Code = code[0]

	clientID, exists := r.URL.Query()["client_id"]
	if !exists || len(clientID[0]) < 1 {
		clientID, exists := os.LookupEnv(ghClientIDEnv)
		if exists {
			authInfo.ClientID = clientID
		} else {
			msg := "Missing query parameter `client_id`"
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
	} else {
		authInfo.ClientID = clientID[0]
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
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	println(response.Status)

	if debug {
		responseDump, err := httputil.DumpResponse(response, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(responseDump))
	}

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

	var userInfo UserInfo
	err = json.NewDecoder(response.Body).Decode(&userInfo)
	if err != nil {
		msg := "Bad response from authentication: " + err.Error()
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	userInfo.Token = authResponse.AccessToken
	userJSON, err := json.Marshal(&userInfo)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(userJSON)
}

func main() {
	if _, exists := os.LookupEnv(ghClientSecretEnv); !exists {
		println("Environment variable " + ghClientSecretEnv + " must be set.")
		return
	}
	app := mux.NewRouter()
	app.HandleFunc("/authenticate/github", GithubAuth).Methods("GET")
	println("Server listening on port 8080 ...")
	err := http.ListenAndServe(":8080", cors.Default().Handler(app))
	if err != nil {
		panic(err)
	}
}
