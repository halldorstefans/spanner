package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/halldorstefans/spanner/server/internal/store"
	"github.com/halldorstefans/spanner/server/internal/telemetry"
	"log/slog"
)

type Handler struct {
	db     *store.Postgres
	logger *slog.Logger
}

func NewHandler(db *store.Postgres, logger *slog.Logger) *Handler {
	return &Handler{
		db:     db,
		logger: logger,
	}
}

func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) HandleGetSignals(w http.ResponseWriter, r *http.Request) {
	vin := chi.URLParam(r, "vin")

	if !telemetry.IsValidVIN(vin) {
		http.Error(w, "invalid VIN format", http.StatusBadRequest)
		return
	}

	signal := chi.URLParam(r, "signal")

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	limitStr := r.URL.Query().Get("limit")

	var from, to time.Time
	var err error

	if fromStr != "" {
		fromSec, err := strconv.ParseFloat(fromStr, 64)
		if err != nil {
			http.Error(w, "invalid 'from' parameter", http.StatusBadRequest)
			return
		}
		from = time.Unix(int64(fromSec), 0)
	} else {
		from = time.Now().Add(-24 * time.Hour)
	}

	if toStr != "" {
		toSec, err := strconv.ParseFloat(toStr, 64)
		if err != nil {
			http.Error(w, "invalid 'to' parameter", http.StatusBadRequest)
			return
		}
		to = time.Unix(int64(toSec), 0)
	} else {
		to = time.Now()
	}

	limit := 500
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	results, err := h.db.QuerySignals(r.Context(), vin, signal, from, to, limit)
	if err != nil {
		h.logger.Error("failed to query signals", "error", err, "vin", vin, "signal", signal)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (h *Handler) HandleGetLatest(w http.ResponseWriter, r *http.Request) {
	vin := chi.URLParam(r, "vin")

	if !telemetry.IsValidVIN(vin) {
		http.Error(w, "invalid VIN format", http.StatusBadRequest)
		return
	}

	results, err := h.db.QueryLatest(r.Context(), vin)
	if err != nil {
		h.logger.Error("failed to query latest", "error", err, "vin", vin)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func NewRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/api/health", h.HandleHealth)

	r.Get("/api/vehicles/{vin}/signals/{signal}", h.HandleGetSignals)
	r.Get("/api/vehicles/{vin}/latest", h.HandleGetLatest)

	return r
}

type Server struct {
	httpServer *http.Server
}

func NewServer(addr string, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) Addr() string {
	return s.httpServer.Addr
}
