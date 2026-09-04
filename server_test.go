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
	createResult  containerInfo
	inspectResult containerInfo
	calls         []string
}

func (f *fakeRuntime) Create(context.Context) (containerInfo, error) {
	f.calls = append(f.calls, "create")
	return f.createResult, nil
}

func (f *fakeRuntime) Start(_ context.Context, id string) error {
	f.calls = append(f.calls, "start:"+id)
	return nil
}

func (f *fakeRuntime) Inspect(_ context.Context, id string) (containerInfo, error) {
	f.calls = append(f.calls, "inspect:"+id)
	return f.inspectResult, nil
}

func (f *fakeRuntime) Stop(_ context.Context, id string) error {
	f.calls = append(f.calls, "stop:"+id)
	return nil
}

func (f *fakeRuntime) Remove(_ context.Context, id string) error {
	f.calls = append(f.calls, "remove:"+id)
	return nil
}

func TestCreateSessionSuccess(t *testing.T) {
	runtime := &fakeRuntime{createResult: containerInfo{ID: "container-1"}}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/sessions", nil)

	newHandler(runtime).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var response struct {
		ContainerID string `json:"container_id"`
		Status      string `json:"status"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.ContainerID != "container-1" || response.Status != "running" {
		t.Fatalf("response = %+v", response)
	}
	wantCalls := []string{"create", "start:container-1"}
	if !reflect.DeepEqual(runtime.calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", runtime.calls, wantCalls)
	}
}

func TestGetSessionSuccess(t *testing.T) {
	runtime := &fakeRuntime{inspectResult: containerInfo{ID: "container-1", Status: "running"}}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/sessions/container-1", nil)

	newHandler(runtime).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var response struct {
		ContainerID string `json:"container_id"`
		Status      string `json:"status"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.ContainerID != "container-1" || response.Status != "running" {
		t.Fatalf("response = %+v", response)
	}
}

func TestDeleteRunningSessionStopsBeforeRemove(t *testing.T) {
	runtime := &fakeRuntime{inspectResult: containerInfo{ID: "container-1", Status: "running"}}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodDelete, "/sessions/container-1", nil)

	newHandler(runtime).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	wantCalls := []string{"inspect:container-1", "stop:container-1", "remove:container-1"}
	if !reflect.DeepEqual(runtime.calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", runtime.calls, wantCalls)
	}
}
