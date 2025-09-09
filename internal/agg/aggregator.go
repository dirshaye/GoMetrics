package agg

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/dirshaye/GoMetrics/internal/collect"
	"github.com/dirshaye/GoMetrics/internal/prom"
)

// Aggregator combines metrics from multiple collectors into periodic samples
type Aggregator struct {
	// Channels to receive metrics from collectors
	metricsChan chan collect.Metric

	// Latest sample storage (thread-safe)
	mu           sync.RWMutex
	latestSample collect.Sample

	// Current metric storage (gets combined into samples)
	currentCPU     *collect.CPUMetric
	currentMemory  *collect.MemoryMetric
	currentDisk    *collect.DiskMetric
	currentNetwork *collect.NetworkMetric

	// Prometheus metrics
	promMetrics *prom.Metrics

	// Configuration
	sampleInterval time.Duration
}

// NewAggregator creates a new metrics aggregator
func NewAggregator(sampleInterval time.Duration, bufferSize int) *Aggregator {
	return &Aggregator{
		metricsChan:    make(chan collect.Metric, bufferSize),
		sampleInterval: sampleInterval,
		promMetrics:    prom.NewMetrics(),
	}
}

// GetMetricsChan returns the channel where collectors should send metrics
func (a *Aggregator) GetMetricsChan() chan<- collect.Metric {
	return a.metricsChan
}

// GetLatestSample returns the most recent complete sample (thread-safe)
func (a *Aggregator) GetLatestSample() collect.Sample {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.latestSample
}

// Start begins the aggregation process
func (a *Aggregator) Start(ctx context.Context) {
	// Ticker for creating samples every 250ms
	ticker := time.NewTicker(a.sampleInterval)
	defer ticker.Stop()

	log.Printf("Aggregator started with %v sample interval", a.sampleInterval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Aggregator stopping...")
			return

		case metric := <-a.metricsChan:
			// Store the latest metric of each type
			a.storeMetric(metric)

		case <-ticker.C:
			// Create and store a new sample every 250ms
			a.createSample()
		}
	}
}

// storeMetric stores the latest metric of each type
func (a *Aggregator) storeMetric(metric collect.Metric) {
	switch metric.Type {
	case "cpu":
		if cpuData, ok := metric.Data.(collect.CPUMetric); ok {
			a.currentCPU = &cpuData
		}
	case "memory":
		if memData, ok := metric.Data.(collect.MemoryMetric); ok {
			a.currentMemory = &memData
		}
	case "disk":
		if diskData, ok := metric.Data.(collect.DiskMetric); ok {
			a.currentDisk = &diskData
		}
	case "network":
		if netData, ok := metric.Data.(collect.NetworkMetric); ok {
			a.currentNetwork = &netData
		}
	default:
		log.Printf("Unknown metric type: %s", metric.Type)
	}
}

// createSample combines current metrics into a sample and stores it
func (a *Aggregator) createSample() {
	// Create sample with current timestamp
	sample := collect.Sample{
		Timestamp: time.Now(),
	}

	// Add available metrics (use zero values if not available)
	if a.currentCPU != nil {
		sample.CPU = *a.currentCPU
	}
	if a.currentMemory != nil {
		sample.Memory = *a.currentMemory
	}
	if a.currentDisk != nil {
		sample.Disk = *a.currentDisk
	}
	if a.currentNetwork != nil {
		sample.Network = *a.currentNetwork
	}

	// Store the sample (thread-safe)
	a.mu.Lock()
	a.latestSample = sample
	a.mu.Unlock()

	// Update Prometheus metrics
	a.promMetrics.UpdateFromSample(sample)

	log.Printf("Sample created at %v (CPU: %.1f%%, Memory: %.1f%%, Disk: %.1f%%)",
		sample.Timestamp.Format("15:04:05.000"),
		sample.CPU.OverallPercent,
		sample.Memory.UsedPercent,
		sample.Disk.UsedPercent)
}
