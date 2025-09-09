package prom

import (
	"fmt"

	"github.com/dirshaye/GoMetrics/internal/collect"
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	// CPU metrics
	cpuUsagePercent *prometheus.GaugeVec
	cpuLoadAverage  *prometheus.GaugeVec

	// Memory metrics
	memoryUsageBytes   *prometheus.GaugeVec
	memoryUsagePercent prometheus.Gauge
	swapUsageBytes     *prometheus.GaugeVec
	swapUsagePercent   prometheus.Gauge

	// Disk metrics
	diskUsageBytes   *prometheus.GaugeVec
	diskUsagePercent prometheus.Gauge
	diskIOBytes      *prometheus.GaugeVec
	diskIOOperations *prometheus.GaugeVec

	// Network metrics
	networkBytes   *prometheus.GaugeVec
	networkPackets *prometheus.GaugeVec
	networkErrors  *prometheus.GaugeVec
	networkDrops   *prometheus.GaugeVec
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics() *Metrics {
	m := &Metrics{
		// CPU metrics with labels for overall vs per-core
		cpuUsagePercent: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gometrics_cpu_usage_percent",
				Help: "CPU usage percentage",
			},
			[]string{"type"}, // "overall" or "core_N"
		),

		cpuLoadAverage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gometrics_cpu_load_average",
				Help: "CPU load average",
			},
			[]string{"period"}, // "1m", "5m", "15m"
		),

		// Memory metrics with labels for type
		memoryUsageBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gometrics_memory_usage_bytes",
				Help: "Memory usage in bytes",
			},
			[]string{"type"}, // "total", "used", "available"
		),

		memoryUsagePercent: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "gometrics_memory_usage_percent",
				Help: "Memory usage percentage",
			},
		),

		swapUsageBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gometrics_swap_usage_bytes",
				Help: "Swap usage in bytes",
			},
			[]string{"type"}, // "total", "used"
		),

		swapUsagePercent: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "gometrics_swap_usage_percent",
				Help: "Swap usage percentage",
			},
		),

		// Disk metrics
		diskUsageBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gometrics_disk_usage_bytes",
				Help: "Disk usage in bytes",
			},
			[]string{"type"}, // "total", "used", "free"
		),

		diskUsagePercent: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "gometrics_disk_usage_percent",
				Help: "Disk usage percentage",
			},
		),

		diskIOBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gometrics_disk_io_bytes_total",
				Help: "Total disk I/O bytes",
			},
			[]string{"direction"}, // "read", "write"
		),

		diskIOOperations: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gometrics_disk_io_operations_total",
				Help: "Total disk I/O operations",
			},
			[]string{"direction"}, // "read", "write"
		),

		// Network metrics
		networkBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gometrics_network_bytes_total",
				Help: "Total network bytes",
			},
			[]string{"direction"}, // "sent", "received"
		),

		networkPackets: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gometrics_network_packets_total",
				Help: "Total network packets",
			},
			[]string{"direction"}, // "sent", "received"
		),

		networkErrors: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gometrics_network_errors_total",
				Help: "Total network errors",
			},
			[]string{"direction"}, // "in", "out"
		),

		networkDrops: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gometrics_network_drops_total",
				Help: "Total network packet drops",
			},
			[]string{"direction"}, // "in", "out"
		),
	}

	// Register all metrics with Prometheus
	prometheus.MustRegister(
		m.cpuUsagePercent,
		m.cpuLoadAverage,
		m.memoryUsageBytes,
		m.memoryUsagePercent,
		m.swapUsageBytes,
		m.swapUsagePercent,
		m.diskUsageBytes,
		m.diskUsagePercent,
		m.diskIOBytes,
		m.diskIOOperations,
		m.networkBytes,
		m.networkPackets,
		m.networkErrors,
		m.networkDrops,
	)

	return m
}

// UpdateFromSample updates all Prometheus metrics from a Sample
func (m *Metrics) UpdateFromSample(sample collect.Sample) {
	// Update CPU metrics
	m.cpuUsagePercent.WithLabelValues("overall").Set(sample.CPU.OverallPercent)

	// Update per-core CPU usage
	for i, corePercent := range sample.CPU.PerCorePercent {
		label := fmt.Sprintf("core_%d", i)
		m.cpuUsagePercent.WithLabelValues(label).Set(corePercent)
	}

	// Update load averages (if available)
	if len(sample.CPU.LoadAverage) >= 3 {
		m.cpuLoadAverage.WithLabelValues("1m").Set(sample.CPU.LoadAverage[0])
		m.cpuLoadAverage.WithLabelValues("5m").Set(sample.CPU.LoadAverage[1])
		m.cpuLoadAverage.WithLabelValues("15m").Set(sample.CPU.LoadAverage[2])
	}

	// Update memory metrics
	m.memoryUsageBytes.WithLabelValues("total").Set(float64(sample.Memory.TotalBytes))
	m.memoryUsageBytes.WithLabelValues("used").Set(float64(sample.Memory.UsedBytes))
	m.memoryUsageBytes.WithLabelValues("available").Set(float64(sample.Memory.AvailableBytes))
	m.memoryUsagePercent.Set(sample.Memory.UsedPercent)

	// Update swap metrics
	m.swapUsageBytes.WithLabelValues("total").Set(float64(sample.Memory.SwapTotalBytes))
	m.swapUsageBytes.WithLabelValues("used").Set(float64(sample.Memory.SwapUsedBytes))
	m.swapUsagePercent.Set(sample.Memory.SwapUsedPercent)

	// Update disk metrics
	m.diskUsageBytes.WithLabelValues("total").Set(float64(sample.Disk.TotalBytes))
	m.diskUsageBytes.WithLabelValues("used").Set(float64(sample.Disk.UsedBytes))
	m.diskUsageBytes.WithLabelValues("free").Set(float64(sample.Disk.FreeBytes))
	m.diskUsagePercent.Set(sample.Disk.UsedPercent)

	m.diskIOBytes.WithLabelValues("read").Set(float64(sample.Disk.ReadBytes))
	m.diskIOBytes.WithLabelValues("write").Set(float64(sample.Disk.WriteBytes))
	m.diskIOOperations.WithLabelValues("read").Set(float64(sample.Disk.ReadOps))
	m.diskIOOperations.WithLabelValues("write").Set(float64(sample.Disk.WriteOps))

	// Update network metrics
	m.networkBytes.WithLabelValues("sent").Set(float64(sample.Network.BytesSent))
	m.networkBytes.WithLabelValues("received").Set(float64(sample.Network.BytesRecv))
	m.networkPackets.WithLabelValues("sent").Set(float64(sample.Network.PacketsSent))
	m.networkPackets.WithLabelValues("received").Set(float64(sample.Network.PacketsRecv))

	m.networkErrors.WithLabelValues("in").Set(float64(sample.Network.ErrorsIn))
	m.networkErrors.WithLabelValues("out").Set(float64(sample.Network.ErrorsOut))
	m.networkDrops.WithLabelValues("in").Set(float64(sample.Network.DropsIn))
	m.networkDrops.WithLabelValues("out").Set(float64(sample.Network.DropsOut))
}
