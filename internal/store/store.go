package store

import (
	"context"

	"example.com/acoustic-calibration/internal/model"
)

type AuditEntry struct {
	Kind string
	ID   string
}

type Repository interface {
	RegisterInstrument(context.Context, model.Instrument) error
	GetInstrument(context.Context, string) (model.Instrument, error)
	UpdateInstrument(context.Context, model.Instrument) error
	SaveRun(context.Context, model.CalibrationRun) error
	GetRun(context.Context, string) (model.CalibrationRun, error)
	UpdateRun(context.Context, model.CalibrationRun) error
	ListRuns(context.Context, string) ([]model.CalibrationRun, error)
	AppendAudit(context.Context, AuditEntry) error
	AuditLog(context.Context) ([]AuditEntry, error)
}
