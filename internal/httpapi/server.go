package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"example.com/acoustic-calibration/internal/model"
	"example.com/acoustic-calibration/internal/service"
)

type Server struct {
	service *service.Service
}

func New(service *service.Service) *Server {
	return &Server{service: service}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.health)
	mux.HandleFunc("/v1/instruments", s.instruments)
	mux.HandleFunc("/v1/instruments/", s.instrumentActions)
	mux.HandleFunc("/v1/runs/", s.runActions)
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) instruments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var input model.RegisterInstrumentInput
	if !decodeJSON(w, r, &input) {
		return
	}
	instrument, err := s.service.RegisterInstrument(context.Background(), input)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, instrument)
}

func (s *Server) instrumentActions(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 4 || parts[0] != "v1" || parts[1] != "instruments" {
		writeError(w, http.StatusNotFound, model.ErrNotFound)
		return
	}
	id, action := parts[2], parts[3]
	switch {
	case r.Method == http.MethodPut && action == "limits":
		var input struct {
			Name string  `json:"name"`
			Min  float64 `json:"min"`
			Max  float64 `json:"max"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		instrument, err := s.service.SetLimit(context.Background(), id, input.Name, model.Limit{Min: input.Min, Max: input.Max})
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, instrument)
	case r.Method == http.MethodPost && action == "runs":
		var input model.SubmitRunInput
		if !decodeJSON(w, r, &input) {
			return
		}
		run, err := s.service.SubmitRun(context.Background(), id, input)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, run)
	case r.Method == http.MethodGet && action == "runs":
		runs, err := s.service.ListRuns(context.Background(), id)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, runs)
	default:
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
	}
}

func (s *Server) runActions(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 4 || parts[0] != "v1" || parts[1] != "runs" || parts[3] != "review" || r.Method != http.MethodPost {
		writeError(w, http.StatusNotFound, model.ErrNotFound)
		return
	}
	var input model.ReviewInput
	if !decodeJSON(w, r, &input) {
		return
	}
	run, err := s.service.ReviewRun(context.Background(), parts[2], input)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return false
	}
	return true
}

func writeServiceError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, model.ErrInvalidInstrument), errors.Is(err, model.ErrInvalidRun), errors.Is(err, model.ErrInvalidReview):
		status = http.StatusBadRequest
	case errors.Is(err, model.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, model.ErrAlreadyExists):
		status = http.StatusConflict
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		status = http.StatusRequestTimeout
	}
	writeError(w, status, err)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
