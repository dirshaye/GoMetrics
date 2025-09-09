package collect

import (
	"context"
	"log"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)

// MemoryCollector collects memory usage metrics
type MemoryCollector struct {
	interval time.Duration
	output   chan<- Metric
}

// NewMemoryCollector creates a new memory collector
func NewMemoryCollector(interval time.Duration, output chan<- Metric) *MemoryCollector {
	return &MemoryCollector{
		interval: interval,
		output:   output,
	}
}

// Start begins collecting memory metrics in a goroutine
func (m *MemoryCollector) Start(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	// Collect initial metric immediately
	m.collectAndSend()

	for {
		select {
		case <-ctx.Done():
			log.Println("Memory collector stopping...")
			return
		case <-ticker.C:
			m.collectAndSend()
		}
	}
}

// collectAndSend gathers memory metrics and sends them through the channel
func (m *MemoryCollector) collectAndSend() {
	// Get virtual memory (RAM) statistics
	vmem, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("Error collecting virtual memory metrics: %v", err)
		return
	}

	// Get swap memory statistics
	swap, err := mem.SwapMemory()
	if err != nil {
		log.Printf("Error collecting swap memory metrics: %v", err)
		return
	}

	// Create memory metric
	memMetric := MemoryMetric{
		// Virtual Memory (RAM)
		TotalBytes:     vmem.Total,
		AvailableBytes: vmem.Available,
		UsedBytes:      vmem.Used,
		UsedPercent:    vmem.UsedPercent,

		// Swap Memory
		SwapTotalBytes:  swap.Total,
		SwapUsedBytes:   swap.Used,
		SwapUsedPercent: swap.UsedPercent,
	}

	// Create metric wrapper
	metric := Metric{
		Type:      "memory",
		Timestamp: time.Now(),
		Data:      memMetric,
	}

	// Try to send metric (non-blocking)
	select {
	case m.output <- metric:
		// Successfully sent
	default:
		// Channel full, drop metric
		log.Println("Memory metrics channel full, dropping metric")
	}
}
