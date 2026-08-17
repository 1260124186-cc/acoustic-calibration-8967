# Acoustic Calibration Record Service

Acoustic Calibration Record Service stores calibration instruments and evaluates
measurement runs for a small laboratory. Operators can register instruments,
submit batches of readings, review the resulting status, and query recent runs.
The service is self-contained and uses an in-memory repository so it can be
built and tested without external services.

## Structure

- `internal/model`: instrument, run, validation, and domain calculations.
- `internal/store`: thread-safe in-memory persistence and audit entries.
- `internal/service`: registration, batch processing, review, and query flows.
- `internal/httpapi`: JSON HTTP handlers for the public API.
- `cmd/calibrated`: executable HTTP server entrypoint.

## Commands

```bash
go build ./...
go test ./...
go run ./cmd/calibrated
```

The server listens on `:8080` by default. Set `CALIBRATION_ADDR` to change the
listen address.

## API

- `POST /v1/instruments` registers an instrument.
- `PUT /v1/instruments/{id}/limits` changes the reading limits.
- `POST /v1/instruments/{id}/runs` evaluates a batch of readings.
- `POST /v1/runs/{id}/review` accepts or rejects a run.
- `GET /v1/instruments/{id}/runs` lists recent runs.
- `GET /healthz` reports service health.
