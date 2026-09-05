package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type fakeRuntime struct {
	createResult containerInfo
	findResult   containerInfo
	findErr      error
	calls        []string
}

func (f *fakeRuntime) Create(_ context.Context, sessionID string) (containerInfo, error) {
	f.calls = append(f.calls, "create:"+sessionID)
	return f.createResult, nil
}

func (f *fakeRuntime) Start(_ context.Context, id string) error {
	f.calls = append(f.calls, "start:"+id)
	return nil
}

func (f *fakeRuntime) FindBySessionID(_ context.Context, sessionID string) (containerInfo, error) {
	f.calls = append(f.calls, "find:"+sessionID)
	return f.findResult, f.findErr
}

func (f *fakeRuntime) Stop(_ context.Context, id string) error {
	f.calls = append(f.calls, "stop:"+id)
	return nil
}

func (f *fakeRuntime) Remove(_ context.Context, id string) error {
	f.calls = append(f.calls, "remove:"+id)
	return nil
}

func TestCreateSessionReturnsSessionID(t *testing.T) {
	runtime := &fakeRuntime{createResult: containerInfo{ID: "container-1"}}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/sessions", nil)
	handler := newHandlerWithIDGenerator(runtime, func() string { return "session-1" })

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if got := recorder.Header().Get("Location"); got != "/sessions/session-1" {
		t.Fatalf("Location = %q, want %q", got, "/sessions/session-1")
	}
	var response sessionResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.SessionID != "session-1" || response.Status != sessionStatusRunning {
		t.Fatalf("response = %+v", response)
	}
	wantCalls := []string{"create:session-1", "start:container-1"}
	if !reflect.DeepEqual(runtime.calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", runtime.calls, wantCalls)
	}
}

func TestGetSessionFindsManagedContainerBySessionID(t *testing.T) {
	runtime := &fakeRuntime{findResult: containerInfo{ID: "container-1", Status: "running"}}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/sessions/session-1", nil)

	newHandler(runtime).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var response sessionResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.SessionID != "session-1" || response.Status != sessionStatusRunning {
		t.Fatalf("response = %+v", response)
	}
	wantCalls := []string{"find:session-1"}
	if !reflect.DeepEqual(runtime.calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", runtime.calls, wantCalls)
	}
}

func TestGetSessionMapsNonRunningContainerToStoppedSession(t *testing.T) {
	runtime := &fakeRuntime{findResult: containerInfo{ID: "container-1", Status: "exited"}}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/sessions/session-1", nil)

	newHandler(runtime).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var response sessionResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.Status != sessionStatusStopped {
		t.Fatalf("status = %q, want %q", response.Status, sessionStatusStopped)
	}
}

func TestGetSessionReturnsNotFoundWhenRuntimeFindsNoManagedContainer(t *testing.T) {
	runtime := &fakeRuntime{findErr: errContainerNotFound}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/sessions/not-managed", nil)

	newHandler(runtime).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
	wantCalls := []string{"find:not-managed"}
	if !reflect.DeepEqual(runtime.calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", runtime.calls, wantCalls)
	}
}

func TestDeleteRunningSessionStopsBeforeRemove(t *testing.T) {
	runtime := &fakeRuntime{findResult: containerInfo{ID: "container-1", Status: "running"}}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodDelete, "/sessions/session-1", nil)

	newHandler(runtime).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	wantCalls := []string{"find:session-1", "stop:container-1", "remove:container-1"}
	if !reflect.DeepEqual(runtime.calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", runtime.calls, wantCalls)
	}
}
