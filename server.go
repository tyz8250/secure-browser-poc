package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

var errContainerNotFound = errors.New("container not found")

const (
	sessionStatusRunning = "running"
	sessionStatusStopped = "stopped"
)

type containerInfo struct {
	ID     string
	Status string
}

type containerRuntime interface {
	Create(context.Context, string) (containerInfo, error)
	Start(context.Context, string) error
	FindBySessionID(context.Context, string) (containerInfo, error)
	Stop(context.Context, string) error
	Remove(context.Context, string) error
}

type server struct {
	runtime      containerRuntime
	newSessionID func() string
}

func newHandler(runtime containerRuntime) http.Handler {
	return newHandlerWithIDGenerator(runtime, uuid.NewString)
}

func newHandlerWithIDGenerator(runtime containerRuntime, newSessionID func() string) http.Handler {
	s := &server{runtime: runtime, newSessionID: newSessionID}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /sessions", s.handleCreateSession)
	mux.HandleFunc("GET /sessions/{id}", s.handleGetSession)
	mux.HandleFunc("DELETE /sessions/{id}", s.handleDeleteSession)
	return mux
}

func (s *server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	sessionID := s.newSessionID()
	container, err := s.runtime.Create(r.Context(), sessionID)
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
	w.Header().Set("Location", "/sessions/"+sessionID)
	w.WriteHeader(http.StatusCreated)
	response := sessionResponse{
		SessionID: sessionID,
		Status:    sessionStatusRunning,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *server) handleGetSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	if sessionID == "" {
		http.Error(w, "session id is required", http.StatusBadRequest)
		return
	}

	container, err := s.runtime.FindBySessionID(r.Context(), sessionID)
	if err != nil {
		if errors.Is(err, errContainerNotFound) {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := sessionResponse{
		SessionID: sessionID,
		Status:    toSessionStatus(container.Status),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *server) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	if sessionID == "" {
		http.Error(w, "session id is required", http.StatusBadRequest)
		return
	}

	container, err := s.runtime.FindBySessionID(r.Context(), sessionID)
	if err != nil {
		if errors.Is(err, errContainerNotFound) {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if container.Status == "running" {
		if err := s.runtime.Stop(r.Context(), container.ID); err != nil {
			http.Error(w, "failed to stop container", http.StatusInternalServerError)
			return
		}
	}
	if err := s.runtime.Remove(r.Context(), container.ID); err != nil {
		http.Error(w, "failed to remove container", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type sessionResponse struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
}

func toSessionStatus(containerStatus string) string {
	if containerStatus == "running" {
		return sessionStatusRunning
	}
	return sessionStatusStopped
}
