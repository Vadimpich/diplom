package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"diplom/internal/observability"
)

type Handler struct {
	db              healthChecker
	readinessChecks map[string]func(context.Context) error
	metrics         *observability.MetricsRegistry
}

type healthChecker interface {
	Ping(context.Context) error
}

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

type readyResponse struct {
	Status       string            `json:"status"`
	Service      string            `json:"service"`
	Dependencies map[string]string `json:"dependencies"`
}

func (h Handler) Health(w http.ResponseWriter, r *http.Request) {
	response := healthResponse{
		Status:  "ok",
		Service: "core-backend",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func (h Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	response := readyResponse{
		Status:       "ready",
		Service:      "core-backend",
		Dependencies: map[string]string{},
	}
	statusCode := http.StatusOK
	if len(h.readinessChecks) == 0 {
		response.Status = "degraded"
		statusCode = http.StatusServiceUnavailable
	}
	for name, check := range h.readinessChecks {
		if err := check(ctx); err != nil {
			response.Dependencies[name] = "down"
			statusCode = http.StatusServiceUnavailable
			response.Status = "degraded"
			if h.metrics != nil {
				h.metrics.SetDependencyUp(name, false)
			}
			continue
		}
		response.Dependencies[name] = "up"
		if h.metrics != nil {
			h.metrics.SetDependencyUp(name, true)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(response)
}

func (h Handler) Metrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	if h.metrics == nil {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(""))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(h.metrics.Render()))
}
