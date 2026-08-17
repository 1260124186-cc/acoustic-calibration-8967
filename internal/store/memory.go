package store

import (
	"context"
	"sync"

	"example.com/acoustic-calibration/internal/model"
)

type MemoryRepository struct {
	mu           sync.RWMutex
	instruments  map[string]model.Instrument
	runs         map[string]model.CalibrationRun
	byInstrument map[string][]string
	audit        []AuditEntry
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		instruments:  make(map[string]model.Instrument),
		runs:         make(map[string]model.CalibrationRun),
		byInstrument: make(map[string][]string),
	}
}

// checkContext 在执行持久化操作前确认上下文尚未取消，
// 使客户端取消请求时能尽早中止写入，避免残留校准记录。
func checkContext(ctx context.Context) error {
	return ctx.Err()
}

func (r *MemoryRepository) RegisterInstrument(ctx context.Context, instrument model.Instrument) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.instruments[instrument.ID]; exists {
		return model.ErrAlreadyExists
	}
	r.instruments[instrument.ID] = model.CloneInstrument(instrument)
	return nil
}

func (r *MemoryRepository) GetInstrument(ctx context.Context, id string) (model.Instrument, error) {
	if err := checkContext(ctx); err != nil {
		return model.Instrument{}, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	instrument, ok := r.instruments[id]
	if !ok {
		return model.Instrument{}, model.ErrNotFound
	}
	return model.CloneInstrument(instrument), nil
}

func (r *MemoryRepository) UpdateInstrument(ctx context.Context, instrument model.Instrument) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.instruments[instrument.ID]; !ok {
		return model.ErrNotFound
	}
	r.instruments[instrument.ID] = model.CloneInstrument(instrument)
	return nil
}

func (r *MemoryRepository) SaveRun(ctx context.Context, run model.CalibrationRun) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.instruments[run.InstrumentID]; !ok {
		return model.ErrNotFound
	}
	if _, exists := r.runs[run.ID]; exists {
		return model.ErrAlreadyExists
	}
	r.runs[run.ID] = model.CloneRun(run)
	r.byInstrument[run.InstrumentID] = append(r.byInstrument[run.InstrumentID], run.ID)
	return nil
}

func (r *MemoryRepository) GetRun(ctx context.Context, id string) (model.CalibrationRun, error) {
	if err := checkContext(ctx); err != nil {
		return model.CalibrationRun{}, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	run, ok := r.runs[id]
	if !ok {
		return model.CalibrationRun{}, model.ErrNotFound
	}
	return model.CloneRun(run), nil
}

func (r *MemoryRepository) UpdateRun(ctx context.Context, run model.CalibrationRun) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.runs[run.ID]; !ok {
		return model.ErrNotFound
	}
	r.runs[run.ID] = model.CloneRun(run)
	return nil
}

func (r *MemoryRepository) ListRuns(ctx context.Context, instrumentID string) ([]model.CalibrationRun, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := append([]string(nil), r.byInstrument[instrumentID]...)
	runs := make([]model.CalibrationRun, 0, len(ids))
	for _, id := range ids {
		runs = append(runs, model.CloneRun(r.runs[id]))
	}
	return runs, nil
}

func (r *MemoryRepository) AppendAudit(ctx context.Context, entry AuditEntry) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.audit = append(r.audit, entry)
	return nil
}

func (r *MemoryRepository) AuditLog(ctx context.Context) ([]AuditEntry, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]AuditEntry(nil), r.audit...), nil
}
