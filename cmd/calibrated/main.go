package main

import (
	"log"
	"net/http"
	"os"

	"example.com/acoustic-calibration/internal/httpapi"
	"example.com/acoustic-calibration/internal/service"
	"example.com/acoustic-calibration/internal/store"
)

func main() {
	addr := os.Getenv("CALIBRATION_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	repository := store.NewMemoryRepository()
	app := httpapi.New(service.New(repository))
	log.Printf("acoustic calibration service listening on %s", addr)
	if err := http.ListenAndServe(addr, app.Handler()); err != nil {
		log.Fatal(err)
	}
}
