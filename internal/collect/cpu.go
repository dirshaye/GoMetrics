package collect

import (
	"context"
	"log"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/load"
)

// CPUCollector collects CPU usage metrics
type CPUCollector struct {
	interval time.Duration // How often to collect metrics
	output   chan<- Metric // Channel to send metrics to
}

// NewCPUCollector creates a new CPU collector
func NewCPUCollector(interval time.Duration, output chan<- Metric) *CPUCollector {
	return &CPUCollector{
		interval: interval,
		output:   output,
	}
}

// Start begins collecting CPU metrics in a goroutine
// ctx is used for graceful cancellation
func (c *CPUCollector) Start(ctx context.Context) {
	// Create a ticker that fires every interval
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop() // Clean up ticker when function exits

	log.Printf("CPU collector started with interval %v", c.interval)

	// Infinite loop until context is cancelled
	for {
		select {
		case <-ctx.Done():
			// Context was cancelled - shut down gracefully
			log.Println("CPU collector stopping...")
			return
		case <-ticker.C:
			// Ticker fired - collect metrics
			metric, err := c.collectCPUMetrics()
			if err != nil {
				log.Printf("Error collecting CPU metrics: %v", err)
				continue // Skip this collection cycle
			}

			// Try to send metric to output channel
			select {
			case c.output <- metric:
				// Successfully sent
			case <-ctx.Done():
				// Context cancelled while trying to send
				return
			default:
				// Channel is full - drop this metric to prevent blocking
				log.Println("CPU metric dropped: output channel full")
			}
		}
	}
}

// collectCPUMetrics gathers CPU usage data using gopsutil
func (c *CPUCollector) collectCPUMetrics() (Metric, error) {
	// Get overall CPU percentage (1-second sampling)
	overallPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return Metric{}, err
	}

	// Get per-core CPU percentages (1-second sampling)
	perCorePercent, err := cpu.Percent(time.Second, true)
	if err != nil {
		return Metric{}, err
	}

	// Get load averages (1, 5, 15 minutes)
	loadAvg, err := load.Avg()
	if err != nil {
		log.Printf("Warning: could not get load average: %v", err)
		// Don't fail completely if load average unavailable
		loadAvg = &load.AvgStat{Load1: 0, Load5: 0, Load15: 0}
	}

	// Create CPU metric struct
	cpuMetric := CPUMetric{
		OverallPercent: overallPercent[0], // First (and only) element for overall
		PerCorePercent: perCorePercent,    // Slice of per-core percentages
		LoadAverage:    []float64{loadAvg.Load1, loadAvg.Load5, loadAvg.Load15},
	}

	// Wrap in generic Metric struct
	return Metric{
		Type:      "cpu",
		Timestamp: time.Now(),
		Data:      cpuMetric,
	}, nil
}
