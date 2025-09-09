package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/dirshaye/GoMetrics/internal/agg"
	"github.com/dirshaye/GoMetrics/internal/collect"
	"github.com/dirshaye/GoMetrics/internal/rest"
)

func main() {
	// Configuration from environment variables
	port := getEnv("PORT", "8080")
	collectorInterval := getEnvDuration("COLLECTOR_INTERVAL", 5*time.Second)
	sampleInterval := getEnvDuration("SAMPLE_INTERVAL", 250*time.Millisecond)
	bufferSize := getEnvInt("BUFFER_SIZE", 100)

	log.Printf("Starting GoMetrics server...")
	log.Printf("Port: %s", port)
	log.Printf("Collector interval: %v", collectorInterval)
	log.Printf("Sample interval: %v", sampleInterval)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create aggregator
	aggregator := agg.NewAggregator(sampleInterval, bufferSize)
	metricsChan := aggregator.GetMetricsChan()

	// Create collectors
	cpuCollector := collect.NewCPUCollector(collectorInterval, metricsChan)
	memoryCollector := collect.NewMemoryCollector(collectorInterval, metricsChan)
	diskCollector := collect.NewDiskCollector(collectorInterval, metricsChan)
	networkCollector := collect.NewNetworkCollector(collectorInterval, metricsChan)

	// Start aggregator
	go aggregator.Start(ctx)

	// Start all collectors
	go cpuCollector.Start(ctx)
	go memoryCollector.Start(ctx)
	go diskCollector.Start(ctx)
	go networkCollector.Start(ctx)

	// Create REST handlers
	handlers := rest.NewHandlers(aggregator)

	// Create HTTP router using chi
	r := chi.NewRouter()

	// Add middleware (functions that run before your handlers)
	r.Use(middleware.Logger)    // Logs every HTTP request
	r.Use(middleware.Recoverer) // Recovers from panics and returns 500

	// Health endpoints (required for Kubernetes)
	r.Get("/healthz", handlers.HealthzHandler) // Liveness probe
	r.Get("/readyz", handlers.ReadyzHandler)   // Readiness probe

	// Metrics endpoints
	r.Get("/metrics/latest", handlers.MetricsLatestHandler) // JSON metrics
	r.Handle("/metrics", handlers.PrometheusHandler())      // Prometheus metrics

	addr := ":" + port

	// Create HTTP server
	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine so it doesn't block
	go func() {
		log.Printf("Starting server on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	log.Println("Shutdown signal received")

	// Cancel context to stop collectors and aggregator
	cancel()

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Gracefully shutdown the server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server shutdown complete")
}

// getEnv - Helper function to get environment variable with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvDuration - Helper function to get duration from environment variable
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getEnvInt - Helper function to get integer from environment variable
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
