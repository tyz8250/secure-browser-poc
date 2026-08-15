package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/containerd/errdefs"
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
	http.HandleFunc("POST /sessions", handleCreateSession(apiClient))

	// GET /sessions/{id} 用handlerを登録する
	http.HandleFunc("GET /sessions/{id}", handleGetSession(apiClient))

	// DELETE /sessions/{id} 用handlerを登録する
	http.HandleFunc("DELETE /sessions/{id}", handleDeleteSession(apiClient))

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

// GET /sessions/{id} を処理するhandlerを作る
func handleGetSession(apiClient *client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// GET以外なら拒否する
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Container IDを取得する
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "session id is required", http.StatusBadRequest)
			return
		}

		// HTTPリクエストのcontextを取得する
		ctx := r.Context()

		// ContainerをInspectする
		container, err := apiClient.ContainerInspect(ctx, id, client.ContainerInspectOptions{})

		if err != nil {
			// Not Found判定
			if errdefs.IsNotFound(err) {
				http.Error(w, "session not found", http.StatusNotFound)
				return
			}

			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// Container Statusをレスポンスとして返す
		w.Header().Set("Content-Type", "application/json")

		type getSessionResponse struct {
			ContainerID string `json:"container_id"`
			Status      string `json:"status"`
		}

		response := getSessionResponse{
			ContainerID: container.Container.ID,
			Status:      string(container.Container.State.Status),
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// DELETE /sessions/{id} を処理するhandlerを作る
func handleDeleteSession(apiClient *client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Container IDを取得する
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "session id is required", http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		// ContainerをInspectする
		container, err := apiClient.ContainerInspect(ctx, id, client.ContainerInspectOptions{})

		if err != nil {
			// Not Found判定
			if errdefs.IsNotFound(err) {
				http.Error(w, "session not found", http.StatusNotFound)
				return
			}

			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if container.Container.State.Status == "running" {
			// Containerを停止する
			_, err := apiClient.ContainerStop(
				ctx,
				id,
				client.ContainerStopOptions{},
			)

			if err != nil {
				http.Error(w, "failed to stop container", http.StatusInternalServerError)
				return
			}
		}
		if _, err := apiClient.ContainerRemove(ctx, id, client.ContainerRemoveOptions{}); err != nil {
			http.Error(w, "failed to remove container", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
