package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"example.com/acoustic-calibration/internal/model"
	"example.com/acoustic-calibration/internal/store"
)

type Service struct {
	repo store.Repository
	seq  atomic.Uint64
}

func New(repo store.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) nextID(prefix string) string {
	return fmt.Sprintf("%s-%06d", prefix, s.seq.Add(1))
}

func (s *Service) RegisterInstrument(ctx context.Context, input model.RegisterInstrumentInput) (model.Instrument, error) {
	if err := input.Validate(); err != nil {
		return model.Instrument{}, err
	}
	instrument := input.Normalized()
	if err := s.repo.RegisterInstrument(ctx, instrument); err != nil {
		return model.Instrument{}, fmt.Errorf("register instrument %q: %w", instrument.ID, err)
	}
	if err := s.repo.AppendAudit(ctx, store.AuditEntry{Kind: "instrument_registered", ID: instrument.ID}); err != nil {
		return model.Instrument{}, fmt.Errorf("record instrument audit: %w", err)
	}
	return instrument, nil
}

func (s *Service) SetLimit(ctx context.Context, instrumentID, name string, limit model.Limit) (model.Instrument, error) {
	if name == "" || limit.Min > limit.Max {
		return model.Instrument{}, fmt.Errorf("%w: invalid limit", model.ErrInvalidInstrument)
	}
	instrument, err := s.repo.GetInstrument(ctx, instrumentID)
	if err != nil {
		return model.Instrument{}, fmt.Errorf("load instrument %q: %w", instrumentID, err)
	}
	if instrument.Limits == nil {
		instrument.Limits = model.DefaultLimits()
	}
	instrument.Limits[name] = limit
	if err := s.repo.UpdateInstrument(ctx, instrument); err != nil {
		return model.Instrument{}, fmt.Errorf("update instrument %q: %w", instrumentID, err)
	}
	if err := s.repo.AppendAudit(ctx, store.AuditEntry{Kind: "limit_updated", ID: instrumentID}); err != nil {
		return model.Instrument{}, fmt.Errorf("record limit audit: %w", err)
	}
	return instrument, nil
}

func (s *Service) SubmitRun(ctx context.Context, instrumentID string, input model.SubmitRunInput) (model.CalibrationRun, error) {
	if err := input.Validate(); err != nil {
		return model.CalibrationRun{}, err
	}
	if err := contextErr(ctx); err != nil {
		return model.CalibrationRun{}, err
	}
	instrument, err := s.repo.GetInstrument(ctx, instrumentID)
	if err != nil {
		return model.CalibrationRun{}, fmt.Errorf("load instrument %q: %w", instrumentID, err)
	}
	limit, ok := instrument.Limits["reading"]
	if !ok {
		limit = model.DefaultLimits()["reading"]
	}
	mean, peak, accepted, err := model.EvaluateReadings(input.Readings, limit)
	if err != nil {
		return model.CalibrationRun{}, fmt.Errorf("evaluate readings: %w", err)
	}
	if err := contextErr(ctx); err != nil {
		return model.CalibrationRun{}, err
	}
	run := model.CalibrationRun{
		ID:           s.nextID("run"),
		InstrumentID: instrumentID,
		Readings:     append([]float64(nil), input.Readings...),
		Mean:         mean,
		Peak:         peak,
		Status:       statusFor(accepted),
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.repo.SaveRun(ctx, run); err != nil {
		return model.CalibrationRun{}, fmt.Errorf("save calibration run: %w", err)
	}
	if err := s.repo.AppendAudit(ctx, store.AuditEntry{Kind: "run_submitted", ID: run.ID}); err != nil {
		return model.CalibrationRun{}, fmt.Errorf("record run audit: %w", err)
	}
	return run, nil
}

func (s *Service) SubmitBatch(ctx context.Context, instrumentID string, inputs []model.SubmitRunInput, workers int) ([]model.CalibrationRun, error) {
	if workers < 1 {
		workers = 1
	}
	type job struct {
		input model.SubmitRunInput
	}
	jobs := make(chan job)
	errs := make(chan error, len(inputs))
	runs := make([]model.CalibrationRun, 0, len(inputs))
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for item := range jobs {
				run, err := s.SubmitRun(ctx, instrumentID, item.input)
				if err != nil {
					errs <- err
					continue
				}
				runs = append(runs, run)
			}
		}()
	}
	go func() {
		defer close(jobs)
		for _, input := range inputs {
			select {
			case jobs <- job{input: input}:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		wg.Wait()
		close(errs)
	}()
	for err := range errs {
		return nil, err
	}
	if len(runs) != len(inputs) {
		return nil, fmt.Errorf("batch produced %d runs, want %d", len(runs), len(inputs))
	}
	if err := contextErr(ctx); err != nil {
		return nil, err
	}
	return runs, nil
}

func (s *Service) ReviewRun(ctx context.Context, runID string, input model.ReviewInput) (model.CalibrationRun, error) {
	if err := input.Validate(); err != nil {
		return model.CalibrationRun{}, err
	}
	run, err := s.repo.GetRun(ctx, runID)
	if err != nil {
		return model.CalibrationRun{}, fmt.Errorf("load run %q: %w", runID, err)
	}
	if input.Accepted {
		run.Status = model.RunAccepted
	} else {
		run.Status = model.RunRejected
	}
	run.ReviewNote = input.Note
	run.ReviewedAt = time.Now().UTC()
	if err := s.repo.UpdateRun(ctx, run); err != nil {
		return model.CalibrationRun{}, fmt.Errorf("update run %q: %w", runID, err)
	}
	if err := s.repo.AppendAudit(ctx, store.AuditEntry{Kind: "run_reviewed", ID: runID}); err != nil {
		return model.CalibrationRun{}, fmt.Errorf("record review audit: %w", err)
	}
	return run, nil
}

func (s *Service) ListRuns(ctx context.Context, instrumentID string) ([]model.CalibrationRun, error) {
	runs, err := s.repo.ListRuns(ctx, instrumentID)
	if err != nil {
		return nil, fmt.Errorf("list runs for %q: %w", instrumentID, err)
	}
	return runs, nil
}

func contextErr(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func statusFor(accepted bool) model.RunStatus {
	if accepted {
		return model.RunAccepted
	}
	return model.RunRejected
}
