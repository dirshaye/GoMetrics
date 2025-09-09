package collect

import (
	"context"
	"log"
	"time"

	"github.com/shirou/gopsutil/v3/net"
)

// NetworkCollector collects network interface metrics
type NetworkCollector struct {
	interval time.Duration
	output   chan<- Metric
}

// NewNetworkCollector creates a new network collector
func NewNetworkCollector(interval time.Duration, output chan<- Metric) *NetworkCollector {
	return &NetworkCollector{
		interval: interval,
		output:   output,
	}
}

// Start begins collecting network metrics in a goroutine
func (n *NetworkCollector) Start(ctx context.Context) {
	ticker := time.NewTicker(n.interval)
	defer ticker.Stop()

	// Collect initial metric immediately
	n.collectAndSend()

	for {
		select {
		case <-ctx.Done():
			log.Println("Network collector stopping...")
			return
		case <-ticker.C:
			n.collectAndSend()
		}
	}
}

// collectAndSend gathers network metrics and sends them through the channel
func (n *NetworkCollector) collectAndSend() {
	// Get network I/O statistics for all interfaces
	ioCounters, err := net.IOCounters(false) // false = aggregate all interfaces
	if err != nil {
		log.Printf("Error collecting network metrics: %v", err)
		return
	}

	// Since we passed false, we get one aggregated result
	if len(ioCounters) == 0 {
		log.Println("No network interfaces found")
		return
	}

	// Take the first (and only) aggregated result
	netStats := ioCounters[0]

	// Create network metric
	networkMetric := NetworkMetric{
		BytesSent:   netStats.BytesSent,
		BytesRecv:   netStats.BytesRecv,
		PacketsSent: netStats.PacketsSent,
		PacketsRecv: netStats.PacketsRecv,
		ErrorsIn:    netStats.Errin,
		ErrorsOut:   netStats.Errout,
		DropsIn:     netStats.Dropin,
		DropsOut:    netStats.Dropout,
	}

	// Create metric wrapper
	metric := Metric{
		Type:      "network",
		Timestamp: time.Now(),
		Data:      networkMetric,
	}

	// Try to send metric (non-blocking)
	select {
	case n.output <- metric:
		// Successfully sent
	default:
		// Channel full, drop metric
		log.Println("Network metrics channel full, dropping metric")
	}
}
