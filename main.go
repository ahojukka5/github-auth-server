package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// GithubAuth is route that takes `client_id` and `code` and returns `login`,
// `email` and `token`, if authentication is succesful.
func GithubAuth(w http.ResponseWriter, r *http.Request) {
}

func main() {
	const ghSecretEnv = "GITHUB_CLIENT_SECRET"
	if _, exists := os.LookupEnv(ghSecretEnv); !exists {
		println("Environment variable " + ghSecretEnv + " must be set.")
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
