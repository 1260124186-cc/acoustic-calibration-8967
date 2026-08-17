package store_test

import (
	"context"
	"testing"

	"example.com/acoustic-calibration/internal/model"
	"example.com/acoustic-calibration/internal/store"
)

func TestMemoryRepositoryCopiesMutableFields(t *testing.T) {
	repo := store.NewMemoryRepository()
	instrument := model.Instrument{
		ID: "mic-copy", Label: "copy check", ChannelCount: 1,
		Limits: map[string]model.Limit{"reading": {Min: 0, Max: 10}},
		Active: true,
	}
	if err := repo.RegisterInstrument(context.Background(), instrument); err != nil {
		t.Fatal(err)
	}
	instrument.Limits["reading"] = model.Limit{Min: 90, Max: 100}
	stored, err := repo.GetInstrument(context.Background(), "mic-copy")
	if err != nil {
		t.Fatal(err)
	}
	if stored.Limits["reading"].Max != 10 {
		t.Fatalf("repository exposed mutable map: %+v", stored.Limits)
	}
}
