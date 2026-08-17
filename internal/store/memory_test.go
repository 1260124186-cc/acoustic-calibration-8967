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

func TestMemoryRepositoryIsolatesRunReadings(t *testing.T) {
	repo := store.NewMemoryRepository()
	if err := repo.RegisterInstrument(context.Background(), model.Instrument{
		ID: "mic-runs", Label: "runs", ChannelCount: 1,
		Limits: map[string]model.Limit{"reading": {Min: 0, Max: 120}},
	}); err != nil {
		t.Fatal(err)
	}

	readings := []float64{40, 45}
	if err := repo.SaveRun(context.Background(), model.CalibrationRun{
		ID: "run-1", InstrumentID: "mic-runs", Readings: readings,
	}); err != nil {
		t.Fatal(err)
	}

	// 提交时的切片改动不应影响已保存数据
	readings[0] = 999
	listed, err := repo.ListRuns(context.Background(), "mic-runs")
	if err != nil {
		t.Fatal(err)
	}
	if listed[0].Readings[0] != 40 {
		t.Fatalf("SaveRun shared slice with caller: %+v", listed[0].Readings)
	}

	// GetRun 返回值改动不应影响已保存数据
	got, err := repo.GetRun(context.Background(), "run-1")
	if err != nil {
		t.Fatal(err)
	}
	got.Readings[1] = 777
	again, err := repo.GetRun(context.Background(), "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if again.Readings[1] != 45 {
		t.Fatalf("GetRun exposed mutable slice: %+v", again.Readings)
	}

	// ListRuns 返回值改动不应影响已保存数据
	listed[0].Readings[0] = 555
	relisted, err := repo.ListRuns(context.Background(), "mic-runs")
	if err != nil {
		t.Fatal(err)
	}
	if relisted[0].Readings[0] != 40 {
		t.Fatalf("ListRuns exposed mutable slice: %+v", relisted[0].Readings)
	}
}
