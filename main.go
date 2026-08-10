package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/moby/moby/client"
)

func main() {
	// Docker Clientを作る
	apiClient, err := client.New(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}
	defer apiClient.Close()

	// POST /sessions 用handlerを登録する
	http.HandleFunc("/sessions", handleCreateSession(apiClient))

	// :8080でHTTP Serverを起動する
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

// POST /sessions を処理するhandlerを作る
func handleCreateSession(apiClient *client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// POST以外なら拒否する
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// HTTPリクエストのcontextを取得する
		ctx := r.Context()

		// ContainerをCreateする
		resp, err := apiClient.ContainerCreate(
			ctx, client.ContainerCreateOptions{
				Image: "nginx:alpine",
			},
		)
		if err != nil {
			http.Error(w, "failed to create container", http.StatusInternalServerError)
			return
		}

		fmt.Println("Container created:", resp.ID)

		// Docker start
		_, err = apiClient.ContainerStart(
			ctx,
			resp.ID,
			client.ContainerStartOptions{},
		)
		if err != nil {
			http.Error(w, "failed to start container", http.StatusInternalServerError)
			return
		}

		fmt.Println("Container started:", resp.ID)

		// Container IDをレスポンスとして返す
		w.Header().Set("Content-Type", "application/json")

		type createSessionResponse struct {
			ContainerID string `json:"container_id"`
			Status      string `json:"status"`
		}

		response := createSessionResponse{
			ContainerID: resp.ID,
			Status:      "running",
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
