package service_test

import (
	"context"
	"errors"
	"testing"

	"example.com/acoustic-calibration/internal/model"
	"example.com/acoustic-calibration/internal/service"
	"example.com/acoustic-calibration/internal/store"
)

func newService(t *testing.T) *service.Service {
	t.Helper()
	return service.New(store.NewMemoryRepository())
}

func register(t *testing.T, app *service.Service, id string, limits map[string]model.Limit) {
	t.Helper()
	_, err := app.RegisterInstrument(context.Background(), model.RegisterInstrumentInput{
		ID:           id,
		Label:        "reference microphone",
		ChannelCount: 2,
		Limits:       limits,
	})
	if err != nil {
		t.Fatalf("register instrument: %v", err)
	}
}

func TestRegisterAndEvaluateRun(t *testing.T) {
	app := newService(t)
	register(t, app, "mic-01", map[string]model.Limit{"reading": {Min: 20, Max: 80}})

	run, err := app.SubmitRun(context.Background(), "mic-01", model.SubmitRunInput{
		Readings: []float64{30, 50, 70},
	})
	if err != nil {
		t.Fatalf("submit run: %v", err)
	}
	if run.Status != model.RunAccepted || run.Mean != 50 || run.Peak != 70 {
		t.Fatalf("unexpected run: %+v", run)
	}
}

func TestSetLimitWithoutExplicitLimits(t *testing.T) {
	app := newService(t)
	register(t, app, "mic-02", nil)

	instrument, err := app.SetLimit(context.Background(), "mic-02", "reading", model.Limit{Min: 10, Max: 90})
	if err != nil {
		t.Fatalf("set limit: %v", err)
	}
	if instrument.Limits["reading"].Min != 10 || instrument.Limits["reading"].Max != 90 {
		t.Fatalf("limit not persisted: %+v", instrument.Limits)
	}
}

func TestRunSamplesAreIsolated(t *testing.T) {
	app := newService(t)
	register(t, app, "mic-03", nil)
	created, err := app.SubmitRun(context.Background(), "mic-03", model.SubmitRunInput{Readings: []float64{40, 45}})
	if err != nil {
		t.Fatalf("submit run: %v", err)
	}

	runs, err := app.ListRuns(context.Background(), "mic-03")
	if err != nil {
		t.Fatalf("list runs: %v", err)
	}
	runs[0].Readings[0] = 999
	again, err := app.ListRuns(context.Background(), "mic-03")
	if err != nil {
		t.Fatalf("list runs again: %v", err)
	}
	if again[0].Readings[0] != 40 || created.Readings[0] != 40 {
		t.Fatalf("stored readings were changed through a returned slice: %+v", again[0].Readings)
	}
}

func TestRunReadingsIsolatedFromInput(t *testing.T) {
	app := newService(t)
	register(t, app, "mic-iso", nil)

	input := model.SubmitRunInput{Readings: []float64{40, 45}}
	created, err := app.SubmitRun(context.Background(), "mic-iso", input)
	if err != nil {
		t.Fatalf("submit run: %v", err)
	}

	// 返回的 run 和请求体都不应与底层切片共享
	created.Readings[0] = 999
	input.Readings[0] = 888

	runs, err := app.ListRuns(context.Background(), "mic-iso")
	if err != nil {
		t.Fatalf("list runs: %v", err)
	}
	if runs[0].Readings[0] != 40 {
		t.Fatalf("stored readings shared with submit input or returned run: %+v", runs[0].Readings)
	}
}

func TestMissingInstrumentPreservesNotFound(t *testing.T) {
	app := newService(t)
	_, err := app.SubmitRun(context.Background(), "missing", model.SubmitRunInput{Readings: []float64{1}})
	if !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("expected errors.Is(..., ErrNotFound), got %v", err)
	}
}

func TestConcurrentBatchSubmission(t *testing.T) {
	app := newService(t)
	register(t, app, "mic-04", nil)
	inputs := make([]model.SubmitRunInput, 96)
	for i := range inputs {
		inputs[i] = model.SubmitRunInput{Readings: []float64{35, 40, 45}}
	}

	runs, err := app.SubmitBatch(context.Background(), "mic-04", inputs, 8)
	if err != nil {
		t.Fatalf("submit batch: %v", err)
	}
	if len(runs) != len(inputs) {
		t.Fatalf("got %d runs, want %d", len(runs), len(inputs))
	}
	seen := make(map[string]bool, len(runs))
	for _, run := range runs {
		if run.ID == "" || seen[run.ID] {
			t.Fatalf("duplicate or empty run id: %q", run.ID)
		}
		seen[run.ID] = true
	}
}

func TestSubmitRunStopsAfterCancellation(t *testing.T) {
	app := newService(t)
	register(t, app, "mic-05", nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := app.SubmitRun(ctx, "mic-05", model.SubmitRunInput{Readings: []float64{42}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}
