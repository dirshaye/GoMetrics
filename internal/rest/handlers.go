package rest

import (
	"encoding/json"
	"net/http"

	"github.com/dirshaye/GoMetrics/internal/agg"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handlers holds the aggregator and provides HTTP handlers
type Handlers struct {
	aggregator *agg.Aggregator
}

// NewHandlers creates new REST handlers
func NewHandlers(aggregator *agg.Aggregator) *Handlers {
	return &Handlers{
		aggregator: aggregator,
	}
}

// HealthzHandler - Kubernetes liveness probe
func (h *Handlers) HealthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// ReadyzHandler - Kubernetes readiness probe
func (h *Handlers) ReadyzHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Add checks for dependencies (database, etc.)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}

// MetricsLatestHandler returns the latest metrics sample
func (h *Handlers) MetricsLatestHandler(w http.ResponseWriter, r *http.Request) {
	// Get the latest sample from aggregator
	sample := h.aggregator.GetLatestSample()

	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Check if we have any data
	if sample.Timestamp.IsZero() {
		// No data yet
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "No metrics data available yet",
		})
		return
	}

	// Return the sample as JSON
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(sample); err != nil {
		http.Error(w, "Error encoding metrics", http.StatusInternalServerError)
		return
	}
}

// PrometheusHandler returns the Prometheus metrics handler
func (h *Handlers) PrometheusHandler() http.Handler {
	return promhttp.Handler()
}
