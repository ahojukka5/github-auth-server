package main

import "os"

func main() {
	const ghSecretEnv = "GITHUB_CLIENT_SECRET"
	if _, exists := os.LookupEnv(ghSecretEnv); !exists {
		println("Environment variable " + ghSecretEnv + " + must be set.")
		return
	}
}
