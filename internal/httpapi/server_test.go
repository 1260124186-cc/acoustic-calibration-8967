package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/acoustic-calibration/internal/httpapi"
	"example.com/acoustic-calibration/internal/service"
	"example.com/acoustic-calibration/internal/store"
)

func TestCalibrationHTTPWorkflow(t *testing.T) {
	app := httpapi.New(service.New(store.NewMemoryRepository()))
	server := httptest.NewServer(app.Handler())
	defer server.Close()

	createBody := bytes.NewBufferString(`{"id":"mic-http","label":"bench mic","channel_count":1}`)
	response, err := http.Post(server.URL+"/v1/instruments", "application/json", createBody)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("create status: %s", response.Status)
	}
	_ = response.Body.Close()

	runBody := bytes.NewBufferString(`{"readings":[35,40,45]}`)
	response, err = http.Post(server.URL+"/v1/instruments/mic-http/runs", "application/json", runBody)
	if err != nil {
		t.Fatalf("run request: %v", err)
	}
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("run status: %s", response.Status)
	}
	_ = response.Body.Close()

	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/v1/instruments/mic-http/runs", nil)
	if err != nil {
		t.Fatalf("list request: %v", err)
	}
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("list request: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("list status: %s", response.Status)
	}
	var runs []map[string]any
	if err := json.NewDecoder(response.Body).Decode(&runs); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("got %d runs, want 1", len(runs))
	}
}

func TestMissingInstrumentHTTPStatus(t *testing.T) {
	app := httpapi.New(service.New(store.NewMemoryRepository()))
	request := httptest.NewRequest(http.MethodPost, "/v1/instruments/unknown/runs", bytes.NewBufferString(`{"readings":[1]}`))
	recorder := httptest.NewRecorder()
	app.Handler().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusNotFound)
	}
}
