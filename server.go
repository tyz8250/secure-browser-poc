package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var errContainerNotFound = errors.New("container not found")

type containerInfo struct {
	ID     string
	Status string
}

type containerRuntime interface {
	Create(context.Context) (containerInfo, error)
	Start(context.Context, string) error
	Inspect(context.Context, string) (containerInfo, error)
	Stop(context.Context, string) error
	Remove(context.Context, string) error
}

type server struct {
	runtime containerRuntime
}

func newHandler(runtime containerRuntime) http.Handler {
	s := &server{runtime: runtime}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /sessions", s.handleCreateSession)
	mux.HandleFunc("GET /sessions/{id}", s.handleGetSession)
	mux.HandleFunc("DELETE /sessions/{id}", s.handleDeleteSession)
	return mux
}

func (s *server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	container, err := s.runtime.Create(r.Context())
	if err != nil {
		http.Error(w, "failed to create container", http.StatusInternalServerError)
		return
	}
	fmt.Println("Container created:", container.ID)

	if err := s.runtime.Start(r.Context(), container.ID); err != nil {
		http.Error(w, "failed to start container", http.StatusInternalServerError)
		return
	}
	fmt.Println("Container started:", container.ID)

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		ContainerID string `json:"container_id"`
		Status      string `json:"status"`
	}{
		ContainerID: container.ID,
		Status:      "running",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *server) handleGetSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "session id is required", http.StatusBadRequest)
		return
	}

	container, err := s.runtime.Inspect(r.Context(), id)
	if err != nil {
		if errors.Is(err, errContainerNotFound) {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		ContainerID string `json:"container_id"`
		Status      string `json:"status"`
	}{
		ContainerID: container.ID,
		Status:      container.Status,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *server) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "session id is required", http.StatusBadRequest)
		return
	}

	container, err := s.runtime.Inspect(r.Context(), id)
	if err != nil {
		if errors.Is(err, errContainerNotFound) {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if container.Status == "running" {
		if err := s.runtime.Stop(r.Context(), id); err != nil {
			http.Error(w, "failed to stop container", http.StatusInternalServerError)
			return
		}
	}
	if err := s.runtime.Remove(r.Context(), id); err != nil {
		http.Error(w, "failed to remove container", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
