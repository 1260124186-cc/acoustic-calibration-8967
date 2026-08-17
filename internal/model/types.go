package model

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidInstrument = errors.New("invalid instrument")
	ErrInvalidRun        = errors.New("invalid calibration run")
	ErrNotFound          = errors.New("record not found")
	ErrAlreadyExists     = errors.New("record already exists")
	ErrInvalidReview     = errors.New("invalid review")
)

type RunStatus string

const (
	RunAccepted RunStatus = "accepted"
	RunRejected RunStatus = "rejected"
)

type Limit struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type Instrument struct {
	ID           string           `json:"id"`
	Label        string           `json:"label"`
	ChannelCount int              `json:"channel_count"`
	Limits       map[string]Limit `json:"limits"`
	Active       bool             `json:"active"`
	RegisteredAt time.Time        `json:"registered_at"`
}

type CalibrationRun struct {
	ID           string    `json:"id"`
	InstrumentID string    `json:"instrument_id"`
	Readings     []float64 `json:"readings"`
	Mean         float64   `json:"mean"`
	Peak         float64   `json:"peak"`
	Status       RunStatus `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	ReviewedAt   time.Time `json:"reviewed_at,omitempty"`
	ReviewNote   string    `json:"review_note,omitempty"`
}

type RegisterInstrumentInput struct {
	ID           string           `json:"id"`
	Label        string           `json:"label"`
	ChannelCount int              `json:"channel_count"`
	Limits       map[string]Limit `json:"limits"`
}

type SubmitRunInput struct {
	Readings []float64 `json:"readings"`
}

type ReviewInput struct {
	Accepted bool   `json:"accepted"`
	Note     string `json:"note"`
}

func (i RegisterInstrumentInput) Validate() error {
	if strings.TrimSpace(i.ID) == "" || strings.TrimSpace(i.Label) == "" || i.ChannelCount < 1 {
		return fmt.Errorf("%w: id, label, and channel_count are required", ErrInvalidInstrument)
	}
	if i.Limits == nil {
		i.Limits = DefaultLimits()
	}
	for name, limit := range i.Limits {
		if strings.TrimSpace(name) == "" || limit.Min > limit.Max {
			return fmt.Errorf("%w: invalid limit %q", ErrInvalidInstrument, name)
		}
	}
	return nil
}

func (i RegisterInstrumentInput) Normalized() Instrument {
	limits := CloneLimits(i.Limits)
	if limits == nil {
		limits = DefaultLimits()
	}
	return Instrument{
		ID:           strings.TrimSpace(i.ID),
		Label:        strings.TrimSpace(i.Label),
		ChannelCount: i.ChannelCount,
		Limits:       limits,
		Active:       true,
		RegisteredAt: time.Now().UTC(),
	}
}

func (r SubmitRunInput) Validate() error {
	if len(r.Readings) == 0 {
		return fmt.Errorf("%w: readings are required", ErrInvalidRun)
	}
	for _, reading := range r.Readings {
		if reading != reading {
			return fmt.Errorf("%w: readings cannot contain NaN", ErrInvalidRun)
		}
	}
	return nil
}

func (r ReviewInput) Validate() error {
	if strings.TrimSpace(r.Note) == "" {
		return fmt.Errorf("%w: review note is required", ErrInvalidReview)
	}
	return nil
}
