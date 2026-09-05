package main

import (
	"log"
	"net/http"

	"github.com/moby/moby/client"
)

func main() {
	apiClient, err := client.New(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}
	defer apiClient.Close()

	runtime := &dockerRuntime{client: apiClient}
	if err := http.ListenAndServe(":8080", newHandler(runtime)); err != nil {
		log.Fatal(err)
	}
}
