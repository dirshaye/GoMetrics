package collect

import (
	"context"
	"log"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
)

// DiskCollector collects disk usage and I/O metrics
type DiskCollector struct {
	interval time.Duration
	output   chan<- Metric
}

// NewDiskCollector creates a new disk collector
func NewDiskCollector(interval time.Duration, output chan<- Metric) *DiskCollector {
	return &DiskCollector{
		interval: interval,
		output:   output,
	}
}

// Start begins collecting disk metrics in a goroutine
func (d *DiskCollector) Start(ctx context.Context) {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	// Collect initial metric immediately
	d.collectAndSend()

	for {
		select {
		case <-ctx.Done():
			log.Println("Disk collector stopping...")
			return
		case <-ticker.C:
			d.collectAndSend()
		}
	}
}

// collectAndSend gathers disk metrics and sends them through the channel
func (d *DiskCollector) collectAndSend() {
	// Get disk usage for root filesystem
	usage, err := disk.Usage("/")
	if err != nil {
		log.Printf("Error collecting disk usage metrics: %v", err)
		return
	}

	// Get disk I/O statistics
	ioCounters, err := disk.IOCounters()
	if err != nil {
		log.Printf("Error collecting disk I/O metrics: %v", err)
		return
	}

	// Aggregate I/O stats across all disks
	var totalReadBytes, totalWriteBytes, totalReadOps, totalWriteOps uint64
	for _, io := range ioCounters {
		totalReadBytes += io.ReadBytes
		totalWriteBytes += io.WriteBytes
		totalReadOps += io.ReadCount
		totalWriteOps += io.WriteCount
	}

	// Create disk metric
	diskMetric := DiskMetric{
		// Disk Usage
		TotalBytes:  usage.Total,
		FreeBytes:   usage.Free,
		UsedBytes:   usage.Used,
		UsedPercent: usage.UsedPercent,

		// Disk I/O
		ReadBytes:  totalReadBytes,
		WriteBytes: totalWriteBytes,
		ReadOps:    totalReadOps,
		WriteOps:   totalWriteOps,
	}

	// Create metric wrapper
	metric := Metric{
		Type:      "disk",
		Timestamp: time.Now(),
		Data:      diskMetric,
	}

	// Try to send metric (non-blocking)
	select {
	case d.output <- metric:
		// Successfully sent
	default:
		// Channel full, drop metric
		log.Println("Disk metrics channel full, dropping metric")
	}
}
