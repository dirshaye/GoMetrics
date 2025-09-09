package collect

import "time"

// Metric represents a single metric measurement with timestamp
type Metric struct {
	Type      string    `json:"type"`      // "cpu", "memory", "disk", "network"
	Timestamp time.Time `json:"timestamp"` // When the metric was collected
	Data      any       `json:"data"`      // The actual metric data (CPU, Memory, etc.)
}

// CPUMetric represents CPU usage information
type CPUMetric struct {
	OverallPercent float64   `json:"overall_percent"`  // Overall CPU usage percentage
	PerCorePercent []float64 `json:"per_core_percent"` // CPU usage per core
	LoadAverage    []float64 `json:"load_average"`     // 1, 5, 15 minute load averages
}

// MemoryMetric represents memory usage information
type MemoryMetric struct {
	// Virtual Memory (RAM)
	TotalBytes     uint64  `json:"total_bytes"`     // Total RAM in bytes
	AvailableBytes uint64  `json:"available_bytes"` // Available RAM in bytes
	UsedBytes      uint64  `json:"used_bytes"`      // Used RAM in bytes
	UsedPercent    float64 `json:"used_percent"`    // Used RAM percentage

	// Swap Memory
	SwapTotalBytes  uint64  `json:"swap_total_bytes"`  // Total swap in bytes
	SwapUsedBytes   uint64  `json:"swap_used_bytes"`   // Used swap in bytes
	SwapUsedPercent float64 `json:"swap_used_percent"` // Used swap percentage
}

// DiskMetric represents disk usage and I/O information
type DiskMetric struct {
	// Disk Usage (for root filesystem)
	TotalBytes  uint64  `json:"total_bytes"`  // Total disk space in bytes
	FreeBytes   uint64  `json:"free_bytes"`   // Free disk space in bytes
	UsedBytes   uint64  `json:"used_bytes"`   // Used disk space in bytes
	UsedPercent float64 `json:"used_percent"` // Used disk space percentage

	// Disk I/O Statistics
	ReadBytes  uint64 `json:"read_bytes"`  // Bytes read from disk
	WriteBytes uint64 `json:"write_bytes"` // Bytes written to disk
	ReadOps    uint64 `json:"read_ops"`    // Number of read operations
	WriteOps   uint64 `json:"write_ops"`   // Number of write operations
}

// NetworkMetric represents network interface statistics
type NetworkMetric struct {
	// Total across all interfaces
	BytesSent   uint64 `json:"bytes_sent"`   // Total bytes sent
	BytesRecv   uint64 `json:"bytes_recv"`   // Total bytes received
	PacketsSent uint64 `json:"packets_sent"` // Total packets sent
	PacketsRecv uint64 `json:"packets_recv"` // Total packets received

	// Error statistics
	ErrorsIn  uint64 `json:"errors_in"`  // Input errors
	ErrorsOut uint64 `json:"errors_out"` // Output errors
	DropsIn   uint64 `json:"drops_in"`   // Input packet drops
	DropsOut  uint64 `json:"drops_out"`  // Output packet drops
}

// Sample represents a complete snapshot of all metrics at a point in time
// This is what gets sent to clients and stored as "latest"
type Sample struct {
	Timestamp time.Time     `json:"timestamp"` // When this sample was created
	CPU       CPUMetric     `json:"cpu"`       // CPU metrics
	Memory    MemoryMetric  `json:"memory"`    // Memory metrics
	Disk      DiskMetric    `json:"disk"`      // Disk metrics
	Network   NetworkMetric `json:"network"`   // Network metrics
}

// IsComplete checks if a sample has all required metrics
func (s *Sample) IsComplete() bool {
	// For now, we consider a sample complete if it has a timestamp
	// Later we might add more sophisticated validation
	return !s.Timestamp.IsZero()
}
