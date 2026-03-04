package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/halldor03/omnidrive/persistence-service/internal/db"
)

type Handler struct {
	db *db.DB
}

func New(database *db.DB) *Handler {
	return &Handler{db: database}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) GetTelemetry(w http.ResponseWriter, r *http.Request) {
	vin := r.URL.Query().Get("vin")
	if vin == "" {
		http.Error(w, "vin is required", http.StatusBadRequest)
		return
	}

	secondsStr := r.URL.Query().Get("seconds")
	seconds := 10
	if secondsStr != "" {
		s, err := strconv.Atoi(secondsStr)
		if err != nil || s <= 0 {
			http.Error(w, "invalid seconds parameter", http.StatusBadRequest)
			return
		}
		seconds = s
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	results, err := h.db.QueryLastSeconds(ctx, vin, seconds)
	if err != nil {
		http.Error(w, "failed to query telemetry", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"vin":     vin,
		"seconds": seconds,
		"count":   len(results),
		"data":    results,
	})
}
